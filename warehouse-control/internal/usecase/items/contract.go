package items_usecase

import (
	"context"

	"warehouse-control/internal/domain"
)

type itemsRepository interface {
	CreateItem(ctx context.Context, item *domain.Item, username string) (int64, error)
	GetItems(ctx context.Context, limit, offset int, search string) ([]*domain.Item, int, error)
	GetItemByID(ctx context.Context, id int64) (*domain.Item, error)
	UpdateItem(ctx context.Context, id int64, item *domain.Item, username string) error
	DeleteItem(ctx context.Context, id int64, username string) error
	BulkDeleteItems(ctx context.Context, ids []int64, username string) error
}
