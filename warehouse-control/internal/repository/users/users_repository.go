package users_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

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
	user := &domain.User{}
	query := `SELECT id, username, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, username)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	return user, nil
}

func (r *UsersPostgresRepository) GetUsers(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User
	query := `SELECT id, username, role, created_at, updated_at FROM users ORDER BY username`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	defer rows.Close()
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
		}
		users = append(users, user)
	}
	return users, nil
}
