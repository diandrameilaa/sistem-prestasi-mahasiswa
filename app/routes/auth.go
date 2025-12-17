package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"sistem-prestasi-mhs/app/helpers"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"
)

type AuthHandler struct {
	service  *service.AuthService
	userRepo *repository.UserRepository
}

func NewAuthHandler(s *service.AuthService, u *repository.UserRepository) *AuthHandler {
	return &AuthHandler{service: s, userRepo: u}
}

func SetupAuthRoutes(router fiber.Router, db *sql.DB) {
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	handler := NewAuthHandler(authService, userRepo)

	router.Post("/login", handler.Login)
	router.Post("/refresh", handler.Refresh)
	router.Get("/profile", middleware.AuthMiddleware(authService), handler.Profile)
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with tokens"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 422 {object} map[string]interface{} "Validation error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.ErrInvalidRequest
	}

	if req.Username == "" || req.Password == "" {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "username and password required")
	}

	user, token, refreshToken, err := h.service.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusUnauthorized, "login failed", err.Error())
	}

	return h.respondWithUser(c, user, token, refreshToken, "login successful")
}

// Refresh godoc
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param token body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} map[string]interface{} "New access token"
// @Failure 401 {object} map[string]interface{} "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return helpers.ErrInvalidRequest
	}

	if req.RefreshToken == "" {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "refresh_token required")
	}

	userID, roleID, err := h.service.ValidateToken(req.RefreshToken)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusUnauthorized, "invalid refresh token", err.Error())
	}

	token, err := h.service.GenerateToken(userID, roleID)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "token generation failed", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "token refreshed", fiber.Map{
		"token": token,
	})
}

// Profile godoc
// @Summary Get current user profile
// @Description Retrieve authenticated user's profile information
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /auth/profile [get]
func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	userID, err := helpers.GetUserIDFromLocals(c)
	if err != nil {
		return helpers.ErrUnauthorized
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusNotFound, "user not found", err.Error())
	}

	return h.respondWithUser(c, user, "", "", "profile retrieved")
}

func (h *AuthHandler) respondWithUser(c *fiber.Ctx, user *models.User, token, refreshToken, msg string) error {
	var roleName string
	if role, err := h.userRepo.GetRole(c.Context(), user.RoleID); err == nil && role != nil {
		roleName = role.Name
	}

	permissions, _ := h.userRepo.GetPermissions(c.Context(), user.ID)

	data := fiber.Map{
		"user": fiber.Map{
			"id":          user.ID,
			"username":    user.Username,
			"full_name":   user.FullName,
			"email":       user.Email,
			"role":        roleName,
			"permissions": permissions,
		},
	}

	if token != "" {
		data["token"] = token
		data["refresh_token"] = refreshToken
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, msg, data)
}
