package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// Konfigurasi Database
const (
	DefaultDSN = "postgres://postgres:postgre@localhost:5432/achievement_db?sslmode=disable"
)

func main() {
	// Flags untuk opsi command line
	action := flag.String("action", "up", "Options: up (migrate + seed), down (drop tables), fresh (drop + migrate + seed)")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "Database connection string")
	flag.Parse()

	if *dsn == "" {
		*dsn = DefaultDSN
	}

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		log.Fatalf("❌ Connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Database unreachable: %v", err)
	}
	log.Println("✅ Connected to Database")

	switch *action {
	case "down":
		dropTables(db)
	case "fresh":
		dropTables(db)
		createTables(db)
		seedDatabase(db)
	case "up":
		createTables(db)
		seedDatabase(db)
	default:
		log.Fatal("Invalid action. Use -action=up, -action=down, or -action=fresh")
	}
}

// ============================================================================
// 1. MIGRATION (DDL)
// ============================================================================

func dropTables(db *sql.DB) {
	log.Println("🔥 Dropping all tables...")
	query := `
		DROP TABLE IF EXISTS achievement_references CASCADE;
		DROP TABLE IF EXISTS students CASCADE;
		DROP TABLE IF EXISTS lecturers CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
		DROP TABLE IF EXISTS role_permissions CASCADE;
		DROP TABLE IF EXISTS permissions CASCADE;
		DROP TABLE IF EXISTS roles CASCADE;
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("❌ Failed to drop tables: %v", err)
	}
	log.Println("✅ Tables dropped.")
}

func createTables(db *sql.DB) {
	log.Println("🏗️  Migrating database schema...")

	queries := []string{
		// 1. Roles
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		);`,

		// 2. Permissions
		`CREATE TABLE IF NOT EXISTS permissions (
			id UUID PRIMARY KEY,
			name VARCHAR(50) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);`,

		// 3. Role Permissions (Many-to-Many)
		`CREATE TABLE IF NOT EXISTS role_permissions (
			role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
			permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
			PRIMARY KEY (role_id, permission_id)
		);`,

		// 4. Users
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(100) NOT NULL,
			role_id UUID REFERENCES roles(id),
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		// 5. Lecturers (Extension of Users)
		`CREATE TABLE IF NOT EXISTS lecturers (
			id UUID PRIMARY KEY,
			user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			lecturer_id VARCHAR(20) UNIQUE NOT NULL, -- NIP
			department VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		);`,

		// 6. Students (Extension of Users)
		`CREATE TABLE IF NOT EXISTS students (
			id UUID PRIMARY KEY,
			user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
			student_id VARCHAR(20) UNIQUE NOT NULL, -- NIM
			program_study VARCHAR(100) NOT NULL,
			academic_year VARCHAR(10) NOT NULL,
			advisor_id UUID REFERENCES lecturers(id),
			created_at TIMESTAMP DEFAULT NOW()
		);`,

		// 7. Achievement References (Links to MongoDB)
		`CREATE TABLE IF NOT EXISTS achievement_references (
			id UUID PRIMARY KEY,
			student_id UUID REFERENCES students(id) ON DELETE CASCADE,
			mongo_achievement_id VARCHAR(50) NOT NULL,
			status VARCHAR(20) DEFAULT 'draft', -- draft, submitted, verified, rejected, deleted
			submitted_at TIMESTAMP,
			verified_at TIMESTAMP,
			verified_by UUID REFERENCES users(id),
			rejection_note TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatalf("❌ Migration failed on query:\n%s\nError: %v", q, err)
		}
	}
	log.Println("✅ Migration completed.")
}

// ============================================================================
// 2. SEEDER (Data Population with Dummy Data)
// ============================================================================

func seedDatabase(db *sql.DB) {
	log.Println("🌱 Seeding initial data...")

	// Cek apakah data sudah ada (prevent duplicate seeding)
	var count int
	db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&count)
	if count > 0 {
		log.Println("⚠️  Data already exists. Skipping seeder.")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("❌ Failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // Rollback jika ada error/panic

	// --- 1. Seed Permissions ---
	log.Println("📝 Seeding permissions...")
	perms := []string{
		"achievement:create",
		"achievement:read",
		"achievement:update",
		"achievement:delete",
		"achievement:verify",
		"report:view",
		"user:manage",
	}
	permMap := make(map[string]uuid.UUID)

	for _, p := range perms {
		id := uuid.New()
		permMap[p] = id
		_, err := tx.Exec("INSERT INTO permissions (id, name) VALUES ($1, $2)", id, p)
		if err != nil {
			log.Fatalf("❌ Failed to seed permission %s: %v", p, err)
		}
	}

	// --- 2. Seed Roles ---
	log.Println("👥 Seeding roles...")
	roles := map[string]uuid.UUID{
		"Admin":      uuid.New(),
		"Dosen Wali": uuid.New(),
		"Mahasiswa":  uuid.New(),
	}

	for name, id := range roles {
		_, err := tx.Exec("INSERT INTO roles (id, name, description) VALUES ($1, $2, $3)",
			id, name, "Role for "+name)
		if err != nil {
			log.Fatalf("❌ Failed to seed role %s: %v", name, err)
		}
	}

	// --- 3. Assign Permissions to Roles ---
	log.Println("🔗 Assigning permissions to roles...")

	// Admin: Full access
	assignPermission(tx, roles["Admin"],
		permMap["user:manage"],
		permMap["report:view"],
		permMap["achievement:read"],
		permMap["achievement:verify"])

	// Dosen Wali: Read, verify, and view reports
	assignPermission(tx, roles["Dosen Wali"],
		permMap["achievement:read"],
		permMap["achievement:verify"],
		permMap["report:view"])

	// Mahasiswa: CRUD their own achievements
	assignPermission(tx, roles["Mahasiswa"],
		permMap["achievement:create"],
		permMap["achievement:read"],
		permMap["achievement:update"],
		permMap["achievement:delete"])

	// --- 4. Seed Users & Profiles ---
	password, _ := hashPassword("password123") // Default password

	// Store IDs for relationships
	var lecturerIDs []uuid.UUID
	var lecturerUserIDs []uuid.UUID // ✅ ADDED: Store user IDs of lecturers
	var studentIDs []uuid.UUID

	// A. Admin Users (2 admins)
	log.Println("👨‍💼 Seeding admin users...")
	adminUsers := []struct {
		username string
		email    string
		fullname string
	}{
		{"admin", "admin@university.ac.id", "Super Admin"},
		{"admin2", "admin2@university.ac.id", "Admin Sistem"},
	}

	for _, admin := range adminUsers {
		adminID := uuid.New()
		mustInsertUser(tx, adminID, admin.username, admin.email, password, admin.fullname, roles["Admin"])
		log.Printf("   ✅ Admin: %s (%s)", admin.username, admin.email)
	}

	// B. Lecturer Users (5 lecturers)
	log.Println("👨‍🏫 Seeding lecturer users...")
	lecturers := []struct {
		username   string
		email      string
		fullname   string
		lecturerID string
		department string
	}{
		{"dosen1", "budi.santoso@university.ac.id", "Dr. Budi Santoso, M.Kom", "NIP198001011", "Teknik Informatika"},
		{"dosen2", "siti.rahayu@university.ac.id", "Dr. Siti Rahayu, M.T", "NIP198502022", "Teknik Informatika"},
		{"dosen3", "ahmad.wijaya@university.ac.id", "Dr. Ahmad Wijaya, M.Sc", "NIP199003033", "Sistem Informasi"},
		{"dosen4", "rina.kusuma@university.ac.id", "Rina Kusuma, M.Kom", "NIP198804044", "Teknik Informatika"},
		{"dosen5", "dedi.prakoso@university.ac.id", "Dedi Prakoso, M.T", "NIP199205055", "Sistem Informasi"},
	}

	for _, lec := range lecturers {
		lecUserID := uuid.New()
		lecID := uuid.New()

		mustInsertUser(tx, lecUserID, lec.username, lec.email, password, lec.fullname, roles["Dosen Wali"])

		_, err = tx.Exec(`INSERT INTO lecturers (id, user_id, lecturer_id, department) 
			VALUES ($1, $2, $3, $4)`,
			lecID, lecUserID, lec.lecturerID, lec.department)
		if err != nil {
			log.Fatalf("❌ Failed to seed lecturer profile: %v", err)
		}

		lecturerIDs = append(lecturerIDs, lecID)
		lecturerUserIDs = append(lecturerUserIDs, lecUserID) // ✅ FIXED: Store user IDs
		log.Printf("   ✅ Lecturer: %s - %s (%s)", lec.lecturerID, lec.fullname, lec.department)
	}

	// C. Student Users (15 students)
	log.Println("🎓 Seeding student users...")
	students := []struct {
		username     string
		email        string
		fullname     string
		studentID    string
		programStudy string
		academicYear string
		advisorIndex int // Index to lecturerIDs
	}{
		{"mahasiswa1", "ahmad.fadli@student.univ.ac.id", "Ahmad Fadli", "2241760001", "Teknik Informatika", "2024/2025", 0},
		{"mahasiswa2", "dewi.lestari@student.univ.ac.id", "Dewi Lestari", "2241760002", "Teknik Informatika", "2024/2025", 0},
		{"mahasiswa3", "rizki.pratama@student.univ.ac.id", "Rizki Pratama", "2241760003", "Teknik Informatika", "2024/2025", 1},
		{"mahasiswa4", "sarah.amelia@student.univ.ac.id", "Sarah Amelia", "2241760004", "Teknik Informatika", "2024/2025", 1},
		{"mahasiswa5", "fajar.nugroho@student.univ.ac.id", "Fajar Nugroho", "2241760005", "Sistem Informasi", "2024/2025", 2},

		{"mahasiswa6", "linda.permata@student.univ.ac.id", "Linda Permata", "2341760006", "Teknik Informatika", "2023/2024", 0},
		{"mahasiswa7", "bayu.wicaksono@student.univ.ac.id", "Bayu Wicaksono", "2341760007", "Teknik Informatika", "2023/2024", 1},
		{"mahasiswa8", "maya.sari@student.univ.ac.id", "Maya Sari", "2341760008", "Sistem Informasi", "2023/2024", 2},
		{"mahasiswa9", "aldi.firmansyah@student.univ.ac.id", "Aldi Firmansyah", "2341760009", "Teknik Informatika", "2023/2024", 3},
		{"mahasiswa10", "nina.ananda@student.univ.ac.id", "Nina Ananda", "2341760010", "Sistem Informasi", "2023/2024", 4},

		{"mahasiswa11", "dimas.prasetyo@student.univ.ac.id", "Dimas Prasetyo", "2441760011", "Teknik Informatika", "2022/2023", 0},
		{"mahasiswa12", "kartika.putri@student.univ.ac.id", "Kartika Putri", "2441760012", "Sistem Informasi", "2022/2023", 2},
		{"mahasiswa13", "eko.saputra@student.univ.ac.id", "Eko Saputra", "2441760013", "Teknik Informatika", "2022/2023", 3},
		{"mahasiswa14", "wulan.dari@student.univ.ac.id", "Wulan Dari", "2441760014", "Sistem Informasi", "2022/2023", 4},
		{"mahasiswa15", "yoga.aditya@student.univ.ac.id", "Yoga Aditya", "2441760015", "Teknik Informatika", "2022/2023", 1},
	}

	for _, stud := range students {
		studUserID := uuid.New()
		studID := uuid.New()

		mustInsertUser(tx, studUserID, stud.username, stud.email, password, stud.fullname, roles["Mahasiswa"])

		_, err = tx.Exec(`INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id) 
			VALUES ($1, $2, $3, $4, $5, $6)`,
			studID, studUserID, stud.studentID, stud.programStudy, stud.academicYear, lecturerIDs[stud.advisorIndex])
		if err != nil {
			log.Fatalf("❌ Failed to seed student profile: %v", err)
		}

		studentIDs = append(studentIDs, studID)
		log.Printf("   ✅ Student: %s - %s (%s)", stud.studentID, stud.fullname, stud.programStudy)
	}

	// D. Seed Achievement References (Dummy MongoDB References)
	log.Println("🏆 Seeding achievement references...")

	// Status distribution untuk realism
	achievementRefs := []struct {
		studentIndex  int
		mongoID       string
		status        string
		submittedAt   string
		verifiedAt    string
		verifiedBy    *uuid.UUID
		rejectionNote string
	}{
		// Verified achievements - ✅ FIXED: Use lecturerUserIDs instead of lecturerIDs
		{0, "507f1f77bcf86cd799439011", "verified", "2024-01-15 10:00:00", "2024-01-16 14:00:00", &lecturerUserIDs[0], ""},
		{0, "507f1f77bcf86cd799439012", "verified", "2024-02-10 09:30:00", "2024-02-11 11:00:00", &lecturerUserIDs[0], ""},
		{1, "507f1f77bcf86cd799439013", "verified", "2024-01-20 14:00:00", "2024-01-22 10:00:00", &lecturerUserIDs[0], ""},
		{2, "507f1f77bcf86cd799439014", "verified", "2024-03-05 08:00:00", "2024-03-06 13:00:00", &lecturerUserIDs[1], ""},
		{3, "507f1f77bcf86cd799439015", "verified", "2024-02-25 11:00:00", "2024-02-26 15:00:00", &lecturerUserIDs[1], ""},
		{4, "507f1f77bcf86cd799439016", "verified", "2024-03-10 10:00:00", "2024-03-11 09:00:00", &lecturerUserIDs[2], ""},
		{5, "507f1f77bcf86cd799439017", "verified", "2024-01-18 13:00:00", "2024-01-19 10:00:00", &lecturerUserIDs[0], ""},
		{6, "507f1f77bcf86cd799439018", "verified", "2024-02-14 09:00:00", "2024-02-15 14:00:00", &lecturerUserIDs[1], ""},

		// Submitted (waiting for verification)
		{0, "507f1f77bcf86cd799439019", "submitted", "2024-03-20 10:00:00", "", nil, ""},
		{1, "507f1f77bcf86cd799439020", "submitted", "2024-03-21 11:00:00", "", nil, ""},
		{2, "507f1f77bcf86cd799439021", "submitted", "2024-03-22 09:00:00", "", nil, ""},
		{7, "507f1f77bcf86cd799439022", "submitted", "2024-03-19 14:00:00", "", nil, ""},
		{8, "507f1f77bcf86cd799439023", "submitted", "2024-03-23 10:30:00", "", nil, ""},

		// Draft (not yet submitted)
		{3, "507f1f77bcf86cd799439024", "draft", "", "", nil, ""},
		{4, "507f1f77bcf86cd799439025", "draft", "", "", nil, ""},
		{9, "507f1f77bcf86cd799439026", "draft", "", "", nil, ""},
		{10, "507f1f77bcf86cd799439027", "draft", "", "", nil, ""},

		// Rejected achievements - ✅ FIXED: Use lecturerUserIDs instead of lecturerIDs
		{5, "507f1f77bcf86cd799439028", "rejected", "2024-03-01 10:00:00", "", &lecturerUserIDs[0], "Dokumen pendukung tidak lengkap. Mohon upload sertifikat asli."},
		{11, "507f1f77bcf86cd799439029", "rejected", "2024-03-05 11:00:00", "", &lecturerUserIDs[2], "Tingkat kompetisi tidak sesuai dengan bukti yang dilampirkan."},
		{12, "507f1f77bcf86cd799439030", "rejected", "2024-03-10 09:00:00", "", &lecturerUserIDs[3], "Data prestasi tidak lengkap, harap lengkapi detail kompetisi."},
	}

	for _, ref := range achievementRefs {
		refID := uuid.New()

		var submittedAt, verifiedAt sql.NullString
		if ref.submittedAt != "" {
			submittedAt.String = ref.submittedAt
			submittedAt.Valid = true
		}
		if ref.verifiedAt != "" {
			verifiedAt.String = ref.verifiedAt
			verifiedAt.Valid = true
		}

		var verifiedByUUID sql.NullString
		if ref.verifiedBy != nil {
			verifiedByUUID.String = ref.verifiedBy.String()
			verifiedByUUID.Valid = true
		}

		_, err = tx.Exec(`INSERT INTO achievement_references 
			(id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`,
			refID, studentIDs[ref.studentIndex], ref.mongoID, ref.status,
			submittedAt, verifiedAt, verifiedByUUID, ref.rejectionNote)
		if err != nil {
			log.Fatalf("❌ Failed to seed achievement reference: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("❌ Failed to commit transaction: %v", err)
	}

	log.Println("\n✅ Seeding successful!")
	log.Println("\n📊 Summary:")
	log.Println("   • 7 Permissions")
	log.Println("   • 3 Roles (Admin, Dosen Wali, Mahasiswa)")
	log.Println("   • 2 Admin users")
	log.Println("   • 5 Lecturer users")
	log.Println("   • 15 Student users")
	log.Println("   • 20 Achievement references")
	log.Println("      - 8 Verified")
	log.Println("      - 5 Submitted (waiting)")
	log.Println("      - 4 Draft")
	log.Println("      - 3 Rejected")
	log.Println("\n🔑 Login Credentials:")
	log.Println("   Admin:")
	log.Println("      • username: admin | password: password123")
	log.Println("      • username: admin2 | password: password123")
	log.Println("   Dosen Wali:")
	log.Println("      • username: dosen1 | password: password123")
	log.Println("      • username: dosen2-5 | password: password123")
	log.Println("   Mahasiswa:")
	log.Println("      • username: mahasiswa1 | password: password123")
	log.Println("      • username: mahasiswa2-15 | password: password123")
}

// Helpers
func hashPassword(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(bytes), err
}

func mustInsertUser(tx *sql.Tx, id uuid.UUID, username, email, pwd, fullname string, roleID uuid.UUID) {
	_, err := tx.Exec(`INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())`,
		id, username, email, pwd, fullname, roleID,
	)
	if err != nil {
		log.Fatalf("❌ Failed to insert user %s: %v", username, err)
	}
}

func assignPermission(tx *sql.Tx, roleID uuid.UUID, permIDs ...uuid.UUID) {
	for _, pid := range permIDs {
		_, err := tx.Exec("INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)", roleID, pid)
		if err != nil {
			log.Fatalf("❌ Failed to assign permission: %v", err)
		}
	}
}
