package routes

import (
	"sistem-prestasi-mhs/app/helper"
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
	"sistem-prestasi-mhs/app/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func AchievementRoutes(r fiber.Router) {
	achievementRepo := repository.NewAchievementRepository()
	studentRepo := repository.NewStudentRepository()
	achievementService := service.NewAchievementService(achievementRepo, studentRepo)

	achievements := r.Group("/achievements")

	// Pastikan middleware.AuthMiddleware() sudah disesuaikan untuk Fiber
	achievements.Use(middleware.AuthMiddleware())

	{
		// @Summary Get Achievements List
		// @Description Get list of achievements based on user role
		// @Tags Achievements
		// @Produce json
		// @Security BearerAuth
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(10)
		// @Param status query string false "Filter by status"
		// @Success 200 {object} object{status=string,data=object{items=array,total=int,page=int,limit=int}}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /achievements [get]
		achievements.Get("/", func(c *fiber.Ctx) error {
			// Mengambil data dari Locals (Middleware) dan melakukan Type Assertion ke string
			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)
			role, _ := c.Locals("role").(string)

			// Fiber QueryInt memudahkan parsing pagination
			page := c.QueryInt("page", 1)
			limit := c.QueryInt("limit", 10)
			status := c.Query("status")

			offset := (page - 1) * limit

			var items []map[string]interface{}
			var total int64
			var err error

			switch role {
			case "Mahasiswa":
				// Get student record
				student, err := studentRepo.FindByUserID(userID)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"status":  "error",
						"message": "Student record not found",
					})
				}
				items, total, err = achievementService.GetStudentAchievements(student.ID, limit, offset)

			case "Dosen Wali":
				// Get lecturer record
				lecturer, err := studentRepo.FindLecturerByUserID(userID)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"status":  "error",
						"message": "Lecturer record not found",
					})
				}
				items, total, err = achievementService.GetAdvisorAchievements(lecturer.ID, limit, offset)

			case "Admin":
				items, total, err = achievementService.GetAllAchievements(limit, offset, status)

			default:
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"status":  "error",
					"message": "Unauthorized role",
				})
			}

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"items": items,
					"total": total,
					"page":  page,
					"limit": limit,
				},
			})
		})

		// @Summary Create Achievement
		// @Description Create new achievement (Mahasiswa only)
		// @Tags Achievements
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param achievement body models.Achievement true "Achievement data"
		// @Success 201 {object} object{status=string,data=object}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /achievements [post]
		achievements.Post("/", middleware.RequirePermission("achievement:create"), func(c *fiber.Ctx) error {
			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)

			// Get student record
			student, err := studentRepo.FindByUserID(userID)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Student record not found",
				})
			}

			var achievement models.Achievement
			if err := c.BodyParser(&achievement); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			ref, err := achievementService.CreateAchievement(student.ID, &achievement)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"id":                   ref.ID,
					"mongo_achievement_id": ref.MongoAchievementID,
					"status":               ref.Status,
					"achievement":          achievement,
				},
			})
		})

		// @Summary Get Achievement Detail
		// @Description Get achievement detail by ID
		// @Tags Achievements
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Achievement ID"
		// @Success 200 {object} object{status=string,data=object}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /achievements/{id} [get]
		achievements.Get("/:id", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			achievement, ref, err := achievementService.GetAchievementByID(id)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Achievement not found",
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"reference":   ref,
					"achievement": achievement,
				},
			})
		})

		// @Summary Update Achievement
		// @Description Update achievement (Mahasiswa only, draft status only)
		// @Tags Achievements
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Achievement ID"
		// @Param achievement body models.Achievement true "Achievement data"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /achievements/{id} [put]
		achievements.Put("/:id", middleware.RequirePermission("achievement:update"), func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)

			student, err := studentRepo.FindByUserID(userID)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Student record not found",
				})
			}

			var achievement models.Achievement
			if err := c.BodyParser(&achievement); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			if err := achievementService.UpdateAchievement(id, student.ID, &achievement); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Achievement updated successfully",
			})
		})

		// @Summary Delete Achievement
		// @Description Soft delete achievement (Mahasiswa only, draft status only)
		// @Tags Achievements
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Achievement ID"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /achievements/{id} [delete]
		achievements.Delete("/:id", middleware.RequirePermission("achievement:delete"), func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)

			student, err := studentRepo.FindByUserID(userID)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Student record not found",
				})
			}

			if err := achievementService.DeleteAchievement(id, student.ID); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Achievement deleted successfully",
			})
		})

		// @Summary Submit Achievement for Verification
		// @Description Submit draft achievement for verification by advisor
		// @Tags Achievements
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Achievement ID"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /achievements/{id}/submit [post]
		achievements.Post("/:id/submit", middleware.RequirePermission("achievement:create"), func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)

			student, err := studentRepo.FindByUserID(userID)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Student record not found",
				})
			}

			if err := achievementService.SubmitForVerification(id, student.ID); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Achievement submitted for verification",
			})
		})

		// @Summary Verify Achievement
		// @Description Verify submitted achievement (Dosen Wali only)
		// @Tags Achievements
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Achievement ID"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /achievements/{id}/verify [post]
		achievements.Post("/:id/verify", middleware.RequirePermission("achievement:verify"), func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)

			if err := achievementService.VerifyAchievement(id, userID); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Achievement verified successfully",
			})
		})

		// @Summary Reject Achievement
		// @Description Reject submitted achievement with note (Dosen Wali only)
		// @Tags Achievements
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Achievement ID"
		// @Param request body object{rejection_note=string} true "Rejection note"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /achievements/{id}/reject [post]
		achievements.Post("/:id/reject", middleware.RequirePermission("achievement:verify"), func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			var req struct {
				RejectionNote string `json:"rejection_note"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			// Validasi manual karena Fiber BodyParser tidak support tag binding otomatis
			if req.RejectionNote == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "rejection_note is required",
				})
			}

			userIDStr, _ := c.Locals("user_id").(string)
			userID := helper.ParseUUID(userIDStr)

			if err := achievementService.RejectAchievement(id, userID, req.RejectionNote); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Achievement rejected",
			})
		})
	}
}
