package auth_usecase

import (
	"context"
	"sso/internal/domain"
)

type userRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error
}

type appUsecase interface {
	GetByID(ctx context.Context, id int) (*domain.App, error)
}
