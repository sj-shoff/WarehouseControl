package auth_usecase

import (
	"context"
	"warehouse-control/internal/domain"
)

type usersRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
}
