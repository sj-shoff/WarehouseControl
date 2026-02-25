package history_handler

import (
	"context"
	"warehouse-control/internal/domain"
)

type historyUsecase interface {
	GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error)
	GetHistoryByItemID(ctx context.Context, itemID int64) ([]*domain.HistoryRecord, error)
}
