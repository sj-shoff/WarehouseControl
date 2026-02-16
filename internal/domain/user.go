package domain

import "time"

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
	RoleViewer  UserRole = "viewer"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         UserRole
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
