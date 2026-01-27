package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
)

// CreateInvitation creates a single invitation
func CreateInvitation(ctx context.Context, sessionID, inviterID, inviteeID uuid.UUID) (model.InvitationDAO, error) {
	commons.Log(ctx, "[dao.CreateInvitation [session_id=%s [inviter_id=%s [invitee_id=%s", sessionID, inviterID, inviteeID)

	invitation := model.InvitationDAO{}
	qry := `INSERT INTO live_session_invitations (session_id, inviter_id, invitee_id, status, invited_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())
			RETURNING id, session_id, inviter_id, invitee_id, status, invited_at, responded_at, created_at, updated_at`

	err := db.QueryRowContext(ctx, qry, sessionID, inviterID, inviteeID, model.InvitationStatusPending).Scan(
		&invitation.ID,
		&invitation.SessionID,
		&invitation.InviterID,
		&invitation.InviteeID,
		&invitation.Status,
		&invitation.InvitedAt,
		&invitation.RespondedAt,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate
		if err.Error() == "pq: duplicate key value violates unique constraint \"live_session_invitations_session_id_invitee_id_key\"" {
			return invitation, commons.Conflict("already-invited")
		}
		commons.LogE(ctx, "[dao.CreateInvitation::SQLErr [err=%v", err)
		return invitation, err
	}

	commons.Log(ctx, "[dao.CreateInvitation::OK [invitation_id=%s", invitation.ID)
	return invitation, nil
}

// CreateBulkInvitations creates invitations for multiple users
func CreateBulkInvitations(ctx context.Context, sessionID, inviterID uuid.UUID, inviteeIDs []uuid.UUID) error {
	commons.Log(ctx, "[dao.CreateBulkInvitations [session_id=%s [inviter_id=%s [count=%d", sessionID, inviterID, len(inviteeIDs))

	if len(inviteeIDs) == 0 {
		return nil
	}

	qry := `INSERT INTO live_session_invitations (session_id, inviter_id, invitee_id, status, invited_at, created_at, updated_at)
			SELECT $1, $2, unnest($3::uuid[]), $4, NOW(), NOW(), NOW()
			ON CONFLICT (session_id, invitee_id) DO NOTHING`

	_, err := db.ExecContext(ctx, qry, sessionID, inviterID, pq.Array(inviteeIDs), model.InvitationStatusPending)
	if err != nil {
		commons.LogE(ctx, "[dao.CreateBulkInvitations::SQLErr [err=%v", err)
		return err
	}

	commons.Log(ctx, "[dao.CreateBulkInvitations::OK [count=%d", len(inviteeIDs))
	return nil
}

// FindInvitation finds an invitation by ID
func FindInvitation(ctx context.Context, invitationID uuid.UUID) (model.InvitationDAO, error) {
	commons.Log(ctx, "[dao.FindInvitation [invitation_id=%s", invitationID)

	invitation := model.InvitationDAO{}
	qry := `SELECT id, session_id, inviter_id, invitee_id, status, invited_at, responded_at, created_at, updated_at
			FROM live_session_invitations
			WHERE id = $1`

	err := db.QueryRowContext(ctx, qry, invitationID).Scan(
		&invitation.ID,
		&invitation.SessionID,
		&invitation.InviterID,
		&invitation.InviteeID,
		&invitation.Status,
		&invitation.InvitedAt,
		&invitation.RespondedAt,
		&invitation.CreatedAt,
		&invitation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return invitation, commons.NotFound("invitation-not-found")
	}
	if err != nil {
		commons.LogE(ctx, "[dao.FindInvitation::SQLErr [err=%v", err)
		return invitation, err
	}

	return invitation, nil
}

