package mysql

import (
	"context"
	"fmt"
)

func (r *Repository) DeleteExpiredSessions(ctx context.Context) error {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	if err := r.queries(opCtx).DeleteExpiredSessions(opCtx); err != nil {
		return fmt.Errorf("delete expired sessions: %w", r.mapError(err))
	}

	return nil
}
