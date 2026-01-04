package repositories

import (
	"errors"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID (excludes soft deleted)
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByIDWithPassword finds a user by ID with password (excludes soft deleted)
func (r *UserRepository) FindByIDWithPassword(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Password").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email (excludes soft deleted)
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmailWithPassword finds a user by email with password (excludes soft deleted)
func (r *UserRepository) FindByEmailWithPassword(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Password").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByIDIncludingDeleted finds a user by ID (includes soft deleted)
func (r *UserRepository) FindByIDIncludingDeleted(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Unscoped().Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates user information
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// UpdateFields updates specific fields
func (r *UserRepository) UpdateFields(id uuid.UUID, fields map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(fields).Error
}

// SoftDelete performs soft delete on a user
func (r *UserRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, id).Error
}

// HardDelete permanently deletes a user (use with caution)
func (r *UserRepository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.User{}, id).Error
}

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(id uuid.UUID) error {
	return r.db.Model(&models.User{}).Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// EmailExists checks if email already exists (excludes soft deleted)
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// EmailExistsIncludingDeleted checks if email exists including soft deleted users
func (r *UserRepository) EmailExistsIncludingDeleted(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Unscoped().Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// List retrieves users with pagination and filters
func (r *UserRepository) List(page, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error

	return users, total, err
}

// ListIncludingDeleted retrieves users including soft deleted
func (r *UserRepository) ListIncludingDeleted(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).Unscoped()

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error

	return users, total, err
}

// CreatePassword creates user password record
func (r *UserRepository) CreatePassword(password *models.UserPassword) error {
	return r.db.Create(password).Error
}

// FindPasswordByUserID finds password by user ID
func (r *UserRepository) FindPasswordByUserID(userID uuid.UUID) (*models.UserPassword, error) {
	var password models.UserPassword
	err := r.db.Where("user_id = ?", userID).First(&password).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("password not found")
		}
		return nil, err
	}
	return &password, nil
}

// UpdatePassword updates user password record
func (r *UserRepository) UpdatePassword(password *models.UserPassword) error {
	return r.db.Save(password).Error
}

// UpdatePasswordFields updates specific password fields
func (r *UserRepository) UpdatePasswordFields(userID uuid.UUID, fields map[string]interface{}) error {
	return r.db.Model(&models.UserPassword{}).Where("user_id = ?", userID).Updates(fields).Error
}

// FindByVerificationToken finds user by email verification token
func (r *UserRepository) FindByVerificationToken(token string) (*models.User, error) {
	var password models.UserPassword
	err := r.db.Preload("User").
		Where("email_verification_token = ? AND email_token_expires_at > ?", token, time.Now()).
		First(&password).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired verification token")
		}
		return nil, err
	}

	return &password.User, nil
}

// FindByResetToken finds user by password reset token
func (r *UserRepository) FindByResetToken(token string) (*models.User, error) {
	var password models.UserPassword
	err := r.db.Preload("User").
		Where("password_reset_token = ? AND reset_token_expires_at > ?", token, time.Now()).
		First(&password).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired reset token")
		}
		return nil, err
	}

	return &password.User, nil
}

// CountUsers returns total user count (excludes soft deleted)
func (r *UserRepository) CountUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}

// CountDeletedUsers returns count of soft deleted users
func (r *UserRepository) CountDeletedUsers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Unscoped().Where("deleted_at IS NOT NULL").Count(&count).Error
	return count, err
}
