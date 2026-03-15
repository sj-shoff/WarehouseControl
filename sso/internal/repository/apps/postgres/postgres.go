package apps_postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"sso/internal/domain"
	customErr "sso/internal/domain/errors"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type AppsPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *AppsPostgresRepository {
	return &AppsPostgresRepository{db: db, retries: retries}
}

func (r *AppsPostgresRepository) GetByID(ctx context.Context, id int) (*domain.App, error) {
	query := `SELECT id, name, secret FROM apps WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to query app: %v", customErr.ErrDatabase, err)
	}

	app := &domain.App{}
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrInvalidInput
		}
		return nil, fmt.Errorf("%w: scan app error: %v", customErr.ErrDatabase, err)
	}

	return app, nil
}
