package domain

import "time"

type Setting struct {
	ID             int       `json:"id"`
	BusinessName   string    `json:"business_name"`
	Address        string    `json:"address"`
	Phone          string    `json:"phone"`
	CurrencySymbol string    `json:"currency_symbol"`
	Timezone       string    `json:"timezone"`
	UpdatedAt      time.Time `json:"updated_at"`
}
