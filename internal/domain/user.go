package domain

import "time"

type User struct {
	ID              int64
	TelegramUserID  int64
	Phone           string
	FirstName       string
	LastName        string
	Username        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
