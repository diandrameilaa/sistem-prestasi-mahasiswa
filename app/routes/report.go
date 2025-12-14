package routes

import (
	"context"
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/helper"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ReportRoutes(r fiber.Router) {
	studentRepo := repository.NewStudentRepository()
	achievementRepo := repository.NewAchievementRepository()

	reports := r.Group("/reports")

	// Pastikan middleware sudah disesuaikan untuk Fiber
	reports.Use(middleware.AuthMiddleware())

	{
		// @Summary Get Achievement Statistics
		// @Description Get statistics of achievements based on user role
		// @Tags Reports
		// @Produce json
		// @Security BearerAuth
		// @Success 200 {object} object{status=string,data=object}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /reports/statistics [get]
		reports.Get("/statistics", func(c *fiber.Ctx) error {
			// Mengambil data dari Middleware (c.Locals)
			// Kita perlu melakukan type assertion .(string) karena Locals mengembalikan interface{}
			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)
			role, _ := c.Locals("role").(string)

			var filter bson.M
			switch role {
			case "Mahasiswa":
				student, err := studentRepo.FindByUserID(userID)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"status":  "error",
						"message": "Student record not found",
					})
				}
				filter = bson.M{"studentId": student.ID.String()}

			case "Dosen Wali":
				lecturer, err := studentRepo.FindLecturerByUserID(userID)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"status":  "error",
						"message": "Lecturer record not found",
					})
				}
				students, _ := studentRepo.FindByAdvisorID(lecturer.ID)
				studentIDs := []string{}
				for _, s := range students {
					studentIDs = append(studentIDs, s.ID.String())
				}
				filter = bson.M{"studentId": bson.M{"$in": studentIDs}}

			case "Admin":
				filter = bson.M{}

			default:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"status":  "error",
					"message": "Unauthorized role",
				})
			}

			// Get statistics from MongoDB
			collection := config.MongoDatabase.Collection("achievements")

			// Total achievements
			totalAchievements, _ := collection.CountDocuments(context.Background(), filter)

			// Achievements by type
			pipeline := []bson.M{
				{"$match": filter},
				{"$group": bson.M{
					"_id":   "$achievementType",
					"count": bson.M{"$sum": 1},
				}},
			}
			cursor, _ := collection.Aggregate(context.Background(), pipeline)
			var typeStats []bson.M
			cursor.All(context.Background(), &typeStats)

			// Achievements by competition level
			competitionPipeline := []bson.M{
				{"$match": bson.M{
					"$and": []bson.M{
						filter,
						{"achievementType": "competition"},
					},
				}},
				{"$group": bson.M{
					"_id":   "$details.competitionLevel",
					"count": bson.M{"$sum": 1},
				}},
			}
			compCursor, _ := collection.Aggregate(context.Background(), competitionPipeline)
			var competitionStats []bson.M
			compCursor.All(context.Background(), &competitionStats)

			// Get status statistics from PostgreSQL
			var statusStats []struct {
				Status string
				Count  int64
			}

			query := config.DB.Model(&models.AchievementReference{}).
				Select("status, count(*) as count").
				Group("status")

			switch role {
			case "Mahasiswa":
				student, _ := studentRepo.FindByUserID(userID)
				query = query.Where("student_id = ?", student.ID)
			case "Dosen Wali":
				lecturer, _ := studentRepo.FindLecturerByUserID(userID)
				students, _ := studentRepo.FindByAdvisorID(lecturer.ID)
				studentIDs := []uuid.UUID{}
				for _, s := range students {
					studentIDs = append(studentIDs, s.ID)
				}
				query = query.Where("student_id IN ?", studentIDs)
			}

			query.Scan(&statusStats)

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"total_achievements":   totalAchievements,
					"by_type":              typeStats,
					"by_competition_level": competitionStats,
					"by_status":            statusStats,
				},
			})
		})

		// @Summary Get Student Report
		// @Description Get detailed achievement report for specific student
		// @Tags Reports
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Student ID"
		// @Success 200 {object} object{status=string,data=object}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /reports/student/{id} [get]
		reports.Get("/student/:id", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			// Get student info
			student, err := studentRepo.FindByID(id)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Student not found",
				})
			}

			// Get achievements
			// Note: Limit diset tinggi (1000) untuk laporan
			refs, total, _ := achievementRepo.FindReferencesByStudentID(id, 1000, 0)

			// Calculate statistics
			var verified, submitted, draft, rejected int64
			var totalPoints float64

			collection := config.MongoDatabase.Collection("achievements")
			for _, ref := range refs {
				switch ref.Status {
				case models.AchievementStatusVerified:
					verified++
				case models.AchievementStatusSubmitted:
					submitted++
				case models.AchievementStatusDraft:
					draft++
				case models.AchievementStatusRejected:
					rejected++
				}

				// Get points from MongoDB
				var achievement models.Achievement
				objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
				if err == nil {
					collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&achievement)
					totalPoints += achievement.Points
				}
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"student": student,
					"summary": fiber.Map{
						"total_achievements": total,
						"verified":           verified,
						"submitted":          submitted,
						"draft":              draft,
						"rejected":           rejected,
						"total_points":       totalPoints,
					},
					"achievements": refs,
				},
			})
		})
	}
}
