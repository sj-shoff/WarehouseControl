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
	ErrAuthToken          = errors.New("authorization token is missing")
	ErrAdminPrivilege     = errors.New("admin privilege required")
	ErrGetClaims          = errors.New("failed to get claims")
	ErrMetadataMiss       = errors.New("metadata is missing")
	ErrBootstrap          = errors.New("failed to bootstrap")
)
