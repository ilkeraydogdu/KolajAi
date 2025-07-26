package models

import "time"

// Profile represents a user profile
type Profile struct {
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
