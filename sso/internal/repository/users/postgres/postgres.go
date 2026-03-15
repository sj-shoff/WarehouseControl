package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sso/internal/domain"
	customErr "sso/internal/domain/errors"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type UsersPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *UsersPostgresRepository {
	return &UsersPostgresRepository{db: db, retries: retries}
}

func (r *UsersPostgresRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, username)
	if err != nil {
		return nil, fmt.Errorf("%w: query user error: %v", customErr.ErrDatabase, err)
	}

	user := &domain.User{}
	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: scan user error: %v", customErr.ErrDatabase, err)
	}

	return user, nil
}

func (r *UsersPostgresRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: query user by id error: %v", customErr.ErrDatabase, err)
	}

	user := &domain.User{}
	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: scan user by id error: %v", customErr.ErrDatabase, err)
	}

	return user, nil
}

func (r *UsersPostgresRepository) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	query := `INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3) RETURNING id`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		return 0, fmt.Errorf("%w: insert user failed: %v", customErr.ErrDatabase, err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%w: scan created user id failed: %v", customErr.ErrDatabase, err)
	}

	return id, nil
}

func (r *UsersPostgresRepository) GetUsers(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, username, role, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, fmt.Errorf("%w: select users failed: %v", customErr.ErrDatabase, err)
	}
	defer func() { _ = rows.Close() }()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%w: scan user in list failed: %v", customErr.ErrDatabase, err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: rows iteration error: %v", customErr.ErrDatabase, err)
	}

	if users == nil {
		return []*domain.User{}, nil
	}

	return users, nil
}

func (r *UsersPostgresRepository) UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`

	res, err := r.db.ExecWithRetry(ctx, r.retries, query, role, userID)
	if err != nil {
		return fmt.Errorf("%w: update role failed: %v", customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: rows affected error: %v", customErr.ErrDatabase, err)
	}

	if rows == 0 {
		return customErr.ErrUserNotFound
	}

	return nil
}

func (r *UsersPostgresRepository) SaveRefreshToken(ctx context.Context, userID int64, tokenHash string, appID int, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token_hash, app_id, expires_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecWithRetry(ctx, r.retries, query, userID, tokenHash, appID, expiresAt)
	if err != nil {
		return fmt.Errorf("%w: save refresh token failed: %v", customErr.ErrDatabase, err)
	}

	return nil
}

func (r *UsersPostgresRepository) GetRefreshToken(ctx context.Context, tokenHash string) (int64, int, time.Time, error) {
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

func (r *UsersPostgresRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
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
