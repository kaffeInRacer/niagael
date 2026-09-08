package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (app *application) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method

		event := app.logger.Info()
		if status >= 400 && status < 500 {
			event = app.logger.Warn()
		} else if status >= 500 {
			event = app.logger.Error()
		}

		event.
			Int("status", status).
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP()).
			Int("body_size", c.Writer.Size()).
			Msg("HTTP Request")
	}
}
