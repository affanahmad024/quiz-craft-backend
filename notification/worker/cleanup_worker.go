package worker

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/notification/dao"
)

const (
	CleanupPollInterval   = 24 * time.Hour // Check once per day
	NotificationRetention = 90             // Keep notifications for 90 days
)

// CleanupWorker handles notification cleanup tasks
type CleanupWorker struct {
	stopChan chan struct{}
}

// NewCleanupWorker creates a new cleanup worker
func NewCleanupWorker() *CleanupWorker {
	return &CleanupWorker{
		stopChan: make(chan struct{}),
	}
}

// Start starts the worker
func (w *CleanupWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(CleanupPollInterval)
	defer ticker.Stop()

	commons.Log(ctx, "[CleanupWorker.Start] Notification cleanup worker started, polling every %v", CleanupPollInterval)

	// Run immediately on start (optional, can comment out for production)
	// w.ProcessCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			commons.Log(ctx, "[CleanupWorker.Start] Context cancelled, stopping cleanup worker")
			return
		case <-w.stopChan:
			commons.Log(ctx, "[CleanupWorker.Start] Stop signal received, stopping cleanup worker")
			return
		case <-ticker.C:
			w.ProcessCleanup(ctx)
		}
	}
}

// Stop stops the worker
func (w *CleanupWorker) Stop() {
	close(w.stopChan)
}

// ProcessCleanup deletes old read notifications
func (w *CleanupWorker) ProcessCleanup(ctx context.Context) {
	commons.Log(ctx, "[CleanupWorker.ProcessCleanup] Starting notification cleanup")

	// Delete read notifications older than retention period
	deleted, err := dao.DeleteOldNotifications(ctx, NotificationRetention)
	if err != nil {
		commons.LogE(ctx, "[CleanupWorker.ProcessCleanup::DeleteErr [err=%v", err)
		return
	}

	if deleted > 0 {
		commons.Log(ctx, "[CleanupWorker.ProcessCleanup::Complete [deleted=%d", deleted)
	} else {
		commons.Log(ctx, "[CleanupWorker.ProcessCleanup::Complete [no_old_notifications]")
	}
}
