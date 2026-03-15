package history_postgres

import (
	"context"
	"fmt"
	"strings"

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

func (r *HistoryPostgresRepository) GetHistory(ctx context.Context, filter domain.HistoryFilter) ([]*domain.HistoryRecord, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ItemID != nil {
		conditions = append(conditions, fmt.Sprintf("item_id = $%d", argIndex))
		args = append(args, *filter.ItemID)
		argIndex++
	}
	if filter.Action != nil {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, *filter.Action)
		argIndex++
	}
	if filter.Username != nil {
		conditions = append(conditions, fmt.Sprintf("changed_by = $%d", argIndex))
		args = append(args, *filter.Username)
		argIndex++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("changed_at >= $%d", argIndex))
		args = append(args, *filter.DateFrom)
		argIndex++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("changed_at <= $%d", argIndex))
		args = append(args, *filter.DateTo)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM items_history %s", whereClause)
	var total int
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: count history error: %v", customErr.ErrDatabase, err)
	}
	if err := row.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("%w: scan total history error: %v", customErr.ErrDatabase, err)
	}

	if total == 0 {
		return []*domain.HistoryRecord{}, 0, nil
	}

	query := fmt.Sprintf(`
		SELECT id, item_id, action, old_data, new_data, changed_by, changed_at
		FROM items_history
		%s
		ORDER BY changed_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	finalArgs := append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryWithRetry(ctx, r.retries, query, finalArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: select history error: %v", customErr.ErrDatabase, err)
	}
	defer func() { _ = rows.Close() }()

	records := make([]*domain.HistoryRecord, 0, filter.Limit)
	for rows.Next() {
		rec := &domain.HistoryRecord{}
		err := rows.Scan(
			&rec.ID,
			&rec.ItemID,
			&rec.Action,
			&rec.OldData,
			&rec.NewData,
			&rec.ChangedBy,
			&rec.ChangedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("%w: scan history record: %v", customErr.ErrDatabase, err)
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("%w: history rows error: %v", customErr.ErrDatabase, err)
	}

	return records, total, nil
}

func (r *HistoryPostgresRepository) GetHistoryByItemID(ctx context.Context, itemID int64, limit, offset int) ([]*domain.HistoryRecord, error) {
	records, _, err := r.GetHistory(ctx, domain.HistoryFilter{
		ItemID: &itemID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("GetHistoryByItemID: %w", err)
	}
	return records, nil
}
