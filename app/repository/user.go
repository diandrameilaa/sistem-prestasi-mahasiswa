package repository

import (
	"sistem-prestasi-mhs/app/config"
	"sistem-prestasi-mhs/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindAll(limit, offset int) ([]models.User, int64, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	UpdateRole(userID, roleID uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{db: config.DB}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Preload("Role.Permissions").First(&user, "id = ?", id).Error
	return &user, err
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Preload("Role.Permissions").First(&user, "username = ?", username).Error
	return &user, err
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").Preload("Role.Permissions").First(&user, "email = ?", email).Error
	return &user, err
}

func (r *userRepository) FindAll(limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	r.db.Model(&models.User{}).Count(&total)
	err := r.db.Preload("Role").Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *userRepository) UpdateRole(userID, roleID uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}
