package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
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
		return fmt.Errorf("listen: %w", err)
	}

	a.logger.Info().Str("port", a.cfg.GRPC.Port).Msg("gRPC SSO started")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := a.server.Serve(lis); err != nil {
			a.logger.Error().Err(err).Msg("gRPC serve failed")
		}
	}()

	<-ctx.Done()
	a.logger.Info().Msg("Graceful shutdown...")
	a.server.GracefulStop()
	a.db.Master.Close()
	return nil
}
