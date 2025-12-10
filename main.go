package main

import (
	"log"
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/helper"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	// Correct import for fiber-swagger
	fiberSwagger "github.com/swaggo/fiber-swagger"

	// Import swagger docs
	_ "sistem-prestasi-mhs/docs"
)

// @title Student Achievement Reporting System API
// @version 1.0
// @description API for managing student achievements with RBAC
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@university.ac.id
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Load configuration
	config.LoadConfig()

	// 2. Connect to databases
	config.ConnectDB()
	config.ConnectMongo()

	// 3. Auto migrate PostgreSQL models
	err := config.DB.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.User{},
		&models.Student{},
		&models.Lecturer{},
		&models.AchievementReference{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 4. Seed initial data
	helper.SeedDatabase()

	// 5. Initialize Fiber app
	app := fiber.New()

	// 6. Middlewares
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, HEAD, PUT, DELETE, PATCH",
	}))

	// 7. Swagger Documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// 8. API Routes
	api := app.Group("/api/v1")
	{
		routes.AuthRoutes(api)
		routes.UserRoutes(api)
		routes.AchievementRoutes(api)
		routes.StudentRoutes(api)
	}

	// 9. Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status":  "success",
			"message": "Server is running",
		})
	})

	// 10. Start server
	port := config.GetEnv("PORT", "8080")
	log.Printf("🚀 Server running on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
