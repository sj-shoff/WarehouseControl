package dto

import (
	"time"
	"warehouse-control/internal/domain"
)

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

func ToHistoryRecordResponse(rec *domain.HistoryRecord) *HistoryRecordResponse {
	resp := &HistoryRecordResponse{
		ID:        rec.ID,
		ItemID:    rec.ItemID,
		Action:    rec.Action,
		ChangedBy: rec.ChangedBy,
		ChangedAt: rec.ChangedAt,
	}
	if rec.OldData != nil {
		resp.OldData = &ItemData{
			ID:       rec.OldData.ID,
			Name:     rec.OldData.Name,
			SKU:      rec.OldData.SKU,
			Quantity: rec.OldData.Quantity,
			Price:    rec.OldData.Price,
		}
	}
	if rec.NewData != nil {
		resp.NewData = &ItemData{
			ID:       rec.NewData.ID,
			Name:     rec.NewData.Name,
			SKU:      rec.NewData.SKU,
			Quantity: rec.NewData.Quantity,
			Price:    rec.NewData.Price,
		}
	}
	return resp
}
