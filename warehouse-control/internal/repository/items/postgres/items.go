package items_postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.setAuditUser(ctx, tx, username); err != nil {
		return 0, err
	}

	query := `INSERT INTO items (name, sku, quantity, price, category, location) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id int64
	err = tx.QueryRowContext(ctx, query,
		item.Name, item.SKU, item.Quantity, item.Price, item.Category, item.Location,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to insert item: %v", customErr.ErrDatabase, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return id, nil
}

func (r *ItemsPostgresRepository) GetItems(ctx context.Context, limit, offset int, search string) ([]*domain.Item, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR sku ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM items %s", whereClause)
	var total int
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: total count error: %v", customErr.ErrDatabase, err)
	}
	if err := row.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%w: scan total error: %v", customErr.ErrDatabase, err)
	}

	if total == 0 {
		return []*domain.Item{}, 0, nil
	}

	query := fmt.Sprintf(`
		SELECT id, name, sku, quantity, price, category, location, created_at, updated_at 
		FROM items %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	finalArgs := append(args, limit, offset)
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, finalArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: query items error: %v", customErr.ErrDatabase, err)
	}
	defer func() { _ = rows.Close() }()

	items := make([]*domain.Item, 0, limit)
	for rows.Next() {
		i := &domain.Item{}
		err := rows.Scan(&i.ID, &i.Name, &i.SKU, &i.Quantity, &i.Price, &i.Category, &i.Location, &i.CreatedAt, &i.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: scan item error: %v", customErr.ErrDatabase, err)
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%w: rows iteration error: %v", customErr.ErrDatabase, err)
	}

	return items, total, nil
}

func (r *ItemsPostgresRepository) GetItemByID(ctx context.Context, id int64) (*domain.Item, error) {
	query := `SELECT id, name, sku, quantity, price, category, location, created_at, updated_at FROM items WHERE id = $1`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}

	item := &domain.Item{}
	err = row.Scan(&item.ID, &item.Name, &item.SKU, &item.Quantity, &item.Price, &item.Category, &item.Location, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, customErr.ErrItemNotFound
		}
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	return item, nil
}

func (r *ItemsPostgresRepository) UpdateItem(ctx context.Context, id int64, item *domain.Item, username string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.setAuditUser(ctx, tx, username); err != nil {
		return err
	}

	query := `UPDATE items SET name=$1, sku=$2, quantity=$3, price=$4, category=$5, location=$6, updated_at=NOW() WHERE id=$7`
	res, err := tx.ExecContext(ctx, query, item.Name, item.SKU, item.Quantity, item.Price, item.Category, item.Location, id)
	if err != nil {
		return fmt.Errorf("%w: update failed: %v", customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return customErr.ErrItemNotFound
	}

	return tx.Commit()
}

func (r *ItemsPostgresRepository) DeleteItem(ctx context.Context, id int64, username string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.setAuditUser(ctx, tx, username); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM items WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("%w: delete failed: %v", customErr.ErrDatabase, err)
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return customErr.ErrItemNotFound
	}

	return tx.Commit()
}

func (r *ItemsPostgresRepository) BulkDeleteItems(ctx context.Context, ids []int64, username string) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := r.setAuditUser(ctx, tx, username); err != nil {
		return err
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM items WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%w: bulk delete failed: %v", customErr.ErrDatabase, err)
	}

	return tx.Commit()
}

func (r *ItemsPostgresRepository) setAuditUser(ctx context.Context, tx *sql.Tx, username string) error {
	_, err := tx.ExecContext(ctx, "SET LOCAL warehouse_control.changed_by = $1", username)
	if err != nil {
		return fmt.Errorf("%w: failed to set audit user: %v", customErr.ErrDatabase, err)
	}
	return nil
}
