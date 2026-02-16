package items_handler

import (
	"context"
	"warehouse-control/internal/domain"
)

type itemsUsecase interface {
	CreateItem(ctx context.Context, item *domain.Item, username string) (int64, error)
	GetItems(ctx context.Context) ([]*domain.Item, error)
	GetItemByID(ctx context.Context, id int64) (*domain.Item, error)
	UpdateItem(ctx context.Context, id int64, item *domain.Item, username string) error
	DeleteItem(ctx context.Context, id int64, username string) error
}
