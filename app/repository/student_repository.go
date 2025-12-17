package repository

import (
	"context"
	"database/sql"
	"fmt"

	"sistem-prestasi-mhs/app/models"

	"github.com/google/uuid"
)

const (
	// Kolom standar untuk query SELECT
	studentCols = `id, user_id, student_id, program_study, academic_year, advisor_id, created_at`
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Create inserts a new student
func (r *StudentRepository) Create(ctx context.Context, s *models.Student) error {
	query := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	return r.db.QueryRowContext(ctx, query,
		s.ID, s.UserID, s.StudentID, s.ProgramStudy, s.AcademicYear, s.AdvisorID,
	).Scan(&s.CreatedAt)
}

// GetByUserID finds a student by their generic user ID
func (r *StudentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Student, error) {
	query := fmt.Sprintf("SELECT %s FROM students WHERE user_id = $1", studentCols)
	return r.fetchOne(ctx, query, userID)
}

// GetByID finds a student by their primary key
func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	query := fmt.Sprintf("SELECT %s FROM students WHERE id = $1", studentCols)
	return r.fetchOne(ctx, query, id)
}

// SetAdvisor updates the advisor for a specific student
func (r *StudentRepository) SetAdvisor(ctx context.Context, studentID, advisorID uuid.UUID) error {
	query := "UPDATE students SET advisor_id = $1 WHERE id = $2"

	res, err := r.db.ExecContext(ctx, query, advisorID, studentID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("student not found")
	}

	return nil
}

// GetAdvisees retrieves all students under a specific advisor
func (r *StudentRepository) GetAdvisees(ctx context.Context, advisorID uuid.UUID) ([]*models.Student, error) {
	query := fmt.Sprintf("SELECT %s FROM students WHERE advisor_id = $1 ORDER BY created_at DESC", studentCols)
	return r.fetchAll(ctx, query, advisorID)
}

// List retrieves all students with pagination
func (r *StudentRepository) List(ctx context.Context, limit, offset int) ([]*models.Student, error) {
	query := fmt.Sprintf("SELECT %s FROM students ORDER BY created_at DESC LIMIT $1 OFFSET $2", studentCols)
	return r.fetchAll(ctx, query, limit, offset)
}

// ============================================================================
// INTERNAL HELPERS
// ============================================================================

// fetchOne executes a query expecting a single row result
func (r *StudentRepository) fetchOne(ctx context.Context, query string, args ...interface{}) (*models.Student, error) {
	var s models.Student
	err := r.scan(r.db.QueryRowContext(ctx, query, args...), &s)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("student not found")
		}
		return nil, err
	}
	return &s, nil
}

// fetchAll executes a query expecting multiple rows
func (r *StudentRepository) fetchAll(ctx context.Context, query string, args ...interface{}) ([]*models.Student, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		var s models.Student
		if err := r.scan(rows, &s); err != nil {
			return nil, err
		}
		students = append(students, &s)
	}
	return students, nil
}

// scan maps a database row to the Student struct
func (r *StudentRepository) scan(scanner interface {
	Scan(dest ...interface{}) error
}, s *models.Student) error {
	return scanner.Scan(
		&s.ID,
		&s.UserID,
		&s.StudentID,
		&s.ProgramStudy,
		&s.AcademicYear,
		&s.AdvisorID,
		&s.CreatedAt,
	)
}
