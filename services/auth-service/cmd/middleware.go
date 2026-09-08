package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (app *application) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		app.logger.Info().Str("method", c.Request.Method).Str("path", c.Request.URL.Path).Int("status", c.Writer.Status()).Dur("latency", time.Since(start)).Msg("request")
	}
}
