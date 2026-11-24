package service

import (
	"context"
	"fmt"
)

func (s *Service) Create(ctx context.Context, email string) (uint64, error) {
	id, err := s.repo.Create(ctx, email)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}

	return id, nil
}
