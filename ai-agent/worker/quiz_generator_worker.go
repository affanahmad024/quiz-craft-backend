package worker

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/core"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/dao"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/services"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/google/uuid"
)

const (
	PollInterval    = 5 * time.Second
	MaxConcurrent   = 5
	ProcessingLimit = 5
)

// QuizGeneratorWorker processes quiz generation jobs
type QuizGeneratorWorker struct {
	db               *sql.DB
	aiService        *core.AIService
	fileParser       *services.FileParserService
	quizGenerator    *services.QuizGeneratorService
	stopChan         chan struct{}
	concurrencySem   chan struct{}
}

// NewQuizGeneratorWorker creates a new worker
func NewQuizGeneratorWorker(db *sql.DB) *QuizGeneratorWorker {
	apiKey := os.Getenv("GOOGLE_AI_API_KEY")
	if apiKey == "" {
		apiKey = "AIzaSyDkTtFFjLXOkrlW6JRfgxXu3kuXMsIjptc"
	}

	model := os.Getenv("GOOGLE_AI_MODEL")
	if model == "" {
		model = "gemini-1.5-pro"
	}

	return &QuizGeneratorWorker{
		db:             db,
		aiService:      core.NewAIService(apiKey, model),
		fileParser:     services.NewFileParserService(),
		quizGenerator:  services.NewQuizGeneratorService(db),
		stopChan:       make(chan struct{}),
		concurrencySem: make(chan struct{}, MaxConcurrent),
	}
}

// Start starts the worker
func (w *QuizGeneratorWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	commons.Log(ctx, "[worker.Start] Quiz generation worker started")

	for {
		select {
		case <-ctx.Done():
			commons.Log(ctx, "[worker.Start] Context cancelled, stopping worker")
			return
		case <-w.stopChan:
			commons.Log(ctx, "[worker.Start] Stop signal received, stopping worker")
			return
		case <-ticker.C:
			w.ProcessPendingJobs(ctx)
		}
	}
}

// Stop stops the worker
func (w *QuizGeneratorWorker) Stop() {
	close(w.stopChan)
}

// ProcessPendingJobs fetches and processes pending jobs
func (w *QuizGeneratorWorker) ProcessPendingJobs(ctx context.Context) {
	jobs, err := dao.GetPendingJobs(ctx, ProcessingLimit)
	if err != nil {
		commons.LogE(ctx, "[worker.ProcessPendingJobs::FetchErr [err=%v", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	commons.Log(ctx, "[worker.ProcessPendingJobs [pending_jobs=%d", len(jobs))

	for _, job := range jobs {
		// Acquire semaphore slot
		w.concurrencySem <- struct{}{}

		// Process job in goroutine
		go func(j *model.GenerationJob) {
			defer func() { <-w.concurrencySem }() // Release slot

			jobCtx := context.Background()
			w.ProcessJob(jobCtx, j)
		}(job)
	}
}

// ProcessJob processes a single generation job
func (w *QuizGeneratorWorker) ProcessJob(ctx context.Context, job *model.GenerationJob) {
	jobUUID, _ := uuid.Parse(job.ID)

	commons.Log(ctx, "[worker.ProcessJob [job_id=%s [user_id=%s", job.ID, job.UserID)

	// Update status to processing
	if err := dao.UpdateJobStatus(ctx, jobUUID, "processing"); err != nil {
		commons.LogE(ctx, "[worker.ProcessJob::StatusErr [job_id=%s [err=%v", job.ID, err)
		return
	}

	// Update progress: 10%
	dao.UpdateJobProgress(ctx, jobUUID, 10)

	// Get content (either from file or direct text)
	var content string
	var err error

	if job.FileID != nil && *job.FileID != "" {
		// Parse file
		fileUUID, _ := uuid.Parse(*job.FileID)
		file, err := dao.GetFileByID(ctx, fileUUID)
		if err != nil {
			w.failJob(ctx, jobUUID, "Failed to get file: "+err.Error())
			return
		}

		dao.UpdateJobProgress(ctx, jobUUID, 30)

		content, err = w.fileParser.ParseFile(ctx, file.FilePath)
		if err != nil {
			w.failJob(ctx, jobUUID, "Failed to parse file: "+err.Error())
			return
		}
	} else if job.TextContent != nil {
		content = *job.TextContent
	} else {
		w.failJob(ctx, jobUUID, "No content provided")
		return
	}

	dao.UpdateJobProgress(ctx, jobUUID, 40)

	// Generate quiz using AI
	aiResponse, err := w.aiService.GenerateQuizFromText(ctx, content, job.Options)
	if err != nil {
		w.failJob(ctx, jobUUID, "AI generation failed: "+err.Error())
		return
	}

	dao.UpdateJobProgress(ctx, jobUUID, 70)

	// Validate AI response
	if err := w.quizGenerator.ValidateAIQuizResponse(aiResponse); err != nil {
		w.failJob(ctx, jobUUID, "Invalid AI response: "+err.Error())
		return
	}

	dao.UpdateJobProgress(ctx, jobUUID, 80)

	// Create quiz in database
	quizID, err := w.quizGenerator.CreateQuizFromAIResponse(ctx, aiResponse, job.UserID, job.Options)
	if err != nil {
		w.failJob(ctx, jobUUID, "Failed to create quiz: "+err.Error())
		return
	}

	dao.UpdateJobProgress(ctx, jobUUID, 95)

	// Mark job as completed
	if err := dao.CompleteJob(ctx, jobUUID, quizID); err != nil {
		commons.LogE(ctx, "[worker.ProcessJob::CompleteErr [job_id=%s [err=%v", job.ID, err)
		return
	}

	commons.Log(ctx, "[worker.ProcessJob::OK [job_id=%s [quiz_id=%s", job.ID, quizID)
}

// failJob marks a job as failed
func (w *QuizGeneratorWorker) failJob(ctx context.Context, jobID uuid.UUID, errorMessage string) {
	commons.LogE(ctx, "[worker.failJob [job_id=%s [error=%s", jobID, errorMessage)
	dao.FailJob(ctx, jobID, errorMessage)
}
