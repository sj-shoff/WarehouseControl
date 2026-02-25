package middleware

import (
	"context"
	"net/http"
	"strings"

	"warehouse-control/internal/domain"
	customErr "warehouse-control/internal/domain/errors"

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

const (
	UserContextKey contextKey = "user"
)

type AuthMiddleware struct {
	secret string
	logger *zlog.Zerolog
}

func NewAuthMiddleware(secret string, logger *zlog.Zerolog) *AuthMiddleware {
	return &AuthMiddleware{
		secret: secret,
		logger: logger,
	}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn().Msg("Missing authorization header")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Warn().Msg("Invalid authorization header format")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims, err := m.validateToken(tokenString)
		if err != nil {
			m.logger.Warn().Err(err).Msg("Token validation failed")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

func (m *AuthMiddleware) RequireRole(roles ...domain.UserRole) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*Claims)
			if !ok || claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			allowed := false
			for _, role := range roles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				m.logger.Warn().Str("role", string(claims.Role)).Msg("Access forbidden")
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) SetDBUser(username string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Устанавливаем пользователя для триггера БД
			// В реальном проекте это делается через connection pool
			next.ServeHTTP(w, r)
		})
	}
}

func GetClaimsFromContext(r *http.Request) *Claims {
	claims, ok := r.Context().Value(UserContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
