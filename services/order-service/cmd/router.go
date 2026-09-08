package main

import (
	"context"
	"kaffein/order-service/internal/handler"
	"kaffein/order-service/internal/repository"
	"kaffein/order-service/internal/usecase"
	"kaffein/order-service/pkg/midtrans"
	"kaffein/order-service/pkg/postgresql"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (app *application) routes(ctx context.Context) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		app.requestLogger(),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": true, "message": "ok"})
	})

	store := postgresql.NewStore(app.pgx)
	midtransClient := midtrans.New(
		app.config.Midtrans.ServerKey,
		app.config.Midtrans.ClientKey,
		app.config.Midtrans.Environment,
	)

	// Cart
	cartRepo := repository.NewCartRepository(store)
	cartUseCase := usecase.NewCartUseCase(cartRepo, app.productClient, app.pricingClient)
	handler.NewCartHandler(cartUseCase, app.logger, r, app.auth)

	// Address
	addressRepo := repository.NewAddressRepository(store)
	addressUseCase := usecase.NewAddressUseCase(addressRepo)
	handler.NewAddressHandler(addressUseCase, app.logger, r, app.auth)

	// Order
	orderRepo := repository.NewOrderRepository(store)
	orderItemRepo := repository.NewOrderItemRepository(store)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, app.productClient, app.pricingClient)
	handler.NewOrderHandler(orderUseCase, addressUseCase, app.logger, r, app.auth)

	// Payment
	paymentRepo := repository.NewPaymentRepository(store)
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo, orderRepo, orderItemRepo, midtransClient, app.productClient)
	handler.NewPaymentHandler(paymentUseCase, orderUseCase, app.logger, r, app.auth)

	return r
}
