package domain

import "time"

type ProductBatch struct {
	ID            int       `json:"id"`
	ProductID     int       `json:"product_id"`
	BatchNumber   string    `json:"batch_number"`
	ExpiryDate    string    `json:"expiry_date"` // Using string for date YYYY-MM-DD
	InitialQty    int       `json:"initial_qty"`
	CurrentStock  int       `json:"current_stock"`
	PurchasePrice float64   `json:"purchase_price"`
	SellingPrice  float64   `json:"selling_price"`
	IsVerified    bool      `json:"is_verified"`
	CreatedAt     time.Time `json:"created_at"`
}
