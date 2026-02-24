package domain

import "time"

type StockEntry struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	BatchID     int       `json:"batch_id"`
	Quantity    int       `json:"quantity"`
	Status      string    `json:"status"` // pending, approved, rejected
	IsVerified  bool      `json:"is_verified"`
	RequestedBy int       `json:"requested_by"`
	CreatedAt   time.Time `json:"created_at"`
}
