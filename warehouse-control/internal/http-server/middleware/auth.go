package middleware

import (
	"net/http"
	"strings"

	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/zlog"
)

type Claims struct {
	jwt.RegisteredClaims
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	AppID    int    `json:"app_id"`
}

const UserContextKey = "user_claims"

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

		c.Set(UserContextKey, claims)
		c.Next()
	}
}

func (m *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, customErr.ErrInvalidToken
		}
		return []byte(m.secret), nil
	})

	if err != nil || !token.Valid {
		return nil, customErr.ErrInvalidToken
	}

	return claims, nil
}

func (m *AuthMiddleware) RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(UserContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrUnauthorized.Error()})
			return
		}

		claims, ok := val.(*Claims)
		if !ok || claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": customErr.ErrUnauthorized.Error()})
			return
		}

		hasAccess := false
		for _, role := range roles {
			if claims.Role == string(role) {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": customErr.ErrForbidden.Error()})
			return
		}

		c.Next()
	}
}

func GetClaimsFromContext(c *gin.Context) *Claims {
	val, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}

	claims, ok := val.(*Claims)
	if !ok {
		return nil
	}

	return claims
}
