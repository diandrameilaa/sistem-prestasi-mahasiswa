package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"sistem-prestasi-mhs/app/helpers"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"
)

type ReportHandler struct {
	achRepo *repository.AchievementRepository
}

func NewReportHandler(ar *repository.AchievementRepository) *ReportHandler {
	return &ReportHandler{achRepo: ar}
}

// ============================================================================
// ROUTE SETUP
// ============================================================================

func SetupReportRoutes(router fiber.Router, db *sql.DB, mongoDB *mongo.Database) {
	// Initialize Dependencies
	userRepo := repository.NewUserRepository(db)
	achRepo := repository.NewAchievementRepository(db, mongoDB)

	authService := service.NewAuthService(userRepo)
	handler := NewReportHandler(achRepo)

	// Apply Auth Middleware
	router.Use(middleware.AuthMiddleware(authService))

	// Routes
	router.Get("/statistics", handler.GetSystemStatistics)
	router.Get("/student/:id", handler.GetStudentReport)
}

// ============================================================================
// HANDLER METHODS
// ============================================================================

// GetSystemStatistics aggregates achievement data to show dashboard statistics
// @Summary Get System Statistics
// @Description Get aggregated statistics of achievements including counts by type and status. Returns a snapshot of recent data (limited to 100 items for performance).
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string,data=object{sample_size=int,by_achievement_type=object,by_status=object}} "Successfully retrieved system statistics"
// @Failure 401 {object} object{status=string,message=string,error=string} "Unauthorized - Invalid or missing token"
// @Failure 403 {object} object{status=string,message=string,error=string} "Forbidden - Insufficient permissions"
// @Failure 500 {object} object{status=string,message=string,error=string} "Internal server error"
// @Router /reports/statistics [get]
func (h *ReportHandler) GetSystemStatistics(c *fiber.Ctx) error {
	limit := 100 // Hard limit for analysis snapshot
	offset := 0

	// 1. Fetch recent references
	refs, err := h.achRepo.ListAll(c.Context(), limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve data", err.Error())
	}

	// 2. Aggregate Data
	stats := h.aggregateStats(c, refs)

	return helpers.SuccessResponse(c, fiber.StatusOK, "statistics retrieved", stats)
}

// GetStudentReport retrieves detailed achievement report for a specific student
// @Summary Get Student Achievement Report
// @Description Get comprehensive report of achievements for a specific student including all their achievement records with pagination support
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student UUID" format(uuid) example(770e8400-e29b-41d4-a716-446655440002)
// @Param page query int false "Page number for pagination" default(1) minimum(1)
// @Param limit query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} object{status=string,message=string,data=object{total_shown=int,achievements=[]object}} "Successfully retrieved student report"
// @Failure 400 {object} object{status=string,message=string,error=string} "Bad Request - Invalid student ID format"
// @Failure 401 {object} object{status=string,message=string,error=string} "Unauthorized - Invalid or missing token"
// @Failure 403 {object} object{status=string,message=string,error=string} "Forbidden - Cannot access other student's report"
// @Failure 404 {object} object{status=string,message=string,error=string} "Student not found"
// @Failure 500 {object} object{status=string,message=string,error=string} "Internal server error"
// @Router /reports/student/{id} [get]
func (h *ReportHandler) GetStudentReport(c *fiber.Ctx) error {
	studentID, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	limit, offset := helpers.GetPaginationParams(c)

	data, err := h.achRepo.GetStudentAchievements(c.Context(), studentID, limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve report", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "student report retrieved", fiber.Map{
		"total_shown":  len(data),
		"achievements": data,
	})
}

// ============================================================================
// INTERNAL HELPERS
// ============================================================================

// aggregateStats processes achievement references and returns aggregated statistics
func (h *ReportHandler) aggregateStats(c *fiber.Ctx, refs []*models.AchievementReference) fiber.Map {
	typeCount := make(map[string]int)
	statusCount := make(map[string]int)

	// Process aggregation
	for _, ref := range refs {
		// Count Status
		statusCount[ref.Status]++

		// Convert string Hex ID to primitive.ObjectID
		objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			// Skip if MongoDB ID is invalid/corrupted
			continue
		}

		// Count Type (Requires fetching detail from MongoDB)
		if detail, err := h.achRepo.GetAchievementByID(c.Context(), objID); err == nil && detail != nil {
			typeCount[detail.AchievementType]++
		}
	}

	return fiber.Map{
		"sample_size":         len(refs),
		"by_achievement_type": typeCount,
		"by_status":           statusCount,
	}
}
