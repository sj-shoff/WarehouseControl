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

// ✅ ДОБАВЛЕНЫ CATEGORY И LOCATION
type ItemData struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	Location string  `json:"location"`
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
			Category: rec.OldData.Category,
			Location: rec.OldData.Location,
		}
	}
	if rec.NewData != nil {
		resp.NewData = &ItemData{
			ID:       rec.NewData.ID,
			Name:     rec.NewData.Name,
			SKU:      rec.NewData.SKU,
			Quantity: rec.NewData.Quantity,
			Price:    rec.NewData.Price,
			Category: rec.NewData.Category,
			Location: rec.NewData.Location,
		}
	}
	return resp
}

type DiffFieldDTO struct {
	Field  string      `json:"field"`
	Label  string      `json:"label"`
	Old    interface{} `json:"old"`
	New    interface{} `json:"new"`
	Status string      `json:"status"`
}

type DiffResponseDTO struct {
	RecordID  int64          `json:"record_id,omitempty"`
	Action    string         `json:"action"`
	ChangedBy string         `json:"changed_by"`
	ChangedAt string         `json:"changed_at"`
	Fields    []DiffFieldDTO `json:"fields"`
}

func MapToDTO(d domain.DiffResponse) DiffResponseDTO {
	dtoFields := make([]DiffFieldDTO, len(d.Fields))
	for i, f := range d.Fields {
		dtoFields[i] = DiffFieldDTO{
			Field:  f.Field,
			Label:  f.Label,
			Old:    f.Old,
			New:    f.New,
			Status: string(f.Status),
		}
	}
	return DiffResponseDTO{
		RecordID:  d.RecordID,
		Action:    d.Action,
		ChangedBy: d.ChangedBy,
		ChangedAt: d.ChangedAt,
		Fields:    dtoFields,
	}
}
