package routes

import (
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func UserRoutes(r fiber.Router) {
	userRepo := repository.NewUserRepository()

	users := r.Group("/users")

	// Pastikan middleware ini sudah disesuaikan dengan format Fiber (return error)
	users.Use(middleware.AuthMiddleware())
	users.Use(middleware.RequirePermission("user:manage"))

	{
		// @Summary Get All Users
		// @Description Get list of all users (Admin only)
		// @Tags Users
		// @Produce json
		// @Security BearerAuth
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(10)
		// @Success 200 {object} object{status=string,data=object{items=array,total=int,page=int,limit=int}}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /users [get]
		users.Get("/", func(c *fiber.Ctx) error {
			// Fiber memiliki helper QueryInt untuk parsing integer langsung dengan default value
			page := c.QueryInt("page", 1)
			limit := c.QueryInt("limit", 10)
			offset := (page - 1) * limit

			userList, total, err := userRepo.FindAll(limit, offset)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"items": userList,
					"total": total,
					"page":  page,
					"limit": limit,
				},
			})
		})

		// @Summary Get User by ID
		// @Description Get user detail by ID (Admin only)
		// @Tags Users
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "User ID"
		// @Success 200 {object} object{status=string,data=models.User}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /users/{id} [get]
		users.Get("/:id", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			user, err := userRepo.FindByID(id)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "User not found",
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data":   user,
			})
		})

		// @Summary Create User
		// @Description Create new user (Admin only)
		// @Tags Users
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param user body object{username=string,email=string,password=string,full_name=string,role_id=string} true "User data"
		// @Success 201 {object} object{status=string,data=models.User}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /users [post]
		users.Post("/", func(c *fiber.Ctx) error {
			var req struct {
				Username string    `json:"username"`
				Email    string    `json:"email"`
				Password string    `json:"password"`
				FullName string    `json:"full_name"`
				RoleID   uuid.UUID `json:"role_id"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			// Validasi Manual (Fiber BodyParser tidak memproses tag 'binding' secara otomatis)
			if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" || req.RoleID == uuid.Nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "All fields are required",
				})
			}
			if len(req.Password) < 8 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Password must be at least 8 characters",
				})
			}

			// Hash password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": "Failed to hash password",
				})
			}

			user := &models.User{
				Username:     req.Username,
				Email:        req.Email,
				PasswordHash: string(hashedPassword),
				FullName:     req.FullName,
				RoleID:       req.RoleID,
				IsActive:     true,
			}

			if err := userRepo.Create(user); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"status": "success",
				"data":   user,
			})
		})

		// @Summary Update User
		// @Description Update user data (Admin only)
		// @Tags Users
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "User ID"
		// @Param user body object{username=string,email=string,full_name=string,is_active=boolean} true "User data"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /users/{id} [put]
		users.Put("/:id", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			user, err := userRepo.FindByID(id)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "User not found",
				})
			}

			var req struct {
				Username string `json:"username"`
				Email    string `json:"email"`
				FullName string `json:"full_name"`
				IsActive *bool  `json:"is_active"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			if req.Username != "" {
				user.Username = req.Username
			}
			if req.Email != "" {
				user.Email = req.Email
			}
			if req.FullName != "" {
				user.FullName = req.FullName
			}
			if req.IsActive != nil {
				user.IsActive = *req.IsActive
			}

			if err := userRepo.Update(user); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "User updated successfully",
			})
		})

		// @Summary Delete User
		// @Description Delete user (Admin only)
		// @Tags Users
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "User ID"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /users/{id} [delete]
		users.Delete("/:id", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			if err := userRepo.Delete(id); err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "User not found",
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "User deleted successfully",
			})
		})

		// @Summary Update User Role
		// @Description Update user role (Admin only)
		// @Tags Users
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "User ID"
		// @Param request body object{role_id=string} true "Role ID"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /users/{id}/role [put]
		users.Put("/:id/role", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			var req struct {
				RoleID uuid.UUID `json:"role_id"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			// Validasi Manual Role ID
			if req.RoleID == uuid.Nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "role_id is required",
				})
			}

			if err := userRepo.UpdateRole(id, req.RoleID); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "User role updated successfully",
			})
		})
	}
}
