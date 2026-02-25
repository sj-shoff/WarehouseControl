package history_usecase

import (
	"context"
	"warehouse-control/internal/domain"
)

type historyRepository interface {
	GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error)
	GetHistoryByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*domain.HistoryRecord, error)
}
