package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/educator/model"
)

// Follow creates a new subscription (follow relationship)
func Follow(ctx context.Context, subscriberID, educatorID uuid.UUID) error {
	commons.Log(ctx, "[dao.Follow [subscriber_id=%s [educator_id=%s", subscriberID, educatorID)

	// Check if already following
	exists, _ := IsFollowing(ctx, subscriberID, educatorID)
	if exists {
		return commons.Conflict("already-following")
	}

	qry := `INSERT INTO subscriptions (id, subscriber_id, educator_id, subscribed_at)
			VALUES ($1, $2, $3, NOW())`

	_, err := db.ExecContext(ctx, qry, uuid.New(), subscriberID, educatorID)
	if err != nil {
		commons.LogE(ctx, "[dao.Follow::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.Follow::OK [subscriber_id=%s [educator_id=%s", subscriberID, educatorID)
	return nil
}

// Unfollow removes a subscription
func Unfollow(ctx context.Context, subscriberID, educatorID uuid.UUID) error {
	commons.Log(ctx, "[dao.Unfollow [subscriber_id=%s [educator_id=%s", subscriberID, educatorID)

	qry := `DELETE FROM subscriptions
			WHERE subscriber_id = $1 AND educator_id = $2`

	result, err := db.ExecContext(ctx, qry, subscriberID, educatorID)
	if err != nil {
		commons.LogE(ctx, "[dao.Unfollow::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("subscription-not-found")
	}

	commons.Log(ctx, "[dao.Unfollow::OK [subscriber_id=%s [educator_id=%s", subscriberID, educatorID)
	return nil
}

// IsFollowing checks if subscriber is following educator
func IsFollowing(ctx context.Context, subscriberID, educatorID uuid.UUID) (bool, error) {
	var exists bool
	qry := `SELECT EXISTS(
				SELECT 1 FROM subscriptions
				WHERE subscriber_id = $1 AND educator_id = $2
			)`

	err := db.QueryRowContext(ctx, qry, subscriberID, educatorID).Scan(&exists)
	if err != nil {
		commons.LogE(ctx, "[dao.IsFollowing::SQLErr [err=%v", err)
		return false, err
	}

	return exists, nil
}

// GetFollowers gets all followers of an educator
func GetFollowers(ctx context.Context, educatorID uuid.UUID, limit, offset int) ([]model.FollowerInfo, int, error) {
	commons.Log(ctx, "[dao.GetFollowers [educator_id=%s", educatorID)

	// Get total count
	var total int
	qry := `SELECT COUNT(*) FROM subscriptions WHERE educator_id = $1`
	err := db.QueryRowContext(ctx, qry, educatorID).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.GetFollowers::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Get followers
	followers := []model.FollowerInfo{}
	qry = `SELECT u.id, u.username, u.full_name, u.avatar_url, s.subscribed_at
			FROM subscriptions s
			JOIN users u ON s.subscriber_id = u.id
			WHERE s.educator_id = $1
			ORDER BY s.subscribed_at DESC
			LIMIT $2 OFFSET $3`

	rows, err := db.QueryContext(ctx, qry, educatorID, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.GetFollowers::SQLErr [err=%v", err)
		return followers, total, err
	}
	defer rows.Close()

	for rows.Next() {
		var follower model.FollowerInfo
		var fullName, avatarURL sql.NullString
		var subscribedAt time.Time

		err := rows.Scan(
			&follower.UserID,
			&follower.Username,
			&fullName,
			&avatarURL,
			&subscribedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetFollowers::ScanErr [err=%v", err)
			continue
		}

		if fullName.Valid {
			follower.FullName = fullName.String
		}
		if avatarURL.Valid {
			follower.AvatarURL = avatarURL.String
		}
		follower.SubscribedAt = subscribedAt.Format(time.RFC3339)

		followers = append(followers, follower)
	}

	commons.Log(ctx, "[dao.GetFollowers::OK [count=%d [total=%d", len(followers), total)
	return followers, total, nil
}

// GetFollowing gets all educators that a user is following
func GetFollowing(ctx context.Context, subscriberID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	commons.Log(ctx, "[dao.GetFollowing [subscriber_id=%s", subscriberID)

	// Get total count
	var total int
	qry := `SELECT COUNT(*) FROM subscriptions WHERE subscriber_id = $1`
	err := db.QueryRowContext(ctx, qry, subscriberID).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.GetFollowing::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Get followed educator IDs
	educatorIDs := []uuid.UUID{}
	qry = `SELECT educator_id
			FROM subscriptions
			WHERE subscriber_id = $1
			ORDER BY subscribed_at DESC
			LIMIT $2 OFFSET $3`

	rows, err := db.QueryContext(ctx, qry, subscriberID, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.GetFollowing::SQLErr [err=%v", err)
		return educatorIDs, total, err
	}
	defer rows.Close()

	for rows.Next() {
		var educatorID uuid.UUID
		if err := rows.Scan(&educatorID); err != nil {
			commons.LogE(ctx, "[dao.GetFollowing::ScanErr [err=%v", err)
			continue
		}
		educatorIDs = append(educatorIDs, educatorID)
	}

	commons.Log(ctx, "[dao.GetFollowing::OK [count=%d [total=%d", len(educatorIDs), total)
	return educatorIDs, total, nil
}

// GetFollowerStats gets follower statistics for an educator
func GetFollowerStats(ctx context.Context, educatorID uuid.UUID) (model.FollowerStatsResponse, error) {
	commons.Log(ctx, "[dao.GetFollowerStats [educator_id=%s", educatorID)

	stats := model.FollowerStatsResponse{}

	// Total followers
	qry := `SELECT COUNT(*) FROM subscriptions WHERE educator_id = $1`
	err := db.QueryRowContext(ctx, qry, educatorID).Scan(&stats.TotalFollowers)
	if err != nil {
		commons.LogE(ctx, "[dao.GetFollowerStats::TotalErr [err=%v", err)
		return stats, err
	}

	// New followers today
	qry = `SELECT COUNT(*) FROM subscriptions
			WHERE educator_id = $1
			AND subscribed_at >= CURRENT_DATE`
	err = db.QueryRowContext(ctx, qry, educatorID).Scan(&stats.NewFollowersToday)
	if err != nil {
		stats.NewFollowersToday = 0
	}

	// New followers this week
	qry = `SELECT COUNT(*) FROM subscriptions
			WHERE educator_id = $1
			AND subscribed_at >= DATE_TRUNC('week', CURRENT_DATE)`
	err = db.QueryRowContext(ctx, qry, educatorID).Scan(&stats.NewFollowersWeek)
	if err != nil {
		stats.NewFollowersWeek = 0
	}

	// New followers this month
	qry = `SELECT COUNT(*) FROM subscriptions
			WHERE educator_id = $1
			AND subscribed_at >= DATE_TRUNC('month', CURRENT_DATE)`
	err = db.QueryRowContext(ctx, qry, educatorID).Scan(&stats.NewFollowersMonth)
	if err != nil {
		stats.NewFollowersMonth = 0
	}

	// Growth trend (last 30 days)
	qry = `SELECT DATE(subscribed_at) as date, COUNT(*) as count
			FROM subscriptions
			WHERE educator_id = $1
			AND subscribed_at >= CURRENT_DATE - INTERVAL '30 days'
			GROUP BY DATE(subscribed_at)
			ORDER BY date DESC`

	rows, err := db.QueryContext(ctx, qry, educatorID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var growth model.DailyGrowth
			var date time.Time
			if err := rows.Scan(&date, &growth.Count); err == nil {
				growth.Date = date.Format("2006-01-02")
				stats.GrowthTrend = append(stats.GrowthTrend, growth)
			}
		}
	}

	commons.Log(ctx, "[dao.GetFollowerStats::OK [educator_id=%s [total=%d", educatorID, stats.TotalFollowers)
	return stats, nil
}

// GetSubscribers gets all subscriber IDs for an educator (for notifications)
func GetSubscribers(ctx context.Context, educatorID uuid.UUID) ([]uuid.UUID, error) {
	commons.Log(ctx, "[dao.GetSubscribers [educator_id=%s", educatorID)

	subscriberIDs := []uuid.UUID{}
	qry := `SELECT subscriber_id
			FROM subscriptions
			WHERE educator_id = $1 AND notification_enabled = true`

	rows, err := db.QueryContext(ctx, qry, educatorID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetSubscribers::SQLErr [err=%v", err)
		return subscriberIDs, err
	}
	defer rows.Close()

	for rows.Next() {
		var subscriberID uuid.UUID
		if err := rows.Scan(&subscriberID); err != nil {
			commons.LogE(ctx, "[dao.GetSubscribers::ScanErr [err=%v", err)
			continue
		}
		subscriberIDs = append(subscriberIDs, subscriberID)
	}

	commons.Log(ctx, "[dao.GetSubscribers::OK [count=%d", len(subscriberIDs))
	return subscriberIDs, nil
}
