package service

import (
	"errors"
	"sistem-prestasi-mhs/app/models"
	"sistem-prestasi-mhs/app/repository"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementService interface {
	CreateAchievement(studentID uuid.UUID, achievement *models.Achievement) (*models.AchievementReference, error)
	GetAchievementByID(id uuid.UUID) (*models.Achievement, *models.AchievementReference, error)
	UpdateAchievement(id uuid.UUID, studentID uuid.UUID, achievement *models.Achievement) error
	DeleteAchievement(id uuid.UUID, studentID uuid.UUID) error
	SubmitForVerification(id uuid.UUID, studentID uuid.UUID) error
	VerifyAchievement(id uuid.UUID, verifierID uuid.UUID) error
	RejectAchievement(id uuid.UUID, verifierID uuid.UUID, note string) error
	GetStudentAchievements(studentID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error)
	GetAdvisorAchievements(advisorID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error)
	GetAllAchievements(limit, offset int, status string) ([]map[string]interface{}, int64, error)
}

type achievementService struct {
	achievementRepo repository.AchievementRepository
	studentRepo     repository.StudentRepository
}

func NewAchievementService(achievementRepo repository.AchievementRepository, studentRepo repository.StudentRepository) AchievementService {
	return &achievementService{
		achievementRepo: achievementRepo,
		studentRepo:     studentRepo,
	}
}

func (s *achievementService) CreateAchievement(studentID uuid.UUID, achievement *models.Achievement) (*models.AchievementReference, error) {
	// Set student ID
	achievement.StudentID = studentID.String()

	// Save to MongoDB
	if err := s.achievementRepo.CreateMongo(achievement); err != nil {
		return nil, err
	}

	// Create reference in PostgreSQL
	ref := &models.AchievementReference{
		StudentID:          studentID,
		MongoAchievementID: achievement.ID.Hex(),
		Status:             models.AchievementStatusDraft,
	}

	if err := s.achievementRepo.CreateReference(ref); err != nil {
		return nil, err
	}

	return ref, nil
}

func (s *achievementService) GetAchievementByID(id uuid.UUID) (*models.Achievement, *models.AchievementReference, error) {
	// Get reference from PostgreSQL
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return nil, nil, err
	}

	// Get achievement from MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return nil, nil, err
	}

	achievement, err := s.achievementRepo.FindMongoByID(mongoID)
	if err != nil {
		return nil, nil, err
	}

	return achievement, ref, nil
}

func (s *achievementService) UpdateAchievement(id uuid.UUID, studentID uuid.UUID, achievement *models.Achievement) error {
	// Get reference
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	// Check ownership
	if ref.StudentID != studentID {
		return errors.New("unauthorized: not your achievement")
	}

	// Check status - only draft can be updated
	if ref.Status != models.AchievementStatusDraft {
		return errors.New("only draft achievements can be updated")
	}

	// Update MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return err
	}
	achievement.ID = mongoID

	return s.achievementRepo.UpdateMongo(achievement)
}

func (s *achievementService) DeleteAchievement(id uuid.UUID, studentID uuid.UUID) error {
	// Get reference
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	// Check ownership
	if ref.StudentID != studentID {
		return errors.New("unauthorized: not your achievement")
	}

	// Check status - only draft can be deleted
	if ref.Status != models.AchievementStatusDraft {
		return errors.New("only draft achievements can be deleted")
	}

	// Soft delete in MongoDB
	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return err
	}
	if err := s.achievementRepo.SoftDeleteMongo(mongoID); err != nil {
		return err
	}

	// Update status to deleted in PostgreSQL
	ref.Status = models.AchievementStatusDeleted
	return s.achievementRepo.UpdateReference(ref)
}

func (s *achievementService) SubmitForVerification(id uuid.UUID, studentID uuid.UUID) error {
	// Get reference
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	// Check ownership
	if ref.StudentID != studentID {
		return errors.New("unauthorized: not your achievement")
	}

	// Check status
	if ref.Status != models.AchievementStatusDraft {
		return errors.New("only draft achievements can be submitted")
	}

	// Update status
	now := time.Now()
	ref.Status = models.AchievementStatusSubmitted
	ref.SubmittedAt = &now

	return s.achievementRepo.UpdateReference(ref)
}

func (s *achievementService) VerifyAchievement(id uuid.UUID, verifierID uuid.UUID) error {
	// Get reference
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	// Check status
	if ref.Status != models.AchievementStatusSubmitted {
		return errors.New("only submitted achievements can be verified")
	}

	// Update status
	now := time.Now()
	ref.Status = models.AchievementStatusVerified
	ref.VerifiedAt = &now
	ref.VerifiedBy = &verifierID

	return s.achievementRepo.UpdateReference(ref)
}

func (s *achievementService) RejectAchievement(id uuid.UUID, verifierID uuid.UUID, note string) error {
	// Get reference
	ref, err := s.achievementRepo.FindReferenceByID(id)
	if err != nil {
		return err
	}

	// Check status
	if ref.Status != models.AchievementStatusSubmitted {
		return errors.New("only submitted achievements can be rejected")
	}

	// Update status
	ref.Status = models.AchievementStatusRejected
	ref.RejectionNote = note
	ref.VerifiedBy = &verifierID

	return s.achievementRepo.UpdateReference(ref)
}

func (s *achievementService) GetStudentAchievements(studentID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error) {
	refs, total, err := s.achievementRepo.FindReferencesByStudentID(studentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return s.combineAchievementsWithReferences(refs), total, nil
}

func (s *achievementService) GetAdvisorAchievements(advisorID uuid.UUID, limit, offset int) ([]map[string]interface{}, int64, error) {
	refs, total, err := s.achievementRepo.FindReferencesByAdvisorID(advisorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return s.combineAchievementsWithReferences(refs), total, nil
}

func (s *achievementService) GetAllAchievements(limit, offset int, status string) ([]map[string]interface{}, int64, error) {
	refs, total, err := s.achievementRepo.FindAllReferences(limit, offset, status)
	if err != nil {
		return nil, 0, err
	}

	return s.combineAchievementsWithReferences(refs), total, nil
}

func (s *achievementService) combineAchievementsWithReferences(refs []models.AchievementReference) []map[string]interface{} {
	var results []map[string]interface{}

	for _, ref := range refs {
		mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		if err != nil {
			continue
		}

		achievement, err := s.achievementRepo.FindMongoByID(mongoID)
		if err != nil {
			continue
		}

		result := map[string]interface{}{
			"id":             ref.ID,
			"student_id":     ref.StudentID,
			"student":        ref.Student,
			"status":         ref.Status,
			"submitted_at":   ref.SubmittedAt,
			"verified_at":    ref.VerifiedAt,
			"verified_by":    ref.VerifiedBy,
			"rejection_note": ref.RejectionNote,
			"achievement":    achievement,
			"created_at":     ref.CreatedAt,
			"updated_at":     ref.UpdatedAt,
		}

		results = append(results, result)
	}

	return results
}
