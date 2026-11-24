package mysql

import (
	"context"
	"fmt"

	"server/internal/core/domain"
)

func (r *Repository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	u, err := r.queries(opCtx).GetUserByID(opCtx, id)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", r.mapError(err))
	}

	return toDomainUser(u), nil
}
