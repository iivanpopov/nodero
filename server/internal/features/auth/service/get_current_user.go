package service

import (
	"context"
	"fmt"

	"server/internal/core/domain"
)

func (s *Service) GetCurrentUser(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	return user, nil
}
