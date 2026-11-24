package service

import (
	"context"
	"fmt"

	"server/internal/core/domain"
)

func (s *Service) GetByID(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
