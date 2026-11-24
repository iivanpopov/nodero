package service

import (
	"context"

	"server/internal/core/domain"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uint64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, email string) (uint64, error)
}

type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{
		repo: repo,
	}
}
