package service

import (
	"context"
	"fmt"

	"server/internal/core/domain"
)

func (s *Service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}
