package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	RoleID       uuid.UUID `json:"role_id"`
	Role         *Role     `json:"role,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Role struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Permissions []*Permission   `json:"permissions,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Status       string `json:"status"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         *UserResponse `json:"user"`
}

type UserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	FullName    string     `json:"full_name"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
}

type CreateUserRequest struct {
	Username string    `json:"username" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
	Password string    `json:"password" validate:"required,min=8"`
	FullName string    `json:"full_name" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
}
