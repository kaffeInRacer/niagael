package dto

import "kaffein/auth-service/internal/domain"

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserList struct {
	Data     []domain.User `json:"data"`
	Count    int64         `json:"count"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

type RoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type StatusRequest struct {
	IsActive *bool `json:"is_active" binding:"required"`
}
