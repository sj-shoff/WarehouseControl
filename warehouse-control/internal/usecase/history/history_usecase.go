package history_usecase

import (
	"context"
	"errors"
	"fmt"
	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

	"github.com/wb-go/wbf/zlog"
)

type HistoryUsecase struct {
	repo   historyRepository
	logger *zlog.Zerolog
}

func NewService(repo historyRepository, logger *zlog.Zerolog) *HistoryUsecase {
	return &HistoryUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (s *HistoryUsecase) GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	s.logger.Info().Msg("Getting history")
	records, err := s.repo.GetHistory(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get history")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int("count", len(records)).Msg("History retrieved")
	return records, nil
}

func (s *HistoryUsecase) GetHistoryByItemID(ctx context.Context, itemID int64) ([]*domain.HistoryRecord, error) {
	if itemID <= 0 {
		return nil, customErr.ErrInvalidInput
	}
	s.logger.Info().Int64("item_id", itemID).Msg("Getting item history")
	records, err := s.repo.GetHistoryByItemID(ctx, itemID, 100, 0)
	if err != nil {
		s.logger.Error().Err(err).Int64("item_id", itemID).Msg("Failed to get item history")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	s.logger.Info().Int("count", len(records)).Msg("Item history retrieved")
	return records, nil
}
