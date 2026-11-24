package mysql

import (
	"context"
	"errors"
	"fmt"

	coreerrors "server/internal/core/errors"
	"server/internal/core/repository/mysql/pool"
	"server/internal/features/users/repository/mysql/sqlc"
)

type Repository struct {
	pool pool.Pool
}

func NewRepository(p pool.Pool) *Repository {
	return &Repository{pool: p}
}

func (r *Repository) queries(ctx context.Context) *sqlc.Queries {
	return sqlc.New(r.pool.Executor(ctx))
}

func (r *Repository) withOpTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.pool.OpTimeout())
}

func (r *Repository) mapError(err error) error {
	mapped := r.pool.MapError(err)

	switch {
	case errors.Is(mapped, pool.ErrNoRows):
		return fmt.Errorf("%w: %v", coreerrors.ErrNotFound, mapped)
	case errors.Is(mapped, pool.ErrDuplicateEntry):
		return fmt.Errorf("%w: %v", coreerrors.ErrConflict, mapped)
	case errors.Is(mapped, pool.ErrViolatesForeignKey):
		return fmt.Errorf("%w: %v", coreerrors.ErrInvalidArgument, mapped)
	default:
		return mapped
	}
}
