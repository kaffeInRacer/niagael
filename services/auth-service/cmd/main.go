package main

import (
	"context"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"kaffein/auth-service/config"
	"kaffein/auth-service/pkg/logger"
	"kaffein/auth-service/pkg/postgresql"
	redisclient "kaffein/auth-service/pkg/redis"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	l := logger.New(cfg.Logger)
	db, err := postgresql.NewPool(ctx, cfg.Postgres)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect postgres")
	}
	defer db.Close()
	rdb, err := redisclient.New(cfg.Redis)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect redis")
	}
	defer rdb.Close()

	productDB := connectServiceDB(ctx, cfg, "product_service", l)
	if productDB != nil {
		defer productDB.Close()
	}
	pricingDB := connectServiceDB(ctx, cfg, "dynamic_pricing", l)
	if pricingDB != nil {
		defer pricingDB.Close()
	}
	orderDB := connectServiceDB(ctx, cfg, "order_service", l)
	if orderDB != nil {
		defer orderDB.Close()
	}

	app := &application{
		config:    cfg,
		logger:    l,
		pgx:       db,
		redis:     rdb,
		productDB: productDB,
		pricingDB: pricingDB,
		orderDB:   orderDB,
	}
	if err := app.ServeHTTP(ctx); err != nil {
		l.Fatal().Err(err).Msg("HTTP server failed")
	}
}

func connectServiceDB(ctx context.Context, cfg *config.Config, dbName string, l zerolog.Logger) *pgxpool.Pool {
	dsn := cfg.Postgres.DSN

	if idx := strings.Index(dsn, "dbname="); idx != -1 {
		endIdx := strings.IndexAny(dsn[idx:], " \n")
		if endIdx == -1 {
			dsn = dsn[:idx] + "dbname=" + dbName
		} else {
			dsn = dsn[:idx] + "dbname=" + dbName + dsn[idx+endIdx:]
		}
	} else {
		dsn = dsn + " dbname=" + dbName
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		l.Warn().Err(err).Str("database", dbName).Msg("failed to parse DSN for service database")
		return nil
	}

	poolCfg.MaxConns = int32(cfg.Postgres.MaxConns)
	poolCfg.MinConns = int32(cfg.Postgres.MinConns)
	poolCfg.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		l.Warn().Err(err).Str("database", dbName).Msg("failed to connect to service database")
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		l.Warn().Err(err).Str("database", dbName).Msg("failed to ping service database")
		pool.Close()
		return nil
	}

	l.Info().Str("database", dbName).Msg("connected to service database")
	return pool
}
