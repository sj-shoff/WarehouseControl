package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		zlog.Logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("query", c.Request.URL.RawQuery).
			Msg("Request started")
		c.Next()
		duration := time.Since(start)
		zlog.Logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Dur("duration", duration).
			Msg("Request completed")
	}
}
