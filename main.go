package main

import (
	"log"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/config"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Load Config
	config.LoadConfig()

	// 2. Connect Database
	config.ConnectDB()    // Postgres
	config.ConnectMongo() // MongoDB

	// 3. Auto Migration (PostgreSQL)
	log.Println("Running Auto Migration...")
	err := config.DB.AutoMigrate(
		&models.Role{},
		&models.Permission{},
		&models.User{},
		&models.Lecturer{},
		&models.Student{},
		&models.AchievementReference{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("✅ Database Migrated Successfully")

	// 4. Init Fiber App
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Sistem Pelaporan Prestasi Mahasiswa API is Running",
			"status":  "success",
		})
	})

	// 5. Listen
	port := config.GetEnv("APP_PORT", "3000")
	log.Fatal(app.Listen(":" + port))
}
