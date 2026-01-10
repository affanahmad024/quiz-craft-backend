package core

import (
	"context"

	"github.com/affanahmad024/quiz-craft/backend/admin/dao"
	"github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/google/uuid"
)

// ListUsers retrieves users with admin view (includes stats and filters)
func ListUsers(ctx context.Context, adminID string, page, limit int, role *string, enabled *bool, emailVerified *bool, search *string, sortBy, sortOrder string) ([]model.AdminUserDetail, int, error) {
	commons.Log(ctx, "[core.ListUsers [admin_id=%s [page=%d", adminID, page)

	// Validate admin (already done in middleware, but double-check)
	if !isAdmin(ctx, adminID) {
		return nil, 0, commons.Forbidden("admin-access-required")
	}

	// Set defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}

	users, total, err := dao.ListUsersAdmin(ctx, page, limit, role, enabled, emailVerified, search, sortBy, sortOrder)
	if err != nil {
		commons.LogE(ctx, "[core.ListUsers::DAOErr [err=%v", err)
		return nil, 0, err
	}

	return users, total, nil
}

// GetUserDetail retrieves detailed user info for admin
func GetUserDetail(ctx context.Context, adminID string, targetUserID string) (*model.AdminUserDetail, error) {
	commons.Log(ctx, "[core.GetUserDetail [admin_id=%s [target_user_id=%s", adminID, targetUserID)

	// Validate admin
	if !isAdmin(ctx, adminID) {
		return nil, commons.Forbidden("admin-access-required")
	}

	userUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return nil, commons.BadInput("invalid-user-id")
	}

	user, err := dao.GetUserDetailAdmin(ctx, userUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetUserDetail::DAOErr [err=%v", err)
		return nil, err
	}

	return user, nil
}

// DisableUser disables a user account
func DisableUser(ctx context.Context, adminID string, targetUserID string, reason string) error {
	commons.Log(ctx, "[core.DisableUser [admin_id=%s [target_user_id=%s", adminID, targetUserID)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return commons.Forbidden("admin-access-required")
	}

	// Prevent self-disable
	if adminID == targetUserID {
		return commons.BadInput("cannot-disable-self")
	}

	// Get target user
	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return commons.BadInput("invalid-user-id")
	}

	targetUser, err := userdao.FindByID(ctx, targetUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DisableUser::FindErr [err=%v", err)
		return err
	}

	// Disable user
	err = userdao.UpdateEnabled(ctx, targetUUID, false)
	if err != nil {
		commons.LogE(ctx, "[core.DisableUser::UpdateErr [err=%v", err)
		return err
	}

	// Log activity
	LogActivity(ctx, adminID, model.ActionUserDisabled, "user", targetUserID, map[string]interface{}{
		"reason":      reason,
		"admin_name":  admin.Name,
		"target_name": targetUser.Name,
	})

	commons.Log(ctx, "[core.DisableUser::OK [target_user_id=%s", targetUserID)
	return nil
}

// EnableUser enables a user account
func EnableUser(ctx context.Context, adminID string, targetUserID string) error {
	commons.Log(ctx, "[core.EnableUser [admin_id=%s [target_user_id=%s", adminID, targetUserID)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return commons.Forbidden("admin-access-required")
	}

	// Get target user
	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return commons.BadInput("invalid-user-id")
	}

	targetUser, err := userdao.FindByID(ctx, targetUUID)
	if err != nil {
		commons.LogE(ctx, "[core.EnableUser::FindErr [err=%v", err)
		return err
	}

	// Enable user
	err = userdao.UpdateEnabled(ctx, targetUUID, true)
	if err != nil {
		commons.LogE(ctx, "[core.EnableUser::UpdateErr [err=%v", err)
		return err
	}

	// Log activity
	LogActivity(ctx, adminID, model.ActionUserEnabled, "user", targetUserID, map[string]interface{}{
		"admin_name":  admin.Name,
		"target_name": targetUser.Name,
	})

	commons.Log(ctx, "[core.EnableUser::OK [target_user_id=%s", targetUserID)
	return nil
}

// ChangeUserRole changes a user's role
func ChangeUserRole(ctx context.Context, adminID string, targetUserID string, newRole string) error {
	commons.Log(ctx, "[core.ChangeUserRole [admin_id=%s [target_user_id=%s [new_role=%s", adminID, targetUserID, newRole)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return commons.Forbidden("admin-access-required")
	}

	// Prevent self role change
	if adminID == targetUserID {
		return commons.BadInput("cannot-change-own-role")
	}

	// Validate new role
	if newRole != "user" && newRole != "admin" {
		return commons.BadInput("invalid-role")
	}

	// Get target user
	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return commons.BadInput("invalid-user-id")
	}

	targetUser, err := userdao.FindByID(ctx, targetUUID)
	if err != nil {
		commons.LogE(ctx, "[core.ChangeUserRole::FindErr [err=%v", err)
		return err
	}

	oldRole := targetUser.Role

	// Update role
	err = userdao.UpdateRole(ctx, targetUUID, newRole)
	if err != nil {
		commons.LogE(ctx, "[core.ChangeUserRole::UpdateErr [err=%v", err)
		return err
	}

	// Log activity
	LogActivity(ctx, adminID, model.ActionUserRoleChanged, "user", targetUserID, map[string]interface{}{
		"old_role":    oldRole,
		"new_role":    newRole,
		"admin_name":  admin.Name,
		"target_name": targetUser.Name,
	})

	commons.Log(ctx, "[core.ChangeUserRole::OK [target_user_id=%s [new_role=%s", targetUserID, newRole)
	return nil
}

// DeleteUser permanently deletes a user
func DeleteUser(ctx context.Context, adminID string, targetUserID string, permanent bool) error {
	commons.Log(ctx, "[core.DeleteUser [admin_id=%s [target_user_id=%s [permanent=%v", adminID, targetUserID, permanent)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return commons.Forbidden("admin-access-required")
	}

	// Prevent self-delete
	if adminID == targetUserID {
		return commons.BadInput("cannot-delete-self")
	}

	// Permanent delete requires explicit confirmation
	if !permanent {
		return commons.BadInput("permanent-flag-required")
	}

	// Get target user
	targetUUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return commons.BadInput("invalid-user-id")
	}

	targetUser, err := userdao.FindByID(ctx, targetUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteUser::FindErr [err=%v", err)
		return err
	}

	// Delete user (hard delete)
	err = userdao.HardDelete(ctx, targetUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteUser::DeleteErr [err=%v", err)
		return err
	}

	// Log activity (use sync version since user is being deleted)
	LogActivitySync(ctx, adminID, model.ActionUserDeleted, "user", targetUserID, map[string]interface{}{
		"admin_name":  admin.Name,
		"target_name": targetUser.Name,
		"target_email": targetUser.Email,
		"permanent":   permanent,
	})

	commons.Log(ctx, "[core.DeleteUser::OK [target_user_id=%s", targetUserID)
	return nil
}

// isAdmin checks if the user is an admin
func isAdmin(ctx context.Context, userID string) bool {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false
	}

	user, err := userdao.FindByID(ctx, userUUID)
	if err != nil {
		return false
	}

	return user.Role == "admin"
}
