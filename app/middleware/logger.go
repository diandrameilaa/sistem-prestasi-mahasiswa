package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ANSI Color Codes for terminal output
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// RequestLogger middleware logs incoming HTTP requests with detailed info
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Capture metrics
		latency := time.Since(start)
		statusCode := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		clientIP := c.IP()

		// Determine status color
		statusColor := colorGreen
		switch {
		case statusCode >= 500:
			statusColor = colorRed
		case statusCode >= 400:
			statusColor = colorYellow
		case statusCode >= 300:
			statusColor = colorCyan
		}

		// Get UserID if authenticated (safe type assertion)
		var userStr string
		if userID, ok := c.Locals("userID").(uuid.UUID); ok {
			userStr = fmt.Sprintf(" | User: %s", colorBlue+userID.String()+colorReset)
		} else {
			userStr = " | Guest"
		}

		// Format: [TIME] | STATUS | LATENCY | IP | METHOD | PATH | USER
		logMsg := fmt.Sprintf("%s[%s]%s | %s%3d%s | %10v | %15s | %-7s %s%s\n",
			colorCyan,
			start.Format("15:04:05"),
			colorReset,
			statusColor, statusCode, colorReset,
			latency,
			clientIP,
			method,
			path,
			userStr,
		)

		// Write to stdout
		_, _ = os.Stdout.WriteString(logMsg)

		return err
	}
}
