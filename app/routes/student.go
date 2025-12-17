package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"

	"sistem-prestasi-mhs/app/helpers"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"
)

type StudentHandler struct {
	service *service.StudentService
	achRepo *repository.AchievementRepository
}

func NewStudentHandler(s *service.StudentService, ar *repository.AchievementRepository) *StudentHandler {
	return &StudentHandler{
		service: s,
		achRepo: ar,
	}
}

func SetupStudentRoutes(router fiber.Router, db *sql.DB, mongoDB *mongo.Database) {
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	achRepo := repository.NewAchievementRepository(db, mongoDB)

	authService := service.NewAuthService(userRepo)
	studentService := service.NewStudentService(studentRepo)

	handler := NewStudentHandler(studentService, achRepo)

	router.Use(middleware.AuthMiddleware(authService))

	router.Get("/", handler.List)
	router.Get("/:id", handler.GetDetail)
	router.Get("/:id/achievements", handler.GetAchievements)
	router.Put("/:id/advisor", handler.SetAdvisor)
}

// List godoc
// @Summary List all students
// @Description Get paginated list of students
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "List of students"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /students [get]
func (h *StudentHandler) List(c *fiber.Ctx) error {
	limit, offset := helpers.GetPaginationParams(c)

	data, err := h.service.ListStudents(c.Context(), limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve students", err.Error())
	}

	return helpers.ListResponse(c, data, len(data), limit, offset)
}

// GetDetail godoc
// @Summary Get student detail
// @Description Get detailed information about a specific student
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Success 200 {object} map[string]interface{} "Student details"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Router /students/{id} [get]
func (h *StudentHandler) GetDetail(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	data, err := h.service.GetStudent(c.Context(), id)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusNotFound, "student not found", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "student retrieved", data)
}

// GetAchievements godoc
// @Summary Get student achievements
// @Description Get all achievements for a specific student
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "Student achievements"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /students/{id}/achievements [get]
func (h *StudentHandler) GetAchievements(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	limit, offset := helpers.GetPaginationParams(c)

	data, err := h.achRepo.GetStudentAchievements(c.Context(), id, limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve achievements", err.Error())
	}

	return helpers.ListResponse(c, data, len(data), limit, offset)
}

// SetAdvisor godoc
// @Summary Assign advisor to student
// @Description Assign a lecturer as academic advisor for a student
// @Tags Students
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID (UUID)"
// @Param advisor body models.SetAdvisorRequest true "Advisor ID"
// @Success 200 {object} map[string]interface{} "Advisor assigned"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 422 {object} map[string]interface{} "Validation error"
// @Router /students/{id}/advisor [put]
func (h *StudentHandler) SetAdvisor(c *fiber.Ctx) error {
	studentID, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	var req models.SetAdvisorRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.ErrInvalidRequest
	}

	if req.AdvisorID == uuid.Nil {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "advisor_id is required")
	}

	if err := h.service.SetAdvisor(c.Context(), studentID, req.AdvisorID); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "failed to set advisor", err.Error())
	}

	return helpers.SuccessResponseWithoutData(c, fiber.StatusOK, "advisor assigned successfully")
}
