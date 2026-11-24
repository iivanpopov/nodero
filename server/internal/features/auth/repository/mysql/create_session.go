package mysql

import (
	"context"
	"fmt"
	"time"

	"server/internal/features/auth/repository/mysql/sqlc"
)

func (r *Repository) CreateSession(ctx context.Context, userID uint64, tokenHash []byte, expiresAt time.Time) error {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	err := r.queries(opCtx).CreateSession(opCtx, sqlc.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return fmt.Errorf("create session: %w", r.mapError(err))
	}

	return nil
}
