package model

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionDAO represents subscription data from database
type SubscriptionDAO struct {
	ID                  uuid.UUID
	SubscriberID        uuid.UUID
	EducatorID          uuid.UUID
	SubscribedAt        time.Time
	NotificationEnabled bool
}

// SubscriptionAPI represents subscription data for API responses
type SubscriptionAPI struct {
	ID                  uuid.UUID `json:"id"`
	SubscriberID        uuid.UUID `json:"subscriber_id"`
	EducatorID          uuid.UUID `json:"educator_id"`
	SubscribedAt        string    `json:"subscribed_at"`
	NotificationEnabled bool      `json:"notification_enabled"`
}

// ToAPI converts SubscriptionDAO to SubscriptionAPI
func (s *SubscriptionDAO) ToAPI() SubscriptionAPI {
	return SubscriptionAPI{
		ID:                  s.ID,
		SubscriberID:        s.SubscriberID,
		EducatorID:          s.EducatorID,
		SubscribedAt:        s.SubscribedAt.Format(time.RFC3339),
		NotificationEnabled: s.NotificationEnabled,
	}
}

// FollowResponse represents response after following an educator
type FollowResponse struct {
	Message  string `json:"message"`
	Followed bool   `json:"followed"`
}

// FollowerInfo represents follower information
type FollowerInfo struct {
	UserID       uuid.UUID `json:"user_id"`
	Username     string    `json:"username"`
	FullName     string    `json:"full_name,omitempty"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	SubscribedAt string    `json:"subscribed_at"`
}

// FollowerListResponse represents a list of followers
type FollowerListResponse struct {
	Followers []FollowerInfo `json:"followers"`
	Total     int            `json:"total"`
}

// FollowingListResponse represents a list of followed educators
type FollowingListResponse struct {
	Following []EducatorWithUserInfo `json:"following"`
	Total     int                    `json:"total"`
}

// FollowerStatsResponse represents follower statistics
type FollowerStatsResponse struct {
	TotalFollowers     int            `json:"total_followers"`
	NewFollowersToday  int            `json:"new_followers_today"`
	NewFollowersWeek   int            `json:"new_followers_week"`
	NewFollowersMonth  int            `json:"new_followers_month"`
	GrowthTrend        []DailyGrowth  `json:"growth_trend,omitempty"`
}

// DailyGrowth represents daily follower growth
type DailyGrowth struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
