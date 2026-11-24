package stdsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"server/internal/core/repository/mysql/pool"

	_ "github.com/go-sql-driver/mysql"
)

const defaultOpTimeout = 5 * time.Second

type txContextKey struct{}

type Config struct {
	DSN             string
	OpTimeout       time.Duration
	ConnMaxLifetime time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
}

type Pool struct {
	db        *sql.DB
	opTimeout time.Duration
}

var _ pool.Pool = (*Pool)(nil)

func New(cfg Config) (*Pool, error) {
	conn, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	if cfg.ConnMaxLifetime > 0 {
		conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	opTimeout := cfg.OpTimeout
	if opTimeout <= 0 {
		opTimeout = defaultOpTimeout
	}

	return &Pool{
		db:        conn,
		opTimeout: opTimeout,
	}, nil
}

func (p *Pool) OpTimeout() time.Duration {
	return p.opTimeout
}

func (p *Pool) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Pool) Close() error {
	return p.db.Close()
}

func (p *Pool) Executor(ctx context.Context) pool.DBTX {
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}

	return p.db
}

func (p *Pool) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
