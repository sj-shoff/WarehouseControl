package auth_handler

import (
	"context"
	"warehouse-control/internal/domain"
)

type authUsecase interface {
	Authenticate(ctx context.Context, username, password string) (*domain.User, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
}
