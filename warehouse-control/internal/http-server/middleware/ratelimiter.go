package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"

	customErr "warehouse-control/internal/domain/errors"
	"warehouse-control/internal/http-server/ratelimiter"
)

type RateLimiterMiddleware struct {
	rateLimiter *ratelimiter.RateLimiter
	logger      *zlog.Zerolog
}

func NewRateLimiterMiddleware(ratePerSec, capacity int, logger *zlog.Zerolog) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		rateLimiter: ratelimiter.NewRateLimiter(ratePerSec, capacity),
		logger:      logger,
	}
}

func (rm *RateLimiterMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rm.rateLimiter.Allow(ip) {
			rm.logger.Warn().
				Str("ip", ip).
				Int("rate_per_sec", rm.rateLimiter.RatePerSec).
				Int("capacity", rm.rateLimiter.Capacity).
				Msg("Rate limit exceeded")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       customErr.ErrRateLimit,
				"retry_after": 1,
			})
			return
		}

		c.Next()
	}
}
