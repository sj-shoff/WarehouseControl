package middleware

import (
	"net/http"

	customErr "warehouse-control/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/zlog"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zlog.Logger.Error().
					Interface("error", err).
					Msg("Panic recovered")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": customErr.ErrInternal})
			}
		}()
		c.Next()
	}
}
