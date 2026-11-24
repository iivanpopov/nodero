package domain

import "time"

type User struct {
	ID        uint64
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
