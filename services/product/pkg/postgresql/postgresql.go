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

func NewPool(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*pgxpool.Pool, error) {
	if cfg.Postgres.DSN == "" {
		return nil, fmt.Errorf(constants.ErrPostgresDSNEmpty)
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.Postgres.DSN)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrPostgresParseDSN, err)
	}

	if cfg.Postgres.MaxConns > 0 {
		poolCfg.MaxConns = int32(cfg.Postgres.MaxConns)
	}
	if cfg.Postgres.MinConns > 0 {
		poolCfg.MinConns = int32(cfg.Postgres.MinConns)
	}
	if cfg.Postgres.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
	}
	if cfg.Postgres.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime
	}
	// health check interval default 1m
	poolCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrPostgresCreatePool, err)
	}

	// ping
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
