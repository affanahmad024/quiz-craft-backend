package services

import (
	"errors"
	"fmt"

	"github.com/affanahmad024/quiz-craft/backend/internal/models"
	"github.com/affanahmad024/quiz-craft/backend/internal/repositories"
	"github.com/affanahmad024/quiz-craft/backend/internal/utils"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"github.com/google/uuid"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Name        *string           `json:"name"`
	PhoneNumber *string           `json:"phone_number"`
	Meta        *models.Meta      `json:"meta"`
}

// GetProfile retrieves user profile by ID
func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(userID uuid.UUID, req UpdateProfileRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		name := utils.SanitizeString(*req.Name)
		if err := utils.ValidateName(name); err != nil {
			return nil, err
		}
		user.Name = name
	}

	if req.PhoneNumber != nil {
		if err := utils.ValidatePhoneNumber(*req.PhoneNumber); err != nil {
			return nil, err
		}
		user.PhoneNumber = *req.PhoneNumber
	}

	if req.Meta != nil {
		// Merge with existing meta
		if user.Meta == nil {
			user.Meta = models.Meta{}
		}
		for key, value := range *req.Meta {
			user.Meta[key] = value
		}
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.GetLogger().Infof("User profile updated: %s", user.Email)
	return user, nil
}

// DeleteAccount performs soft delete on user account
func (s *UserService) DeleteAccount(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Perform soft delete
	if err := s.userRepo.SoftDelete(userID); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	logger.GetLogger().Infof("User account deleted (soft): %s", user.Email)
	return nil
}

// RestoreAccount restores a soft-deleted account
func (s *UserService) RestoreAccount(userID uuid.UUID) error {
	// Check if user exists (including deleted)
	user, err := s.userRepo.FindByIDIncludingDeleted(userID)
	if err != nil {
		return err
	}

	// Check if user is actually deleted
	if user.DeletedAt.Valid == false {
		return errors.New("account is not deleted")
	}

	// Restore account
	if err := s.userRepo.Restore(userID); err != nil {
		return fmt.Errorf("failed to restore account: %w", err)
	}

	logger.GetLogger().Infof("User account restored: %s", user.Email)
	return nil
}

// PermanentlyDeleteAccount performs hard delete (DANGEROUS - cannot be undone)
func (s *UserService) PermanentlyDeleteAccount(userID uuid.UUID) error {
	user, err := s.userRepo.FindByIDIncludingDeleted(userID)
	if err != nil {
		return err
	}

	// Perform hard delete
	if err := s.userRepo.HardDelete(userID); err != nil {
		return fmt.Errorf("failed to permanently delete account: %w", err)
	}

	logger.GetLogger().Warnf("User account permanently deleted: %s", user.Email)
	return nil
}

// ListUsers retrieves users with pagination
func (s *UserService) ListUsers(page, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	users, total, err := s.userRepo.List(page, limit, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// GetUserStats retrieves user statistics
func (s *UserService) GetUserStats() (map[string]interface{}, error) {
	totalUsers, err := s.userRepo.CountUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	deletedUsers, err := s.userRepo.CountDeletedUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to count deleted users: %w", err)
	}

	stats := map[string]interface{}{
		"total_users":   totalUsers,
		"active_users":  totalUsers,
		"deleted_users": deletedUsers,
	}

	return stats, nil
}

// DisableAccount disables a user account (admin function)
func (s *UserService) DisableAccount(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if !user.Enabled {
		return errors.New("account is already disabled")
	}

	user.Enabled = false

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to disable account: %w", err)
	}

	logger.GetLogger().Infof("User account disabled: %s", user.Email)
	return nil
}

// EnableAccount enables a user account (admin function)
func (s *UserService) EnableAccount(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if user.Enabled {
		return errors.New("account is already enabled")
	}

	user.Enabled = true

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to enable account: %w", err)
	}

	logger.GetLogger().Infof("User account enabled: %s", user.Email)
	return nil
}
