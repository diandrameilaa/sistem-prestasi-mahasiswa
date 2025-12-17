package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"sistem-prestasi-mhs/app/helpers"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"
)

// Keys for fiber.Locals to avoid typos
const (
	LocalsUserID      = "userID"
	LocalsRoleID      = "roleID"
	LocalsRoleName    = "roleName"
	LocalsPermissions = "permissions"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helpers.ErrUnauthorized
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return helpers.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", "invalid token format")
		}

		userID, roleID, err := authService.ValidateToken(token)
		if err != nil {
			return helpers.ErrorResponse(c, fiber.StatusUnauthorized, "unauthorized", err.Error())
		}

		c.Locals(LocalsUserID, userID)
		c.Locals(LocalsRoleID, roleID)

		return c.Next()
	}
}

// RequireRole checks if user belongs to one of the required roles
func RequireRole(userRepo *repository.UserRepository, requiredRoles ...string) fiber.Handler {
	// Optimization: Convert required roles to map for O(1) lookup
	requiredMap := make(map[string]struct{}, len(requiredRoles))
	for _, r := range requiredRoles {
		requiredMap[r] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		roleID, ok := c.Locals(LocalsRoleID).(uuid.UUID)
		if !ok {
			return helpers.ErrUnauthorized
		}

		// Fetch role details
		role, err := userRepo.GetRole(c.Context(), roleID)
		if err != nil || role == nil {
			return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "user role not found")
		}

		// Check if user's role exists in required map
		if _, exists := requiredMap[role.Name]; !exists {
			return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "insufficient role permissions")
		}

		c.Locals(LocalsRoleName, role.Name)
		return c.Next()
	}
}

// RequirePermission ensures user has the specific permission
func RequirePermission(userRepo *repository.UserRepository, requiredPermission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms, err := getOrFetchPermissions(c, userRepo)
		if err != nil {
			return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "unable to verify permissions")
		}

		// Check permission
		if _, exists := perms[requiredPermission]; !exists {
			return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "access denied: missing permission")
		}

		return c.Next()
	}
}

// RequireAnyPermission ensures user has AT LEAST ONE of the required permissions
func RequireAnyPermission(userRepo *repository.UserRepository, requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms, err := getOrFetchPermissions(c, userRepo)
		if err != nil {
			return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "unable to verify permissions")
		}

		for _, req := range requiredPermissions {
			if _, exists := perms[req]; exists {
				return c.Next()
			}
		}

		return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "insufficient permissions")
	}
}

// RequireAllPermissions ensures user has ALL required permissions
func RequireAllPermissions(userRepo *repository.UserRepository, requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms, err := getOrFetchPermissions(c, userRepo)
		if err != nil {
			return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "unable to verify permissions")
		}

		for _, req := range requiredPermissions {
			if _, exists := perms[req]; !exists {
				return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "missing required permission: "+req)
			}
		}

		return c.Next()
	}
}

// RecoveryMiddleware handles panic recovery gracefully
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// In a real app, log the stack trace here
				helpers.ErrorResponse(c, fiber.StatusInternalServerError, "internal server error", "unexpected system failure")
			}
		}()
		return c.Next()
	}
}

// ----------------------------------------------------------------------------
// Internal Helpers
// ----------------------------------------------------------------------------

// getOrFetchPermissions retrieves permissions from Locals if available,
// otherwise fetches from DB and caches them in Locals as a Map for O(1) lookup.
func getOrFetchPermissions(c *fiber.Ctx, repo *repository.UserRepository) (map[string]struct{}, error) {
	// 1. Check if already cached in context
	if cached, ok := c.Locals(LocalsPermissions).(map[string]struct{}); ok {
		return cached, nil
	}

	// 2. Get User ID
	userID, ok := c.Locals(LocalsUserID).(uuid.UUID)
	if !ok {
		return nil, helpers.ErrUnauthorized
	}

	// 3. Fetch from DB
	permList, err := repo.GetPermissions(c.Context(), userID)
	if err != nil {
		return nil, err
	}

	// 4. Convert to Map for efficiency
	permMap := make(map[string]struct{}, len(permList))
	for _, p := range permList {
		permMap[p] = struct{}{}
	}

	// 5. Cache in Locals for next middlewares
	c.Locals(LocalsPermissions, permMap)

	return permMap, nil
}
