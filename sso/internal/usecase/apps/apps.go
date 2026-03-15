package apps_usecase

import (
	"context"
	"fmt"

	"sso/internal/domain"
	customErr "sso/internal/domain/errors"

	"github.com/wb-go/wbf/zlog"
)

type AppUsecase struct {
	repo   appRepository
	logger *zlog.Zerolog
}

func NewAppUsecase(repo appRepository, logger *zlog.Zerolog) *AppUsecase {
	return &AppUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (s *AppUsecase) GetByID(ctx context.Context, id int) (*domain.App, error) {
	if id <= 0 {
		s.logger.Warn().Int("app_id", id).Msg("invalid app_id provided")
		return nil, customErr.ErrInvalidInput
	}

	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetch app by id %d: %w", id, err)
	}

	return app, nil
}
