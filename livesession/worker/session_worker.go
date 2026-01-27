package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/dao"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
	"github.com/affanahmad024/quiz-craft/backend/livesession/websocket"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	quizmodel "github.com/affanahmad024/quiz-craft/backend/quiz/model"
)

const (
	WorkerPollInterval = 1 * time.Second // Check every second for precise timing
)

// SessionWorker handles automatic session start and progression
type SessionWorker struct {
	stopChan chan struct{}
	wsHub    *websocket.Hub
}

// NewSessionWorker creates a new session worker
func NewSessionWorker(wsHub *websocket.Hub) *SessionWorker {
	return &SessionWorker{
		stopChan: make(chan struct{}),
		wsHub:    wsHub,
	}
}

// Start starts the worker
func (w *SessionWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(WorkerPollInterval)
	defer ticker.Stop()

	commons.Log(ctx, "[SessionWorker.Start] Session worker started, polling every %v", WorkerPollInterval)

	for {
		select {
		case <-ctx.Done():
			commons.Log(ctx, "[SessionWorker.Start] Context cancelled, stopping session worker")
			return
		case <-w.stopChan:
			commons.Log(ctx, "[SessionWorker.Start] Stop signal received, stopping session worker")
			return
		case <-ticker.C:
			w.ProcessScheduledStarts(ctx)
			w.ProcessAutomaticProgression(ctx)
		}
	}
}

// Stop stops the worker
func (w *SessionWorker) Stop() {
	close(w.stopChan)
}

// ProcessScheduledStarts checks for sessions that should auto-start
func (w *SessionWorker) ProcessScheduledStarts(ctx context.Context) {
	// Get sessions scheduled to start now or in the past
	sessions, err := dao.GetScheduledSessions(ctx, time.Now())
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.ProcessScheduledStarts::GetErr [err=%v", err)
		return
	}

	if len(sessions) == 0 {
		return
	}

	commons.Log(ctx, "[SessionWorker.ProcessScheduledStarts] Found %d sessions to start", len(sessions))

	for _, session := range sessions {
		w.StartSession(ctx, session.ID)
	}
}

