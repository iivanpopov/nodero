package mysql

import (
	"context"
	"fmt"

	"server/internal/core/domain"
)

func (r *Repository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	u, err := r.queries(opCtx).GetUserByEmail(opCtx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", r.mapError(err))
	}

	return toDomainUser(u), nil
}
