package domain

import "time"

type Session struct {
	ID         uint64
	UserID     uint64
	ExpiresAt  time.Time
	LastUsedAt time.Time
	CreatedAt  time.Time
}

type AuthSession struct {
	Token     string
	ExpiresAt time.Time
}
