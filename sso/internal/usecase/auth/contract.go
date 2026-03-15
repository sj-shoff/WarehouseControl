package auth_usecase

import (
	"context"
	"sso/internal/domain"
	"time"
)

type userRepository interface {
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error

	SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, appID int, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (userID int64, appID int, expiresAt time.Time, err error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type appUsecase interface {
	GetByID(ctx context.Context, id int) (*domain.App, error)
}
