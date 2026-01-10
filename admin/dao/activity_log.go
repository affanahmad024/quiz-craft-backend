package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/google/uuid"
)

// CreateActivityLog creates a new activity log entry
func CreateActivityLog(ctx context.Context, log *model.ActivityLog) error {
	commons.Log(ctx, "[dao.CreateActivityLog [action=%s", log.Action)

	db := database.GetSQLDB()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	detailsJSON, err := json.Marshal(log.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}

	query := `
		INSERT INTO activity_logs (
			id, user_id, action, resource_type, resource_id,
			details, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING created_at
	`

	var userID *uuid.UUID
	if log.UserID != nil {
		uid, err := uuid.Parse(*log.UserID)
		if err == nil {
			userID = &uid
		}
	}

	var resourceID *uuid.UUID
	if log.ResourceID != nil {
		rid, err := uuid.Parse(*log.ResourceID)
		if err == nil {
			resourceID = &rid
		}
	}

	logID := uuid.New()
	err = db.QueryRowContext(
		ctx,
		query,
		logID,
		userID,
		log.Action,
		log.ResourceType,
		resourceID,
		detailsJSON,
		log.IPAddress,
		log.UserAgent,
	).Scan(&log.CreatedAt)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateActivityLog::QueryErr [err=%v", err)
		return err
	}

	log.ID = logID.String()
	return nil
}

// GetActivityLogs retrieves activity logs with filters and pagination
func GetActivityLogs(ctx context.Context, filter model.ActivityLogFilter) ([]model.ActivityLog, int, error) {
	commons.Log(ctx, "[dao.GetActivityLogs [page=%d [limit=%d", filter.Page, filter.Limit)

	db := database.GetSQLDB()
	if db == nil {
		return nil, 0, fmt.Errorf("database connection is nil")
	}

	// Build WHERE clause
	where := []string{}
	args := []interface{}{}
	argCount := 0

	if filter.UserID != nil {
		argCount++
		where = append(where, fmt.Sprintf("user_id = $%d", argCount))
		userUUID, _ := uuid.Parse(*filter.UserID)
		args = append(args, userUUID)
	}

	if filter.Action != nil {
		argCount++
		where = append(where, fmt.Sprintf("action = $%d", argCount))
		args = append(args, *filter.Action)
	}

	if filter.ResourceType != nil {
		argCount++
		where = append(where, fmt.Sprintf("resource_type = $%d", argCount))
		args = append(args, *filter.ResourceType)
	}

	if filter.StartDate != nil {
		argCount++
		where = append(where, fmt.Sprintf("created_at >= $%d", argCount))
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		argCount++
		where = append(where, fmt.Sprintf("created_at <= $%d", argCount))
		args = append(args, *filter.EndDate)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activity_logs %s", whereClause)
	var total int
	err := db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.GetActivityLogs::CountErr [err=%v", err)
		return nil, 0, err
	}

	// Get logs
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := fmt.Sprintf(`
		SELECT
			id, user_id, action, resource_type, resource_id,
			details, ip_address, user_agent, created_at
		FROM activity_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitArg, offsetArg)

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.GetActivityLogs::QueryErr [err=%v", err)
		return nil, 0, err
	}
	defer rows.Close()

	logs := []model.ActivityLog{}
	for rows.Next() {
		var log model.ActivityLog
		var userID, resourceID sql.NullString
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.Action,
			&log.ResourceType,
			&resourceID,
			&detailsJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			commons.LogW(ctx, "[dao.GetActivityLogs::ScanErr [err=%v", err)
			continue
		}

		if userID.Valid {
			log.UserID = &userID.String
		}
		if resourceID.Valid {
			log.ResourceID = &resourceID.String
		}

		// Unmarshal details
		if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
			log.Details = make(map[string]interface{})
		}

		logs = append(logs, log)
	}

	commons.Log(ctx, "[dao.GetActivityLogs::OK [count=%d [total=%d", len(logs), total)
	return logs, total, nil
}

// CountActivityLogsByAction counts activity logs by action type
func CountActivityLogsByAction(ctx context.Context, action string) (int, error) {
	db := database.GetSQLDB()
	if db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	query := "SELECT COUNT(*) FROM activity_logs WHERE action = $1"
	var count int
	err := db.QueryRowContext(ctx, query, action).Scan(&count)
	return count, err
}
