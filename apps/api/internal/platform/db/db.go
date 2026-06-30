// Package db provides the PostgreSQL connection pool and transaction manager.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool for dependency injection and testability.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a configured pgxpool from a connection string.
// maxOpen and maxIdle are used to size the pool; values <= 0 fall back to pgx defaults.
func NewPool(ctx context.Context, connString string, maxOpen, maxIdle int) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	if maxOpen > 0 {
		cfg.MaxConns = int32(maxOpen)
	}
	if maxIdle > 0 {
		cfg.MinConns = int32(maxIdle)
	}
	cfg.MaxConnLifetime = time.Hour
	cfg.HealthCheckPeriod = time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Pool{Pool: pool}, nil
}

// Ping is a thin wrapper for readiness checks.
func (p *Pool) Ping(ctx context.Context) error {
	if p == nil || p.Pool == nil {
		return fmt.Errorf("pool not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return p.Pool.Ping(ctx)
}

// Close closes the pool.
func (p *Pool) Close() {
	if p != nil && p.Pool != nil {
		p.Pool.Close()
	}
}

// TxManager executes functions inside a database transaction.
type TxManager struct {
	pool *Pool
}

// NewTxManager creates a transaction manager from a pool.
func NewTxManager(pool *Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx runs fn inside a transaction and commits if fn returns nil.
// It rolls back on error and returns the original error.
func (tm *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	if tm.pool == nil || tm.pool.Pool == nil {
		return fmt.Errorf("pool not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
