package errors

import "errors"

var (
	ErrItemNotFound       = errors.New("item not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrDatabase           = errors.New("database error")
	ErrInternal           = errors.New("internal error")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRateLimit          = errors.New("rate limit exceeded")
)
