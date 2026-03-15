package authgrpc

import (
	"context"
	"sso/internal/domain"
	"time"
)

type authProvider interface {
	Login(ctx context.Context, username, password string, appID int) (*domain.UserClaim, string, time.Time, error)
	RegisterNewUser(ctx context.Context, username, password string, role domain.UserRole, appID int) (int64, error)
	GetUsers(ctx context.Context) ([]*domain.User, error)
	UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error
}
