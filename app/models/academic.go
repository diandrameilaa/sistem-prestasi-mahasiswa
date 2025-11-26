package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Student struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;unique;not null"`
	User         User       `gorm:"foreignKey:UserID"`
	StudentID    string     `gorm:"type:varchar(20);unique;not null"` // NIM
	ProgramStudy string     `gorm:"type:varchar(100)"`
	AcademicYear string     `gorm:"type:varchar(10)"`
	AdvisorID    *uuid.UUID `gorm:"type:uuid"` // Bisa null jika belum dapet dospem
	Lecturer     *Lecturer  `gorm:"foreignKey:AdvisorID"`
	CreatedAt    time.Time
}

type Lecturer struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;unique;not null"`
	User       User      `gorm:"foreignKey:UserID"`
	LecturerID string    `gorm:"type:varchar(20);unique;not null"` // NIP
	Department string    `gorm:"type:varchar(100)"`
	CreatedAt  time.Time
}

func (s *Student) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New()
	return
}

func (l *Lecturer) BeforeCreate(tx *gorm.DB) (err error) {
	l.ID = uuid.New()
	return
}
