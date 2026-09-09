package main

import (
	"context"
	"kaffein/product-service/internal/auth"
	"kaffein/product-service/pkg/logger"
	"kaffein/product-service/pkg/postgresql"
	"kaffein/product-service/pkg/rdb"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"kaffein/product-service/config"
	"kaffein/product-service/internal/repository"
	grpcclient "kaffein/product-service/pkg/grpc/client"
	kafkaevent "kaffein/product-service/pkg/kafka/event"
	"kaffein/product-service/pkg/kafka"

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

	redisClient, err := rdb.NewCache(c, "product-service", 15*time.Minute)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Client.Close()

	authService, err := auth.New(ctx, c, authPool, redisClient.Client, l)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to initialize authorization")
	}
	defer authService.Close()

	pricingClient, err := grpcclient.NewDynamicPricingClient(c.GRPC.DynamicPricing)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to dynamic-pricing service")
	}
	defer pricingClient.Close()

	app := &application{
		config:        c,
		logger:        l,
		pgx:           pool,
		cache:         redisClient,
		pricingClient: pricingClient,
		auth:          authService,
		wg:            &sync.WaitGroup{},
	}

	l.Info().Str("mode", c.HTTP.Mode).Msg("application started")

	if len(c.Kafka.Brokers) > 0 && c.Kafka.FlashSaleTopic != "" {
		projectionRepo := repository.NewFlashSaleProjectionRepository(pool)
		flashSaleConsumer := kafka.NewFlashSaleConsumer(projectionRepo)
		go kafkaevent.Consume(ctx, c.Kafka.Brokers, c.Kafka.FlashSaleTopic, c.Kafka.GroupID+"-flashsale", flashSaleConsumer.Handle)
	}

	l.Info().
		Str("addr", c.HTTP.Addr).
		Str("grpc", c.GRPC.Server).
		Str("log_level", c.Logger.Level).
		Msg("loaded config")

	if err := app.ServeHTTP(ctx); err != nil {
		l.Error().Err(err).Msg("server failed")
	}
}
