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
	lecturerCols = `id, user_id, lecturer_id, department, created_at`
)

type LecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{db: db}
}

// Create inserts a new lecturer and returns the created_at timestamp
func (r *LecturerRepository) Create(ctx context.Context, lecturer *models.Lecturer) error {
	query := `
		INSERT INTO lecturers (id, user_id, lecturer_id, department)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	return r.db.QueryRowContext(ctx, query,
		lecturer.ID, lecturer.UserID, lecturer.LecturerID, lecturer.Department,
	).Scan(&lecturer.CreatedAt)
}

// GetByUserID finds a lecturer by their generic user_id
func (r *LecturerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Lecturer, error) {
	query := fmt.Sprintf("SELECT %s FROM lecturers WHERE user_id = $1", lecturerCols)
	return r.fetchOne(ctx, query, userID)
}

// GetByID finds a lecturer by their primary key (uuid)
func (r *LecturerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lecturer, error) {
	query := fmt.Sprintf("SELECT %s FROM lecturers WHERE id = $1", lecturerCols)
	return r.fetchOne(ctx, query, id)
}

// List retrieves a paginated list of lecturers
func (r *LecturerRepository) List(ctx context.Context, limit, offset int) ([]*models.Lecturer, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM lecturers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, lecturerCols)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lecturers []*models.Lecturer
	for rows.Next() {
		var l models.Lecturer
		if err := r.scan(rows, &l); err != nil {
			return nil, err
		}
		lecturers = append(lecturers, &l)
	}

	return lecturers, nil
}

// ============================================================================
// INTERNAL HELPERS
// ============================================================================

// fetchOne executes a query expecting a single row result
func (r *LecturerRepository) fetchOne(ctx context.Context, query string, arg interface{}) (*models.Lecturer, error) {
	var l models.Lecturer
	err := r.scan(r.db.QueryRowContext(ctx, query, arg), &l)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("lecturer not found")
		}
		return nil, err
	}
	return &l, nil
}

// scan maps a database row to the Lecturer struct
// It accepts any interface that has a Scan method (sql.Row or *sql.Rows)
func (r *LecturerRepository) scan(scanner interface {
	Scan(dest ...interface{}) error
}, l *models.Lecturer) error {
	return scanner.Scan(
		&l.ID,
		&l.UserID,
		&l.LecturerID,
		&l.Department,
		&l.CreatedAt,
	)
}
