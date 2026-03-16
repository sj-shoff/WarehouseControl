package tokens_postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	customErr "sso/internal/domain/errors"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type refreshTokensPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *refreshTokensPostgresRepository {
	return &refreshTokensPostgresRepository{db: db, retries: retries}
}

func (r *refreshTokensPostgresRepository) SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, appID int, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, app_id, expires_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecWithRetry(ctx, r.retries, query, userID, tokenHash, appID, expiresAt)
	if err != nil {
		return fmt.Errorf("%w: save refresh token failed: %v", customErr.ErrDatabase, err)
	}

	return nil
}

func (r *refreshTokensPostgresRepository) GetRefreshToken(ctx context.Context, tokenHash string) (int64, int, time.Time, error) {
	query := `SELECT user_id, app_id, expires_at FROM refresh_tokens WHERE token_hash = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, tokenHash)
	if err != nil {
		return 0, 0, time.Time{}, fmt.Errorf("%w: query refresh token failed: %v", customErr.ErrDatabase, err)
	}

	var userID int64
	var appID int
	var expiresAt time.Time

	if err := row.Scan(&userID, &appID, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, time.Time{}, customErr.ErrInvalidCredentials
		}
		return 0, 0, time.Time{}, fmt.Errorf("%w: scan refresh token failed: %v", customErr.ErrDatabase, err)
	}

	return userID, appID, expiresAt, nil
}

func (r *refreshTokensPostgresRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	res, err := r.db.ExecWithRetry(ctx, r.retries, query, tokenHash)
	if err != nil {
		return fmt.Errorf("%w: delete refresh token failed: %v", customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: rows affected error on delete: %v", customErr.ErrDatabase, err)
	}

	if rows == 0 {
		return nil
	}

	return nil
}
