package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"sistem-prestasi-mhs/app/helpers"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"
)

type AchievementHandler struct {
	service     *service.AchievementService
	studentRepo *repository.StudentRepository
}

func NewAchievementHandler(s *service.AchievementService, sr *repository.StudentRepository) *AchievementHandler {
	return &AchievementHandler{service: s, studentRepo: sr}
}

func SetupAchievementRoutes(router fiber.Router, db *sql.DB, mongoDB *mongo.Database) {
	userRepo := repository.NewUserRepository(db)
	studentRepo := repository.NewStudentRepository(db)
	achRepo := repository.NewAchievementRepository(db, mongoDB)

	authService := service.NewAuthService(userRepo)
	achService := service.NewAchievementService(achRepo, studentRepo, userRepo)

	handler := NewAchievementHandler(achService, studentRepo)

	router.Use(middleware.AuthMiddleware(authService))

	router.Get("/", handler.List)
	router.Get("/:id", handler.GetDetail)
	router.Post("/", handler.Create)
	router.Delete("/:id", handler.Delete)
	router.Post("/:id/submit", handler.Submit)

	adminGroup := router.Group("/:id", middleware.RequireRole(userRepo, "lecturer", "admin"))
	adminGroup.Post("/verify", handler.Verify)
	adminGroup.Post("/reject", handler.Reject)
}

// List godoc
// @Summary List all achievements
// @Description Get paginated list of achievements
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{} "List of achievements"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /achievements [get]
func (h *AchievementHandler) List(c *fiber.Ctx) error {
	limit, offset := helpers.GetPaginationParams(c)

	data, err := h.service.ListAchievements(c.Context(), limit, offset)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusInternalServerError, "failed to fetch data", err.Error())
	}

	return helpers.ListResponse(c, data, len(data), limit, offset)
}

// GetDetail godoc
// @Summary Get achievement detail
// @Description Get detailed information about a specific achievement
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]interface{} "Achievement details"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Router /achievements/{id} [get]
func (h *AchievementHandler) GetDetail(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	data, err := h.service.GetAchievementDetail(c.Context(), id)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusNotFound, "achievement not found", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusOK, "achievement retrieved", data)
}

// Create godoc
// @Summary Create new achievement
// @Description Students can create a new achievement entry
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param achievement body models.CreateAchievementRequest true "Achievement data"
// @Success 201 {object} map[string]interface{} "Achievement created"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 403 {object} map[string]interface{} "Forbidden - only students can create"
// @Router /achievements [post]
func (h *AchievementHandler) Create(c *fiber.Ctx) error {
	userID, err := helpers.GetUserIDFromLocals(c)
	if err != nil {
		return helpers.ErrUnauthorized
	}

	student, err := h.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "only students can create achievements")
	}

	var req models.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return helpers.ErrInvalidRequest
	}
	if req.Title == "" || req.AchievementType == "" {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "title and type are required")
	}

	result, err := h.service.CreateAchievement(c.Context(), student.ID, &req)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "failed to create achievement", err.Error())
	}

	return helpers.SuccessResponse(c, fiber.StatusCreated, "achievement created", result)
}

// Submit godoc
// @Summary Submit achievement for verification
// @Description Student submits their achievement for lecturer verification
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]interface{} "Achievement submitted"
// @Failure 403 {object} map[string]interface{} "Forbidden - not owner"
// @Failure 400 {object} map[string]interface{} "Invalid state transition"
// @Router /achievements/{id}/submit [post]
func (h *AchievementHandler) Submit(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	userID, _ := helpers.GetUserIDFromLocals(c)

	if err := h.service.ValidateOwnership(c.Context(), id, userID); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", err.Error())
	}

	if err := h.service.SubmitForVerification(c.Context(), id); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "submission failed", err.Error())
	}

	return helpers.SuccessResponseWithoutData(c, fiber.StatusOK, "achievement submitted")
}

// Verify godoc
// @Summary Verify achievement
// @Description Lecturer or admin approves an achievement
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]interface{} "Achievement verified"
// @Failure 400 {object} map[string]interface{} "Invalid state"
// @Router /achievements/{id}/verify [post]
func (h *AchievementHandler) Verify(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	verifierID, _ := helpers.GetUserIDFromLocals(c)

	if err := h.service.VerifyAchievement(c.Context(), id, verifierID); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "verification failed", err.Error())
	}

	return helpers.SuccessResponseWithoutData(c, fiber.StatusOK, "achievement verified")
}

// Reject godoc
// @Summary Reject achievement
// @Description Lecturer or admin rejects an achievement with a note
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID (UUID)"
// @Param rejection body object{note=string} true "Rejection note"
// @Success 200 {object} map[string]interface{} "Achievement rejected"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Router /achievements/{id}/reject [post]
func (h *AchievementHandler) Reject(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	verifierID, _ := helpers.GetUserIDFromLocals(c)

	var req struct {
		Note string `json:"note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return helpers.ErrInvalidRequest
	}
	if req.Note == "" {
		return helpers.ErrorResponse(c, fiber.StatusUnprocessableEntity, "validation error", "rejection note is required")
	}

	if err := h.service.RejectAchievement(c.Context(), id, verifierID, req.Note); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "rejection failed", err.Error())
	}

	return helpers.SuccessResponseWithoutData(c, fiber.StatusOK, "achievement rejected")
}

// Delete godoc
// @Summary Delete achievement
// @Description Student deletes their own draft or rejected achievement
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID (UUID)"
// @Success 200 {object} map[string]interface{} "Achievement deleted"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /achievements/{id} [delete]
func (h *AchievementHandler) Delete(c *fiber.Ctx) error {
	id, err := helpers.GetUUIDFromParams(c, "id")
	if err != nil {
		return helpers.ErrInvalidRequest
	}

	userID, _ := helpers.GetUserIDFromLocals(c)
	student, err := h.studentRepo.GetByUserID(c.Context(), userID)
	if err != nil {
		return helpers.ErrorResponse(c, fiber.StatusForbidden, "forbidden", "user profile not found")
	}

	if err := h.service.DeleteAchievement(c.Context(), id, student.ID); err != nil {
		return helpers.ErrorResponse(c, fiber.StatusBadRequest, "delete failed", err.Error())
	}

	return helpers.SuccessResponseWithoutData(c, fiber.StatusOK, "achievement deleted")
}
