package repository

import (
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepository interface {
	Create(student *models.Student) error
	FindByID(id uuid.UUID) (*models.Student, error)
	FindByUserID(userID uuid.UUID) (*models.Student, error)
	FindByStudentID(studentID string) (*models.Student, error)
	FindAll(limit, offset int) ([]models.Student, int64, error)
	Update(student *models.Student) error
	UpdateAdvisor(studentID, advisorID uuid.UUID) error
	FindByAdvisorID(advisorID uuid.UUID) ([]models.Student, error)

	// Lecturer operations
	CreateLecturer(lecturer *models.Lecturer) error
	FindLecturerByID(id uuid.UUID) (*models.Lecturer, error)
	FindLecturerByUserID(userID uuid.UUID) (*models.Lecturer, error)
	FindAllLecturers(limit, offset int) ([]models.Lecturer, int64, error)
}

type studentRepository struct {
	db *gorm.DB
}

func NewStudentRepository() StudentRepository {
	return &studentRepository{db: config.DB}
}

func (r *studentRepository) Create(student *models.Student) error {
	return r.db.Create(student).Error
}

func (r *studentRepository) FindByID(id uuid.UUID) (*models.Student, error) {
	var student models.Student
	err := r.db.Preload("User").Preload("Lecturer").Preload("Lecturer.User").First(&student, "id = ?", id).Error
	return &student, err
}

func (r *studentRepository) FindByUserID(userID uuid.UUID) (*models.Student, error) {
	var student models.Student
	err := r.db.Preload("User").Preload("Lecturer").Preload("Lecturer.User").First(&student, "user_id = ?", userID).Error
	return &student, err
}

func (r *studentRepository) FindByStudentID(studentID string) (*models.Student, error) {
	var student models.Student
	err := r.db.Preload("User").First(&student, "student_id = ?", studentID).Error
	return &student, err
}

func (r *studentRepository) FindAll(limit, offset int) ([]models.Student, int64, error) {
	var students []models.Student
	var total int64

	r.db.Model(&models.Student{}).Count(&total)
	err := r.db.Preload("User").Preload("Lecturer").Limit(limit).Offset(offset).Find(&students).Error

	return students, total, err
}

func (r *studentRepository) Update(student *models.Student) error {
	return r.db.Save(student).Error
}

func (r *studentRepository) UpdateAdvisor(studentID, advisorID uuid.UUID) error {
	return r.db.Model(&models.Student{}).Where("id = ?", studentID).Update("advisor_id", advisorID).Error
}

func (r *studentRepository) FindByAdvisorID(advisorID uuid.UUID) ([]models.Student, error) {
	var students []models.Student
	err := r.db.Preload("User").Where("advisor_id = ?", advisorID).Find(&students).Error
	return students, err
}

// Lecturer operations
func (r *studentRepository) CreateLecturer(lecturer *models.Lecturer) error {
	return r.db.Create(lecturer).Error
}

func (r *studentRepository) FindLecturerByID(id uuid.UUID) (*models.Lecturer, error) {
	var lecturer models.Lecturer
	err := r.db.Preload("User").First(&lecturer, "id = ?", id).Error
	return &lecturer, err
}

func (r *studentRepository) FindLecturerByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	var lecturer models.Lecturer
	err := r.db.Preload("User").First(&lecturer, "user_id = ?", userID).Error
	return &lecturer, err
}

func (r *studentRepository) FindAllLecturers(limit, offset int) ([]models.Lecturer, int64, error) {
	var lecturers []models.Lecturer
	var total int64

	r.db.Model(&models.Lecturer{}).Count(&total)
	err := r.db.Preload("User").Limit(limit).Offset(offset).Find(&lecturers).Error

	return lecturers, total, err
}
