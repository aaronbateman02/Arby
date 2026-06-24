package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	p *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Pool{p: p}, nil
}

func (pool *Pool) HealthCheck(ctx context.Context) error {
	if pool == nil || pool.p == nil {
		return fmt.Errorf("pool not initialized")
	}
	return pool.p.Ping(ctx)
}

func (pool *Pool) Close() {
	if pool != nil && pool.p != nil {
		pool.p.Close()
	}
}

func (pool *Pool) P() *pgxpool.Pool {
	return pool.p
}
