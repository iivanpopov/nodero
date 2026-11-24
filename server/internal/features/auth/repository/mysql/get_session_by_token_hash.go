package mysql

import (
	"context"
	"fmt"

	"server/internal/core/domain"
)

func (r *Repository) GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (*domain.Session, error) {
	opCtx, cancel := r.withOpTimeout(ctx)
	defer cancel()

	s, err := r.queries(opCtx).GetSessionByTokenHash(opCtx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("get session by token hash: %w", r.mapError(err))
	}

	return toDomainSession(s), nil
}
