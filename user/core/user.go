package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/constants"
	"github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/affanahmad024/quiz-craft/backend/user/model"
)

// GetUserByID gets user by ID (excludes soft deleted)
func GetUserByID(ctx context.Context, id uuid.UUID) (model.UserAPI, error) {
	commons.Log(ctx, "[core.GetUserByID [id=%s", id)

	userDAO, err := dao.FindByID(ctx, id)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserByID::KO [id=%s [err=%v", id, err)
		return model.UserAPI{}, err
	}

	return userDAO.ToAPI(), nil
}

// GetUserProfile gets user profile (same as GetUserByID but semantically clearer)
func GetUserProfile(ctx context.Context, userID uuid.UUID) (model.UserAPI, error) {
	commons.Log(ctx, "[core.GetUserProfile [user_id=%s", userID)
	return GetUserByID(ctx, userID)
}

// UpdateUserProfile updates user profile
func UpdateUserProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) error {
	commons.Log(ctx, "[core.UpdateUserProfile [user_id=%s", userID)

	// Get current user data
	userDAO, err := dao.FindByID(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateUserProfile::FindErr [user_id=%s [err=%v", userID, err)
		return err
	}

	// Apply updates
	name := userDAO.Name
	phoneNumber := ""
	if userDAO.PhoneNumber.Valid {
		phoneNumber = userDAO.PhoneNumber.String
	}

	// Parse current meta
	meta := make(map[string]interface{})
	if len(userDAO.Meta) > 0 {
		// Would need to unmarshal existing meta, but for simplicity keeping it simple
	}

	if req.Name != nil {
		name = *req.Name
	}
	if req.PhoneNumber != nil {
		phoneNumber = *req.PhoneNumber
	}
	if req.Meta != nil {
		// Merge meta
		for k, v := range *req.Meta {
			meta[k] = v
		}
	}

	// Update in database
	err = dao.UpdateUser(ctx, userID, name, phoneNumber, meta)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateUserProfile::UpdateErr [user_id=%s [err=%v", userID, err)
		return err
	}

	commons.Log(ctx, "[core.UpdateUserProfile::OK [user_id=%s", userID)
	return nil
}

// DeleteUser soft deletes a user
func DeleteUser(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[core.DeleteUser [user_id=%s", userID)

	// Check if user exists
	_, err := dao.FindByID(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteUser::FindErr [user_id=%s [err=%v", userID, err)
		return err
	}

	// Soft delete
	err = dao.SoftDelete(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteUser::DeleteErr [user_id=%s [err=%v", userID, err)
		return err
	}

	commons.Log(ctx, "[core.DeleteUser::OK [user_id=%s", userID)
	return nil
}

// RestoreUser restores a soft deleted user
func RestoreUser(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[core.RestoreUser [user_id=%s", userID)

	// Check if user exists and is deleted
	userDAO, err := dao.FindByIDIncludingDeleted(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.RestoreUser::FindErr [user_id=%s [err=%v", userID, err)
		return err
	}

	if !userDAO.DeletedAt.Valid {
		commons.LogW(ctx, "[core.RestoreUser::NotDeleted [user_id=%s", userID)
		return commons.BadInput("user-not-deleted")
	}

	// Restore
	err = dao.Restore(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.RestoreUser::RestoreErr [user_id=%s [err=%v", userID, err)
		return err
	}

	commons.Log(ctx, "[core.RestoreUser::OK [user_id=%s", userID)
	return nil
}

// DisableUser disables a user account
func DisableUser(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[core.DisableUser [user_id=%s", userID)

	err := dao.UpdateEnabled(ctx, userID, false)
	if err != nil {
		commons.LogE(ctx, "[core.DisableUser::UpdateErr [user_id=%s [err=%v", userID, err)
		return err
	}

	commons.Log(ctx, "[core.DisableUser::OK [user_id=%s", userID)
	return nil
}

// EnableUser enables a user account
func EnableUser(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[core.EnableUser [user_id=%s", userID)

	err := dao.UpdateEnabled(ctx, userID, true)
	if err != nil {
		commons.LogE(ctx, "[core.EnableUser::UpdateErr [user_id=%s [err=%v", userID, err)
		return err
	}

	commons.Log(ctx, "[core.EnableUser::OK [user_id=%s", userID)
	return nil
}

// ListUsers lists users with pagination and filters
func ListUsers(ctx context.Context, filter model.UserListFilter) (model.UserListResponse, error) {
	commons.Log(ctx, "[core.ListUsers [page=%d [limit=%d", filter.Page, filter.Limit)

	// Set defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = constants.DefaultPageSize
	}
	if filter.Limit > constants.MaxPageSize {
		filter.Limit = constants.MaxPageSize
	}

	// Get users from DAO
	usersDAO, total, err := dao.ListUsers(ctx, filter)
	if err != nil {
		commons.LogE(ctx, "[core.ListUsers::DAOErr [err=%v", err)
		return model.UserListResponse{}, err
	}

	// Convert to API models
	users := make([]model.UserAPI, 0, len(usersDAO))
	for _, userDAO := range usersDAO {
		users = append(users, userDAO.ToAPI())
	}

	resp := model.UserListResponse{
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
		Users: users,
	}

	commons.Log(ctx, "[core.ListUsers::OK [total=%d [returned=%d", total, len(users))
	return resp, nil
}

// GetUserStats gets user statistics
func GetUserStats(ctx context.Context) (model.UserStatsResponse, error) {
	commons.Log(ctx, "[core.GetUserStats")

	totalUsers, err := dao.CountUsers(ctx)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserStats::CountUsersErr [err=%v", err)
		return model.UserStatsResponse{}, err
	}

	deletedUsers, err := dao.CountDeletedUsers(ctx)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserStats::CountDeletedErr [err=%v", err)
		return model.UserStatsResponse{}, err
	}

	stats := model.UserStatsResponse{
		TotalUsers:   totalUsers,
		ActiveUsers:  totalUsers, // Active = non-deleted
		DeletedUsers: deletedUsers,
	}

	commons.Log(ctx, "[core.GetUserStats::OK [total=%d [deleted=%d", totalUsers, deletedUsers)
	return stats, nil
}
