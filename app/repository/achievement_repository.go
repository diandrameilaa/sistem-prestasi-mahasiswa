package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"sistem-prestasi-mhs/app/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collectionName = "achievements"
	// Kolom yang sering dipanggil dipisahkan agar konsisten
	refColumns = `id, student_id, mongo_achievement_id, status, 
                  submitted_at, verified_at, verified_by, rejection_note, 
                  created_at, updated_at`
)

type AchievementRepository struct {
	db   *sql.DB
	coll *mongo.Collection
}

// NewAchievementRepository menginisialisasi repository
func NewAchievementRepository(db *sql.DB, mongoDB *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		db:   db,
		coll: mongoDB.Collection(collectionName),
	}
}

// ============================================================================
// MONGODB OPERATIONS
// ============================================================================

func (r *AchievementRepository) CreateAchievement(ctx context.Context, achievement *models.Achievement) error {
	achievement.ID = primitive.NewObjectID() // Ensure ID is generated
	_, err := r.coll.InsertOne(ctx, achievement)
	return err
}

func (r *AchievementRepository) GetAchievementByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
	var achievement models.Achievement
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&achievement); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("mongo achievement not found")
		}
		return nil, err
	}
	return &achievement, nil
}

func (r *AchievementRepository) SoftDeleteAchievement(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// ============================================================================
// POSTGRESQL OPERATIONS
// ============================================================================

func (r *AchievementRepository) CreateReference(ctx context.Context, ref *models.AchievementReference) error {
	query := fmt.Sprintf(`
		INSERT INTO achievement_references (id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING %s`, refColumns)

	return r.scanReference(r.db.QueryRowContext(ctx, query,
		ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status,
	), ref)
}

func (r *AchievementRepository) GetReferenceByID(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
	query := fmt.Sprintf("SELECT %s FROM achievement_references WHERE id = $1", refColumns)

	var ref models.AchievementReference
	err := r.scanReference(r.db.QueryRowContext(ctx, query, id), &ref)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reference not found")
		}
		return nil, err
	}
	return &ref, nil
}

// UpdateStatus handles generic status updates (Submit, Verify, etc)
func (r *AchievementRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifierID *uuid.UUID, note *string) error {
	// Build query dynamically based on args to keep it optimal
	var query string
	var args []interface{}

	args = append(args, status, id)

	baseQuery := "UPDATE achievement_references SET status = $1, updated_at = NOW()"

	// Add conditional updates
	switch status {
	case "submitted":
		baseQuery += ", submitted_at = NOW()"
	case "verified", "rejected":
		baseQuery += ", verified_at = NOW()"
		if verifierID != nil {
			baseQuery += ", verified_by = $3"
			args = append(args, *verifierID)
		}
		if note != nil {
			// Adjust argument index based on whether verifierID was added
			noteIdx := 3
			if verifierID != nil {
				noteIdx = 4
			}
			baseQuery += fmt.Sprintf(", rejection_note = $%d", noteIdx)
			args = append(args, *note)
		}
	case "deleted":
		// No extra fields for deleted
	}

	query = baseQuery + " WHERE id = $2"

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("reference not found or no change")
	}

	return nil
}

// ============================================================================
// LIST OPERATIONS
// ============================================================================

func (r *AchievementRepository) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*models.AchievementReference, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM achievement_references 
		WHERE student_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, refColumns)

	rows, err := r.db.QueryContext(ctx, query, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.scanRows(rows)
}

func (r *AchievementRepository) GetAdviseeAchievements(ctx context.Context, adviseeIDs []uuid.UUID, limit, offset int) ([]*models.AchievementReference, error) {
	// Note: Requires lib/pq for ANY($1) with slice, or use pgx driver naturally.
	// Ensure adviseeIDs is passed as pq.Array(adviseeIDs) in service layer if using lib/pq
	query := fmt.Sprintf(`
		SELECT %s FROM achievement_references 
		WHERE student_id = ANY($1) AND status = 'submitted'
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, refColumns)

	rows, err := r.db.QueryContext(ctx, query, adviseeIDs, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.scanRows(rows)
}

func (r *AchievementRepository) ListAll(ctx context.Context, limit, offset int) ([]*models.AchievementReference, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM achievement_references 
		WHERE status != 'deleted'
		ORDER BY created_at DESC LIMIT $1 OFFSET $2`, refColumns)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return r.scanRows(rows)
}

// ============================================================================
// HELPERS (DRY)
// ============================================================================

// scanReference helps scanning a single row into the struct
func (r *AchievementRepository) scanReference(scanner interface {
	Scan(dest ...interface{}) error
}, ref *models.AchievementReference) error {
	return scanner.Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)
}

// scanRows helps scanning multiple rows
func (r *AchievementRepository) scanRows(rows *sql.Rows) ([]*models.AchievementReference, error) {
	defer rows.Close()
	var results []*models.AchievementReference

	for rows.Next() {
		var ref models.AchievementReference
		if err := r.scanReference(rows, &ref); err != nil {
			return nil, err
		}
		results = append(results, &ref)
	}

	return results, rows.Err()
}
