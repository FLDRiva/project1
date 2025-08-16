package models

import "time"

type User struct {
	ID            int    `json:"id"`
	TelegramID    int64  `json:"telegram_id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Username      string `json:"username"`
	LanguageCode  string `json:"language_code"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}