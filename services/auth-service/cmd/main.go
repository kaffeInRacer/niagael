package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"kaffein/auth-service/config"
	"kaffein/auth-service/pkg/logger"
	"kaffein/auth-service/utils/events"
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

	events.Init(cfg.Kafka.Brokers)

	app := &application{
		config: cfg,
		logger: l,
		pgx:    db,
		redis:  rdb,
	}
	if err := app.ServeHTTP(ctx); err != nil {
		l.Fatal().Err(err).Msg("HTTP server failed")
	}
}
