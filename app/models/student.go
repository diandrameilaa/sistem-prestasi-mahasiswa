package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	StudentID      string    `json:"student_id"`
	ProgramStudy   string    `json:"program_study"`
	AcademicYear   string    `json:"academic_year"`
	AdvisorID      *uuid.UUID `json:"advisor_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Lecturer struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	LecturerID string    `json:"lecturer_id"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateStudentRequest struct {
	Username     string `json:"username" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	FullName     string `json:"full_name" validate:"required"`
	StudentID    string `json:"student_id" validate:"required"`
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
}

type SetAdvisorRequest struct {
	AdvisorID uuid.UUID `json:"advisor_id" validate:"required"`
}
