package routes

import (
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router) {
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)

	auth := r.Group("/auth")
	{
		// @Summary User Login
		// @Description Authenticate user and get JWT token
		// @Tags Authentication
		// @Accept json
		// @Produce json
		// @Param request body object{username=string,password=string} true "Login credentials"
		// @Success 200 {object} object{status=string,data=object{token=string,refreshToken=string,user=object}}
		// @Failure 400 {object} object{status=string,message=string}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /auth/login [post]
		auth.Post("/login", func(c *fiber.Ctx) error {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			// Menggunakan BodyParser di Fiber
			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid request body",
				})
			}

			// Validasi manual jika field kosong (karena BodyParser Fiber tidak se-strict Gin Binding)
			if req.Username == "" || req.Password == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Username and password are required",
				})
			}

			token, refreshToken, user, err := authService.Login(req.Username, req.Password)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			// Extract permissions
			permissions := []string{}
			for _, perm := range user.Role.Permissions {
				permissions = append(permissions, perm.Name)
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"token":        token,
					"refreshToken": refreshToken,
					"user": fiber.Map{
						"id":          user.ID,
						"username":    user.Username,
						"email":       user.Email,
						"full_name":   user.FullName,
						"role":        user.Role.Name,
						"permissions": permissions,
					},
				},
			})
		})

		// @Summary Refresh Token
		// @Description Get new access token using refresh token
		// @Tags Authentication
		// @Accept json
		// @Produce json
		// @Param request body object{refreshToken=string} true "Refresh token"
		// @Success 200 {object} object{status=string,data=object{token=string}}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /auth/refresh [post]
		auth.Post("/refresh", func(c *fiber.Ctx) error {
			var req struct {
				RefreshToken string `json:"refreshToken"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid request body",
				})
			}

			if req.RefreshToken == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "RefreshToken is required",
				})
			}

			token, err := authService.RefreshToken(req.RefreshToken)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"token": token,
				},
			})
		})

		// @Summary Get User Profile
		// @Description Get authenticated user profile
		// @Tags Authentication
		// @Produce json
		// @Security BearerAuth
		// @Success 200 {object} object{status=string,data=object}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /auth/profile [get]
		// NOTE: Pastikan AuthMiddleware juga sudah diubah support Fiber
		auth.Get("/profile", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
			// Di Fiber, data dari middleware diambil menggunakan c.Locals
			userID := c.Locals("user_id")
			username := c.Locals("username")
			role := c.Locals("role")
			permissions := c.Locals("permissions")

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"user_id":     userID,
					"username":    username,
					"role":        role,
					"permissions": permissions,
				},
			})
		})

		// @Summary Logout
		// @Description Logout user (client should delete token)
		// @Tags Authentication
		// @Produce json
		// @Security BearerAuth
		// @Success 200 {object} object{status=string,message=string}
		// @Router /auth/logout [post]
		auth.Post("/logout", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Logged out successfully",
			})
		})
	}
}
