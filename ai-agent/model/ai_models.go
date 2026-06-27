package model

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// QuizGenOptions represents user preferences for AI quiz generation
type QuizGenOptions struct {
	NumQuestions int          `json:"num_questions" binding:"required,min=5,max=50"`
	Difficulty   string       `json:"difficulty" binding:"required,oneof=easy medium hard"`
	QuestionType string       `json:"question_type" binding:"required,oneof=multiple_choice true_false mixed"`
	Category     string       `json:"category" binding:"required,min=2,max=50"`
	Language     string       `json:"language" binding:"required,len=2"` // ISO 639-1 code (e.g., "en", "es")
	Metadata     QuizMetadata `json:"metadata,omitempty"`
}

// QuizMetadata represents quiz fields that should be applied to the created quiz
// after AI has generated the questions.
type QuizMetadata struct {
	Title                  string   `json:"title,omitempty"`
	Description            string   `json:"description,omitempty"`
	Tags                   []string `json:"tags,omitempty"`
	TimeLimitMinutes       *int     `json:"time_limit_minutes,omitempty"`
	PassingScorePercentage *int     `json:"passing_score_percentage,omitempty"`
}

// UnmarshalJSON accepts numeric metadata fields as either JSON numbers or
// numeric strings. Browser number inputs can arrive as strings in some callers.
func (m *QuizMetadata) UnmarshalJSON(data []byte) error {
	type quizMetadataAlias QuizMetadata
	var raw struct {
		quizMetadataAlias
		TimeLimitMinutes       json.RawMessage `json:"time_limit_minutes"`
		PassingScorePercentage json.RawMessage `json:"passing_score_percentage"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*m = QuizMetadata(raw.quizMetadataAlias)

	timeLimit, err := parseOptionalInt(raw.TimeLimitMinutes)
	if err != nil {
		return err
	}
	m.TimeLimitMinutes = timeLimit

	passingScore, err := parseOptionalInt(raw.PassingScorePercentage)
	if err != nil {
		return err
	}
	m.PassingScorePercentage = passingScore

	return nil
}

func parseOptionalInt(raw json.RawMessage) (*int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var number int
	if err := json.Unmarshal(raw, &number); err == nil {
		return &number, nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return nil, err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	parsed, err := strconv.Atoi(text)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// FileUpload represents an uploaded file's metadata
type FileUpload struct {
	ID               string     `json:"id" db:"id"`
	UserID           string     `json:"user_id" db:"user_id"`
	Filename         string     `json:"filename" db:"filename"`
	OriginalFilename string     `json:"original_filename" db:"original_filename"`
	FilePath         string     `json:"file_path" db:"file_path"`
	FileSize         int64      `json:"file_size" db:"file_size"`
	MimeType         string     `json:"mime_type" db:"mime_type"`
	Status           string     `json:"status" db:"status"` // uploaded, processing, processed, failed
	ErrorMessage     *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// GenerationJob represents an async AI quiz generation job
type GenerationJob struct {
	ID              string         `json:"id" db:"id"`
	UserID          string         `json:"user_id" db:"user_id"`
	FileID          *string        `json:"file_id,omitempty" db:"file_id"`
	TextContent     *string        `json:"text_content,omitempty" db:"text_content"`
	JobType         string         `json:"job_type" db:"job_type"`                           // generation, extraction
	ReferenceImages []string       `json:"reference_images,omitempty" db:"reference_images"` // File IDs of reference images (stored as JSONB)
	Options         QuizGenOptions `json:"options" db:"options"`                             // Stored as JSONB
	Status          string         `json:"status" db:"status"`                               // pending, processing, completed, failed
	Progress        int            `json:"progress" db:"progress"`
	QuizID          *string        `json:"quiz_id,omitempty" db:"quiz_id"`
	ErrorMessage    *string        `json:"error_message,omitempty" db:"error_message"`
	StartedAt       *time.Time     `json:"started_at,omitempty" db:"started_at"`
	CompletedAt     *time.Time     `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
}

// AIQuizResponse represents the expected JSON response from Google Gemini AI
type AIQuizResponse struct {
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Difficulty  string       `json:"difficulty"`
	Questions   []AIQuestion `json:"questions"`
}

// AIQuestion represents a single question in the AI response
type AIQuestion struct {
	QuestionText string     `json:"question_text"`
	QuestionType string     `json:"question_type"` // multiple_choice, true_false
	Difficulty   string     `json:"difficulty,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	Points       int        `json:"points"`
	Explanation  string     `json:"explanation"`
	Answers      []AIAnswer `json:"answers"`
	ImageData    string     `json:"image_data,omitempty"`
}

// AIEnrichmentResponse represents AI metadata suggestions for an existing quiz.
type AIEnrichmentResponse struct {
	Quiz      AIQuizEnrichment       `json:"quiz"`
	Questions []AIQuestionEnrichment `json:"questions"`
}

// AIQuizEnrichment contains quiz-level enrichment fields.
type AIQuizEnrichment struct {
	Difficulty  string   `json:"difficulty,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// AIQuestionEnrichment contains question-level enrichment fields keyed by question ID.
type AIQuestionEnrichment struct {
	ID          string   `json:"id"`
	Difficulty  string   `json:"difficulty,omitempty"`
	Explanation string   `json:"explanation,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// AIAnswer represents a single answer option in the AI response
type AIAnswer struct {
	AnswerText string `json:"answer_text"`
	IsCorrect  bool   `json:"is_correct"`
}

// GenerateQuizFromTextRequest represents API request for text-based generation
type GenerateQuizFromTextRequest struct {
	Text            string         `json:"text" binding:"required,min=100,max=50000"`
	Options         QuizGenOptions `json:"options" binding:"required"`
	Metadata        QuizMetadata   `json:"metadata,omitempty"`
	ReferenceImages []string       `json:"reference_images,omitempty"` // File IDs of reference images
}

// GenerateQuizFromFileRequest represents API request for file-based generation
type GenerateQuizFromFileRequest struct {
	FileID          string         `json:"file_id" binding:"required,uuid"`
	Options         QuizGenOptions `json:"options" binding:"required"`
	Metadata        QuizMetadata   `json:"metadata,omitempty"`
	ReferenceImages []string       `json:"reference_images,omitempty"` // File IDs of reference images
}

// GenerationStatusResponse represents the response for status check
type GenerationStatusResponse struct {
	Status          string  `json:"status"`
	Progress        int     `json:"progress"`
	QuizID          *string `json:"quiz_id,omitempty"`
	ErrorMessage    *string `json:"error_message,omitempty"`
	EstimatedTimeMs *int    `json:"estimated_time_ms,omitempty"`
}

// FileUploadResponse represents the response after successful file upload
type FileUploadResponse struct {
	FileID   string `json:"file_id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
}

// ListFilesResponse represents paginated file list response
type ListFilesResponse struct {
	Files []*FileUpload `json:"files"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

// ListJobsResponse represents paginated job list response
type ListJobsResponse struct {
	Jobs  []*GenerationJob `json:"jobs"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// ExtractQuizFromImageRequest represents API request for quiz extraction from image/PDF
type ExtractQuizFromImageRequest struct {
	ExtractImages bool `json:"extract_images"` // Whether to extract and attach images from questions
}
