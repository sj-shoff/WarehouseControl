package auth_usecase

import (
	"context"
	"sso/internal/domain"
	"time"
)

type UserRepository interface {
	GetUserByUsernameAndApp(ctx context.Context, username string, appID int) (*domain.User, error)
	CreateAdminIfNotExist(ctx context.Context, username, passHash string, appID int) (int64, error)
	CreateUser(ctx context.Context, user *domain.User) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error
}

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, appID int, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (userID int64, appID int, expiresAt time.Time, err error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type AppRepository interface {
	GetByNameAndSecret(ctx context.Context, name, secret string) (*domain.App, error)
	UpsertApp(ctx context.Context, name, secret string) (int32, error)
	GetByID(ctx context.Context, id int) (*domain.App, error)
}
