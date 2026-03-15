package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"warehouse-control/internal/config"
	"warehouse-control/internal/grpc/sso"
	authH "warehouse-control/internal/http-server/handler/auth"
	historyH "warehouse-control/internal/http-server/handler/history"
	itemsH "warehouse-control/internal/http-server/handler/items"
	"warehouse-control/internal/http-server/middleware"
	"warehouse-control/internal/http-server/router"
	historyRepo "warehouse-control/internal/repository/history/postgres"
	itemsRepo "warehouse-control/internal/repository/items/postgres"
	historyUc "warehouse-control/internal/usecase/history"
	itemsUc "warehouse-control/internal/usecase/items"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	cfg       *config.Config
	logger    *zlog.Zerolog
	server    *http.Server
	db        *dbpg.DB
	ssoClient *sso.Client
}

func NewApp(cfg *config.Config, logger *zlog.Zerolog) (*App, error) {
	retries := cfg.DefaultRetryStrategy()

	db, err := dbpg.New(cfg.DBDSN(), nil, &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	ssoClient, err := sso.NewClient(cfg)
	if err != nil {
		db.Master.Close()
		return nil, fmt.Errorf("sso: %w", err)
	}

	itemsR := itemsRepo.NewPostgresRepository(db, retries)
	historyR := historyRepo.NewPostgresRepository(db, retries)
	itemsU := itemsUc.NewService(itemsR, logger)
	historyU := historyUc.NewService(historyR, logger)
	iH := itemsH.NewHandler(itemsU, logger)
	hH := historyH.NewHandler(historyU, logger)
	aH := authH.NewHandler(ssoClient, cfg, logger)

	authMW := middleware.NewAuthMiddleware(cfg.JWT.Secret, logger)
	r := router.New(iH, hH, aH, authMW, cfg, logger)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		cfg:       cfg,
		logger:    logger,
		server:    srv,
		db:        db,
		ssoClient: ssoClient,
	}, nil
}

func (a *App) Run() error {
	serverErrors := make(chan error, 1)

	go func() {
		a.logger.Info().Str("port", a.cfg.Server.Addr).Msg("HTTP server starting")
		serverErrors <- a.server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-a.handleSignals():
		a.logger.Info().Str("signal", sig.String()).Msg("shutdown signal received")
		return a.Stop()
	}

	return nil
}

func (a *App) handleSignals() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	return sigChan
}

func (a *App) Stop() error {
	a.logger.Info().Msg("performing graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
	defer cancel()

	var hasError bool

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error().Err(err).Msg("server shutdown failed")
		_ = a.server.Close()
		hasError = true
	}

	if a.ssoClient != nil {
		if err := a.ssoClient.Close(); err != nil {
			a.logger.Error().Err(err).Msg("failed to close SSO gRPC client")
			hasError = true
		} else {
			a.logger.Info().Msg("SSO gRPC client closed")
		}
	}

	if a.db != nil {
		if err := a.db.Master.Close(); err != nil {
			a.logger.Error().Err(err).Msg("failed to close DB")
			hasError = true
		} else {
			a.logger.Info().Msg("database connection closed")
		}
	}

	if hasError {
		return fmt.Errorf("app stopped with errors")
	}

	a.logger.Info().Msg("shutdown complete")
	return nil
}
