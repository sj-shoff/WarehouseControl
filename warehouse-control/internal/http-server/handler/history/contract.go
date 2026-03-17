package history_handler

import (
	"context"

	"warehouse-control/internal/domain"
)

type HistoryUsecase interface {
	GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error)
	GetHistoryByItemID(ctx context.Context, itemID int64) ([]*domain.HistoryRecord, error)
	GetDiff(ctx context.Context, recordID int64) (domain.DiffResponse, error)
}
