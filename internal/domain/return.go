package domain

import "time"

type Return struct {
	ID          int       `json:"id"`
	SaleID      int       `json:"sale_id"`
	UserID      int       `json:"user_id"`
	TotalRefund float64   `json:"total_refund"`
	Reason      *string   `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReturnItem struct {
	ID              int     `json:"id"`
	ReturnID        int     `json:"return_id"`
	SaleItemID      int     `json:"sale_item_id"`
	Quantity        int     `json:"quantity"`
	RefundAmount    float64 `json:"refund_amount"`
	ConditionStatus string  `json:"condition_status"` // good, damaged
}

type ReturnDetail struct {
	Return
	Items []ReturnItemDetail `json:"items"`
}

type ReturnItemDetail struct {
	ReturnItem
	ProductName string `json:"product_name"`
	BatchNumber string `json:"batch_number"`
}
