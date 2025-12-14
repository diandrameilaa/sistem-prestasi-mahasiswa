package routes

import (
	"sistem-prestasi-mhs/app/middleware"
	"sistem-prestasi-mhs/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func LecturerRoutes(r fiber.Router) {
	studentRepo := repository.NewStudentRepository()

	lecturers := r.Group("/lecturers")

	// Pastikan AuthMiddleware sudah disesuaikan untuk return fiber.Handler
	lecturers.Use(middleware.AuthMiddleware())

	{
		// @Summary Get All Lecturers
		// @Description Get list of all lecturers
		// @Tags Lecturers
		// @Produce json
		// @Security BearerAuth
		// @Param page query int false "Page number" default(1)
		// @Param limit query int false "Items per page" default(10)
		// @Success 200 {object} object{status=string,data=object{items=array,total=int,page=int,limit=int}}
		// @Failure 401 {object} object{status=string,message=string}
		// @Router /lecturers [get]
		lecturers.Get("/", func(c *fiber.Ctx) error {
			// Menggunakan helper Fiber untuk parsing query param integer
			page := c.QueryInt("page", 1)
			limit := c.QueryInt("limit", 10)
			offset := (page - 1) * limit

			lecturerList, total, err := studentRepo.FindAllLecturers(limit, offset)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data": fiber.Map{
					"items": lecturerList,
					"total": total,
					"page":  page,
					"limit": limit,
				},
			})
		})

		// @Summary Get Lecturer's Advisees
		// @Description Get list of students advised by specific lecturer
		// @Tags Lecturers
		// @Produce json
		// @Security BearerAuth
		// @Param id path string true "Lecturer ID"
		// @Success 200 {object} object{status=string,data=array}
		// @Failure 404 {object} object{status=string,message=string}
		// @Router /lecturers/{id}/advisees [get]
		lecturers.Get("/:id/advisees", func(c *fiber.Ctx) error {
			id, err := uuid.Parse(c.Params("id"))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"status":  "error",
					"message": "Invalid ID format",
				})
			}

			students, err := studentRepo.FindByAdvisorID(id)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status": "success",
				"data":   students,
			})
		})
	}
}
