package main

import (
	"time"

	"kaffein/auth-service/internal/gateway"
	"kaffein/auth-service/internal/handler"
	"kaffein/auth-service/internal/middleware"
	"kaffein/auth-service/internal/repository"
	"kaffein/auth-service/internal/token"
	"kaffein/auth-service/internal/usecase"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true, MaxAge: 12 * time.Hour}),
		app.requestLogger(),
	)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": true, "message": "ok"}) })

	users := repository.NewUserRepository(app.pgx, app.redis)
	sessions := gateway.NewSessionRepository(app.redis)
	tokens := token.NewManager(app.config.JWT)
	authorization := middleware.NewService(tokens, sessions, users, app.pgx, app.logger)

	handler.NewAuthHandler(
		usecase.NewAuthUseCase(users, sessions, tokens, app.config.JWT, app.config.Kafka.Brokers),
		app.logger, r, authorization, app.config.JWT,
	)

	handler.NewAdminHandler(usecase.NewAdminUseCase(users, sessions, app.config.Kafka.Brokers), app.logger, r, authorization)

	rbacRepo := repository.NewRBACRepository(app.pgx, app.config.Kafka.CasbinTopic, app.config.Kafka.Brokers)
	rbacUseCase := usecase.NewRBACUseCase(rbacRepo)
	handler.NewRBACHandler(rbacUseCase, app.logger, r, authorization)

	return r
}
