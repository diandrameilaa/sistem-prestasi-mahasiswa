package routes

import (
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func StudentRoutes(r fiber.Router) {
	studentRepo := repository.NewStudentRepository()
	achievementRepo := repository.NewAchievementRepository()

	students := r.Group("/students")

	// Pastikan AuthMiddleware sudah return fiber.Handler
	students.Use(middleware.AuthMiddleware())

	{
		// @Summary Get All Students
		// @Description Get list of all students (Admin only)
		// @Tags Students
		// @Produce json
		// @Security BearerAuth
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(10)
		// @Success 200 {object} object{status=string,data=object{items=array,total=int,page=int,limit=int}}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /students [get]
		students.Get("/", middleware.RequirePermission("user:manage"), func(c *fiber.Ctx) error {
			// Fiber QueryInt memudahkan parsing query param dengan default value
			page := c.QueryInt("page", 1)
			limit := c.QueryInt("limit", 10)
			offset := (page - 1) * limit

			studentList, total, err := studentRepo.FindAll(limit, offset)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"items": studentList,
					"total": total,
					"page":  page,
					"limit": limit,
				},
			})
		})

		// @Summary Get Student by ID
		// @Description Get student detail by ID
		// @Tags Students
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Student ID"
		// @Success 200 {object} object{status=string,data=object}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /students/{id} [get]
		students.Get("/:id", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			student, err := studentRepo.FindByID(id)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "Student not found",
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data":   student,
			})
		})

		// @Summary Get Student Achievements
		// @Description Get all achievements for a specific student
		// @Tags Students
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Student ID"
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(10)
		// @Success 200 {object} object{status=string,data=object{items=array,total=int,page=int,limit=int}}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /students/{id}/achievements [get]
		students.Get("/:id/achievements", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			page := c.QueryInt("page", 1)
			limit := c.QueryInt("limit", 10)
			offset := (page - 1) * limit

			refs, total, err := achievementRepo.FindReferencesByStudentID(id, limit, offset)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"items": refs,
					"total": total,
					"page":  page,
					"limit": limit,
				},
			})
		})

		// @Summary Update Student Advisor
		// @Description Assign or update advisor for student (Admin only)
		// @Tags Students
		// @Accept json
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Student ID"
		// @Param request body object{advisor_id=string} true "Advisor ID"
		// @Success 200 {object} object{status=string,message=string}
		// @Failure 400 {object} object{status=string,message=string}
		// @Router /students/{id}/advisor [put]
		students.Put("/:id/advisor", middleware.RequirePermission("user:manage"), func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			var req struct {
				AdvisorID uuid.UUID `json:"advisor_id"`
			}

			if err := c.BodyParser(&req); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			// Validasi Manual: Cek apakah AdvisorID kosong (Nil UUID)
			if req.AdvisorID == uuid.Nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Advisor ID is required",
				})
			}

			if err := studentRepo.UpdateAdvisor(id, req.AdvisorID); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "success",
				"message": "Advisor updated successfully",
			})
		})
	}
}
