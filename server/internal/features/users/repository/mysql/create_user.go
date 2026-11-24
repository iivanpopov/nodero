package mysql

import (
	"context"
	"fmt"
)

func (r *Repository) Create(ctx context.Context, email string) (uint64, error) {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	id, err := r.queries(opCtx).CreateUser(opCtx, email)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", r.mapError(err))
	}

	return uint64(id), nil
}
