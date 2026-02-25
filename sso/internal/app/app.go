package app

import (
	"context"
	"fmt"
	"net"
	"sso/internal/config"

	grpcService "sso/internal/grpc/auth"
	userRepository "sso/internal/repository/auth/postgres"
	authUsecase "sso/internal/usecase/auth"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
	"google.golang.org/grpc"
)

type App struct {
	cfg    *config.Config
	logger *zlog.Zerolog
	server *grpc.Server
}

func NewApp(cfg *config.Config, logger *zlog.Zerolog) (*App, error) {
	retries := cfg.DefaultRetryStrategy()
	dbOpts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}
	db, err := dbpg.New(cfg.DBDSN(), []string{}, dbOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	userRepo := userRepository.NewPostgresRepository(db, retries)
	authUc := authUsecase.NewAuthUsecase(userRepo, logger, cfg.JWT.Secret, cfg.JWT.TokenTTL)

	grpcServer := grpc.NewServer()
	grpcService.Register(grpcServer, authUc)

	return &App{
		cfg:    cfg,
		logger: logger,
		server: grpcServer,
	}, nil
}

func (a *App) Run(ctx context.Context, port string) error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	a.logger.Info().Str("port", port).Msg("gRPC server starting")

	go func() {
		<-ctx.Done()
		a.server.GracefulStop()
		a.db.
	}()

	return a.server.Serve(lis)
}
