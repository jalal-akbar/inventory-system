package domain

import "time"

type ActivityLog struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"` // Joined from users table
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}
