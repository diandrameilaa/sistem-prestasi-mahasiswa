// ============================================================================
// SOLUSI 1: Update lecturer_routes.go dengan inline response definition
// ============================================================================

package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"

	"sistem-prestasi-mhs/app/helpers"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"
)

type LecturerHandler struct {
	lecturerRepo *repository.LecturerRepository
	studentRepo  *repository.StudentRepository
}

func NewLecturerHandler(lr *repository.LecturerRepository, sr *repository.StudentRepository) *LecturerHandler {
	return &LecturerHandler{
		lecturerRepo: lr,
		studentRepo:  sr,
	}
}

// ============================================================================
// ROUTE SETUP
// ============================================================================

func SetupLecturerRoutes(router fiber.Router, db *sql.DB) {
	// Initialize Dependencies
	userRepo := repository.NewUserRepository(db)
	lecturerRepo := repository.NewLecturerRepository(db)
	studentRepo := repository.NewStudentRepository(db)

	authService := service.NewAuthService(userRepo)
	handler := NewLecturerHandler(lecturerRepo, studentRepo)

	// Apply Middleware globally for this group
	router.Use(middleware.AuthMiddleware(authService))

	// Routes
	router.Get("/", handler.List)

	// ID Grouping
	userGroup := router.Group("/:id")
	userGroup.Get("/", handler.GetDetail)
	userGroup.Get("/advisees", handler.GetAdvisees)
}

// ============================================================================
// HANDLER METHODS
// ============================================================================

// List retrieves a paginated list of lecturers
// @Summary Get All Lecturers
// @Description Get paginated list of all lecturers in the system
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number for pagination" default(1) minimum(1)
// @Param limit query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} object{status=string,data=[]models.Lecturer,meta=object{total=int,limit=int,offset=int,page=int}} "Successfully retrieved lecturers list"
// @Failure 401 {object} object{status=string,message=string,error=string} "Unauthorized - Invalid or missing token"
// @Failure 500 {object} object{status=string,message=string,error=string} "Internal server error"
// @Router /lecturers [get]
func (h *LecturerHandler) List(c *fiber.Ctx) error {
	limit, offset := helpers.GetPaginationParams(c)

	data, err := h.lecturerRepo.List(c.Context(), limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve lecturers", err.Error())
	}

	return helpers.ListResponse(c, data, len(data), limit, offset)
}

// GetDetail retrieves a specific lecturer profile
// @Summary Get Lecturer by ID
// @Description Get detailed information about a specific lecturer including their profile and user data
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lecturer UUID" format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Success 200 {object} object{status=string,message=string,data=models.Lecturer} "Successfully retrieved lecturer details"
// @Failure 400 {object} object{status=string,message=string,error=string} "Bad Request - Invalid UUID format"
// @Failure 401 {object} object{status=string,message=string,error=string} "Unauthorized - Invalid or missing token"
// @Failure 404 {object} object{status=string,message=string,error=string} "Lecturer not found"
// @Failure 500 {object} object{status=string,message=string,error=string} "Internal server error"
// @Router /lecturers/{id} [get]
func (h *LecturerHandler) GetDetail(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	data, err := h.lecturerRepo.GetByID(c.Context(), id)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusNotFound, "lecturer not found", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "lecturer retrieved", data)
}

// GetAdvisees retrieves students advised by the lecturer
// @Summary Get Lecturer's Advisees
// @Description Get list of students who are under the supervision of a specific lecturer (academic advisor)
// @Tags Lecturers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lecturer UUID" format(uuid) example(550e8400-e29b-41d4-a716-446655440000)
// @Success 200 {object} object{status=string,message=string,data=[]models.Student} "Successfully retrieved advisees list"
// @Failure 400 {object} object{status=string,message=string,error=string} "Bad Request - Invalid UUID format"
// @Failure 401 {object} object{status=string,message=string,error=string} "Unauthorized - Invalid or missing token"
// @Failure 404 {object} object{status=string,message=string,error=string} "Lecturer not found"
// @Failure 500 {object} object{status=string,message=string,error=string} "Internal server error"
// @Router /lecturers/{id}/advisees [get]
func (h *LecturerHandler) GetAdvisees(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	data, err := h.studentRepo.GetAdvisees(c.Context(), id)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to retrieve advisees", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "advisees retrieved", data)
}
