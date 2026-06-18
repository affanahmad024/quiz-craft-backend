package worker

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/livesession/dao"
	"github.com/affanahmad024/quiz-craft/backend/livesession/model"
	"github.com/affanahmad024/quiz-craft/backend/livesession/websocket"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	quizmodel "github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
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
	questions, err := loadQuestionsWithAnswers(ctx, session.QuizID)
	if err != nil || len(questions) == 0 {
		commons.LogE(ctx, "[SessionWorker.StartSession::NoQuestions [session_id=%s", sessionID)
		return
	}

	// Broadcast session started event
	if w.wsHub != nil {
		eventData := map[string]interface{}{
			"session_id":      sessionID.String(),
			"total_questions": len(questions),
			"first_question":  questions[0],
			"timer_mode":      session.TimerMode,
		}
		if session.QuestionTimerSeconds.Valid {
			eventData["question_timer_seconds"] = session.QuestionTimerSeconds.Int64
		}
		if session.WholeQuizTimerSeconds.Valid {
			eventData["whole_quiz_timer_seconds"] = session.WholeQuizTimerSeconds.Int64
		}

		event := websocket.Event{
			Type: websocket.EventSessionStarted,
			Data: eventData,
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
	// Route to appropriate timer handler based on timer mode
	if session.TimerMode == model.TimerModePerQuestion {
		w.ProcessPerQuestionTimer(ctx, session)
	} else if session.TimerMode == model.TimerModeWholeQuiz {
		w.ProcessWholeQuizTimer(ctx, session)
	}
}

// ProcessPerQuestionTimer handles per-question timer logic
func (w *SessionWorker) ProcessPerQuestionTimer(ctx context.Context, session model.LiveSessionDAO) {
	// Check if current question timer has expired
	if !session.CurrentQuestionStartedAt.Valid {
		return
	}
	if !session.QuestionTimerSeconds.Valid {
		return
	}

	timeElapsed := liveTimerElapsed(session.CurrentQuestionStartedAt.Time)
	timerDuration := time.Duration(session.QuestionTimerSeconds.Int64) * time.Second

	// Send timer updates every second
	remainingSeconds := liveTimerRemaining(session.CurrentQuestionStartedAt.Time, session.QuestionTimerSeconds.Int64)
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
	commons.Log(ctx, "[SessionWorker.ProcessPerQuestionTimer] Timer expired for session_id=%s", session.ID)

	// Get total questions
	questions, err := loadQuestionsWithAnswers(ctx, session.QuizID)
	if err != nil {
		commons.LogE(ctx, "[SessionWorker.ProcessPerQuestionTimer::GetQuestionsErr [err=%v", err)
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

// ProcessWholeQuizTimer handles whole-quiz timer logic
func (w *SessionWorker) ProcessWholeQuizTimer(ctx context.Context, session model.LiveSessionDAO) {
	// Check if quiz timer started
	if !session.QuizStartedAt.Valid {
		return
	}

	// Get whole quiz timer duration
	if !session.WholeQuizTimerSeconds.Valid {
		return
	}

	timeElapsed := liveTimerElapsed(session.QuizStartedAt.Time)
	timerDuration := time.Duration(session.WholeQuizTimerSeconds.Int64) * time.Second

	// Send timer updates every second
	remainingSeconds := liveTimerRemaining(session.QuizStartedAt.Time, session.WholeQuizTimerSeconds.Int64)
	if remainingSeconds >= 0 {
		if w.wsHub != nil {
			event := websocket.Event{
				Type: websocket.EventWholeQuizTimerUpdate,
				Data: map[string]interface{}{
					"seconds_remaining": remainingSeconds,
					"total_seconds":     session.WholeQuizTimerSeconds.Int64,
				},
			}
			w.wsHub.BroadcastToSession(session.ID, event)
		}
	}

	// Check if timer expired
	if timeElapsed < timerDuration {
		return
	}

	// Timer expired, end session immediately
	commons.Log(ctx, "[SessionWorker.ProcessWholeQuizTimer] Timer expired for session_id=%s", session.ID)
	w.EndSession(ctx, session.ID)
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

func loadQuestionsWithAnswers(ctx context.Context, quizID uuid.UUID) ([]*quizmodel.Question, error) {
	questions, err := quizdao.ListByQuizID(ctx, quizID)
	if err != nil {
		return questions, err
	}

	for _, question := range questions {
		questionID, err := uuid.Parse(question.ID)
		if err != nil {
			continue
		}

		answers, err := quizdao.ListByQuestionIDHideCorrect(ctx, questionID)
		if err != nil {
			commons.LogE(ctx, "[SessionWorker.loadQuestionsWithAnswers::AnswersErr [question_id=%s [err=%v", question.ID, err)
			continue
		}

		question.Answers = answers
	}

	return questions, nil
}

func liveTimerElapsed(startedAt time.Time) time.Duration {
	startedAt = normalizeLiveTimerStart(startedAt)
	elapsed := time.Since(startedAt)
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func liveTimerRemaining(startedAt time.Time, totalSeconds int64) int {
	if totalSeconds <= 0 {
		return 0
	}

	remaining := int(totalSeconds) - int(liveTimerElapsed(startedAt).Seconds())
	if remaining < 0 {
		return 0
	}
	if remaining > int(totalSeconds) {
		return int(totalSeconds)
	}
	return remaining
}

func normalizeLiveTimerStart(startedAt time.Time) time.Time {
	now := time.Now()
	if !startedAt.After(now.Add(2 * time.Minute)) {
		return startedAt
	}

	_, offsetSeconds := now.In(time.Local).Zone()
	if offsetSeconds != 0 {
		adjusted := startedAt.Add(-time.Duration(offsetSeconds) * time.Second)
		if !adjusted.After(now.Add(2 * time.Minute)) {
			return adjusted
		}
	}

	return now
}
