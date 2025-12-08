package helper

import (
	"errors"
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User) (string, error) {
	jwtSecret := config.GetEnv("JWT_SECRET", "rahasia_negara_api_123")

	// Extract permissions
	permissions := []string{}
	for _, perm := range user.Role.Permissions {
		permissions = append(permissions, perm.Name)
	}

	claims := Claims{
		UserID:      user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role.Name,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func GenerateRefreshToken(user *models.User) (string, error) {
	jwtSecret := config.GetEnv("JWT_SECRET", "rahasia_negara_api_123")

	claims := Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	jwtSecret := config.GetEnv("JWT_SECRET", "rahasia_negara_api_123")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	return token, err
}

func ParseUUID(uuidStr string) uuid.UUID {
	id, _ := uuid.Parse(uuidStr)
	return id
}
