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
	const op = "apps_repository.GetByID"
	app := &domain.App{}
	query := `SELECT id, name, secret FROM apps WHERE id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, customErr.ErrDatabase)
	}
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrInvalidInput
		}
		return nil, fmt.Errorf("%s: %w", op, customErr.ErrDatabase)
	}
	return app, nil
}