// StartSession auto-starts a session
func (w *SessionWorker) StartSession(ctx context.Context, sessionID uuid.UUID) {
	commons.Log(ctx, "[SessionWorker.StartSession [session_id=%s", sessionID)

	// Start session
	err := dao.StartSession(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.StartSession::StartErr [session_id=%s [err=%v", sessionID, err)
		return
	}

	// Get session details to know quiz ID
	session, err := dao.FindSessionByID(ctx, sessionID)
	if err != nil {
		return
	}

	// Get first question to send in the event
	questions, err := quizdao.ListByQuizID(ctx, session.QuizID)
	if err != nil || len(questions) == 0 {
		commons.LogE(ctx, "[SessionWorker.StartSession::NoQuestions [session_id=%s", sessionID)
		return
	}

	// Broadcast session started event
	if w.wsHub != nil {
		event := websocket.Event{
			Type: websocket.EventSessionStarted,
			Data: map[string]interface{}{
				"session_id":      sessionID.String(),
				"total_questions": len(questions),
				"first_question":  questions[0],
			},
		}
		w.wsHub.BroadcastToSession(sessionID, event)
	}

	commons.Log(ctx, "[SessionWorker.StartSession::OK [session_id=%s", sessionID)
}

// ProcessAutomaticProgression handles automatic question progression
func (w *SessionWorker) ProcessAutomaticProgression(ctx context.Context) {
	// Get all live sessions with automatic progression
	sessions, err := dao.GetLiveAutomaticSessions(ctx)
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.ProcessAutomaticProgression::GetErr [err=%v", err)
		return
	}

	if len(sessions) == 0 {
		return
	}

	for _, session := range sessions {
		w.ProcessSession(ctx, session)
	}
}

// ProcessSession processes a single session for automatic progression
func (w *SessionWorker) ProcessSession(ctx context.Context, session model.LiveSessionDAO) {
	// Check if current question timer has expired
	if !session.CurrentQuestionStartedAt.Valid {
		return
	}

	timeElapsed := time.Since(session.CurrentQuestionStartedAt.Time)
	timerDuration := time.Duration(session.QuestionTimerSeconds) * time.Second

	// Send timer updates every second
	remainingSeconds := int(timerDuration.Seconds() - timeElapsed.Seconds())
	if remainingSeconds >= 0 {
		if w.wsHub != nil {
			event := websocket.Event{
				Type: websocket.EventTimerUpdate,
				Data: map[string]interface{}{
					"seconds_remaining":      remainingSeconds,
					"current_question_index": session.CurrentQuestionIndex,
				},
			}
			w.wsHub.BroadcastToSession(session.ID, event)
		}
	}

	// Check if timer expired
	if timeElapsed < timerDuration {
		return
	}

	// Timer expired, advance to next question or end session
	commons.Log(ctx, "[SessionWorker.ProcessSession] Timer expired for session_id=%s", session.ID)

	// Get total questions
	questions, err := quizdao.ListByQuizID(ctx, session.QuizID)
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.ProcessSession::GetQuestionsErr [err=%v", err)
		return
	}

	totalQuestions := len(questions)

	// Check if this was the last question
	if session.CurrentQuestionIndex >= totalQuestions-1 {
		// Last question completed, end session
		w.EndSession(ctx, session.ID)
	} else {
		// Advance to next question
		w.AdvanceQuestion(ctx, session.ID, session.CurrentQuestionIndex+1, questions)
	}
}

// AdvanceQuestion advances to the next question
func (w *SessionWorker) AdvanceQuestion(ctx context.Context, sessionID uuid.UUID, nextIndex int, questions []*quizmodel.Question) {
	commons.Log(ctx, "[SessionWorker.AdvanceQuestion [session_id=%s [next_index=%d", sessionID, nextIndex)

	// Advance question in database
	err := dao.AdvanceQuestion(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.AdvanceQuestion::AdvanceErr [err=%v", err)
		return
	}

	// Broadcast question auto-advance event
	if w.wsHub != nil && nextIndex < len(questions) {
		event := websocket.Event{
			Type: websocket.EventQuestionAutoAdvance,
			Data: map[string]interface{}{
				"next_question_index": nextIndex,
				"next_question":       questions[nextIndex],
			},
		}
		w.wsHub.BroadcastToSession(sessionID, event)
	}

	commons.Log(ctx, "[SessionWorker.AdvanceQuestion::OK [session_id=%s", sessionID)
}

// EndSession ends a session
func (w *SessionWorker) EndSession(ctx context.Context, sessionID uuid.UUID) {
	commons.Log(ctx, "[SessionWorker.EndSession [session_id=%s", sessionID)

	// End session in database
	err := dao.EndSession(ctx, sessionID)
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.EndSession::EndErr [err=%v", err)
		return
	}

	// Mark all active participants as completed
	participants, _ := dao.ListParticipants(ctx, sessionID)
	for _, p := range participants {
		if p.Status == model.ParticipantStatusActive {
			dao.UpdateParticipantStatus(ctx, p.ID, model.ParticipantStatusCompleted)
		}
	}

	// Get final leaderboard
	leaderboard, _ := dao.GetLeaderboard(ctx, sessionID)

	// Broadcast session ended event
	if w.wsHub != nil {
		event := websocket.Event{
			Type: websocket.EventSessionEnded,
			Data: map[string]interface{}{
				"session_id":        sessionID.String(),
				"final_leaderboard": leaderboard,
			},
		}
		w.wsHub.BroadcastToSession(sessionID, event)
	}

	commons.Log(ctx, "[SessionWorker.EndSession::OK [session_id=%s", sessionID)
}
