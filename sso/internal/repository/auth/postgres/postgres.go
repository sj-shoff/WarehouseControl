package users_repository

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

type UsersPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *UsersPostgresRepository {
	return &UsersPostgresRepository{
		db:      db,
		retries: retries,
	}
}

func (r *UsersPostgresRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	const op = "users_repository.GetUserByUsername"

	user := &domain.User{}
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, username)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	return user, nil
}

func (r *UsersPostgresRepository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	const op = "users_repository.GetUserByID"

	user := &domain.User{}
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE id = $1`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	return user, nil
}

func (r *UsersPostgresRepository) CreateUser(ctx context.Context, user *domain.User) (int64, error) {
	const op = "users_repository.CreateUser"

	var id int64
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		return 0, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	err = row.Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	return id, nil
}

func (r *UsersPostgresRepository) GetUsers(ctx context.Context) ([]*domain.User, error) {
	const op = "users_repository.GetUsers"

	var users []*domain.User
	query := `SELECT id, username, role, created_at, updated_at FROM users ORDER BY created_at DESC`

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UsersPostgresRepository) UpdateUser(ctx context.Context, id int64, username string, role domain.UserRole) error {
	const op = "users_repository.UpdateUser"

	query := `UPDATE users SET username = $1, role = $2, updated_at = NOW() WHERE id = $3`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query, username, role, id)
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	if rows == 0 {
		return customErr.ErrUserNotFound
	}

	return nil
}

func (r *UsersPostgresRepository) DeleteUser(ctx context.Context, id int64) error {
	const op = "users_repository.DeleteUser"

	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	if rows == 0 {
		return customErr.ErrUserNotFound
	}

	return nil
}

func (r *UsersPostgresRepository) UpdateUserRole(ctx context.Context, id int64, role domain.UserRole) error {
	const op = "users_repository.UpdateUserRole"

	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query, role, id)
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w: %v", op, customErr.ErrDatabase, err)
	}

	if rows == 0 {
		return customErr.ErrUserNotFound
	}

	return nil
}
