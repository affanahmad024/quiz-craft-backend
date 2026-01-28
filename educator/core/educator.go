package core

import (
	"context"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/educator/dao"
	"github.com/affanahmad024/quiz-craft/backend/educator/model"
	notificationcore "github.com/affanahmad024/quiz-craft/backend/notification/core"
	notificationmodel "github.com/affanahmad024/quiz-craft/backend/notification/model"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
)

// PromoteToEducator promotes a user to educator role
func PromoteToEducator(ctx context.Context, userID uuid.UUID, req model.PromoteToEducatorRequest) (model.EducatorProfileAPI, error) {
	commons.Log(ctx, "[core.PromoteToEducator [user_id=%s", userID)

	// Get user details
	user, err := userdao.FindByID(ctx, userID)
	if err != nil {
		return model.EducatorProfileAPI{}, commons.NotFound("user-not-found")
	}

	// Check if already an educator
	if user.Role == "educator" {
		return model.EducatorProfileAPI{}, commons.Conflict("already-educator")
	}

	// Validate requirements
	if !user.EmailVerified {
		return model.EducatorProfileAPI{}, commons.BadInput("email-not-verified")
	}

	if !user.PhoneNumber.Valid || user.PhoneNumber.String == "" {
		return model.EducatorProfileAPI{}, commons.BadInput("phone-number-required")
	}

	// Update user role to educator
	err = userdao.UpdateRole(ctx, userID, "educator")
	if err != nil {
		commons.LogE(ctx, "[core.PromoteToEducator::UpdateRoleErr [err=%v", err)
		return model.EducatorProfileAPI{}, err
	}

	// Create educator profile
	profileReq := model.CreateEducatorProfileRequest{
		Bio:            req.Bio,
		Tagline:        req.Tagline,
		ExpertiseAreas: req.ExpertiseAreas,
	}

	profile, err := dao.CreateProfile(ctx, userID, profileReq)
	if err != nil {
		commons.LogE(ctx, "[core.PromoteToEducator::CreateProfileErr [err=%v", err)
		return model.EducatorProfileAPI{}, err
	}

	commons.Log(ctx, "[core.PromoteToEducator::OK [user_id=%s [profile_id=%s", userID, profile.ID)
	return profile.ToAPI(), nil
}

// GetEducatorProfile gets educator profile by user ID
func GetEducatorProfile(ctx context.Context, userID uuid.UUID) (model.EducatorProfileAPI, error) {
	commons.Log(ctx, "[core.GetEducatorProfile [user_id=%s", userID)

	profile, err := dao.FindByUserID(ctx, userID)
	if err != nil {
		return model.EducatorProfileAPI{}, err
	}

	return profile.ToAPI(), nil
}

// UpdateEducatorProfile updates educator profile
func UpdateEducatorProfile(ctx context.Context, userID uuid.UUID, req model.UpdateEducatorProfileRequest) error {
	commons.Log(ctx, "[core.UpdateEducatorProfile [user_id=%s", userID)

	// Verify educator profile exists
	_, err := dao.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return dao.UpdateProfile(ctx, userID, req)
}

// ListEducators lists all educators with user information
func ListEducators(ctx context.Context, limit, offset int) (model.EducatorListResponse, error) {
	commons.Log(ctx, "[core.ListEducators")

	profiles, total, err := dao.ListEducators(ctx, limit, offset)
	if err != nil {
		return model.EducatorListResponse{}, err
	}

	// Get user info for each educator
	educators := make([]model.EducatorWithUserInfo, 0, len(profiles))
	for _, profile := range profiles {
		user, err := userdao.FindByID(ctx, profile.UserID)
		if err != nil {
			continue
		}

		educator := model.EducatorWithUserInfo{
			UserID:   user.ID,
			Username: user.Name,
			Email:    user.Email,
			Profile:  profile.ToAPI(),
		}

		educators = append(educators, educator)
	}

	return model.EducatorListResponse{
		Educators: educators,
		Total:     total,
	}, nil
}

