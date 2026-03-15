package authgrpc

import (
	"context"
	"sso/internal/domain"
	"time"
)

type authProvider interface {
	Login(ctx context.Context, user, pass string, appID int) (*domain.UserClaim, string, string, time.Time, error)
	Refresh(ctx context.Context, token string, appID int) (string, string, error)
	RegisterNewUser(ctx context.Context, user, pass string, role domain.UserRole, appID int) (int64, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUserRole(ctx context.Context, uid int64, role domain.UserRole) error
}
