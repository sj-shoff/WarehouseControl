package users_postgres

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

type usersPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *usersPostgresRepository {
	return &usersPostgresRepository{db: db, retries: retries}
}

func (r *usersPostgresRepository) GetUserByUsernameAndApp(ctx context.Context, username string, appID int) (*domain.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username = $1 AND app_id = $2`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, username, appID)
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

func (r *usersPostgresRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
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

func (r *usersPostgresRepository) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	query := `INSERT INTO users (username, password_hash, role, app_id) VALUES ($1, $2, $3, $4) RETURNING id`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, user.Username, user.PasswordHash, user.Role, user.AppID)
	if err != nil {
		return 0, fmt.Errorf("%w: insert user failed: %v", customErr.ErrDatabase, err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%w: scan created user id failed: %v", customErr.ErrDatabase, err)
	}

	return id, nil
}

func (r *usersPostgresRepository) GetUsers(ctx context.Context) ([]*domain.User, error) {
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

func (r *usersPostgresRepository) UpdateUserRole(ctx context.Context, userID int64, role domain.UserRole) error {
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

func (r *usersPostgresRepository) CreateAdminIfNotExist(ctx context.Context, username, passHash string, appID int) (int64, error) {
	query := `
		INSERT INTO users (username, password_hash, role, app_id)
		VALUES ($1, $2, 'admin', $3)
		ON CONFLICT (username, app_id) DO UPDATE SET password_hash = EXCLUDED.password_hash
		RETURNING id
	`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, username, passHash, appID)
	if err != nil {
		return 0, fmt.Errorf("%w: bootstrap admin failed: %v", customErr.ErrDatabase, err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%w: scan bootstrapped admin id failed: %v", customErr.ErrDatabase, err)
	}

	return id, nil
}
