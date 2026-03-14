package dto

import "time"

type HistoryResponse struct {
	Records []*HistoryRecordResponse `json:"history"`
	Total   int                      `json:"total"`
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
