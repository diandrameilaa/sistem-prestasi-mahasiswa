package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
)

// Const untuk Status Prestasi
const (
	AchievementStatusDraft     = "draft"
	AchievementStatusSubmitted = "submitted"
	AchievementStatusVerified  = "verified"
	AchievementStatusRejected  = "rejected"
	AchievementStatusDeleted   = "deleted" // Requirement tambahan untuk soft delete
)

// PostgreSQL Model: AchievementReference
type AchievementReference struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StudentID          uuid.UUID `gorm:"type:uuid;not null"`
	Student            Student   `gorm:"foreignKey:StudentID"`
	MongoAchievementID string    `gorm:"type:varchar(24);not null"`
	Status             string    `gorm:"type:varchar(20);default:'draft'"` // Enum handled by logic
	SubmittedAt        *time.Time
	VerifiedAt         *time.Time
	VerifiedBy         *uuid.UUID `gorm:"type:uuid"`
	Verifier           *User      `gorm:"foreignKey:VerifiedBy"`
	RejectionNote      string     `gorm:"type:text"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (ar *AchievementReference) BeforeCreate(tx *gorm.DB) (err error) {
	ar.ID = uuid.New()
	return
}

// MongoDB Model: Achievement
// Disimpan di MongoDB karena field 'Details' sangat dinamis
type Achievement struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	StudentID       string                 `bson:"studentId" json:"student_id"` // Reference UUID string
	AchievementType string                 `bson:"achievementType" json:"achievement_type"`
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description" json:"description"`
	Details         map[string]interface{} `bson:"details" json:"details"` // Field Dinamis
	Attachments     []Attachment           `bson:"attachments" json:"attachments"`
	Tags            []string               `bson:"tags" json:"tags"`
	Points          float64                `bson:"points" json:"points"`
	CreatedAt       time.Time              `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time              `bson:"updatedAt" json:"updated_at"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"file_name"`
	FileURL    string    `bson:"fileUrl" json:"file_url"`
	FileType   string    `bson:"fileType" json:"file_type"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploaded_at"`
}
