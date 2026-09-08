package postgresql

import (
	"context"
	"fmt"
	"kaffein/product-service/config"
	"kaffein/product-service/utils/constants"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Option func(*config.PostgresConfig)

func WithDSN(dsn string) Option {
	return func(cfg *config.PostgresConfig) {
		cfg.DSN = dsn
	}
}

func NewPool(ctx context.Context, cfg *config.Config, logger zerolog.Logger, opts ...Option) (*pgxpool.Pool, error) {
	pc := cfg.Postgres
	for _, opt := range opts {
		opt(&pc)
	}

	if pc.DSN == "" {
		return nil, fmt.Errorf(constants.ErrPostgresDSNEmpty)
	}

	poolCfg, err := pgxpool.ParseConfig(pc.DSN)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrPostgresParseDSN, err)
	}

	if pc.MaxConns > 0 {
		poolCfg.MaxConns = int32(pc.MaxConns)
	}
	if pc.MinConns > 0 {
		poolCfg.MinConns = int32(pc.MinConns)
	}
	if pc.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = pc.MaxConnLifetime
	}
	if pc.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = pc.MaxConnIdleTime
	}
	poolCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrPostgresCreatePool, err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf(constants.ErrPostgresPing, err)
	}

	logger.Info().
		Int32("max_conns", poolCfg.MaxConns).
		Int32("min_conns", poolCfg.MinConns).
		Str("dsn_host", poolCfg.ConnConfig.Host).
		Msg("postgres pool connected")

	return pool, nil
}
