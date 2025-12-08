package helper

import (
	"log"
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SeedDatabase() {
	log.Println("🌱 Starting database seeding...")

	// Seed Roles
	roles := []models.Role{
		{Name: "Admin", Description: "System administrator with full access"},
		{Name: "Mahasiswa", Description: "Student who can create and manage their achievements"},
		{Name: "Dosen Wali", Description: "Academic advisor who can verify student achievements"},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := config.DB.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			config.DB.Create(&role)
			log.Printf("✅ Created role: %s", role.Name)
		}
	}

	// Seed Permissions
	permissions := []models.Permission{
		{Name: "achievement:create", Resource: "achievement", Action: "create", Description: "Create achievement"},
		{Name: "achievement:read", Resource: "achievement", Action: "read", Description: "Read achievement"},
		{Name: "achievement:update", Resource: "achievement", Action: "update", Description: "Update achievement"},
		{Name: "achievement:delete", Resource: "achievement", Action: "delete", Description: "Delete achievement"},
		{Name: "achievement:verify", Resource: "achievement", Action: "verify", Description: "Verify achievement"},
		{Name: "user:manage", Resource: "user", Action: "manage", Description: "Manage users"},
	}

	for _, perm := range permissions {
		var existingPerm models.Permission
		if err := config.DB.Where("name = ?", perm.Name).First(&existingPerm).Error; err != nil {
			config.DB.Create(&perm)
			log.Printf("✅ Created permission: %s", perm.Name)
		}
	}

	// Assign Permissions to Roles
	var adminRole, mahasiswaRole, dosenRole models.Role
	config.DB.Where("name = ?", "Admin").First(&adminRole)
	config.DB.Where("name = ?", "Mahasiswa").First(&mahasiswaRole)
	config.DB.Where("name = ?", "Dosen Wali").First(&dosenRole)

	var allPerms []models.Permission
	config.DB.Find(&allPerms)

	// Admin gets all permissions
	config.DB.Model(&adminRole).Association("Permissions").Replace(allPerms)

	// Mahasiswa permissions
	var mahasiswaPerms []models.Permission
	config.DB.Where("name IN ?", []string{
		"achievement:create",
		"achievement:read",
		"achievement:update",
		"achievement:delete",
	}).Find(&mahasiswaPerms)
	config.DB.Model(&mahasiswaRole).Association("Permissions").Replace(mahasiswaPerms)

	// Dosen Wali permissions
	var dosenPerms []models.Permission
	config.DB.Where("name IN ?", []string{
		"achievement:read",
		"achievement:verify",
	}).Find(&dosenPerms)
	config.DB.Model(&dosenRole).Association("Permissions").Replace(dosenPerms)

	log.Println("✅ Permissions assigned to roles")

	// Create sample users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Admin user
	adminUser := models.User{
		ID:           uuid.New(),
		Username:     "admin",
		Email:        "admin@university.ac.id",
		PasswordHash: string(hashedPassword),
		FullName:     "System Administrator",
		RoleID:       adminRole.ID,
		IsActive:     true,
	}
	var existingAdmin models.User
	if err := config.DB.Where("username = ?", "admin").First(&existingAdmin).Error; err != nil {
		config.DB.Create(&adminUser)
		log.Println("✅ Created admin user (username: admin, password: password123)")
	}

	// Dosen user
	dosenUser := models.User{
		ID:           uuid.New(),
		Username:     "dosen001",
		Email:        "dosen@university.ac.id",
		PasswordHash: string(hashedPassword),
		FullName:     "Dr. John Doe",
		RoleID:       dosenRole.ID,
		IsActive:     true,
	}
	var existingDosen models.User
	if err := config.DB.Where("username = ?", "dosen001").First(&existingDosen).Error; err != nil {
		config.DB.Create(&dosenUser)

		// Create lecturer profile
		lecturer := models.Lecturer{
			ID:         uuid.New(),
			UserID:     dosenUser.ID,
			LecturerID: "NIP001",
			Department: "Teknik Informatika",
		}
		config.DB.Create(&lecturer)
		log.Println("✅ Created dosen user (username: dosen001, password: password123)")
	}

	// Mahasiswa user
	mahasiswaUser := models.User{
		ID:           uuid.New(),
		Username:     "mahasiswa123",
		Email:        "mahasiswa@university.ac.id",
		PasswordHash: string(hashedPassword),
		FullName:     "Jane Smith",
		RoleID:       mahasiswaRole.ID,
		IsActive:     true,
	}
	var existingMahasiswa models.User
	if err := config.DB.Where("username = ?", "mahasiswa123").First(&existingMahasiswa).Error; err != nil {
		config.DB.Create(&mahasiswaUser)

		// Get lecturer for advisor
		var lecturer models.Lecturer
		config.DB.Where("lecturer_id = ?", "NIP001").First(&lecturer)

		// Create student profile
		student := models.Student{
			ID:           uuid.New(),
			UserID:       mahasiswaUser.ID,
			StudentID:    "NIM123",
			ProgramStudy: "D-IV Teknik Informatika",
			AcademicYear: "2023/2024",
			AdvisorID:    &lecturer.ID,
		}
		config.DB.Create(&student)
		log.Println("✅ Created mahasiswa user (username: mahasiswa123, password: password123)")
	}

	log.Println("🎉 Database seeding completed!")
	log.Println("📝 Default credentials:")
	log.Println("   Admin    - username: admin, password: password123")
	log.Println("   Dosen    - username: dosen001, password: password123")
	log.Println("   Mahasiswa - username: mahasiswa123, password: password123")
}
