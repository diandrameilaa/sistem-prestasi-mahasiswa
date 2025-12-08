package middleware

import (
	"sistem-prestasi-mhs/app/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Di Fiber, c.Get mengambil header value
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Authorization header required",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid authorization format",
			})
		}

		tokenString := parts[1]
		token, err := helper.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or expired token",
			})
		}

		// Set claims to locals (pengganti context di Gin)
		if claims, ok := token.Claims.(*helper.Claims); ok {
			c.Locals("user_id", claims.UserID)
			c.Locals("username", claims.Username)
			c.Locals("role", claims.Role)
			c.Locals("permissions", claims.Permissions)
		}

		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Mengambil data dari Locals
		permissionsInterface := c.Locals("permissions")
		if permissionsInterface == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "No permissions found",
			})
		}

		// Type assertion karena Locals mengembalikan interface{}
		permList, ok := permissionsInterface.([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid permissions format",
			})
		}

		// Check if user has required permission
		hasPermission := false
		for _, p := range permList {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Insufficient permissions",
			})
		}

		return c.Next()
	}
}

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Mengambil dan melakukan type assertion untuk role
		userRole, ok := c.Locals("role").(string)

		if !ok || userRole != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Insufficient role privileges",
			})
		}

		return c.Next()
	}
}
