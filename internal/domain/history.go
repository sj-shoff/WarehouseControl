package domain

import "time"

type HistoryRecord struct {
	ID        int64
	ItemID    int64
	Action    string
	OldData   *Item
	NewData   *Item
	ChangedBy string
	ChangedAt time.Time
}

type HistoryFilter struct {
	ItemID   *int64
	Action   *string
	Username *string
	DateFrom *time.Time
	DateTo   *time.Time
	Limit    int
	Offset   int
}
