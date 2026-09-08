package main

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"kaffein/auth-service/internal/handler"
	"kaffein/auth-service/internal/middleware"
	"kaffein/auth-service/internal/repository"
	"kaffein/auth-service/internal/token"
	"kaffein/auth-service/internal/usecase"
)

func (app *application) routes() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), cors.New(cors.Config{AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"}, AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"}, AllowCredentials: true, MaxAge: 12 * time.Hour}), app.requestLogger())
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": true, "message": "ok"}) })

	users := repository.NewUserRepository(app.pgx)
	sessions := repository.NewSessionRepository(app.redis)
	tokens := token.NewManager(app.config.JWT)
	auth := middleware.Auth(tokens, sessions)

	handler.NewAuthHandler(usecase.NewAuthUseCase(users, sessions, tokens, app.config.JWT), app.logger, r, auth, app.config.JWT)

	handler.NewAdminHandler(usecase.NewAdminUseCase(users, sessions), app.logger, r, auth, middleware.Admin())

	rbacRepo := repository.NewRBACRepository(app.productDB, app.pricingDB, app.orderDB)
	rbacUseCase := usecase.NewRBACUseCase(rbacRepo)
	handler.NewRBACHandler(rbacUseCase, app.logger, r, auth, middleware.Admin())

	return r
}
