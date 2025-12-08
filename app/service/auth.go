package service

import (
	"errors"
	"sistem-prestasi-mhs/app/helper"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Login(username, password string) (string, string, *models.User, error)
	Register(user *models.User, password string) error
	ValidateToken(tokenString string) (*jwt.Token, error)
	RefreshToken(refreshToken string) (string, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(username, password string) (string, string, *models.User, error) {
	// Find user by username or email
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		user, err = s.userRepo.FindByEmail(username)
		if err != nil {
			return "", "", nil, errors.New("invalid credentials")
		}
	}

	// Check if user is active
	if !user.IsActive {
		return "", "", nil, errors.New("user account is inactive")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, errors.New("invalid credentials")
	}

	// Generate JWT tokens
	token, err := helper.GenerateToken(user)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := helper.GenerateRefreshToken(user)
	if err != nil {
		return "", "", nil, err
	}

	return token, refreshToken, user, nil
}

func (s *authService) Register(user *models.User, password string) error {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// Create user
	return s.userRepo.Create(user)
}

func (s *authService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return helper.ValidateToken(tokenString)
}

func (s *authService) RefreshToken(refreshToken string) (string, error) {
	token, err := helper.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	userID := claims["user_id"].(string)
	user, err := s.userRepo.FindByID(helper.ParseUUID(userID))
	if err != nil {
		return "", err
	}

	return helper.GenerateToken(user)
}
