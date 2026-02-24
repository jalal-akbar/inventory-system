package domain

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Never expose password in JSON
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
}
