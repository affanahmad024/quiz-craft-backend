package services

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
)

const (
	CleanupInterval = 24 * time.Hour // Run cleanup once per day
)

// FileCleanupService handles periodic file cleanup
type FileCleanupService struct {
	fileUploadService *FileUploadService
	stopChan          chan struct{}
}

// NewFileCleanupService creates a new file cleanup service
func NewFileCleanupService() *FileCleanupService {
	return &FileCleanupService{
		fileUploadService: NewFileUploadService(),
		stopChan:          make(chan struct{}),
	}
}

// Start starts the cleanup service
func (s *FileCleanupService) Start(ctx context.Context) {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	commons.Log(ctx, "[FileCleanupService.Start] File cleanup service started")

	// Run immediately on start
	s.runCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			commons.Log(ctx, "[FileCleanupService.Start] Context cancelled, stopping cleanup service")
			return
		case <-s.stopChan:
			commons.Log(ctx, "[FileCleanupService.Start] Stop signal received, stopping cleanup service")
			return
		case <-ticker.C:
			s.runCleanup(ctx)
		}
	}
}

// Stop stops the cleanup service
func (s *FileCleanupService) Stop() {
	close(s.stopChan)
}

// runCleanup runs the cleanup process
func (s *FileCleanupService) runCleanup(ctx context.Context) {
	commons.Log(ctx, "[FileCleanupService.runCleanup] Starting file cleanup")

	deleted, err := s.fileUploadService.CleanupExpiredFiles(ctx)
	if err != nil {
		commons.LogE(ctx, "[FileCleanupService.runCleanup::Error [err=%v", err)
		return
	}

	commons.Log(ctx, "[FileCleanupService.runCleanup::Complete [deleted=%d", deleted)
}
