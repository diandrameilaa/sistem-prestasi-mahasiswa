package repository

import (
	"context"
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	// MongoDB operations
	CreateMongo(achievement *models.Achievement) error
	FindMongoByID(id primitive.ObjectID) (*models.Achievement, error)
	UpdateMongo(achievement *models.Achievement) error
	SoftDeleteMongo(id primitive.ObjectID) error

	// PostgreSQL operations
	CreateReference(ref *models.AchievementReference) error
	FindReferenceByID(id uuid.UUID) (*models.AchievementReference, error)
	FindReferenceByMongoID(mongoID string) (*models.AchievementReference, error)
	FindReferencesByStudentID(studentID uuid.UUID, limit, offset int) ([]models.AchievementReference, int64, error)
	FindReferencesByAdvisorID(advisorID uuid.UUID, limit, offset int) ([]models.AchievementReference, int64, error)
	FindAllReferences(limit, offset int, status string) ([]models.AchievementReference, int64, error)
	UpdateReference(ref *models.AchievementReference) error
	UpdateReferenceStatus(id uuid.UUID, status string) error
}

type achievementRepository struct {
	db              *gorm.DB
	mongoCollection *mongo.Collection
}

func NewAchievementRepository() AchievementRepository {
	return &achievementRepository{
		db:              config.DB,
		mongoCollection: config.MongoDatabase.Collection("achievements"),
	}
}

// MongoDB Operations
func (r *achievementRepository) CreateMongo(achievement *models.Achievement) error {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()
	result, err := r.mongoCollection.InsertOne(context.Background(), achievement)
	if err != nil {
		return err
	}
	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *achievementRepository) FindMongoByID(id primitive.ObjectID) (*models.Achievement, error) {
	var achievement models.Achievement
	err := r.mongoCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&achievement)
	return &achievement, err
}

func (r *achievementRepository) UpdateMongo(achievement *models.Achievement) error {
	achievement.UpdatedAt = time.Now()
	filter := bson.M{"_id": achievement.ID}
	update := bson.M{"$set": achievement}
	_, err := r.mongoCollection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *achievementRepository) SoftDeleteMongo(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"isDeleted": true,
			"updatedAt": time.Now(),
		},
	}
	_, err := r.mongoCollection.UpdateOne(context.Background(), filter, update)
	return err
}

// PostgreSQL Operations
func (r *achievementRepository) CreateReference(ref *models.AchievementReference) error {
	return r.db.Create(ref).Error
}

func (r *achievementRepository) FindReferenceByID(id uuid.UUID) (*models.AchievementReference, error) {
	var ref models.AchievementReference
	err := r.db.Preload("Student").Preload("Student.User").Preload("Verifier").First(&ref, "id = ?", id).Error
	return &ref, err
}

func (r *achievementRepository) FindReferenceByMongoID(mongoID string) (*models.AchievementReference, error) {
	var ref models.AchievementReference
	err := r.db.Preload("Student").Preload("Student.User").First(&ref, "mongo_achievement_id = ?", mongoID).Error
	return &ref, err
}

func (r *achievementRepository) FindReferencesByStudentID(studentID uuid.UUID, limit, offset int) ([]models.AchievementReference, int64, error) {
	var refs []models.AchievementReference
	var total int64

	query := r.db.Model(&models.AchievementReference{}).Where("student_id = ? AND status != ?", studentID, models.AchievementStatusDeleted)
	query.Count(&total)
	err := query.Preload("Student").Preload("Student.User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&refs).Error

	return refs, total, err
}

func (r *achievementRepository) FindReferencesByAdvisorID(advisorID uuid.UUID, limit, offset int) ([]models.AchievementReference, int64, error) {
	var refs []models.AchievementReference
	var total int64

	// Find students with this advisor
	var studentIDs []uuid.UUID
	r.db.Model(&models.Student{}).Where("advisor_id = ?", advisorID).Pluck("id", &studentIDs)

	query := r.db.Model(&models.AchievementReference{}).Where("student_id IN ? AND status != ?", studentIDs, models.AchievementStatusDeleted)
	query.Count(&total)
	err := query.Preload("Student").Preload("Student.User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&refs).Error

	return refs, total, err
}

func (r *achievementRepository) FindAllReferences(limit, offset int, status string) ([]models.AchievementReference, int64, error) {
	var refs []models.AchievementReference
	var total int64

	query := r.db.Model(&models.AchievementReference{}).Where("status != ?", models.AchievementStatusDeleted)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Preload("Student").Preload("Student.User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&refs).Error

	return refs, total, err
}

func (r *achievementRepository) UpdateReference(ref *models.AchievementReference) error {
	return r.db.Save(ref).Error
}

func (r *achievementRepository) UpdateReferenceStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.AchievementReference{}).Where("id = ?", id).Update("status", status).Error
}
