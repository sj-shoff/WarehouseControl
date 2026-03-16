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

type appsPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *appsPostgresRepository {
	return &appsPostgresRepository{db: db, retries: retries}
}

func (r *appsPostgresRepository) GetByID(ctx context.Context, id int) (*domain.App, error) {
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

func (r *appsPostgresRepository) GetByNameAndSecret(ctx context.Context, name, secret string) (*domain.App, error) {
	query := `SELECT id, name, secret FROM apps WHERE name = $1 AND secret = $2`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, name, secret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "get by name and secret", err)
	}

	app := &domain.App{}
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", customErr.ErrInvalidInput, err)
		}
		return nil, fmt.Errorf("%s: %w", customErr.ErrDatabase, err)
	}

	return app, nil
}

func (r *appsPostgresRepository) UpsertApp(ctx context.Context, name, secret string) (int32, error) {
	query := `
		INSERT INTO apps (name, secret) 
		VALUES ($1, $2) 
		ON CONFLICT (name) DO UPDATE SET secret = EXCLUDED.secret 
		RETURNING id`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, name, secret)
	if err != nil {
		return 0, fmt.Errorf("%w: upsert app error: %v", customErr.ErrDatabase, err)
	}

	var id int32
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%w: scan upserted app id failed: %v", customErr.ErrDatabase, err)
	}

	return id, nil
}
