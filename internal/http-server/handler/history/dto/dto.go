package dto

import "time"

type HistoryFilterRequest struct {
	ItemID   *int64  `json:"item_id"`
	Action   *string `json:"action"`
	Username *string `json:"username"`
	DateFrom *string `json:"date_from"`
	DateTo   *string `json:"date_to"`
	Limit    int     `json:"limit"`
	Offset   int     `json:"offset"`
}

type HistoryResponse struct {
	Records []*HistoryRecordResponse `json:"history"`
}

type HistoryRecordResponse struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	Action    string    `json:"action"`
	OldData   *ItemData `json:"old_data,omitempty"`
	NewData   *ItemData `json:"new_data,omitempty"`
	ChangedBy string    `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
}

type ItemData struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}
