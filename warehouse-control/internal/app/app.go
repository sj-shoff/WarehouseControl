package app

import (
	"fmt"
	"net/http"

	"warehouse-control/internal/config"
	"warehouse-control/internal/grpc/sso"
	authH "warehouse-control/internal/http-server/handler/auth"
	historyH "warehouse-control/internal/http-server/handler/history"
	itemsH "warehouse-control/internal/http-server/handler/items"
	"warehouse-control/internal/http-server/middleware"
	"warehouse-control/internal/http-server/router"
	historyRepo "warehouse-control/internal/repository/history"
	itemsRepo "warehouse-control/internal/repository/items"
	historyUc "warehouse-control/internal/usecase/history"
	itemsUc "warehouse-control/internal/usecase/items"

	"github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	cfg       *config.Config
	logger    *zlog.Zerolog
	server    *http.Server
	ssoClient *sso.Client
	redis     *redis.Client
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
		return nil, fmt.Errorf("sso: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	itemsR := itemsRepo.NewPostgresRepository(db, retries)
	historyR := historyRepo.NewPostgresRepository(db, retries)
	itemsU := itemsUc.NewService(itemsR, logger, redisClient)
	historyU := historyUc.NewService(historyR, logger, redisClient)
	itemsH := itemsH.NewHandler(itemsU, logger)
	historyH := historyH.NewHandler(historyU, logger)
	authH := authH.NewHandler(ssoClient, logger)
	authMW := middleware.NewAuthMiddleware(cfg.JWT.Secret, logger)
	r := router.New(itemsH, historyH, authH, authMW)
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	return &App{cfg: cfg, logger: logger, server: srv, ssoClient: ssoClient, redis: redisClient}, nil
}

func (a *App) Run() error {
	a.logger.Info().Str("port", a.cfg.Server.Addr).Msg("HTTP server starting")
	return a.server.ListenAndServe()
}
