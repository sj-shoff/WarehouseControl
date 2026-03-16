package errors

import "errors"

var (
	ErrItemNotFound        = errors.New("item not found")
	ErrCreateUser          = errors.New("failed to create user")
	ErrNoTokenProvided     = errors.New("no token provided")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidRequest      = errors.New("invalid request body")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrDatabase            = errors.New("database error")
	ErrInternal            = errors.New("internal error")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrRateLimit           = errors.New("rate limit exceeded")
)
