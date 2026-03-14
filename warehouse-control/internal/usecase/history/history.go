package history_usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"warehouse-control/internal/domain"

	customErr "warehouse-control/internal/domain/errors"

	"github.com/go-redis/redis/v8"
	"github.com/wb-go/wbf/zlog"
)

type HistoryUsecase struct {
	repo   historyRepository
	logger *zlog.Zerolog
	redis  *redis.Client
}

func NewService(repo historyRepository, logger *zlog.Zerolog, redis *redis.Client) *HistoryUsecase {
	return &HistoryUsecase{
		repo:   repo,
		logger: logger,
		redis:  redis,
	}
}

func (s *HistoryUsecase) GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	key := fmt.Sprintf("history:%v", filter)
	cached, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var records []*domain.HistoryRecord
		if err := json.Unmarshal([]byte(cached), &records); err == nil {
			return records, nil
		}
	}
	s.logger.Info().Msg("Getting history")
	records, _, err := s.repo.GetHistory(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get history")
		if errors.Is(err, customErr.ErrDatabase) {
			return nil, customErr.ErrDatabase
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrInternal, err)
	}
	jsonData, _ := json.Marshal(records)
	s.redis.Set(ctx, key, jsonData, 5*time.Minute)
	s.logger.Info().Int("count", len(records)).Msg("History retrieved")
	return records, nil
}

func (s *HistoryUsecase) GetHistoryByItemID(ctx context.Context, itemID int64) ([]*domain.HistoryRecord, error) {
	if itemID <= 0 {
		return nil, customErr.ErrInvalidInput
	}
	key := fmt.Sprintf("history_item:%d", itemID)
	cached, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var records []*domain.HistoryRecord
		if err := json.Unmarshal([]byte(cached), &records); err == nil {
			return records, nil
		}
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
	jsonData, _ := json.Marshal(records)
	s.redis.Set(ctx, key, jsonData, 5*time.Minute)
	s.logger.Info().Int("count", len(records)).Msg("Item history retrieved")
	return records, nil
}
