package pool

import (
	"context"
	"database/sql"
	"time"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type ExecutorProvider interface {
	Executor(ctx context.Context) DBTX
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ErrorMapper interface {
	MapError(err error) error
}

type Pool interface {
	ExecutorProvider
	TxManager
	ErrorMapper
	OpTimeout() time.Duration
	Ping(ctx context.Context) error
	Close() error
}
