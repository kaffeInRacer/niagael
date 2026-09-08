package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleTenant = "tenant"
	RoleStaff  = "staff"
	RoleAdmin  = "admin"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func ValidRole(role string) bool {
	return role == RoleTenant || role == RoleStaff || role == RoleAdmin
}
