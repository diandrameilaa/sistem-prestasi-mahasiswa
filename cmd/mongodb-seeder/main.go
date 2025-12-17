package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB Connection
const MongoURI = "mongodb://localhost:27017"
const DatabaseName = "achievement_db"

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}
	defer client.Disconnect(ctx)

	// Ping MongoDB
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ MongoDB unreachable: %v", err)
	}
	log.Println("✅ Connected to MongoDB")

	db := client.Database(DatabaseName)
	collection := db.Collection("achievements")

	// Drop existing collection for fresh start
	log.Println("🔥 Dropping existing achievements collection...")
	collection.Drop(ctx)

	// Seed achievements
	log.Println("🌱 Seeding achievement data to MongoDB...")
	seedAchievements(ctx, collection)
}

func seedAchievements(ctx context.Context, collection *mongo.Collection) {
	achievements := []interface{}{
		// Achievement 1: Verified - Competition National
		createAchievement(
			"507f1f77bcf86cd799439011",
			"770e8400-e29b-41d4-a716-446655440000", // Student ID (replace with actual)
			"competition",
			"Juara 1 Hackathon Nasional 2024",
			"Memenangkan kompetisi hackathon tingkat nasional dengan tema AI for Social Good",
			map[string]interface{}{
				"competition_name":  "Indonesia Tech Innovation Challenge",
				"competition_level": "national",
				"rank":              1,
				"medal_type":        "gold",
				"event_date":        "2024-01-10",
				"location":          "Jakarta Convention Center",
				"organizer":         "Kementerian Pendidikan dan Kebudayaan",
				"participant_count": 150,
			},
			[]map[string]interface{}{
				{
					"file_name":   "sertifikat_juara1.pdf",
					"file_url":    "https://storage.example.com/certificates/cert001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-45 * 24 * time.Hour),
				},
			},
			[]string{"hackathon", "AI", "national", "programming"},
			100.0,
		),

		// Achievement 2: Verified - Publication
		createAchievement(
			"507f1f77bcf86cd799439012",
			"770e8400-e29b-41d4-a716-446655440000",
			"publication",
			"Publikasi Jurnal Internasional tentang Machine Learning",
			"Paper diterbitkan di IEEE Conference on Computer Vision",
			map[string]interface{}{
				"publication_type":  "journal",
				"publication_title": "Deep Learning Approach for Image Classification",
				"authors":           []string{"Ahmad Fadli", "Dr. Budi Santoso"},
				"publisher":         "IEEE",
				"issn":              "2156-5570",
				"publication_date":  "2024-02-01",
				"indexed_by":        "Scopus",
			},
			[]map[string]interface{}{
				{
					"file_name":   "ieee_paper.pdf",
					"file_url":    "https://storage.example.com/papers/paper001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-30 * 24 * time.Hour),
				},
			},
			[]string{"research", "machine learning", "publication", "international"},
			150.0,
		),

		// Achievement 3: Verified - Competition Regional
		createAchievement(
			"507f1f77bcf86cd799439013",
			"770e8400-e29b-41d4-a716-446655440001",
			"competition",
			"Juara 2 Programming Contest Regional",
			"Meraih juara 2 dalam kompetisi pemrograman tingkat regional Jawa Timur",
			map[string]interface{}{
				"competition_name":  "East Java Programming Championship",
				"competition_level": "regional",
				"rank":              2,
				"medal_type":        "silver",
				"event_date":        "2024-01-15",
				"location":          "Surabaya",
				"organizer":         "Himpunan Mahasiswa Informatika Regional",
				"participant_count": 80,
			},
			[]map[string]interface{}{
				{
					"file_name":   "certificate_regional.pdf",
					"file_url":    "https://storage.example.com/certificates/cert002.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-40 * 24 * time.Hour),
				},
			},
			[]string{"programming", "competition", "regional"},
			75.0,
		),

		// Achievement 4: Verified - Certification
		createAchievement(
			"507f1f77bcf86cd799439014",
			"770e8400-e29b-41d4-a716-446655440002",
			"certification",
			"AWS Certified Solutions Architect",
			"Lulus sertifikasi AWS Solutions Architect Associate",
			map[string]interface{}{
				"certification_name": "AWS Certified Solutions Architect - Associate",
				"issued_by":          "Amazon Web Services",
				"certification_id":   "AWS-CSA-20240305-001",
				"valid_until":        "2027-03-05",
				"score":              850,
			},
			[]map[string]interface{}{
				{
					"file_name":   "aws_certificate.pdf",
					"file_url":    "https://storage.example.com/certificates/aws001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-20 * 24 * time.Hour),
				},
			},
			[]string{"certification", "AWS", "cloud computing"},
			80.0,
		),

		// Achievement 5: Verified - Organization
		createAchievement(
			"507f1f77bcf86cd799439015",
			"770e8400-e29b-41d4-a716-446655440003",
			"organization",
			"Ketua Himpunan Mahasiswa Informatika",
			"Menjabat sebagai Ketua HMIF periode 2023-2024",
			map[string]interface{}{
				"organization_name": "Himpunan Mahasiswa Informatika",
				"position":          "Ketua",
				"period_start":      "2023-01-01",
				"period_end":        "2024-01-01",
				"achievements": []string{
					"Menyelenggarakan 5 workshop teknis",
					"Meningkatkan anggota aktif 40%",
					"Kolaborasi dengan 3 perusahaan IT",
				},
			},
			[]map[string]interface{}{
				{
					"file_name":   "sk_pengurus.pdf",
					"file_url":    "https://storage.example.com/documents/sk001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-25 * 24 * time.Hour),
				},
			},
			[]string{"organization", "leadership", "hmif"},
			60.0,
		),

		// Achievement 6: Verified - Academic
		createAchievement(
			"507f1f77bcf86cd799439016",
			"770e8400-e29b-41d4-a716-446655440004",
			"academic",
			"Beasiswa PPA Tahun 2024",
			"Menerima beasiswa Peningkatan Prestasi Akademik",
			map[string]interface{}{
				"scholarship_name": "Beasiswa PPA",
				"issued_by":        "Kementerian Pendidikan",
				"gpa":              3.85,
				"semester":         5,
				"year":             2024,
			},
			[]map[string]interface{}{
				{
					"file_name":   "sk_beasiswa.pdf",
					"file_url":    "https://storage.example.com/documents/beasiswa001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-15 * 24 * time.Hour),
				},
			},
			[]string{"scholarship", "academic", "achievement"},
			70.0,
		),

		// Achievement 7: Verified - Competition International
		createAchievement(
			"507f1f77bcf86cd799439017",
			"770e8400-e29b-41d4-a716-446655440005",
			"competition",
			"Finalis ICPC Asia Jakarta Regional",
			"Masuk 10 besar dalam International Collegiate Programming Contest",
			map[string]interface{}{
				"competition_name":  "ICPC Asia Jakarta Regional",
				"competition_level": "international",
				"rank":              8,
				"event_date":        "2024-01-12",
				"location":          "Jakarta",
				"organizer":         "ACM ICPC",
				"participant_count": 200,
			},
			[]map[string]interface{}{
				{
					"file_name":   "icpc_certificate.pdf",
					"file_url":    "https://storage.example.com/certificates/icpc001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-35 * 24 * time.Hour),
				},
			},
			[]string{"ICPC", "programming", "international"},
			120.0,
		),

		// Achievement 8: Verified - Other
		createAchievement(
			"507f1f77bcf86cd799439018",
			"770e8400-e29b-41d4-a716-446655440006",
			"other",
			"Pembicara di Tech Talk Community",
			"Menjadi pembicara dalam acara Tech Talk tentang Microservices Architecture",
			map[string]interface{}{
				"event_name": "Tech Talk: Building Scalable Systems",
				"role":       "Speaker",
				"date":       "2024-02-10",
				"organizer":  "Developer Community Surabaya",
				"audience":   150,
				"topic":      "Microservices Architecture with Go",
			},
			[]map[string]interface{}{
				{
					"file_name":   "speaker_certificate.pdf",
					"file_url":    "https://storage.example.com/certificates/speaker001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-10 * 24 * time.Hour),
				},
			},
			[]string{"speaker", "community", "technology"},
			40.0,
		),

		// Achievement 9: Submitted (Waiting)
		createAchievement(
			"507f1f77bcf86cd799439019",
			"770e8400-e29b-41d4-a716-446655440000",
			"competition",
			"Juara 3 Data Science Competition",
			"Meraih juara 3 dalam kompetisi analisis data tingkat nasional",
			map[string]interface{}{
				"competition_name":  "National Data Science Challenge",
				"competition_level": "national",
				"rank":              3,
				"medal_type":        "bronze",
				"event_date":        "2024-03-15",
				"location":          "Bandung",
				"organizer":         "Indonesian Data Science Society",
			},
			[]map[string]interface{}{
				{
					"file_name":   "datasci_certificate.pdf",
					"file_url":    "https://storage.example.com/certificates/datasci001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-5 * 24 * time.Hour),
				},
			},
			[]string{"data science", "competition", "national"},
			85.0,
		),

		// Achievement 10: Submitted (Waiting)
		createAchievement(
			"507f1f77bcf86cd799439020",
			"770e8400-e29b-41d4-a716-446655440001",
			"certification",
			"Google Cloud Professional Data Engineer",
			"Lulus sertifikasi Google Cloud Professional Data Engineer",
			map[string]interface{}{
				"certification_name": "Google Cloud Professional Data Engineer",
				"issued_by":          "Google Cloud",
				"certification_id":   "GCP-PDE-20240320-001",
				"valid_until":        "2026-03-20",
			},
			[]map[string]interface{}{
				{
					"file_name":   "gcp_certificate.pdf",
					"file_url":    "https://storage.example.com/certificates/gcp001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-3 * 24 * time.Hour),
				},
			},
			[]string{"certification", "GCP", "data engineering"},
			80.0,
		),

		// Achievement 11-13: More Submitted
		createAchievement(
			"507f1f77bcf86cd799439021",
			"770e8400-e29b-41d4-a716-446655440002",
			"competition",
			"Best Paper Award - Local Conference",
			"Paper terbaik dalam konferensi nasional teknologi informasi",
			map[string]interface{}{
				"event_name":  "National IT Conference 2024",
				"paper_title": "Blockchain Implementation for Supply Chain",
				"date":        "2024-03-18",
				"location":    "Yogyakarta",
			},
			[]map[string]interface{}{},
			[]string{"research", "conference", "blockchain"},
			65.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439022",
			"770e8400-e29b-41d4-a716-446655440007",
			"competition",
			"UI/UX Design Competition Winner",
			"Juara desain UI/UX tingkat universitas",
			map[string]interface{}{
				"competition_name":  "University Design Fest",
				"competition_level": "institutional",
				"rank":              1,
				"event_date":        "2024-03-14",
			},
			[]map[string]interface{}{},
			[]string{"design", "UI/UX", "competition"},
			50.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439023",
			"770e8400-e29b-41d4-a716-446655440008",
			"organization",
			"Wakil Ketua BEM Fakultas",
			"Menjabat sebagai Wakil Ketua BEM Fakultas Teknik",
			map[string]interface{}{
				"organization_name": "BEM Fakultas Teknik",
				"position":          "Wakil Ketua",
				"period_start":      "2024-01-01",
				"period_end":        "2024-12-31",
			},
			[]map[string]interface{}{},
			[]string{"organization", "BEM", "leadership"},
			55.0,
		),

		// Achievement 14-17: Draft
		createAchievement(
			"507f1f77bcf86cd799439024",
			"770e8400-e29b-41d4-a716-446655440003",
			"competition",
			"CTF Competition Participant",
			"Peserta kompetisi Capture The Flag keamanan siber",
			map[string]interface{}{
				"competition_name":  "Cyber Security CTF Challenge",
				"competition_level": "national",
				"event_date":        "2024-04-01",
			},
			[]map[string]interface{}{},
			[]string{"cybersecurity", "CTF", "competition"},
			45.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439025",
			"770e8400-e29b-41d4-a716-446655440004",
			"certification",
			"Oracle Java SE Programmer",
			"Draft sertifikasi Java Programming",
			map[string]interface{}{
				"certification_name": "Oracle Certified Professional Java SE Programmer",
				"issued_by":          "Oracle",
			},
			[]map[string]interface{}{},
			[]string{"certification", "java", "programming"},
			70.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439026",
			"770e8400-e29b-41d4-a716-446655440009",
			"academic",
			"IPK Semester 6 - 4.00",
			"Mencapai IPK sempurna di semester 6",
			map[string]interface{}{
				"gpa":      4.00,
				"semester": 6,
				"year":     2024,
			},
			[]map[string]interface{}{},
			[]string{"academic", "GPA", "excellence"},
			60.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439027",
			"770e8400-e29b-41d4-a716-446655440010",
			"other",
			"Volunteer di Event Nasional",
			"Menjadi volunteer dalam acara teknologi nasional",
			map[string]interface{}{
				"event_name": "Indonesia Tech Summit 2024",
				"role":       "Technical Volunteer",
				"date":       "2024-03-25",
			},
			[]map[string]interface{}{},
			[]string{"volunteer", "event", "community"},
			30.0,
		),

		// Achievement 18-20: Rejected
		createAchievement(
			"507f1f77bcf86cd799439028",
			"770e8400-e29b-41d4-a716-446655440005",
			"competition",
			"Programming Marathon 2024",
			"Peserta programming marathon",
			map[string]interface{}{
				"competition_name":  "24 Hours Programming Marathon",
				"competition_level": "national",
				"event_date":        "2024-02-20",
			},
			[]map[string]interface{}{
				{
					"file_name":   "participation_cert.pdf",
					"file_url":    "https://storage.example.com/certificates/part001.pdf",
					"file_type":   "application/pdf",
					"uploaded_at": time.Now().Add(-20 * 24 * time.Hour),
				},
			},
			[]string{"programming", "competition"},
			40.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439029",
			"770e8400-e29b-41d4-a716-446655440011",
			"competition",
			"Local Hackathon Winner",
			"Juara hackathon tingkat lokal",
			map[string]interface{}{
				"competition_name":  "City Hackathon 2024",
				"competition_level": "local",
				"rank":              1,
				"event_date":        "2024-02-28",
			},
			[]map[string]interface{}{},
			[]string{"hackathon", "local"},
			35.0,
		),

		createAchievement(
			"507f1f77bcf86cd799439030",
			"770e8400-e29b-41d4-a716-446655440012",
			"organization",
			"Koordinator Divisi IT HMIF",
			"Koordinator divisi IT di himpunan mahasiswa",
			map[string]interface{}{
				"organization_name": "HMIF",
				"position":          "Koordinator Divisi IT",
				"period_start":      "2024-01-01",
			},
			[]map[string]interface{}{},
			[]string{"organization", "IT", "coordinator"},
			45.0,
		),
	}

	// Insert all achievements
	result, err := collection.InsertMany(ctx, achievements)
	if err != nil {
		log.Fatalf("❌ Failed to seed achievements: %v", err)
	}

	log.Printf("✅ Successfully seeded %d achievements to MongoDB\n", len(result.InsertedIDs))
	log.Println("\n📊 Summary:")
	log.Println("   • Competition: 10 achievements")
	log.Println("   • Publication: 1 achievement")
	log.Println("   • Certification: 4 achievements")
	log.Println("   • Organization: 4 achievements")
	log.Println("   • Academic: 2 achievements")
	log.Println("   • Other: 2 achievements")
	log.Println("\n✅ MongoDB seeding complete!")
}

func createAchievement(
	id string,
	studentID string,
	achievementType string,
	title string,
	description string,
	details map[string]interface{},
	attachments []map[string]interface{},
	tags []string,
	points float64,
) bson.M {
	objectID, _ := primitive.ObjectIDFromHex(id)

	return bson.M{
		"_id":             objectID,
		"studentId":       studentID,
		"achievementType": achievementType,
		"title":           title,
		"description":     description,
		"details":         details,
		"attachments":     attachments,
		"tags":            tags,
		"points":          points,
		"isDeleted":       false,
		"createdAt":       time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
		"updatedAt":       time.Now().Add(-60 * 24 * time.Hour),
	}
}
