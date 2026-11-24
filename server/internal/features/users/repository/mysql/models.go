package mysql

import (
	"server/internal/core/domain"
	"server/internal/features/users/repository/mysql/sqlc"
)

func toDomainUser(u sqlc.User) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
