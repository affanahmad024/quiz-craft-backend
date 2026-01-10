package worker

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/core"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/google/uuid"
)

const (
	AttemptPollInterval = 1 * time.Minute // Check every 1 minute
	MaxExpiredAttempts  = 50              // Process up to 50 expired attempts per poll
)

// AttemptWorker processes expired quiz attempts
type AttemptWorker struct {
	stopChan chan struct{}
}

// NewAttemptWorker creates a new attempt worker
func NewAttemptWorker() *AttemptWorker {
	return &AttemptWorker{
		stopChan: make(chan struct{}),
	}
}

// Start starts the worker
func (w *AttemptWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(AttemptPollInterval)
	defer ticker.Stop()

	commons.Log(ctx, "[AttemptWorker.Start] Attempt worker started, polling every %v", AttemptPollInterval)

	// Run immediately on start
	w.ProcessExpiredAttempts(ctx)

	for {
		select {
		case <-ctx.Done():
			commons.Log(ctx, "[AttemptWorker.Start] Context cancelled, stopping attempt worker")
			return
		case <-w.stopChan:
			commons.Log(ctx, "[AttemptWorker.Start] Stop signal received, stopping attempt worker")
			return
		case <-ticker.C:
			w.ProcessExpiredAttempts(ctx)
		}
	}
}

// Stop stops the worker
func (w *AttemptWorker) Stop() {
	close(w.stopChan)
}

// ProcessExpiredAttempts finds and finishes expired attempts
func (w *AttemptWorker) ProcessExpiredAttempts(ctx context.Context) {
	commons.Log(ctx, "[AttemptWorker.ProcessExpiredAttempts] Checking for expired attempts")

	// Get expired attempts
	attempts, err := quizdao.GetExpiredAttempts(ctx, MaxExpiredAttempts)
	if err != nil {
		commons.LogE(ctx, "[AttemptWorker.ProcessExpiredAttempts::GetErr [err=%v", err)
		return
	}

	if len(attempts) == 0 {
		return
	}

	commons.Log(ctx, "[AttemptWorker.ProcessExpiredAttempts] Found %d expired attempts", len(attempts))

	// Process each expired attempt
	processed := 0
	for _, attempt := range attempts {
		if w.FinishExpiredAttempt(ctx, attempt.ID) {
			processed++
		}
	}

	commons.Log(ctx, "[AttemptWorker.ProcessExpiredAttempts::Complete [expired=%d [processed=%d", len(attempts), processed)
}

// FinishExpiredAttempt finishes a single expired attempt
func (w *AttemptWorker) FinishExpiredAttempt(ctx context.Context, attemptID string) bool {
	commons.Log(ctx, "[AttemptWorker.FinishExpiredAttempt [attempt_id=%s", attemptID)

	attemptUUID, err := uuid.Parse(attemptID)
	if err != nil {
		commons.LogE(ctx, "[AttemptWorker.FinishExpiredAttempt::InvalidID [attempt_id=%s", attemptID)
		return false
	}

	// Get attempt
	attempt, err := quizdao.FindAttemptByID(ctx, attemptUUID)
	if err != nil {
		commons.LogE(ctx, "[AttemptWorker.FinishExpiredAttempt::FindErr [attempt_id=%s [err=%v", attemptID, err)
		return false
	}

	// Double check it's expired and in progress
	if !attempt.IsExpired() || !attempt.IsInProgress() {
		commons.LogW(ctx, "[AttemptWorker.FinishExpiredAttempt::NotExpired [attempt_id=%s [status=%s", attemptID, attempt.Status)
		return false
	}

	// Use core logic to finish the attempt
	// This will calculate the score and update the status
	_, err = core.FinishAttempt(ctx, attemptID, attempt.UserID)
	if err != nil {
		commons.LogE(ctx, "[AttemptWorker.FinishExpiredAttempt::FinishErr [attempt_id=%s [err=%v", attemptID, err)
		return false
	}

	commons.Log(ctx, "[AttemptWorker.FinishExpiredAttempt::OK [attempt_id=%s", attemptID)
	return true
}
