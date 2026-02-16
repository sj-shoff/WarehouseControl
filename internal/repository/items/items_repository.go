package items_repository

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

type ItemsPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *ItemsPostgresRepository {
	return &ItemsPostgresRepository{
		db:      db,
		retries: retries,
	}
}

func (r *ItemsPostgresRepository) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {
	var id int64
	query := `
		INSERT INTO items (name, sku, quantity, price, category, location)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query,
		item.Name, item.SKU, item.Quantity, item.Price, item.Category, item.Location)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	err = row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: no rows returned", customErr.ErrDatabase)
		}
		return 0, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	return id, nil
}

func (r *ItemsPostgresRepository) GetItems(ctx context.Context) ([]*domain.Item, error) {
	var items []*domain.Item
	query := `
		SELECT id, name, sku, quantity, price, category, location, created_at, updated_at
		FROM items
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	defer rows.Close()
	for rows.Next() {
		item := &domain.Item{}
		err := rows.Scan(
			&item.ID, &item.Name, &item.SKU, &item.Quantity, &item.Price,
			&item.Category, &item.Location, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	return items, nil
}

func (r *ItemsPostgresRepository) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	item := &domain.Item{}
	query := `
		SELECT id, name, sku, quantity, price, category, location, created_at, updated_at
		FROM items
		WHERE id = $1
	`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	err = row.Scan(
		&item.ID, &item.Name, &item.SKU, &item.Quantity, &item.Price,
		&item.Category, &item.Location, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrItemNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	return item, nil
}

func (r *ItemsPostgresRepository) UpdateItem(ctx context.Context, id int64, item *domain.Item) error {
	query := `
		UPDATE items
		SET name = $1, sku = $2, quantity = $3, price = $4, category = $5, location = $6, updated_at = now()
		WHERE id = $7
	`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query,
		item.Name, item.SKU, item.Quantity, item.Price, item.Category, item.Location, id)
	if err != nil {
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	if rows == 0 {
		return customErr.ErrItemNotFound
	}
	return nil
}

func (r *ItemsPostgresRepository) DeleteItem(ctx context.Context, id int64) error {
	query := `DELETE FROM items WHERE id = $1`
	res, err := r.db.ExecWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	if rows == 0 {
		return customErr.ErrItemNotFound
	}
	return nil
}
