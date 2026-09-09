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

	"kaffein/order-service/internal/repository"
	grpcclient "kaffein/order-service/pkg/grpc/client"
	"kaffein/order-service/pkg/kafka"
	"kaffein/order-service/pkg/kafka/event"

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

	authPool, err := postgresql.NewPool(ctx, c, l, postgresql.WithDSN(c.AuthDB.DSN))
	if err != nil {
		l.Fatal().Err(err).Msg("failed to create auth postgres pool")
	}
	defer authPool.Close()

	redisClient, err := redis.New(c.Redis)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Client.Close()

	authService, err := auth.New(ctx, c, authPool, redisClient.Client, l)
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

	go app.startOrderExpiryJob(ctx)

	if len(c.Kafka.Brokers) > 0 {
		authPoolForSnapshot, err := postgresql.NewPool(ctx, c, l, postgresql.WithDSN(c.AuthDB.DSN))
		if err != nil {
			l.Error().Err(err).Msg("failed to connect auth db for buyers snapshot")
		} else {
			buyersRepo := repository.NewBuyersRepository(pool)
			if err := buyersRepo.Snapshot(ctx, authPoolForSnapshot); err != nil {
				l.Error().Err(err).Msg("failed to snapshot buyers from auth db")
			}
			authPoolForSnapshot.Close()

			buyersConsumer := kafka.NewBuyersConsumer(buyersRepo)
			go event.Consume(ctx, c.Kafka.Brokers, c.Kafka.UserTopic, c.Kafka.GroupID+"-users", buyersConsumer.Handle)
		}

		fulfillment := kafka.NewFulfillmentConsumer(productClient, pricingClient)
		go event.Consume(ctx, c.Kafka.Brokers, c.Kafka.OrderTopic, c.Kafka.GroupID+"-fulfillment", fulfillment.Handle)

		relay := kafka.NewOutboxRelay(c.Kafka.Brokers, c.Kafka.OrderTopic, outboxKafkaAdapter{repository.NewOutboxRepository(pool, postgresql.NewStore(pool))})
		go relay.Run(ctx)
		defer relay.Close()
	}

	l.Info().Str("mode", c.HTTP.Mode).Msg("application started")

	l.Info().
		Str("addr", c.HTTP.Addr).
		Msg("loaded config")

	if err := app.ServeHTTP(ctx); err != nil {
		l.Error().Err(err).Msg("server failed")
	}
}

type outboxKafkaAdapter struct {
	repo *repository.OutboxRepository
}

func (a outboxKafkaAdapter) ListUnprocessed(ctx context.Context, limit int) ([]kafka.OutboxEventRow, error) {
	events, err := a.repo.ListUnprocessed(ctx, limit)
	if err != nil {
		return nil, err
	}
	rows := make([]kafka.OutboxEventRow, 0, len(events))
	for _, e := range events {
		rows = append(rows, kafka.OutboxEventRow{ID: e.ID, Topic: e.Topic, Key: e.Key, Payload: e.Payload})
	}
	return rows, nil
}

func (a outboxKafkaAdapter) MarkProcessed(ctx context.Context, ids []int64) error {
	return a.repo.MarkProcessed(ctx, ids)
}
