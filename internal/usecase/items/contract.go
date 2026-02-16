package items_usecase

import (
	"context"
	"warehouse-control/internal/domain"
)

type itemsRepository interface {
	CreateItem(ctx context.Context, item *domain.Item) (int64, error)
	GetItems(ctx context.Context) ([]*domain.Item, error)
	GetItemByID(ctx context.Context, id int64) (*domain.Item, error)
	UpdateItem(ctx context.Context, id int64, item *domain.Item) error
	DeleteItem(ctx context.Context, id int64) error
}