// ListSessionInvitations lists all invitations for a session
func ListSessionInvitations(ctx context.Context, sessionID uuid.UUID) ([]model.InvitationDAO, error) {
	commons.Log(ctx, "[dao.ListSessionInvitations [session_id=%s", sessionID)

	invitations := []model.InvitationDAO{}
	qry := `SELECT id, session_id, inviter_id, invitee_id, status, invited_at, responded_at, created_at, updated_at
			FROM live_session_invitations
			WHERE session_id = $1
			ORDER BY invited_at DESC`

	rows, err := db.QueryContext(ctx, qry, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.ListSessionInvitations::SQLErr [err=%v", err)
		return invitations, err
	}
	defer rows.Close()

	for rows.Next() {
		var invitation model.InvitationDAO
		err := rows.Scan(
			&invitation.ID,
			&invitation.SessionID,
			&invitation.InviterID,
			&invitation.InviteeID,
			&invitation.Status,
			&invitation.InvitedAt,
			&invitation.RespondedAt,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListSessionInvitations::ScanErr [err=%v", err)
			continue
		}
		invitations = append(invitations, invitation)
	}

	commons.Log(ctx, "[dao.ListSessionInvitations::OK [count=%d", len(invitations))
	return invitations, nil
}

// ListUserInvitations lists pending invitations for a user
func ListUserInvitations(ctx context.Context, userID uuid.UUID) ([]model.InvitationDAO, error) {
	commons.Log(ctx, "[dao.ListUserInvitations [user_id=%s", userID)

	invitations := []model.InvitationDAO{}
	qry := `SELECT id, session_id, inviter_id, invitee_id, status, invited_at, responded_at, created_at, updated_at
			FROM live_session_invitations
			WHERE invitee_id = $1 AND status = $2
			ORDER BY invited_at DESC`

	rows, err := db.QueryContext(ctx, qry, userID, model.InvitationStatusPending)
	if err != nil {
		commons.LogE(ctx, "[dao.ListUserInvitations::SQLErr [err=%v", err)
		return invitations, err
	}
	defer rows.Close()

	for rows.Next() {
		var invitation model.InvitationDAO
		err := rows.Scan(
			&invitation.ID,
			&invitation.SessionID,
			&invitation.InviterID,
			&invitation.InviteeID,
			&invitation.Status,
			&invitation.InvitedAt,
			&invitation.RespondedAt,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			commons.LogE(ctx, "[dao.ListUserInvitations::ScanErr [err=%v", err)
			continue
		}
		invitations = append(invitations, invitation)
	}

	commons.Log(ctx, "[dao.ListUserInvitations::OK [count=%d", len(invitations))
	return invitations, nil
}

// UpdateInvitationStatus updates invitation status
func UpdateInvitationStatus(ctx context.Context, invitationID uuid.UUID, status string) error {
	commons.Log(ctx, "[dao.UpdateInvitationStatus [invitation_id=%s [status=%s", invitationID, status)

	qry := `UPDATE live_session_invitations
			SET status = $1, responded_at = NOW(), updated_at = NOW()
			WHERE id = $2 AND status = $3`

	result, err := db.ExecContext(ctx, qry, status, invitationID, model.InvitationStatusPending)
	if err != nil {
		commons.LogE(ctx, "[dao.UpdateInvitationStatus::SQLErr [err=%v", err)
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return commons.NotFound("invitation-not-found-or-already-responded")
	}

	commons.Log(ctx, "[dao.UpdateInvitationStatus::OK [invitation_id=%s", invitationID)
	return nil
}

// HasInvitation checks if user has an invitation to the session
func HasInvitation(ctx context.Context, sessionID, userID uuid.UUID) (bool, error) {
	var exists bool
	qry := `SELECT EXISTS(SELECT 1 FROM live_session_invitations WHERE session_id = $1 AND invitee_id = $2)`

	err := db.QueryRowContext(ctx, qry, sessionID, userID).Scan(&exists)
	if err != nil {
		commons.LogE(ctx, "[dao.HasInvitation::SQLErr [err=%v", err)
		return false, err
	}

	return exists, nil
}

// GetInvitedUserIDs gets all user IDs invited to a session
func GetInvitedUserIDs(ctx context.Context, sessionID uuid.UUID) ([]uuid.UUID, error) {
	commons.Log(ctx, "[dao.GetInvitedUserIDs [session_id=%s", sessionID)

	userIDs := []uuid.UUID{}
	qry := `SELECT invitee_id FROM live_session_invitations WHERE session_id = $1`

	rows, err := db.QueryContext(ctx, qry, sessionID)
	if err != nil {
		commons.LogE(ctx, "[dao.GetInvitedUserIDs::SQLErr [err=%v", err)
		return userIDs, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}

	commons.Log(ctx, "[dao.GetInvitedUserIDs::OK [count=%d", len(userIDs))
	return userIDs, nil
}
