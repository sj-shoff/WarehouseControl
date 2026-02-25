package dto

import "time"

type CreateItemRequest struct {
	Name     string  `json:"name"`
	SKU      string  `json:"sku"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	Location string  `json:"location"`
}

type UpdateItemRequest struct {
	Name     string   `json:"name,omitempty"`
	SKU      string   `json:"sku,omitempty"`
	Quantity *int     `json:"quantity,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	Category string   `json:"category,omitempty"`
	Location string   `json:"location,omitempty"`
}

type ItemResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Category  string    `json:"category"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemsResponse struct {
	Items []*ItemResponse `json:"items"`
}
