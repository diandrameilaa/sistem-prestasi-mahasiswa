package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser handles new user registration
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest, authService *AuthService) (*models.User, error) {
	// 1. Hash Password using AuthService
	hashedPassword, err := authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 2. Prepare Model
	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 3. Persist
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	return s.repo.GetByID(ctx, id)
}

// UpdateUser handles partial updates for user profile
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, updates *models.User) (*models.User, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid user id")
	}

	// 1. Fetch existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Apply updates (Partial Update Logic)
	// Only update fields that are not empty string/zero value
	if updates.Username != "" {
		user.Username = updates.Username
	}
	if updates.Email != "" {
		user.Email = updates.Email
	}
	if updates.FullName != "" {
		user.FullName = updates.FullName
	}
	// Note: RoleID and Password updates usually require separate secure endpoints,
	// so we typically don't update them here unless explicitly needed.

	// 3. Persist changes
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser performs a soft delete via repository
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid user id")
	}
	return s.repo.Delete(ctx, id)
}

// ListUsers retrieves a paginated list of users
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return s.repo.List(ctx, limit, offset)
}

// GetPermissions retrieves permission strings for a specific user
// Note: Updated to take userID to match the optimized repository JOIN query
func (s *UserService) GetPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	return s.repo.GetPermissions(ctx, userID)
}
