package mysql

import (
	"context"
	"fmt"
	"time"

	"server/internal/features/auth/repository/mysql/sqlc"
)

func (r *Repository) UpdateSession(ctx context.Context, sessionID uint64, expiresAt time.Time) error {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	err := r.queries(opCtx).UpdateSession(opCtx, sqlc.UpdateSessionParams{
		ID:        sessionID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("update session: %w", r.mapError(err))
	}

	return nil
}
