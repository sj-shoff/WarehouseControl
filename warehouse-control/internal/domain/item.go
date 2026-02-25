package domain

import "time"

type Item struct {
	ID        int64
	Name      string
	SKU       string
	Quantity  int
	Price     float64
	Category  string
	Location  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
