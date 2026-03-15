package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrDatabase           = errors.New("database error")
	ErrInternal           = errors.New("internal error")
	ErrInvalidToken       = errors.New("invalid token")
)
