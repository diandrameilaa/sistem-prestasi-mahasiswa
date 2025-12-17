package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"

	"sistem-prestasi-mhs/app/database"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/routes"
	_ "sistem-prestasi-mhs/docs" // Import swagger docs
)

// @title Sistem Prestasi Mahasiswa API
// @version 1.0
// @description API documentation for Student Achievement Management System
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 1. Initialize Databases
	pgDB, err := database.InitPostgres()
	if err != nil {
		log.Fatalf("Postgres init failed: %v", err)
	}
	defer pgDB.Close()

	mongoClient, mongoDB, err := database.InitMongoDB()
	if err != nil {
		log.Fatalf("Mongo init failed: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting mongo: %v", err)
		}
	}()

	// 2. Setup Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Sistem Prestasi Mahasiswa API",
	})

	// 3. Global Middlewares
	app.Use(cors.New())
	app.Use(middleware.RecoveryMiddleware())
	app.Use(middleware.RequestLogger())

	// 4. Swagger Documentation Route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 5. API Routes Grouping
	api := app.Group("/api/v1")

	// 6. Register Routes
	routes.SetupAuthRoutes(api.Group("/auth"), pgDB)
	routes.SetupUserRoutes(api.Group("/users"), pgDB)
	routes.SetupStudentRoutes(api.Group("/students"), pgDB, mongoDB)
	routes.SetupLecturerRoutes(api.Group("/lecturers"), pgDB)
	routes.SetupAchievementRoutes(api.Group("/achievements"), pgDB, mongoDB)
	routes.SetupReportRoutes(api.Group("/reports"), pgDB, mongoDB)

	// 7. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger docs available at http://localhost:%s/swagger/", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