// GetEducatorWithUserInfo gets educator profile with user information
func GetEducatorWithUserInfo(ctx context.Context, userID uuid.UUID) (model.EducatorWithUserInfo, error) {
	commons.Log(ctx, "[core.GetEducatorWithUserInfo [user_id=%s", userID)

	profile, err := dao.FindByUserID(ctx, userID)
	if err != nil {
		return model.EducatorWithUserInfo{}, err
	}

	user, err := userdao.FindByID(ctx, userID)
	if err != nil {
		return model.EducatorWithUserInfo{}, err
	}

	educator := model.EducatorWithUserInfo{
		UserID:   user.ID,
		Username: user.Name,
		Email:    user.Email,
		Profile:  profile.ToAPI(),
	}

	return educator, nil
}

// SearchEducators searches for educators
func SearchEducators(ctx context.Context, query string, limit, offset int) (model.EducatorListResponse, error) {
	commons.Log(ctx, "[core.SearchEducators [query=%s", query)

	profiles, total, err := dao.SearchEducators(ctx, query, limit, offset)
	if err != nil {
		return model.EducatorListResponse{}, err
	}

	// Get user info for each educator
	educators := make([]model.EducatorWithUserInfo, 0, len(profiles))
	for _, profile := range profiles {
		user, err := userdao.FindByID(ctx, profile.UserID)
		if err != nil {
			continue
		}

		educator := model.EducatorWithUserInfo{
			UserID:   user.ID,
			Username: user.Name,
			Email:    user.Email,
			Profile:  profile.ToAPI(),
		}

		educators = append(educators, educator)
	}

	return model.EducatorListResponse{
		Educators: educators,
		Total:     total,
	}, nil
}

