package jwt

import (
	"fmt"
	"time"

	"sso/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	AppID    int    `json:"app_id"`
}

func NewToken(user *domain.User, secret string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		UID:      user.ID,
		Username: user.Username,
		Role:     string(user.Role),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return tokenString, nil
}
