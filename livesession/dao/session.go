package dao

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
)

var db *sql.DB

// Init initializes the DAO with database connection
func Init(database *sql.DB) {
	db = database
}

// CreateSession creates a new live session
func CreateSession(ctx context.Context, req model.CreateLiveSessionRequest, creatorID uuid.UUID) (model.LiveSessionDAO, error) {
	commons.Log(ctx, "[dao.CreateSession [creator_id=%s [quiz_id=%s [type=%s", creatorID, req.QuizID, req.SessionType)

	session := model.LiveSessionDAO{}
	configJSON, _ := json.Marshal(req.Config)

	// Generate access code for private sessions
	accessCode := ""
	if req.SessionType == model.SessionTypePrivate {
		accessCode = generateAccessCode()
	}

	// Parse scheduled_at if provided
	var scheduledAt sql.NullTime
	if req.ScheduledAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err == nil {
			scheduledAt = sql.NullTime{Time: parsedTime, Valid: true}
		}
	}

	var descriptionSQL sql.NullString
	if req.Description != "" {
		descriptionSQL = sql.NullString{String: req.Description, Valid: true}
	}

	var maxParticipantsSQL sql.NullInt64
	if req.MaxParticipants != nil {
		maxParticipantsSQL = sql.NullInt64{Int64: int64(*req.MaxParticipants), Valid: true}
	}

	var accessCodeSQL sql.NullString
	if accessCode != "" {
		accessCodeSQL = sql.NullString{String: accessCode, Valid: true}
	}

	qry := `INSERT INTO live_sessions (quiz_id, creator_id, session_type, title, description, scheduled_at,
			status, access_code, is_discoverable, max_participants, progression_mode, question_timer_seconds,
			current_question_index, config, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 0, $13, NOW(), NOW())
			RETURNING id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
			ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
			question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at`

	err := db.QueryRowContext(ctx, qry, req.QuizID, creatorID, req.SessionType, req.Title, descriptionSQL,
		scheduledAt, model.SessionStatusWaiting, accessCodeSQL, req.IsDiscoverable, maxParticipantsSQL,
		req.ProgressionMode, req.QuestionTimerSeconds, configJSON).Scan(
		&session.ID,
		&session.QuizID,
		&session.CreatorID,
		&session.SessionType,
		&session.Title,
		&session.Description,
		&session.ScheduledAt,
		&session.StartedAt,
		&session.EndedAt,
		&session.Status,
		&session.AccessCode,
		&session.IsDiscoverable,
		&session.MaxParticipants,
		&session.ProgressionMode,
		&session.QuestionTimerSeconds,
		&session.CurrentQuestionIndex,
		&session.CurrentQuestionStartedAt,
		&session.Config,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.CreateSession::SQLErr [err=%v", err)
		return session, err
	}

	commons.Log(ctx, "[dao.CreateSession::OK [session_id=%s", session.ID)
	return session, nil
}

// FindSessionByID finds a session by ID
func FindSessionByID(ctx context.Context, sessionID uuid.UUID) (model.LiveSessionDAO, error) {
	commons.Log(ctx, "[dao.FindSessionByID [session_id=%s", sessionID)

	session := model.LiveSessionDAO{}
	qry := `SELECT id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
			ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
			question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at
			FROM live_sessions
			WHERE id = $1`

	err := db.QueryRowContext(ctx, qry, sessionID).Scan(
		&session.ID,
		&session.QuizID,
		&session.CreatorID,
		&session.SessionType,
		&session.Title,
		&session.Description,
		&session.ScheduledAt,
		&session.StartedAt,
		&session.EndedAt,
		&session.Status,
		&session.AccessCode,
		&session.IsDiscoverable,
		&session.MaxParticipants,
		&session.ProgressionMode,
		&session.QuestionTimerSeconds,
		&session.CurrentQuestionIndex,
		&session.CurrentQuestionStartedAt,
		&session.Config,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return session, commons.NotFound("session-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindSessionByID::SQLErr [session_id=%s [err=%v", sessionID, err)
		return session, err
	}

	return session, nil
}

