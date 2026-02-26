package domain

import "time"

type Sale struct {
	ID                 int       `json:"id"`
	UserID             int       `json:"user_id"`
	TotalAmount        float64   `json:"total_amount"`
	Profit             float64   `json:"profit"`
	Discount           float64   `json:"discount"`
	PaymentMethod      string    `json:"payment_method"` // Cash, Transfer
	CustomerName       *string   `json:"customer_name"`
	DoctorName         *string   `json:"doctor_name"`
	PrescriptionNumber *string   `json:"prescription_number"`
	Status             string    `json:"status"` // active, pending_void, void
	VoidReason         *string   `json:"void_reason"`
	VoidRequestedBy    *int      `json:"void_requested_by"`
	CreatedAt          time.Time `json:"created_at"`
}

type SaleItem struct {
	ID           int     `json:"id"`
	SaleID       int     `json:"sale_id"`
	ProductID    int     `json:"product_id"`
	BatchID      int     `json:"batch_id"`
	Quantity     int     `json:"quantity"`
	Price        float64 `json:"price"`
	Subtotal     float64 `json:"subtotal"`
	SaleUnit     string  `json:"sale_unit"`
	ItemsPerUnit int     `json:"items_per_unit"`
}
