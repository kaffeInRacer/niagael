package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
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

	var kafkaWriter *kafka.Writer
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Topic != "" {
		kafkaWriter = &kafka.Writer{
			Addr:         kafka.TCP(cfg.Kafka.Brokers...),
			Topic:        cfg.Kafka.Topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
		}
		defer kafkaWriter.Close()
	}

	app := &application{
		config: cfg,
		logger: l,
		pgx:    db,
		redis:  rdb,
		kafka:  kafkaWriter,
	}
	if err := app.ServeHTTP(ctx); err != nil {
		l.Fatal().Err(err).Msg("HTTP server failed")
	}
}
