package mysql

import (
	"context"
	"fmt"

	"server/internal/core/domain"
	"server/internal/features/auth/repository/mysql/sqlc"
)

func (r *Repository) CreateOAuthAccount(ctx context.Context, userID uint64, provider domain.OAuthProvider, providerUserID string) error {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	err := r.queries(opCtx).CreateOAuthAccount(opCtx, sqlc.CreateOAuthAccountParams{
		UserID:         userID,
		Provider:       string(provider),
		ProviderUserID: providerUserID,
	})
	if err != nil {
		return fmt.Errorf("create oauth account: %w", r.mapError(err))
	}

	return nil
}
