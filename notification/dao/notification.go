package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/notification/model"
)

var db *sql.DB

// Init initializes the DAO with database connection
func Init(database *sql.DB) {
	db = database
}

// Create creates a single notification
func Create(ctx context.Context, userID uuid.UUID, notifType, title, message string, data map[string]interface{}) (model.NotificationDAO, error) {
	commons.Log(ctx, "[dao.Create [user_id=%s [type=%s", userID, notifType)

	notification := model.NotificationDAO{}
	dataJSON, _ := json.Marshal(data)

	qry := `INSERT INTO notifications (user_id, type, title, message, data, is_read, created_at)
			VALUES ($1, $2, $3, $4, $5, false, NOW())
			RETURNING id, user_id, type, title, message, data, is_read, read_at, created_at`

	err := db.QueryRowContext(ctx, qry, userID, notifType, title, message, dataJSON).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.Create::SQLErr [user_id=%s [err=%v", userID, err)
		return notification, err
	}

	commons.Log(ctx, "[dao.Create::OK [notification_id=%s", notification.ID)
	return notification, nil
}

// CreateBulk creates notifications for multiple users
func CreateBulk(ctx context.Context, userIDs []uuid.UUID, notifType, title, message string, data map[string]interface{}) error {
	commons.Log(ctx, "[dao.CreateBulk [type=%s [user_count=%d", notifType, len(userIDs))

	if len(userIDs) == 0 {
		return nil
	}

	dataJSON, _ := json.Marshal(data)

	qry := `INSERT INTO notifications (user_id, type, title, message, data, is_read, created_at)
			SELECT unnest($1::uuid[]), $2, $3, $4, $5, false, NOW()`

	_, err := db.ExecContext(ctx, qry, pq.Array(userIDs), notifType, title, message, dataJSON)
	if err != nil {
		commons.LogE(ctx, "[dao.CreateBulk::SQLErr [type=%s [err=%v", notifType, err)
		return err
	}

	commons.Log(ctx, "[dao.CreateBulk::OK [user_count=%d", len(userIDs))
	return nil
}

// FindByID finds notification by ID
func FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (model.NotificationDAO, error) {
	commons.Log(ctx, "[dao.FindByID [id=%s [user_id=%s", id, userID)

	notification := model.NotificationDAO{}
	qry := `SELECT id, user_id, type, title, message, data, is_read, read_at, created_at
			FROM notifications
			WHERE id = $1 AND user_id = $2`

	err := db.QueryRowContext(ctx, qry, id, userID).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return notification, commons.NotFound("notification-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindByID::SQLErr [id=%s [err=%v", id, err)
		return notification, err
	}

	return notification, nil
}

// ListByUser lists notifications for a user with filters
func ListByUser(ctx context.Context, filter model.NotificationListFilter) ([]model.NotificationDAO, int, error) {
	commons.Log(ctx, "[dao.ListByUser [user_id=%s [unread_only=%v", filter.UserID, filter.UnreadOnly)

	notifications := []model.NotificationDAO{}

	// Build WHERE clause
	where := "WHERE user_id = $1"
	args := []interface{}{filter.UserID}
	argCount := 2

	if filter.UnreadOnly {
		where += " AND is_read = false"
	}

	if filter.Type != nil {
		where += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, *filter.Type)
		argCount++
	}

	// Get total count
	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByUser::CountErr [err=%v", err)
		return notifications, 0, err
	}

	// Get paginated results
	qry := fmt.Sprintf(`SELECT id, user_id, type, title, message, data, is_read, read_at, created_at
						FROM notifications %s
						ORDER BY created_at DESC
						LIMIT $%d OFFSET $%d`, where, argCount, argCount+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListByUser::SQLErr [err=%v", err)
		return notifications, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var notification model.NotificationDAO
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.Data,
			&notification.IsRead,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListByUser::ScanErr [err=%v", err)
			continue
		}
		notifications = append(notifications, notification)
	}

	commons.Log(ctx, "[dao.ListByUser::OK [total=%d [returned=%d", total, len(notifications))
	return notifications, total, nil
}

// MarkAsRead marks a notification as read
func MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.MarkAsRead [id=%s [user_id=%s", id, userID)

	qry := `UPDATE notifications
			SET is_read = true, read_at = NOW()
			WHERE id = $1 AND user_id = $2 AND is_read = false`

	result, err := db.ExecContext(ctx, qry, id, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.MarkAsRead::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("notification-not-found-or-already-read")
	}

	commons.Log(ctx, "[dao.MarkAsRead::OK [id=%s", id)
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func MarkAllAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	commons.Log(ctx, "[dao.MarkAllAsRead [user_id=%s", userID)

	qry := `UPDATE notifications
			SET is_read = true, read_at = NOW()
			WHERE user_id = $1 AND is_read = false`

	result, err := db.ExecContext(ctx, qry, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.MarkAllAsRead::SQLErr [user_id=%s [err=%v", userID, err)
		return 0, err
	}

	rows, _ := result.RowsAffected()
	commons.Log(ctx, "[dao.MarkAllAsRead::OK [user_id=%s [count=%d", userID, rows)
	return rows, nil
}

// Delete deletes a notification
func Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	commons.Log(ctx, "[dao.Delete [id=%s [user_id=%s", id, userID)

	qry := `DELETE FROM notifications WHERE id = $1 AND user_id = $2`

	result, err := db.ExecContext(ctx, qry, id, userID)
	if err != nil {
		commons.LogE(ctx, "[dao.Delete::SQLErr [id=%s [err=%v", id, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("notification-not-found")
	}

	commons.Log(ctx, "[dao.Delete::OK [id=%s", id)
	return nil
}

// GetUnreadCount gets the count of unread notifications for a user
func GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	err := db.QueryRowContext(ctx, qry, userID).Scan(&count)
	if err != nil {
		commons.LogE(ctx, "[dao.GetUnreadCount::SQLErr [user_id=%s [err=%v", userID, err)
		return 0, err
	}

	return count, nil
}

// DeleteOldNotifications deletes notifications older than a certain duration (cleanup job)
func DeleteOldNotifications(ctx context.Context, daysOld int) (int64, error) {
	commons.Log(ctx, "[dao.DeleteOldNotifications [days_old=%d", daysOld)

	qry := `DELETE FROM notifications
			WHERE created_at < NOW() - INTERVAL '1 day' * $1
			AND is_read = true`

	result, err := db.ExecContext(ctx, qry, daysOld)
	if err != nil {
		commons.LogE(ctx, "[dao.DeleteOldNotifications::SQLErr [err=%v", err)
		return 0, err
	}

	rows, _ := result.RowsAffected()
	commons.Log(ctx, "[dao.DeleteOldNotifications::OK [deleted=%d", rows)
	return rows, nil
}