// FindSessionByAccessCode finds a session by access code
func FindSessionByAccessCode(ctx context.Context, accessCode string) (model.LiveSessionDAO, error) {
	commons.Log(ctx, "[dao.FindSessionByAccessCode [access_code=%s", accessCode)

	session := model.LiveSessionDAO{}
	qry := `SELECT id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
			ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
			question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at
			FROM live_sessions
			WHERE access_code = $1`

	err := db.QueryRowContext(ctx, qry, accessCode).Scan(
		&session.ID,
		&session.QuizID,
		&session.CreatorID,
		&session.SessionType,
		&session.Title,
		&session.Description,
		&session.ScheduledAt,
		&session.StartedAt,
		&session.EndedAt,
		&session.Status,
		&session.AccessCode,
		&session.IsDiscoverable,
		&session.MaxParticipants,
		&session.ProgressionMode,
		&session.QuestionTimerSeconds,
		&session.CurrentQuestionIndex,
		&session.CurrentQuestionStartedAt,
		&session.Config,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return session, commons.NotFound("session-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindSessionByAccessCode::SQLErr [err=%v", err)
		return session, err
	}

	return session, nil
}

// ListSessions lists sessions with filters
func ListSessions(ctx context.Context, filter model.SessionListFilter) ([]model.LiveSessionDAO, int, error) {
	commons.Log(ctx, "[dao.ListSessions [limit=%d [offset=%d", filter.Limit, filter.Offset)

	sessions := []model.LiveSessionDAO{}

	// Build WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.CreatorID != nil {
		where += fmt.Sprintf(" AND creator_id = $%d", argCount)
		args = append(args, *filter.CreatorID)
		argCount++
	}

	if filter.SessionType != nil {
		where += fmt.Sprintf(" AND session_type = $%d", argCount)
		args = append(args, *filter.SessionType)
		argCount++
	}

	if filter.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}

	// Get total count
	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM live_sessions %s", where)
	err := db.QueryRowContext(ctx, countQry, args...).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListSessions::CountErr [err=%v", err)
		return sessions, 0, err
	}

	// Get paginated results
	qry := fmt.Sprintf(`SELECT id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
						ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
						question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at
						FROM live_sessions %s
						ORDER BY created_at DESC
						LIMIT $%d OFFSET $%d`, where, argCount, argCount+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := db.QueryContext(ctx, qry, args...)
	if err != nil {
		commons.LogE(ctx, "[dao.ListSessions::SQLErr [err=%v", err)
		return sessions, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var session model.LiveSessionDAO
		err := rows.Scan(
			&session.ID,
			&session.QuizID,
			&session.CreatorID,
			&session.SessionType,
			&session.Title,
			&session.Description,
			&session.ScheduledAt,
			&session.StartedAt,
			&session.EndedAt,
			&session.Status,
			&session.AccessCode,
			&session.IsDiscoverable,
			&session.MaxParticipants,
			&session.ProgressionMode,
			&session.QuestionTimerSeconds,
			&session.CurrentQuestionIndex,
			&session.CurrentQuestionStartedAt,
			&session.Config,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListSessions::ScanErr [err=%v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	commons.Log(ctx, "[dao.ListSessions::OK [total=%d [returned=%d", total, len(sessions))
	return sessions, total, nil
}

// ListDiscoverableSessions lists public discoverable sessions
func ListDiscoverableSessions(ctx context.Context, limit, offset int) ([]model.LiveSessionDAO, int, error) {
	commons.Log(ctx, "[dao.ListDiscoverableSessions [limit=%d [offset=%d", limit, offset)

	sessions := []model.LiveSessionDAO{}

	where := "WHERE is_discoverable = true AND session_type = 'public' AND status IN ('waiting', 'live')"

	// Get total count
	var total int
	countQry := fmt.Sprintf("SELECT COUNT(*) FROM live_sessions %s", where)
	err := db.QueryRowContext(ctx, countQry).Scan(&total)
	if err != nil {
		commons.LogE(ctx, "[dao.ListDiscoverableSessions::CountErr [err=%v", err)
		return sessions, 0, err
	}

	// Get paginated results
	qry := fmt.Sprintf(`SELECT id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
						ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
						question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at
						FROM live_sessions %s
						ORDER BY scheduled_at ASC, created_at DESC
						LIMIT $1 OFFSET $2`, where)

	rows, err := db.QueryContext(ctx, qry, limit, offset)
	if err != nil {
		commons.LogE(ctx, "[dao.ListDiscoverableSessions::SQLErr [err=%v", err)
		return sessions, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var session model.LiveSessionDAO
		err := rows.Scan(
			&session.ID,
			&session.QuizID,
			&session.CreatorID,
			&session.SessionType,
			&session.Title,
			&session.Description,
			&session.ScheduledAt,
			&session.StartedAt,
			&session.EndedAt,
			&session.Status,
			&session.AccessCode,
			&session.IsDiscoverable,
			&session.MaxParticipants,
			&session.ProgressionMode,
			&session.QuestionTimerSeconds,
			&session.CurrentQuestionIndex,
			&session.CurrentQuestionStartedAt,
			&session.Config,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListDiscoverableSessions::ScanErr [err=%v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	commons.Log(ctx, "[dao.ListDiscoverableSessions::OK [total=%d [returned=%d", total, len(sessions))
	return sessions, total, nil
}

// UpdateSessionStatus updates the session status
func UpdateSessionStatus(ctx context.Context, sessionID uuid.UUID, status string) error {
	commons.Log(ctx, "[dao.UpdateSessionStatus [session_id=%s [status=%s", sessionID, status)

	qry := `UPDATE live_sessions
			SET status = $1, updated_at = NOW()
			WHERE id = $2`

	result, err := db.ExecContext(ctx, qry, status, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateSessionStatus::SQLErr [session_id=%s [err=%v", sessionID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("session-not-found")
	}

	commons.Log(ctx, "[dao.UpdateSessionStatus::OK [session_id=%s", sessionID)
	return nil
}

// StartSession marks session as started
func StartSession(ctx context.Context, sessionID uuid.UUID) error {
	commons.Log(ctx, "[dao.StartSession [session_id=%s", sessionID)

	qry := `UPDATE live_sessions
			SET status = $1, started_at = NOW(), current_question_started_at = NOW(), updated_at = NOW()
			WHERE id = $2 AND status = $3`

	result, err := db.ExecContext(ctx, qry, model.SessionStatusLive, sessionID, model.SessionStatusWaiting)
	if err != nil {
		commons.LogE(ctx, "[dao.StartSession::SQLErr [session_id=%s [err=%v", sessionID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("session-not-found-or-already-started")
	}

	commons.Log(ctx, "[dao.StartSession::OK [session_id=%s", sessionID)
	return nil
}

// AdvanceQuestion advances to the next question
func AdvanceQuestion(ctx context.Context, sessionID uuid.UUID) error {
	commons.Log(ctx, "[dao.AdvanceQuestion [session_id=%s", sessionID)

	qry := `UPDATE live_sessions
			SET current_question_index = current_question_index + 1,
			    current_question_started_at = NOW(),
			    updated_at = NOW()
			WHERE id = $1 AND status = $2`

	result, err := db.ExecContext(ctx, qry, sessionID, model.SessionStatusLive)
	if err != nil {
		commons.LogE(ctx, "[dao.AdvanceQuestion::SQLErr [session_id=%s [err=%v", sessionID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("session-not-found-or-not-live")
	}

	commons.Log(ctx, "[dao.AdvanceQuestion::OK [session_id=%s", sessionID)
	return nil
}

// EndSession marks session as completed
func EndSession(ctx context.Context, sessionID uuid.UUID) error {
	commons.Log(ctx, "[dao.EndSession [session_id=%s", sessionID)

	qry := `UPDATE live_sessions
			SET status = $1, ended_at = NOW(), updated_at = NOW()
			WHERE id = $2 AND status = $3`

	result, err := db.ExecContext(ctx, qry, model.SessionStatusCompleted, sessionID, model.SessionStatusLive)
	if err != nil {
		commons.LogE(ctx, "[dao.EndSession::SQLErr [session_id=%s [err=%v", sessionID, err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("session-not-found-or-not-live")
	}

	commons.Log(ctx, "[dao.EndSession::OK [session_id=%s", sessionID)
	return nil
}

// GetScheduledSessions gets sessions that should be auto-started
func GetScheduledSessions(ctx context.Context, beforeTime time.Time) ([]model.LiveSessionDAO, error) {
	commons.Log(ctx, "[dao.GetScheduledSessions [before_time=%s", beforeTime.Format(time.RFC3339))

	sessions := []model.LiveSessionDAO{}

	qry := `SELECT id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
			ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
			question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at
			FROM live_sessions
			WHERE status = $1 AND scheduled_at IS NOT NULL AND scheduled_at <= $2
			AND progression_mode = $3
			ORDER BY scheduled_at ASC
			LIMIT 50`

	rows, err := db.QueryContext(ctx, qry, model.SessionStatusWaiting, beforeTime, model.ProgressionModeAutomatic)
	if err != nil {
		commons.LogE(ctx, "[dao.GetScheduledSessions::SQLErr [err=%v", err)
		return sessions, err
	}
	defer rows.Close()

	for rows.Next() {
		var session model.LiveSessionDAO
		err := rows.Scan(
			&session.ID,
			&session.QuizID,
			&session.CreatorID,
			&session.SessionType,
			&session.Title,
			&session.Description,
			&session.ScheduledAt,
			&session.StartedAt,
			&session.EndedAt,
			&session.Status,
			&session.AccessCode,
			&session.IsDiscoverable,
			&session.MaxParticipants,
			&session.ProgressionMode,
			&session.QuestionTimerSeconds,
			&session.CurrentQuestionIndex,
			&session.CurrentQuestionStartedAt,
			&session.Config,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetScheduledSessions::ScanErr [err=%v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	commons.Log(ctx, "[dao.GetScheduledSessions::OK [found=%d", len(sessions))
	return sessions, nil
}

// GetLiveAutomaticSessions gets live sessions with automatic progression
func GetLiveAutomaticSessions(ctx context.Context) ([]model.LiveSessionDAO, error) {
	commons.Log(ctx, "[dao.GetLiveAutomaticSessions")

	sessions := []model.LiveSessionDAO{}

	qry := `SELECT id, quiz_id, creator_id, session_type, title, description, scheduled_at, started_at,
			ended_at, status, access_code, is_discoverable, max_participants, progression_mode,
			question_timer_seconds, current_question_index, current_question_started_at, config, created_at, updated_at
			FROM live_sessions
			WHERE status = $1 AND progression_mode = $2
			ORDER BY started_at ASC`

	rows, err := db.QueryContext(ctx, qry, model.SessionStatusLive, model.ProgressionModeAutomatic)
	if err != nil {
		commons.LogE(ctx, "[dao.GetLiveAutomaticSessions::SQLErr [err=%v", err)
		return sessions, err
	}
	defer rows.Close()

	for rows.Next() {
		var session model.LiveSessionDAO
		err := rows.Scan(
			&session.ID,
			&session.QuizID,
			&session.CreatorID,
			&session.SessionType,
			&session.Title,
			&session.Description,
			&session.ScheduledAt,
			&session.StartedAt,
			&session.EndedAt,
			&session.Status,
			&session.AccessCode,
			&session.IsDiscoverable,
			&session.MaxParticipants,
			&session.ProgressionMode,
			&session.QuestionTimerSeconds,
			&session.CurrentQuestionIndex,
			&session.CurrentQuestionStartedAt,
			&session.Config,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetLiveAutomaticSessions::ScanErr [err=%v", err)
			continue
		}
		sessions = append(sessions, session)
	}

	commons.Log(ctx, "[dao.GetLiveAutomaticSessions::OK [found=%d", len(sessions))
	return sessions, nil
}

// generateAccessCode generates a random 6-character access code
func generateAccessCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude similar characters
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
