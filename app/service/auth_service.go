package service

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
)

// AuthClaims extends standard JWT claims with custom fields
type AuthClaims struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
	Type   string `json:"type,omitempty"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	// Load secret once at startup
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-prod" // Fallback for dev safety
	}

	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(secret),
	}
}

// Login verifies credentials and returns user + tokens
func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, string, string, error) {
	// 1. Fetch User
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", "", errors.New("invalid username or password")
	}

	// 2. Verify Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid username or password")
	}

	// 3. Generate Tokens
	accessToken, err := s.GenerateToken(user.ID, user.RoleID)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.GenerateRefreshToken(user.ID, user.RoleID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// GenerateToken creates a short-lived access token
func (s *AuthService) GenerateToken(userID, roleID uuid.UUID) (string, error) {
	claims := AuthClaims{
		UserID: userID.String(),
		RoleID: roleID.String(),
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateRefreshToken creates a long-lived refresh token
func (s *AuthService) GenerateRefreshToken(userID, roleID uuid.UUID) (string, error) {
	claims := AuthClaims{
		UserID: userID.String(),
		RoleID: roleID.String(), // Added RoleID so refresh flow knows the role
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken parses and validates the JWT, returning UserID and RoleID
func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return uuid.Nil, uuid.Nil, errors.New("invalid user id in token")
		}

		roleID, err := uuid.Parse(claims.RoleID)
		if err != nil {
			// For refresh tokens created by older versions (without role_id), this might fail.
			// In strict mode, we return error. In lenient mode, return uuid.Nil.
			return uuid.Nil, uuid.Nil, errors.New("invalid role id in token")
		}

		return userID, roleID, nil
	}

	return uuid.Nil, uuid.Nil, errors.New("invalid token claims")
}

// HashPassword helper
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
