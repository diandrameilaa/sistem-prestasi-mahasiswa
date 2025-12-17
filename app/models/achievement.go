package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementReference struct {
	ID                  uuid.UUID `json:"id"`
	StudentID           uuid.UUID `json:"student_id"`
	MongoAchievementID  string    `json:"mongo_achievement_id"`
	Status              string    `json:"status"` // draft, submitted, verified, rejected
	SubmittedAt         *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt          *time.Time `json:"verified_at,omitempty"`
	VerifiedBy          *uuid.UUID `json:"verified_by,omitempty"`
	RejectionNote       *string   `json:"rejection_note,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Achievement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       uuid.UUID          `bson:"studentId" json:"student_id"`
	AchievementType string             `bson:"achievementType" json:"achievement_type"` // academic, competition, organization, publication, certification, other
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`
	Details         map[string]interface{} `bson:"details" json:"details"`
	Attachments     []Attachment       `bson:"attachments" json:"attachments"`
	Tags            []string           `bson:"tags" json:"tags"`
	Points          int                `bson:"points" json:"points"`
	CreatedAt       time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updated_at"`
}

type Attachment struct {
	FileName   string    `bson:"fileName" json:"file_name"`
	FileURL    string    `bson:"fileUrl" json:"file_url"`
	FileType   string    `bson:"fileType" json:"file_type"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploaded_at"`
}

type CreateAchievementRequest struct {
	AchievementType string                 `json:"achievement_type" validate:"required"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
}

type SubmitAchievementRequest struct {
	Comment string `json:"comment"`
}

type VerifyAchievementRequest struct {
	Approved bool   `json:"approved"`
	Note     string `json:"note"`
}

type AchievementResponse struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Details     map[string]interface{} `json:"details"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
