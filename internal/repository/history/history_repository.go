package history_repository

import (
	"context"
	"fmt"
	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type HistoryPostgresRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewPostgresRepository(db *dbpg.DB, retries retry.Strategy) *HistoryPostgresRepository {
	return &HistoryPostgresRepository{
		db:      db,
		retries: retries,
	}
}

func (r *HistoryPostgresRepository) GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, error) {
	query := `
		SELECT id, item_id, action, old_data, new_data, changed_by, changed_at
		FROM items_history
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.ItemID != nil {
		query += fmt.Sprintf(" AND item_id = $%d", argIndex)
		args = append(args, *filter.ItemID)
		argIndex++
	}
	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *filter.Action)
		argIndex++
	}
	if filter.Username != nil {
		query += fmt.Sprintf(" AND changed_by = $%d", argIndex)
		args = append(args, *filter.Username)
		argIndex++
	}
	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND changed_at >= $%d", argIndex)
		args = append(args, *filter.DateFrom)
		argIndex++
	}
	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND changed_at <= $%d", argIndex)
		args = append(args, *filter.DateTo)
		argIndex++
	}

	query += " ORDER BY changed_at DESC LIMIT $1 OFFSET $2"
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
	}
	defer rows.Close()

	var records []*domain.HistoryRecord
	for rows.Next() {
		rec := &domain.HistoryRecord{}
		var oldDataJSON, newDataJSON []byte
		err := rows.Scan(&rec.ID, &rec.ItemID, &rec.Action, &oldDataJSON, &newDataJSON, &rec.ChangedBy, &rec.ChangedAt)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", customErr.ErrDatabase, err)
		}
		records = append(records, rec)
	}
	return records, nil
}

func (r *HistoryPostgresRepository) GetHistoryByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*domain.HistoryRecord, error) {
	return r.GetHistory(ctx, domain.HistoryFilter{
		ItemID: &itemID,
		Limit:  limit,
		Offset: offset,
	})
}
