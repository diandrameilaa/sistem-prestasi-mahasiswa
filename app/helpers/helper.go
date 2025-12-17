package helpers

import (
	"regexp"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ============================================================================
// 1. CONSTANTS & VARIABLES (Pre-compiled for performance)
// ============================================================================

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Common App Errors
var (
	ErrInvalidRequest = NewAppError(fiber.StatusBadRequest, "invalid request")
	ErrUnauthorized   = NewAppError(fiber.StatusUnauthorized, "unauthorized")
	ErrForbidden      = NewAppError(fiber.StatusForbidden, "forbidden")
	ErrNotFound       = NewAppError(fiber.StatusNotFound, "not found")
	ErrConflict       = NewAppError(fiber.StatusConflict, "conflict")
	ErrValidation     = NewAppError(fiber.StatusUnprocessableEntity, "validation error")
	ErrInternal       = NewAppError(fiber.StatusInternalServerError, "internal server error")
)

// ============================================================================
// 2. ERROR HANDLING
// ============================================================================

// AppError represents a structured application error.
// It implements the standard Go error interface.
type AppError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func NewAppErrorWithDetails(code int, message string, details interface{}) *AppError {
	return &AppError{Code: code, Message: message, Details: details}
}

// ============================================================================
// 3. HTTP RESPONSE HELPERS
// ============================================================================

type apiResponse struct {
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Details    interface{} `json:"details,omitempty"`
	Pagination *pagination `json:"pagination,omitempty"`
}

type pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func jsonResponse(c *fiber.Ctx, code int, status string, msg string, data, details interface{}, page *pagination) error {
	return c.Status(code).JSON(apiResponse{
		Status:     status,
		Message:    msg,
		Data:       data,
		Details:    details,
		Pagination: page,
	})
}

// SuccessResponse sends a standard success response with data
func SuccessResponse(c *fiber.Ctx, code int, message string, data interface{}) error {
	return jsonResponse(c, code, "success", message, data, nil, nil)
}

// SuccessResponseWithoutData sends a success response without payload
func SuccessResponseWithoutData(c *fiber.Ctx, code int, message string) error {
	return jsonResponse(c, code, "success", message, nil, nil, nil)
}

// ErrorResponse sends a standard error response
func ErrorResponse(c *fiber.Ctx, code int, message string, details interface{}) error {
	return jsonResponse(c, code, "error", message, nil, details, nil)
}

// ListResponse sends a paginated list response
func ListResponse(c *fiber.Ctx, data interface{}, total, limit, offset int) error {
	return jsonResponse(c, fiber.StatusOK, "success", "", data, nil, &pagination{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// ============================================================================
// 4. REQUEST & CONTEXT HELPERS
// ============================================================================

// GetPaginationParams extracts limit and offset using Fiber's built-in parser
func GetPaginationParams(c *fiber.Ctx) (limit, offset int) {
	limit = c.QueryInt("limit", 10)
	offset = c.QueryInt("offset", 0)

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return
}

// GetUUIDFromParams parses a UUID from the URL parameters
func GetUUIDFromParams(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	return uuid.Parse(c.Params(paramName))
}

// GetUserIDFromLocals retrieves the user ID (UUID) from Fiber Locals
func GetUserIDFromLocals(c *fiber.Ctx) (uuid.UUID, error) {
	id, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, ErrUnauthorized
	}
	return id, nil
}

// ============================================================================
// 5. VALIDATION HELPERS
// ============================================================================

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	return usernameRegex.MatchString(username)
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var (
		hasUpper, hasLower, hasNumber, hasSpecial bool
	)
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}

// ============================================================================
// 6. PERMISSION HELPERS
// ============================================================================

// CheckPermission verifies if a user has a specific permission
func CheckPermission(userID uuid.UUID, required string, permsMap map[uuid.UUID][]string) bool {
	userPerms, ok := permsMap[userID]
	if !ok {
		return false
	}
	for _, p := range userPerms {
		if p == required {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if user has AT LEAST ONE of the required permissions
func HasAnyPermission(userID uuid.UUID, required []string, permsMap map[uuid.UUID][]string) bool {
	userPerms, ok := permsMap[userID]
	if !ok {
		return false
	}

	// Create a map for O(1) lookup of user permissions
	permSet := make(map[string]struct{}, len(userPerms))
	for _, p := range userPerms {
		permSet[p] = struct{}{}
	}

	for _, req := range required {
		if _, exists := permSet[req]; exists {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if user has ALL required permissions
func HasAllPermissions(userID uuid.UUID, required []string, permsMap map[uuid.UUID][]string) bool {
	userPerms, ok := permsMap[userID]
	if !ok {
		return false
	}

	permSet := make(map[string]struct{}, len(userPerms))
	for _, p := range userPerms {
		permSet[p] = struct{}{}
	}

	for _, req := range required {
		if _, exists := permSet[req]; !exists {
			return false
		}
	}
	return true
}
