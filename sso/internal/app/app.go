package app

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sso/internal/config"
	authgrpc "sso/internal/grpc/auth"
	appsRepository "sso/internal/repository/apps/postgres"
	userRepository "sso/internal/repository/users/postgres"
	appsUsecase "sso/internal/usecase/apps"
	authUsecase "sso/internal/usecase/auth"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
	"google.golang.org/grpc"
)

type App struct {
	cfg    *config.Config
	logger *zlog.Zerolog
	server *grpc.Server
	db     *dbpg.DB
}

func NewApp(cfg *config.Config, logger *zlog.Zerolog) (*App, error) {
	retries := cfg.DefaultRetryStrategy()

	db, err := dbpg.New(cfg.DBDSN(), nil, &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	userRepo := userRepository.NewPostgresRepository(db, retries)
	appsRepo := appsRepository.NewPostgresRepository(db, retries)

	appsUc := appsUsecase.NewAppUsecase(appsRepo, logger)
	authUc := authUsecase.NewAuthUsecase(userRepo, appsUc, logger, cfg.JWT.Secret, cfg.JWT.TokenTTL)

	grpcServer := grpc.NewServer()
	authgrpc.Register(grpcServer, authUc)

	return &App{
		cfg:    cfg,
		logger: logger,
		server: grpcServer,
		db:     db,
	}, nil
}

func (a *App) Run() error {
	lis, err := net.Listen("tcp", ":"+a.cfg.GRPC.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	serverErrors := make(chan error, 1)

	go func() {
		a.logger.Info().Str("port", a.cfg.GRPC.Port).Msg("gRPC SSO server starting")
		if err := a.server.Serve(lis); err != nil {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("gRPC server failed: %w", err)
	case sig := <-a.handleSignals():
		a.logger.Info().Str("signal", sig.String()).Msg("shutdown signal received")
		a.Stop()
	}

	return nil
}

func (a *App) handleSignals() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	return sigChan
}

func (a *App) Stop() {
	a.logger.Info().Msg("performing graceful shutdown...")

	done := make(chan struct{})

	go func() {
		a.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info().Msg("gRPC server stopped gracefully")
	case <-time.After(a.cfg.GRPC.Timeout):
		a.logger.Warn().
			Str("timeout", a.cfg.GRPC.Timeout.String()).
			Msg("graceful shutdown timed out, forcing stop")

		a.server.Stop()
	}

	if a.db != nil {
		if err := a.db.Master.Close(); err != nil {
			a.logger.Error().Err(err).Msg("failed to close DB")
		} else {
			a.logger.Info().Msg("database connection closed")
		}
	}

	a.logger.Info().Msg("shutdown complete")
}
