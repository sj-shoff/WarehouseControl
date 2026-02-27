package items

import (
	"context"
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
	return &ItemsPostgresRepository{db: db, retries: retries}
}

func (r *ItemsPostgresRepository) CreateItem(ctx context.Context, item *domain.Item, username string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, _ = tx.ExecContext(ctx, "SET LOCAL warehouse_control.changed_by = $1", username)

	var id int64
	query := `INSERT INTO items (name, sku, quantity, price, category, location) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = tx.QueryRowContext(ctx, query, item.Name, item.SKU, item.Quantity, item.Price, item.Category, item.Location).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert: %w", err)
	}
	return id, tx.Commit()
}

func (r *ItemsPostgresRepository) GetItems(ctx context.Context) ([]*domain.Item, error) {
	var items []*domain.Item
	query := `SELECT id,name,sku,quantity,price,category,location,created_at,updated_at FROM items ORDER BY created_at DESC`
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		i := &domain.Item{}
		rows.Scan(&i.ID, &i.Name, &i.SKU, &i.Quantity, &i.Price, &i.Category, &i.Location, &i.CreatedAt, &i.UpdatedAt)
		items = append(items, i)
	}
	return items, nil
}

func (r *ItemsPostgresRepository) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	item := &domain.Item{}
	query := `SELECT id,name,sku,quantity,price,category,location,created_at,updated_at FROM items WHERE id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, err
	}
	err = row.Scan(&item.ID, &item.Name, &item.SKU, &item.Quantity, &item.Price, &item.Category, &item.Location, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, customErr.ErrItemNotFound
	}
	return item, nil
}

func (r *ItemsPostgresRepository) UpdateItem(ctx context.Context, id int64, item *domain.Item, username string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.ExecContext(ctx, "SET LOCAL warehouse_control.changed_by = $1", username)

	query := `UPDATE items SET name=$1,sku=$2,quantity=$3,price=$4,category=$5,location=$6,updated_at=NOW() WHERE id=$7`
	res, err := tx.ExecContext(ctx, query, item.Name, item.SKU, item.Quantity, item.Price, item.Category, item.Location, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return customErr.ErrItemNotFound
	}
	return tx.Commit()
}

func (r *ItemsPostgresRepository) DeleteItem(ctx context.Context, id int64, username string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.ExecContext(ctx, "SET LOCAL warehouse_control.changed_by = $1", username)

	res, err := tx.ExecContext(ctx, `DELETE FROM items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return customErr.ErrItemNotFound
	}
	return tx.Commit()
}
