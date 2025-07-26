package models

import "time"

// UserProfile represents a user profile
type UserProfile struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Bio       string    `json:"bio"`
	Avatar    string    `json:"avatar"`
	Company   string    `json:"company"`
	Website   string    `json:"website"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
