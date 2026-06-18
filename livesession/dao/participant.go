package dao

import (
	"context"
	"database/sql"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
	"github.com/google/uuid"
)

// AddParticipant adds a participant to a session
func AddParticipant(ctx context.Context, sessionID, userID uuid.UUID) (model.ParticipantDAO, error) {
	commons.Log(ctx, "[dao.AddParticipant [session_id=%s [user_id=%s", sessionID, userID)

	participant := model.ParticipantDAO{}
	qry := `INSERT INTO live_session_participants (session_id, user_id, joined_at, current_score, rank, status, created_at, updated_at)
			VALUES ($1, $2, NOW(), 0, 0, $3, NOW(), NOW())
			ON CONFLICT (session_id, user_id) DO UPDATE
			SET status = $3,
			    left_at = NULL,
			    updated_at = NOW()
			RETURNING id, session_id, user_id, joined_at, left_at, current_score, rank, status, created_at, updated_at`

	err := db.QueryRowContext(ctx, qry, sessionID, userID, model.ParticipantStatusActive).Scan(
		&participant.ID,
		&participant.SessionID,
		&participant.UserID,
		&participant.JoinedAt,
		&participant.LeftAt,
		&participant.CurrentScore,
		&participant.Rank,
		&participant.Status,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err != nil {
		commons.LogE(ctx, "[dao.AddParticipant::SQLErr [session_id=%s [user_id=%s [err=%v", sessionID, userID, err)
		return participant, err
	}

	commons.Log(ctx, "[dao.AddParticipant::OK [participant_id=%s", participant.ID)
	return participant, nil
}

// FindParticipant finds a participant by session and user ID
func FindParticipant(ctx context.Context, sessionID, userID uuid.UUID) (model.ParticipantDAO, error) {
	commons.Log(ctx, "[dao.FindParticipant [session_id=%s [user_id=%s", sessionID, userID)

	participant := model.ParticipantDAO{}
	qry := `SELECT id, session_id, user_id, joined_at, left_at, current_score, rank, status, created_at, updated_at
			FROM live_session_participants
			WHERE session_id = $1 AND user_id = $2`

	err := db.QueryRowContext(ctx, qry, sessionID, userID).Scan(
		&participant.ID,
		&participant.SessionID,
		&participant.UserID,
		&participant.JoinedAt,
		&participant.LeftAt,
		&participant.CurrentScore,
		&participant.Rank,
		&participant.Status,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return participant, commons.NotFound("participant-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindParticipant::SQLErr [err=%v", err)
		return participant, err
	}

	return participant, nil
}

// FindParticipantByID finds a participant by ID
func FindParticipantByID(ctx context.Context, participantID uuid.UUID) (model.ParticipantDAO, error) {
	commons.Log(ctx, "[dao.FindParticipantByID [participant_id=%s", participantID)

	participant := model.ParticipantDAO{}
	qry := `SELECT id, session_id, user_id, joined_at, left_at, current_score, rank, status, created_at, updated_at
			FROM live_session_participants
			WHERE id = $1`

	err := db.QueryRowContext(ctx, qry, participantID).Scan(
		&participant.ID,
		&participant.SessionID,
		&participant.UserID,
		&participant.JoinedAt,
		&participant.LeftAt,
		&participant.CurrentScore,
		&participant.Rank,
		&participant.Status,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return participant, commons.NotFound("participant-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindParticipantByID::SQLErr [err=%v", err)
		return participant, err
	}

	return participant, nil
}

// ListParticipants lists all participants in a session
func ListParticipants(ctx context.Context, sessionID uuid.UUID) ([]model.ParticipantDAO, error) {
	commons.Log(ctx, "[dao.ListParticipants [session_id=%s", sessionID)

	participants := []model.ParticipantDAO{}
	qry := `SELECT id, session_id, user_id, joined_at, left_at, current_score, rank, status, created_at, updated_at
			FROM live_session_participants
			WHERE session_id = $1
			ORDER BY joined_at ASC`

	rows, err := db.QueryContext(ctx, qry, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.ListParticipants::SQLErr [err=%v", err)
		return participants, err
	}
	defer rows.Close()

	for rows.Next() {
		var participant model.ParticipantDAO
		err := rows.Scan(
			&participant.ID,
			&participant.SessionID,
			&participant.UserID,
			&participant.JoinedAt,
			&participant.LeftAt,
			&participant.CurrentScore,
			&participant.Rank,
			&participant.Status,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListParticipants::ScanErr [err=%v", err)
			continue
		}
		participants = append(participants, participant)
	}

	commons.Log(ctx, "[dao.ListParticipants::OK [session_id=%s [count=%d", sessionID, len(participants))
	return participants, nil
}

// GetLeaderboard gets the current leaderboard for a session
func GetLeaderboard(ctx context.Context, sessionID uuid.UUID) ([]model.LeaderboardEntry, error) {
	commons.Log(ctx, "[dao.GetLeaderboard [session_id=%s", sessionID)

	leaderboard := []model.LeaderboardEntry{}
	qry := `SELECT p.rank, p.user_id, u.name, p.current_score, p.status
			FROM live_session_participants p
			JOIN users u ON p.user_id = u.id
			WHERE p.session_id = $1
			ORDER BY p.current_score DESC, p.joined_at ASC`

	rows, err := db.QueryContext(ctx, qry, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetLeaderboard::SQLErr [err=%v", err)
		return leaderboard, err
	}
	defer rows.Close()

	rank := 1
	for rows.Next() {
		var entry model.LeaderboardEntry
		err := rows.Scan(
			&entry.Rank,
			&entry.UserID,
			&entry.UserName,
			&entry.CurrentScore,
			&entry.Status,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.GetLeaderboard::ScanErr [err=%v", err)
			continue
		}
		entry.Rank = rank
		leaderboard = append(leaderboard, entry)
		rank++
	}

	commons.Log(ctx, "[dao.GetLeaderboard::OK [session_id=%s [count=%d", sessionID, len(leaderboard))
	return leaderboard, nil
}

// UpdateParticipantScore updates participant's score and recalculates ranks
func UpdateParticipantScore(ctx context.Context, participantID uuid.UUID, pointsToAdd int) error {
	commons.Log(ctx, "[dao.UpdateParticipantScore [participant_id=%s [points=%d", participantID, pointsToAdd)

	qry := `UPDATE live_session_participants
			SET current_score = current_score + $1, updated_at = NOW()
			WHERE id = $2`

	result, err := db.ExecContext(ctx, qry, pointsToAdd, participantID)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateParticipantScore::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("participant-not-found")
	}

	// Get session ID to recalculate ranks
	var sessionID uuid.UUID
	err = db.QueryRowContext(ctx, "SELECT session_id FROM live_session_participants WHERE id = $1", participantID).Scan(&sessionID)
	if err == nil {
		RecalculateRanks(ctx, sessionID)
	}

	commons.Log(ctx, "[dao.UpdateParticipantScore::OK [participant_id=%s", participantID)
	return nil
}

// RecalculateRanks recalculates ranks for all participants in a session
func RecalculateRanks(ctx context.Context, sessionID uuid.UUID) error {
	commons.Log(ctx, "[dao.RecalculateRanks [session_id=%s", sessionID)

	qry := `UPDATE live_session_participants
			SET rank = subquery.new_rank
			FROM (
				SELECT id, ROW_NUMBER() OVER (ORDER BY current_score DESC, joined_at ASC) as new_rank
				FROM live_session_participants
				WHERE session_id = $1
			) AS subquery
			WHERE live_session_participants.id = subquery.id`

	_, err := db.ExecContext(ctx, qry, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.RecalculateRanks::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.RecalculateRanks::OK [session_id=%s", sessionID)
	return nil
}

// UpdateParticipantStatus updates participant status
func UpdateParticipantStatus(ctx context.Context, participantID uuid.UUID, status string) error {
	commons.Log(ctx, "[dao.UpdateParticipantStatus [participant_id=%s [status=%s", participantID, status)

	qry := `UPDATE live_session_participants
			SET status = $1, updated_at = NOW()
			WHERE id = $2`

	result, err := db.ExecContext(ctx, qry, status, participantID)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateParticipantStatus::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("participant-not-found")
	}

	commons.Log(ctx, "[dao.UpdateParticipantStatus::OK [participant_id=%s", participantID)
	return nil
}

// MarkParticipantLeft marks participant as left
func MarkParticipantLeft(ctx context.Context, participantID uuid.UUID) error {
	commons.Log(ctx, "[dao.MarkParticipantLeft [participant_id=%s", participantID)

	qry := `UPDATE live_session_participants
			SET left_at = NOW(), status = $1, updated_at = NOW()
			WHERE id = $2`

	result, err := db.ExecContext(ctx, qry, model.ParticipantStatusDisconnected, participantID)
	if err != nil {
		commons.LogE(ctx, "[dao.MarkParticipantLeft::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("participant-not-found")
	}

	commons.Log(ctx, "[dao.MarkParticipantLeft::OK [participant_id=%s", participantID)
	return nil
}

// GetParticipantCount gets the number of participants in a session
func GetParticipantCount(ctx context.Context, sessionID uuid.UUID) (int, error) {
	var count int
	qry := `SELECT COUNT(*) FROM live_session_participants WHERE session_id = $1`

	err := db.QueryRowContext(ctx, qry, sessionID).Scan(&count)
	if err != nil {
		commons.LogE(ctx, "[dao.GetParticipantCount::SQLErr [err=%v", err)
		return 0, err
	}

	return count, nil
}
