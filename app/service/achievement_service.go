package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
)

type AchievementService struct {
	achRepo     *repository.AchievementRepository
	studentRepo *repository.StudentRepository
	userRepo    *repository.UserRepository
}

func NewAchievementService(ar *repository.AchievementRepository, sr *repository.StudentRepository, ur *repository.UserRepository) *AchievementService {
	return &AchievementService{
		achRepo:     ar,
		studentRepo: sr,
		userRepo:    ur,
	}
}

// ============================================================================
// CORE BUSINESS LOGIC
// ============================================================================

func (s *AchievementService) CreateAchievement(ctx context.Context, studentID uuid.UUID, req *models.CreateAchievementRequest) (*models.Achievement, error) {
	// 1. Prepare MongoDB Document
	achievement := &models.Achievement{
		StudentID:       studentID,
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         req.Details,
		Tags:            req.Tags,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 2. Insert into MongoDB
	// Note: ID is generated inside repository/service before insert or by Mongo driver
	if err := s.achRepo.CreateAchievement(ctx, achievement); err != nil {
		return nil, err
	}

	// 3. Prepare PostgreSQL Reference
	ref := &models.AchievementReference{
		ID:                 uuid.New(),
		StudentID:          studentID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// 4. Insert into PostgreSQL with Compensation Logic (Rollback)
	if err := s.achRepo.CreateReference(ctx, ref); err != nil {
		// CRITICAL: If SQL fails, delete the orphan document in Mongo
		_ = s.achRepo.SoftDeleteAchievement(ctx, achievement.ID)
		return nil, err
	}

	return achievement, nil
}

func (s *AchievementService) GetAchievementDetail(ctx context.Context, referenceID uuid.UUID) (*models.Achievement, error) {
	// 1. Get Reference to find Mongo ID
	ref, err := s.achRepo.GetReferenceByID(ctx, referenceID)
	if err != nil {
		return nil, err
	}

	// 2. Parse ID
	oid, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return nil, errors.New("invalid data integrity: malformed mongo id")
	}

	// 3. Get Actual Data
	// Optional: You might want to merge SQL status data into the response here
	return s.achRepo.GetAchievementByID(ctx, oid)
}

func (s *AchievementService) ListAchievements(ctx context.Context, limit, offset int) ([]*models.AchievementReference, error) {
	return s.achRepo.ListAll(ctx, limit, offset)
}

// ============================================================================
// STATE MANAGEMENT (Submit -> Verify/Reject)
// ============================================================================

func (s *AchievementService) SubmitForVerification(ctx context.Context, id uuid.UUID) error {
	// Business Rule: Ensure it is currently a draft
	ref, err := s.achRepo.GetReferenceByID(ctx, id)
	if err != nil {
		return err
	}
	if ref.Status != "draft" {
		return errors.New("only draft achievements can be submitted")
	}

	return s.achRepo.UpdateStatus(ctx, id, "submitted", nil, nil)
}

func (s *AchievementService) VerifyAchievement(ctx context.Context, id, verifierID uuid.UUID) error {
	// Business Rule: Ensure it is submitted
	ref, err := s.achRepo.GetReferenceByID(ctx, id)
	if err != nil {
		return err
	}
	if ref.Status != "submitted" {
		return errors.New("achievement is not in submitted state")
	}

	return s.achRepo.UpdateStatus(ctx, id, "verified", &verifierID, nil)
}

func (s *AchievementService) RejectAchievement(ctx context.Context, id, verifierID uuid.UUID, note string) error {
	ref, err := s.achRepo.GetReferenceByID(ctx, id)
	if err != nil {
		return err
	}
	if ref.Status != "submitted" {
		return errors.New("achievement is not in submitted state")
	}

	return s.achRepo.UpdateStatus(ctx, id, "rejected", &verifierID, &note)
}

func (s *AchievementService) DeleteAchievement(ctx context.Context, id, studentID uuid.UUID) error {
	// 1. Validate Ownership & Status
	ref, err := s.achRepo.GetReferenceByID(ctx, id)
	if err != nil {
		return err
	}

	if ref.StudentID != studentID {
		return errors.New("unauthorized: you do not own this achievement")
	}
	if ref.Status != "draft" && ref.Status != "rejected" {
		return errors.New("cannot delete achievement that is under review or verified")
	}

	// 2. Parse Mongo ID
	oid, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return err
	}

	// 3. Soft Delete in Mongo
	if err := s.achRepo.SoftDeleteAchievement(ctx, oid); err != nil {
		return err
	}

	// 4. Mark as deleted in SQL
	return s.achRepo.UpdateStatus(ctx, id, "deleted", nil, nil)
}

// ============================================================================
// HELPERS
// ============================================================================

// ValidateOwnership checks if the user owns the achievement resource
func (s *AchievementService) ValidateOwnership(ctx context.Context, achievementID, userID uuid.UUID) error {
	ref, err := s.achRepo.GetReferenceByID(ctx, achievementID)
	if err != nil {
		return err
	}

	// Need to check if userID matches the student's UserID.
	// We have ref.StudentID, we need to resolve it to UserID or check if caller is the student.
	student, err := s.studentRepo.GetByID(ctx, ref.StudentID)
	if err != nil {
		return errors.New("student profile not found")
	}

	if student.UserID != userID {
		return errors.New("access denied: resource does not belong to user")
	}

	return nil
}
