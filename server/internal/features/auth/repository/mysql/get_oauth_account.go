package mysql

import (
	"context"
	"fmt"

	"server/internal/core/domain"
	"server/internal/features/auth/repository/mysql/sqlc"
)

func (r *Repository) GetOAuthAccount(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.OAuthAccount, error) {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	acc, err := r.queries(opCtx).GetOAuthAccount(opCtx, sqlc.GetOAuthAccountParams{
		Provider:       string(provider),
		ProviderUserID: providerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("get oauth account: %w", r.mapError(err))
	}

	return toDomainOAuthAccount(acc), nil
}
