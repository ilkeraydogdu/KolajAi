package models

import "time"

// EmailRecord represents a record of an email sent to a user
type EmailRecord struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	EmailTo      string    `json:"email_to"`
	Subject      string    `json:"subject"`
	EmailType    string    `json:"email_type"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}
