package apps_usecase

import (
	"context"
	"sso/internal/domain"
)

type appRepository interface {
	GetByID(ctx context.Context, id int) (*domain.App, error)
}
