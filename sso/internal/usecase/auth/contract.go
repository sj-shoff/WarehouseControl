package auth_usecase

import (
	"context"
	"sso/internal/domain"
)

type userRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUser(ctx context.Context, id int64, username string, role domain.UserRole) error
	DeleteUser(ctx context.Context, id int64) error
	UpdateUserRole(ctx context.Context, id int64, role domain.UserRole) error
}