// GetFeaturedEducators gets featured educators
func GetFeaturedEducators(ctx context.Context, limit int) ([]model.EducatorWithUserInfo, error) {
	commons.Log(ctx, "[core.GetFeaturedEducators")

	profiles, err := dao.GetFeaturedEducators(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Get user info for each educator
	educators := make([]model.EducatorWithUserInfo, 0, len(profiles))
	for _, profile := range profiles {
		user, err := userdao.FindByID(ctx, profile.UserID)
		if err != nil {
			continue
		}

		educator := model.EducatorWithUserInfo{
			UserID:   user.ID,
			Username: user.Name,
			Email:    user.Email,
			Profile:  profile.ToAPI(),
		}

		educators = append(educators, educator)
	}

	return educators, nil
}

// DemoteEducator demotes educator to user (admin only)
func DemoteEducator(ctx context.Context, userID uuid.UUID) error {
	commons.Log(ctx, "[core.DemoteEducator [user_id=%s", userID)

	// Verify educator profile exists
	_, err := dao.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Update user role to user
	err = userdao.UpdateRole(ctx, userID, "user")
	if err != nil {
		commons.LogE(ctx, "[core.DemoteEducator::UpdateRoleErr [err=%v", err)
		return err
	}

	// Delete educator profile
	err = dao.DeleteProfile(ctx, userID)
	if err != nil {
		commons.LogE(ctx, "[core.DemoteEducator::DeleteProfileErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[core.DemoteEducator::OK [user_id=%s", userID)
	return nil
}

// FollowEducator follows an educator
func FollowEducator(ctx context.Context, subscriberID, educatorID uuid.UUID, notificationService *notificationcore.NotificationService) (model.FollowResponse, error) {
	commons.Log(ctx, "[core.FollowEducator [subscriber_id=%s [educator_id=%s", subscriberID, educatorID)

	// Can't follow yourself
	if subscriberID == educatorID {
		return model.FollowResponse{}, commons.BadInput("cannot-follow-yourself")
	}

	// Verify educator exists and has educator role
	educator, err := userdao.FindByID(ctx, educatorID)
	if err != nil {
		return model.FollowResponse{}, commons.NotFound("educator-not-found")
	}

	if educator.Role != "educator" {
		return model.FollowResponse{}, commons.BadInput("user-not-educator")
	}

	// Follow
	err = dao.Follow(ctx, subscriberID, educatorID)
	if err != nil {
		return model.FollowResponse{}, err
	}

	// Send notification to educator about new follower
	if notificationService != nil {
		subscriber, _ := userdao.FindByID(ctx, subscriberID)
		username := "A user"
		if subscriber.Email != "" {
			username = subscriber.Name
		}

		data := map[string]interface{}{
			"subscriber_id": subscriberID.String(),
		}
		notificationService.CreateNotification(ctx, educatorID,
			notificationmodel.NotificationTypeNewFollower,
			"New Follower",
			username+" started following you",
			data)
	}

	return model.FollowResponse{
		Message:  "Successfully followed educator",
		Followed: true,
	}, nil
}

// UnfollowEducator unfollows an educator
func UnfollowEducator(ctx context.Context, subscriberID, educatorID uuid.UUID) (model.FollowResponse, error) {
	commons.Log(ctx, "[core.UnfollowEducator [subscriber_id=%s [educator_id=%s", subscriberID, educatorID)

	err := dao.Unfollow(ctx, subscriberID, educatorID)
	if err != nil {
		return model.FollowResponse{}, err
	}

	return model.FollowResponse{
		Message:  "Successfully unfollowed educator",
		Followed: false,
	}, nil
}

// GetFollowers gets educator's followers
func GetFollowers(ctx context.Context, educatorID uuid.UUID, limit, offset int) (model.FollowerListResponse, error) {
	commons.Log(ctx, "[core.GetFollowers [educator_id=%s", educatorID)

	followers, total, err := dao.GetFollowers(ctx, educatorID, limit, offset)
	if err != nil {
		return model.FollowerListResponse{}, err
	}

	return model.FollowerListResponse{
		Followers: followers,
		Total:     total,
	}, nil
}

// GetFollowing gets user's followed educators
func GetFollowing(ctx context.Context, subscriberID uuid.UUID, limit, offset int) (model.FollowingListResponse, error) {
	commons.Log(ctx, "[core.GetFollowing [subscriber_id=%s", subscriberID)

	educatorIDs, total, err := dao.GetFollowing(ctx, subscriberID, limit, offset)
	if err != nil {
		return model.FollowingListResponse{}, err
	}

	// Get educator profiles with user info
	following := make([]model.EducatorWithUserInfo, 0, len(educatorIDs))
	for _, educatorID := range educatorIDs {
		educator, err := GetEducatorWithUserInfo(ctx, educatorID)
		if err != nil {
			continue
		}
		following = append(following, educator)
	}

	return model.FollowingListResponse{
		Following: following,
		Total:     total,
	}, nil
}

// GetFollowerStats gets follower statistics
func GetFollowerStats(ctx context.Context, educatorID uuid.UUID) (model.FollowerStatsResponse, error) {
	commons.Log(ctx, "[core.GetFollowerStats [educator_id=%s", educatorID)

	return dao.GetFollowerStats(ctx, educatorID)
}

// IsFollowing checks if user is following educator
func IsFollowing(ctx context.Context, subscriberID, educatorID uuid.UUID) (bool, error) {
	return dao.IsFollowing(ctx, subscriberID, educatorID)
}

// NotifyFollowersNewQuiz sends notifications to all followers when educator publishes a quiz
func NotifyFollowersNewQuiz(ctx context.Context, educatorID uuid.UUID, quizID, quizTitle string, notificationService *notificationcore.NotificationService) error {
	commons.Log(ctx, "[core.NotifyFollowersNewQuiz [educator_id=%s [quiz_id=%s", educatorID, quizID)

	if notificationService == nil {
		return nil
	}

	// Get all subscribers
	subscriberIDs, err := dao.GetSubscribers(ctx, educatorID)
	if err != nil {
		commons.LogE(ctx, "[core.NotifyFollowersNewQuiz::GetSubscribersErr [err=%v", err)
		return err
	}

	if len(subscriberIDs) == 0 {
		return nil
	}

	// Get educator name
	educator, _ := userdao.FindByID(ctx, educatorID)
	educatorName := "An educator"
	if educator.Email != "" {
		educatorName = educator.Name
	}

	// Create notifications for all subscribers
	title := "New Quiz Published"
	message := educatorName + " published a new quiz: " + quizTitle
	data := map[string]interface{}{
		"quiz_id":     quizID,
		"educator_id": educatorID.String(),
	}

	// Bulk create notifications
	err = notificationService.CreateBulkNotifications(ctx, subscriberIDs,
		notificationmodel.NotificationTypeNewQuiz,
		title,
		message,
		data)
	if err != nil {
		commons.LogE(ctx, "[core.NotifyFollowersNewQuiz::CreateNotificationsErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[core.NotifyFollowersNewQuiz::OK [sent=%d", len(subscriberIDs))
	return nil
}

// GetUserRole gets user role (helper function)
func GetUserRole(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := userdao.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.Role, nil
}
