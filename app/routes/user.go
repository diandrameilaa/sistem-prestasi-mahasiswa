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

type UserHandler struct {
	userService *service.UserService
	authService *service.AuthService
}

func NewUserHandler(us *service.UserService, as *service.AuthService) *UserHandler {
	return &UserHandler{
		userService: us,
		authService: as,
	}
}

func SetupUserRoutes(router fiber.Router, db *sql.DB) {
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo)

	handler := NewUserHandler(userService, authService)

	router.Use(middleware.AuthMiddleware(authService))

	router.Get("/", handler.List)
	router.Get("/:id", handler.GetDetail)
	router.Post("/", handler.Create)
	router.Put("/:id", handler.Update)
	router.Delete("/:id", handler.Delete)
}

// List godoc
// @Summary List all users
// @Description Get paginated list of users
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [get]
func (h *UserHandler) List(c *fiber.Ctx) error {
	limit, offset := helpers.GetPaginationParams(c)

	users, err := h.userService.ListUsers(c.Context(), limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve users", err.Error())
	}

	return helpers.ListResponse(c, users, len(users), limit, offset)
}

// GetDetail godoc
// @Summary Get user detail
// @Description Get detailed information about a specific user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "User details"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetDetail(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	user, err := h.userService.GetUser(c.Context(), id)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusNotFound, "user not found", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "user retrieved", user)
}

// Create godoc
// @Summary Create new user
// @Description Register a new user in the system
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.CreateUserRequest true "User data"
// @Success 201 {object} map[string]interface{} "User created"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 422 {object} map[string]interface{} "Validation error"
// @Router /users [post]
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.ErrInvalidRequest
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "username, email, and password required")
	}
	if !helpers.ValidateEmail(req.Email) {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "invalid email format")
	}
	if !helpers.ValidatePassword(req.Password) {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "password is too weak (min 8 chars, mixed case, numbers, special chars)")
	}

	user, err := h.userService.CreateUser(c.Context(), &req, h.authService)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "failed to create user", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusCreated, "user created successfully", user)
}

// Update godoc
// @Summary Update user
// @Description Update existing user information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param user body models.User true "Updated user data"
// @Success 200 {object} map[string]interface{} "User updated"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	var updates models.User
	if err := c.BodyParser(&updates); err != nil {
		return helpers.ErrInvalidRequest
	}

	user, err := h.userService.UpdateUser(c.Context(), id, &updates)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "failed to update user", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "user updated successfully", user)
}

// Delete godoc
// @Summary Delete user
// @Description Soft delete a user (sets is_active to false)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "User deleted"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	if err := h.userService.DeleteUser(c.Context(), id); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "failed to delete user", err.Error())
	}

	return helpers.SuccessResponseWithoutData(c, fiber.StatusOK, "user deleted successfully")
}
