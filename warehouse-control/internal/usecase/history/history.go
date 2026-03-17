package history_usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

	"github.com/wb-go/wbf/zlog"
)

type historyUsecase struct {
	repo   HistoryRepository
	logger *zlog.Zerolog
}

func NewService(repo HistoryRepository, logger *zlog.Zerolog) *historyUsecase {
	return &historyUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (s *historyUsecase) GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
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
	s.logger.Info().Int("count", len(records)).Msg("History retrieved")
	return records, nil
}

func (s *historyUsecase) GetHistoryByItemID(ctx context.Context, itemID int64) ([]*domain.HistoryRecord, error) {
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

func (s *historyUsecase) GetDiff(ctx context.Context, recordID int64) (domain.DiffResponse, error) {
	record, err := s.repo.GetByID(ctx, recordID)
	if err != nil {
		s.logger.Error().Err(err).Int64("id", recordID).Msg("Failed to find record for diff")
		return domain.DiffResponse{}, err
	}
	return s.computeDiff(record), nil
}

func (s *historyUsecase) computeDiff(r *domain.HistoryRecord) domain.DiffResponse {
	return s.buildDiff(r.OldData, r.NewData, r.Action, r.ChangedBy, r.ChangedAt.Format(time.RFC3339), r.ID)
}

func (s *historyUsecase) buildDiff(oldItem, newItem *domain.Item, action, by, at string, id int64) domain.DiffResponse {
	config := []struct{ Key, Label string }{
		{"name", "Название"},
		{"sku", "SKU"},
		{"quantity", "Количество"},
		{"price", "Цена"},
		{"category", "Категория"},
		{"location", "Место хранения"},
	}
	var fields []domain.DiffField
	for _, cfg := range config {
		oldVal := s.getVal(oldItem, cfg.Key)
		newVal := s.getVal(newItem, cfg.Key)
		status := domain.StatusUnchanged
		if action == "INSERT" || (oldVal == nil && newVal != nil) {
			status = domain.StatusAdded
		} else if action == "DELETE" || (oldVal != nil && newVal == nil) {
			status = domain.StatusRemoved
		} else if oldVal != newVal {
			status = domain.StatusChanged
		}
		fields = append(fields, domain.DiffField{
			Field: cfg.Key, Label: cfg.Label, Old: oldVal, New: newVal, Status: status,
		})
	}
	return domain.DiffResponse{
		RecordID: id, Action: action, ChangedBy: by, ChangedAt: at, Fields: fields,
	}
}

func (s *historyUsecase) getVal(item *domain.Item, field string) interface{} {
	if item == nil {
		return nil
	}
	switch field {
	case "name":
		return item.Name
	case "sku":
		return item.SKU
	case "quantity":
		return item.Quantity
	case "price":
		return item.Price
	case "category":
		return item.Category
	case "location":
		return item.Location
	}
	return nil
}
