package domain

import "time"

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleViewer  UserRole = "viewer"
)

func IsValidRole(role string) bool {
	return role == string(RoleAdmin) || role == string(RoleManager) || role == string(RoleViewer)
}

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserClaim struct {
	UserID   int64
	Username string
	Role     UserRole
}

type CreateUserRequest struct {
	Username string
	Password string
	Role     UserRole
}
