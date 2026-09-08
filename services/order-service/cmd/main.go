package main

import (
	"context"
	"kaffein/order-service/config"
	"kaffein/order-service/internal/auth"
	"kaffein/order-service/pkg/logger"
	"kaffein/order-service/pkg/postgresql"
	"kaffein/order-service/pkg/redis"
	"os/signal"
	"sync"
	"syscall"

	grpcclient "kaffein/order-service/pkg/grpc/client"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c, err := config.Load("config.yaml")
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

	redisClient, err := redis.New(c.Redis)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Client.Close()

	authService, err := auth.New(ctx, c.JWT.Secret, c.JWT.Issuer, c.JWT.RedisDB, pool, redisClient.Client)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to initialize authorization")
	}
	defer authService.Close()

	productClient, err := grpcclient.NewProductClient(c.GRPC.ProductServer)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to product service")
	}
	defer productClient.Close()

	pricingClient, err := grpcclient.NewDynamicPricingClient(c.GRPC.DynamicPricingServer)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to dynamic-pricing service")
	}
	defer pricingClient.Close()

	app := &application{
		config:        *c,
		logger:        l,
		pgx:           pool,
		redis:         redisClient,
		productClient: productClient,
		pricingClient: pricingClient,
		auth:          authService,
		wg:            &sync.WaitGroup{},
	}

	// Start background job for expired order cancellation
	go app.startOrderExpiryJob(ctx)

	l.Info().Str("mode", c.HTTP.Mode).Msg("application started")

	l.Info().
		Str("addr", c.HTTP.Addr).
		Msg("loaded config")

	if err := app.ServeHTTP(ctx); err != nil {
		l.Error().Err(err).Msg("server failed")
	}
}
