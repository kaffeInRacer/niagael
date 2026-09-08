package main

import (
	"context"
	"kaffein/dynamic-pricing-service/internal/auth"
	"kaffein/dynamic-pricing-service/pkg/logger"
	"kaffein/dynamic-pricing-service/pkg/postgresql"
	"kaffein/dynamic-pricing-service/pkg/redis"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"kaffein/dynamic-pricing-service/config"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	logger.New(c)
	l := logger.Get()

	pool, err := postgresql.NewPool(ctx, c, l)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to create postgres pool")
	}
	defer pool.Close()

	authPool, err := postgresql.NewPool(ctx, c, l, postgresql.WithDSN(c.AuthDB.DSN))
	if err != nil {
		l.Fatal().Err(err).Msg("failed to create auth postgres pool")
	}
	defer authPool.Close()

	redisClient, err := redis.NewCache(c, "dynamic-pricing-service", 15*time.Minute)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Client.Close()

	authService, err := auth.New(ctx, c.JWT.Secret, c.JWT.Issuer, c.JWT.RedisDB, authPool, redisClient.Client, c.Kafka.Brokers, c.Kafka.CasbinTopic, c.Kafka.GroupID)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to initialize authorization")
	}
	defer authService.Close()

	app := &application{
		config: c,
		logger: l,
		pgx:    pool,
		cache:  redisClient,
		auth:   authService,
		wg:     &sync.WaitGroup{},
	}

	l.Info().Str("mode", c.HTTP.Mode).Msg("application started")

	l.Info().
		Str("addr", c.HTTP.Addr).
		Str("grpc", c.GRPC.Server).
		Str("log_level", c.Logger.Level).
		Msg("loaded config")

	if err := app.ServeHTTP(ctx); err != nil {
		l.Error().Err(err).Msg("server failed")
	}
}
