package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
)

type StudentService struct {
	repo *repository.StudentRepository
}

func NewStudentService(repo *repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// CreateStudent registers a new student profile
func (s *StudentService) CreateStudent(ctx context.Context, student *models.Student) error {
	if student.StudentID == "" {
		return errors.New("student ID (NIM) is required")
	}
	return s.repo.Create(ctx, student)
}

// GetStudent retrieves a student by their UUID
func (s *StudentService) GetStudent(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid student id")
	}
	return s.repo.GetByID(ctx, id)
}

// GetStudentByUserID retrieves a student by their generic User UUID
func (s *StudentService) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (*models.Student, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user id")
	}
	return s.repo.GetByUserID(ctx, userID)
}

// SetAdvisor assigns a lecturer to a student
func (s *StudentService) SetAdvisor(ctx context.Context, studentID, advisorID uuid.UUID) error {
	if studentID == uuid.Nil || advisorID == uuid.Nil {
		return errors.New("invalid student or advisor id")
	}

	// Note: Ideally, check if advisorID exists in LecturerRepository here
	// but assuming Foreign Key constraints in DB will handle the integrity.
	return s.repo.SetAdvisor(ctx, studentID, advisorID)
}

// GetAdvisees retrieves all students under a specific advisor
func (s *StudentService) GetAdvisees(ctx context.Context, advisorID uuid.UUID) ([]*models.Student, error) {
	if advisorID == uuid.Nil {
		return nil, errors.New("invalid advisor id")
	}
	return s.repo.GetAdvisees(ctx, advisorID)
}

// ListStudents retrieves paginated students
func (s *StudentService) ListStudents(ctx context.Context, limit, offset int) ([]*models.Student, error) {
	return s.repo.List(ctx, limit, offset)
}
