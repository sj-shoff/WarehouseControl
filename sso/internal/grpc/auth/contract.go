package authgrpc

import (
	"context"
	"sso/internal/domain"
	"time"
)

type AuthProvider interface {
	Login(ctx context.Context, username, password string, appID int, appSecret string) (claim *domain.UserClaim, accessToken string, refreshToken string, expiresAt time.Time, err error)
	RegisterNewUser(ctx context.Context, username, password string, role domain.UserRole, appID int) (userID int64, err error)
	Refresh(ctx context.Context, refreshToken string, appID int) (accessToken string, newRefreshToken string, err error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error
	InitialBootstrap(ctx context.Context, appName, appSecret, adminUser, adminPass string) (uid int64, appID int32, err error)
}
