package mysql

import (
	"context"
	"fmt"
)

func (r *Repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	exists, err := r.queries(opCtx).UserExistsByEmail(opCtx, email)
	if err != nil {
		return false, fmt.Errorf("user exists by email: %w", r.mapError(err))
	}

	return exists, nil
}
