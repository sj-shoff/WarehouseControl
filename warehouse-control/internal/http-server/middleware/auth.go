package middleware

import (
	"context"
	"net/http"
	"strings"

	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/zlog"
)

type Claims struct {
	UserID   int64           `json:"user_id"`
	Username string          `json:"username"`
	Role     domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	secret string
	logger *zlog.Zerolog
}

func NewAuthMiddleware(secret string, logger *zlog.Zerolog) *AuthMiddleware {
	return &AuthMiddleware{secret: secret, logger: logger}
}

func (m *AuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn().Msg("Missing authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrUnauthorized.Error()})
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrUnauthorized.Error()})
			return
		}
		tokenString := parts[1]
		claims, err := m.validateToken(tokenString)
		if err != nil {
			m.logger.Warn().Err(err).Msg("Token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrUnauthorized.Error()})
			return
		}
		ctx := context.WithValue(c.Request.Context(), UserContextKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (m *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, customErr.ErrTokenInvalid
		}
		return []byte(m.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, customErr.ErrTokenInvalid
	}
	return claims, nil
}

func (m *AuthMiddleware) RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Request.Context().Value(UserContextKey).(*Claims)
		if !ok || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrUnauthorized.Error()})
			return
		}
		for _, role := range roles {
			if claims.Role == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": customErr.ErrForbidden.Error()})
	}
}

func GetClaimsFromContext(c *gin.Context) *Claims {
	claims, ok := c.Request.Context().Value(UserContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
