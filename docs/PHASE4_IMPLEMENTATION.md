# Phase 4 Implementation Document: AI Integration & File Upload

**Project:** Quiz-Craft Backend
**Phase:** 4 - AI Integration & File Upload
**Start Date:** January 10, 2026
**Estimated Duration:** 7-8 days
**Status:** 🟡 Ready for Review

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture Decisions](#architecture-decisions)
3. [Database Schema](#database-schema)
4. [Module Structure](#module-structure)
5. [Implementation Timeline](#implementation-timeline)
6. [Security Considerations](#security-considerations)
7. [Testing Strategy](#testing-strategy)
8. [Deployment Checklist](#deployment-checklist)
9. [API Documentation](#api-documentation)
10. [Risk Assessment](#risk-assessment)

---

## 1. Overview

### 1.1 Objectives

Phase 4 introduces **AI-powered quiz generation** using Google's Gemini AI. This allows users to:
- Upload documents (PDF, DOCX, TXT, MD)
- Extract text content automatically
- Generate quizzes with AI-created questions and answers
- Review and edit AI-generated content before publishing

### 1.2 Key Features

✅ **File Upload System**
- Support for PDF, DOCX, TXT, MD files
- Max file size: 50MB
- Local filesystem storage
- 7-day retention policy

✅ **File Parsing**
- PDF text extraction
- DOCX text extraction
- Plain text/Markdown reading
- Error handling for corrupted files

✅ **Google AI Integration**
- Gemini 1.5 Pro model
- Custom prompt engineering
- JSON response parsing
- Rate limiting and cost control

✅ **Async Processing**
- Background job queue (database-backed)
- Progress tracking (0-100%)
- Status polling endpoints
- Graceful error handling

✅ **Quiz Generation**
- Validates AI responses
- Creates quiz + questions + answers
- Database transactions (rollback on error)
- Draft status for review

---

## 2. Architecture Decisions

### 2.1 Module Separation

**Decision:** Create separate `ai-agent` module (not part of `quiz` module)

**Rationale:**
- Clean separation of concerns
- Easy to mock for testing
- Can scale independently
- Isolates AI-specific costs and rate limits
- Future-proof for adding more AI features

**Structure:**
```
backend/
├── user/           (existing)
├── quiz/           (existing - Phase 3)
├── ai-agent/       ← NEW MODULE
│   ├── model/
│   ├── dao/
│   ├── services/
│   ├── core/
│   └── worker/
└── api/
    └── ai.go       ← NEW API HANDLERS
```

### 2.2 File Storage

**Decision:** Local filesystem (`/uploads`) for Phase 4

**Rationale:**
- Simple to implement
- No cloud setup required
- Good for MVP/development
- Easy to migrate to S3/GCS later
- Metadata stored in PostgreSQL

**Storage Path:** `/uploads/{user_id}/{filename}`

**Migration Path:**
- Phase 4: Local filesystem
- Phase 6+: Cloud storage (S3/GCS) when scaling needed

### 2.3 Async Processing

**Decision:** Database-backed job queue (not Redis)

**Rationale:**
- No external dependencies
- Simple for MVP
- Uses existing PostgreSQL
- Scales to hundreds of concurrent jobs
- Can migrate to Redis later if needed

**How it works:**
1. API creates job in `generation_jobs` table (status: pending)
2. Background worker polls table every 5 seconds
3. Worker processes job, updates progress
4. Client polls status endpoint
5. Job completes, quiz_id returned

### 2.4 Quiz Status

**Decision:** AI-generated quizzes start as `draft` with special `ai_generated` flag

**Rationale:**
- User can review AI output before publishing
- Prevents publishing low-quality quizzes
- User can edit questions/answers
- Maintains trust in platform

**Quiz Status Flow:**
```
AI generates → draft + ai_generated flag → User reviews/edits → User publishes
```

---

## 3. Database Schema

### 3.1 New Tables

#### Table: `uploaded_files`

**Purpose:** Track user file uploads

```sql
CREATE TABLE uploaded_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'uploaded',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_uploaded_files_user_id ON uploaded_files(user_id);
CREATE INDEX idx_uploaded_files_status ON uploaded_files(status);
CREATE INDEX idx_uploaded_files_expires_at ON uploaded_files(expires_at)
    WHERE status = 'processed';
```

**Fields:**
- `id` - UUID primary key
- `user_id` - Owner of the file
- `filename` - Generated filename (timestamp + random)
- `original_filename` - User's original filename
- `file_path` - Full path on disk
- `file_size` - Size in bytes
- `mime_type` - MIME type for validation
- `status` - uploaded | processing | processed | failed
- `error_message` - Error details if failed
- `expires_at` - Auto-delete timestamp (7 days)

#### Table: `generation_jobs`

**Purpose:** Track async quiz generation jobs

```sql
CREATE TABLE generation_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_id UUID REFERENCES uploaded_files(id) ON DELETE SET NULL,
    text_content TEXT,
    options JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    progress INT DEFAULT 0,
    quiz_id UUID,
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_generation_jobs_user_id ON generation_jobs(user_id);
CREATE INDEX idx_generation_jobs_status ON generation_jobs(status);
CREATE INDEX idx_generation_jobs_pending ON generation_jobs(status)
    WHERE status = 'pending';
```

**Fields:**
- `id` - UUID primary key (job_id)
- `user_id` - Owner of the generation request
- `file_id` - Reference to uploaded_files (if from file)
- `text_content` - Raw text (if from text input)
- `options` - JSONB with QuizGenOptions (num_questions, difficulty, etc.)
- `status` - pending | processing | completed | failed
- `progress` - 0-100 percentage
- `quiz_id` - Set when quiz is successfully created
- `error_message` - Error details if failed
- `started_at` - When worker picked up job
- `completed_at` - When job finished

### 3.2 Migration Files

**Create migrations:**
- `000006_create_uploaded_files_table.up.sql` / `.down.sql`
- `000007_create_generation_jobs_table.up.sql` / `.down.sql`

**Migration order:**
1. `000006` creates `uploaded_files` (referenced by `generation_jobs`)
2. `000007` creates `generation_jobs` (references `uploaded_files`)

### 3.3 Quiz Table Modification

**No changes needed!** The existing `quizzes` table already has:
- `meta` JSONB field → can add `{"ai_generated": true}`
- `is_published` boolean → AI quizzes start with `false`
- All required fields for AI-generated quizzes

---

## 4. Module Structure

### 4.1 Directory Layout

```
ai-agent/
├── model/
│   └── ai_models.go           (Data structures)
├── dao/
│   ├── file_dao.go            (uploaded_files CRUD)
│   └── job_dao.go             (generation_jobs CRUD)
├── services/
│   ├── file_upload.go         (File handling)
│   ├── file_parser.go         (PDF/DOCX parsing)
│   └── quiz_generator.go      (Quiz creation from AI)
├── core/
│   └── ai_service.go          (Google AI integration)
└── worker/
    └── quiz_generator_worker.go (Background processor)
```

### 4.2 File: `ai-agent/model/ai_models.go`

**Models to create:**

```go
// QuizGenOptions represents options for AI quiz generation
type QuizGenOptions struct {
    NumQuestions   int    `json:"num_questions"`   // 5-50
    Difficulty     string `json:"difficulty"`      // easy, medium, hard
    QuestionType   string `json:"question_type"`   // multiple_choice, true_false, mixed
    Category       string `json:"category"`
    Language       string `json:"language"`        // en, es, fr, etc.
}

// FileUpload represents uploaded file metadata
type FileUpload struct {
    ID               string
    UserID           string
    Filename         string
    OriginalFilename string
    FilePath         string
    FileSize         int64
    MimeType         string
    Status           string
    ErrorMessage     string
    CreatedAt        time.Time
    ExpiresAt        *time.Time
}

// GenerationJob represents an async quiz generation job
type GenerationJob struct {
    ID           string
    UserID       string
    FileID       string
    TextContent  string
    Options      QuizGenOptions
    Status       string
    Progress     int
    QuizID       string
    ErrorMessage string
    StartedAt    *time.Time
    CompletedAt  *time.Time
    CreatedAt    time.Time
}

// AIQuizResponse represents the expected JSON from Google AI
type AIQuizResponse struct {
    Title       string       `json:"title"`
    Description string       `json:"description"`
    Category    string       `json:"category"`
    Difficulty  string       `json:"difficulty"`
    Questions   []AIQuestion `json:"questions"`
}

// AIQuestion represents a question from AI
type AIQuestion struct {
    QuestionText string     `json:"question_text"`
    QuestionType string     `json:"question_type"`
    Points       int        `json:"points"`
    Explanation  string     `json:"explanation"`
    Answers      []AIAnswer `json:"answers"`
}

// AIAnswer represents an answer from AI
type AIAnswer struct {
    AnswerText string `json:"answer_text"`
    IsCorrect  bool   `json:"is_correct"`
}
```

### 4.3 File: `ai-agent/dao/file_dao.go`

**Methods to implement:**

```go
func Init(database *sql.DB)
func CreateFile(ctx context.Context, file *model.FileUpload) error
func GetFileByID(ctx context.Context, id uuid.UUID) (*model.FileUpload, error)
func ListFilesByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.FileUpload, int, error)
func UpdateFileStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error
func DeleteFile(ctx context.Context, id uuid.UUID) error
func GetExpiredFiles(ctx context.Context) ([]*model.FileUpload, error)
```

### 4.4 File: `ai-agent/dao/job_dao.go`

**Methods to implement:**

```go
func CreateJob(ctx context.Context, job *model.GenerationJob) error
func GetJobByID(ctx context.Context, id uuid.UUID) (*model.GenerationJob, error)
func ListJobsByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]*model.GenerationJob, int, error)
func GetPendingJobs(ctx context.Context, limit int) ([]*model.GenerationJob, error)
func UpdateJobProgress(ctx context.Context, id uuid.UUID, progress int) error
func UpdateJobStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error
func CompleteJob(ctx context.Context, id uuid.UUID, quizID string) error
```

### 4.5 File: `ai-agent/services/file_upload.go`

**Functions to implement:**

```go
func ValidateFile(filename string, size int64, mimeType string) error
func SanitizeFilename(filename string) string
func GenerateFilename(userID string, originalFilename string) string
func SaveFile(userID string, filename string, fileData []byte) (string, error)
func DeleteFileFromDisk(filePath string) error
```

### 4.6 File: `ai-agent/services/file_parser.go`

**Functions to implement:**

```go
func ParsePDF(filePath string) (string, error)
func ParseDOCX(filePath string) (string, error)
func ParseText(filePath string) (string, error)
func ParseMarkdown(filePath string) (string, error)
func DetectFileType(filePath string, mimeType string) (string, error)
```

### 4.7 File: `ai-agent/core/ai_service.go`

**Functions to implement:**

```go
func Init(apiKey string, model string)
func GenerateQuizFromText(ctx context.Context, text string, options model.QuizGenOptions) (*model.AIQuizResponse, error)
func buildPrompt(text string, options model.QuizGenOptions) string
func parseAIResponse(response string) (*model.AIQuizResponse, error)
func validateAIResponse(response *model.AIQuizResponse) error
```

### 4.8 File: `ai-agent/services/quiz_generator.go`

**Functions to implement:**

```go
func CreateQuizFromAIResponse(ctx context.Context, userID string, aiResponse *model.AIQuizResponse) (string, error)
func validateQuizStructure(aiResponse *model.AIQuizResponse) error
func createQuizTransaction(ctx context.Context, userID string, aiResponse *model.AIQuizResponse) (string, error)
```

### 4.9 File: `ai-agent/worker/quiz_generator_worker.go`

**Functions to implement:**

```go
func StartQuizGeneratorWorker(ctx context.Context)
func ProcessPendingJobs(ctx context.Context)
func ProcessJob(ctx context.Context, job *model.GenerationJob)
```

---

## 5. Implementation Timeline

### Day 1: Database & DAOs (Friday, Jan 10)

**Tasks:**
- [x] Read this implementation doc
- [ ] Create module structure: `mkdir -p ai-agent/{model,dao,services,core,worker}`
- [ ] Create migration `000006_create_uploaded_files_table.up.sql`
- [ ] Create migration `000006_create_uploaded_files_table.down.sql`
- [ ] Create migration `000007_create_generation_jobs_table.up.sql`
- [ ] Create migration `000007_create_generation_jobs_table.down.sql`
- [ ] Test migrations: Fix dirty state, run migrations
- [ ] Create `ai-agent/model/ai_models.go` with all structs
- [ ] Create `ai-agent/dao/file_dao.go` with all methods
- [ ] Create `ai-agent/dao/job_dao.go` with all methods
- [ ] Add DAO init to `main.go`
- [ ] Write unit tests for DAOs
- [ ] Test build: `go build ./...`

**Acceptance Criteria:**
- Both migrations run successfully
- DAOs compile without errors
- Can create/read records in both tables
- Unit tests pass

**Estimated Time:** 4-5 hours

---

### Day 2: File Upload & Storage (Saturday, Jan 11)

**Tasks:**
- [ ] Create `/uploads` directory: `mkdir -p uploads && chmod 755 uploads`
- [ ] Add to `.env`:
  ```bash
  UPLOAD_MAX_FILE_SIZE=52428800
  UPLOAD_ALLOWED_TYPES=pdf,docx,txt,md
  UPLOAD_STORAGE_PATH=./uploads
  UPLOAD_RETENTION_DAYS=7
  ```
- [ ] Create `ai-agent/services/file_upload.go` with all functions
- [ ] Create `api/ai.go` with:
  - `POST /api/v1/files/upload`
  - `GET /api/v1/files`
  - `DELETE /api/v1/files/:id`
- [ ] Add routes to `main.go`:
  ```go
  files := v1.Group("/files")
  files.Use(middleware.Auth(jwtManager))
  {
      files.POST("/upload", api.UploadFile)
      files.GET("", api.ListFiles)
      files.DELETE("/:id", api.DeleteFile)
  }
  ```
- [ ] Test with curl:
  ```bash
  # Upload file
  curl -X POST -F "file=@test.pdf" \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/v1/files/upload

  # List files
  curl -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/v1/files
  ```

**Acceptance Criteria:**
- Can upload PDF/DOCX/TXT/MD files
- File saved to `/uploads/{user_id}/`
- Metadata saved to database
- Invalid files rejected
- Can list and delete files

**Estimated Time:** 4-5 hours

---

### Day 3: File Parsing (Sunday, Jan 12)

**Tasks:**
- [ ] Install libraries:
  ```bash
  go get github.com/ledongthuc/pdf
  go get github.com/nguyenthenguyen/docx
  ```
- [ ] Create `ai-agent/services/file_parser.go` with all functions
- [ ] Implement `ParsePDF()`
- [ ] Implement `ParseDOCX()`
- [ ] Implement `ParseText()`
- [ ] Implement `ParseMarkdown()`
- [ ] Write unit tests with sample files
- [ ] Test with corrupted/encrypted files
- [ ] Add error handling for edge cases

**Test Files:**
- `test_sample.pdf` - regular PDF
- `test_encrypted.pdf` - password-protected
- `test_empty.pdf` - empty PDF
- `test_sample.docx` - regular DOCX
- `test_large.docx` - large file (>50MB)

**Acceptance Criteria:**
- Successfully extracts text from valid PDFs
- Successfully extracts text from valid DOCX
- Handles encrypted PDFs gracefully
- Handles empty documents gracefully
- Returns error for corrupted files

**Estimated Time:** 4-5 hours

---

### Day 4: Google AI Integration (Monday, Jan 13)

**Tasks:**
- [ ] Install library:
  ```bash
  go get cloud.google.com/go/ai/generativelanguage
  go get cloud.google.com/go/genai
  ```
- [ ] Add to `.env`:
  ```bash
  GOOGLE_AI_API_KEY=AIzaSyDkTtFFjLXOkrlW6JRfgxXu3kuXMsIjptc
  GOOGLE_AI_MODEL=gemini-1.5-pro
  GOOGLE_AI_TIMEOUT=60s
  ```
- [ ] Create `ai-agent/core/ai_service.go`
- [ ] Implement `Init()` - initialize AI client
- [ ] Implement `GenerateQuizFromText()` - main function
- [ ] Implement `buildPrompt()` - create AI prompt
- [ ] Implement `parseAIResponse()` - parse JSON
- [ ] Implement `validateAIResponse()` - validate structure
- [ ] Test with sample text:
  ```go
  text := "The Pythagorean theorem states that a² + b² = c²..."
  options := QuizGenOptions{
      NumQuestions: 5,
      Difficulty: "medium",
      QuestionType: "multiple_choice",
      Category: "Mathematics",
  }
  response, err := GenerateQuizFromText(ctx, text, options)
  ```
- [ ] Log all AI requests/responses
- [ ] Handle timeouts (retry once)
- [ ] Handle rate limits (return error)

**Acceptance Criteria:**
- Successfully calls Google AI API
- Receives valid JSON response
- Parses response into AIQuizResponse
- Handles API errors gracefully
- Logs all requests for debugging

**Estimated Time:** 5-6 hours

---

### Day 5: Quiz Generator & Worker (Tuesday, Jan 14)

**Tasks:**
- [ ] Create `ai-agent/services/quiz_generator.go`
- [ ] Implement `CreateQuizFromAIResponse()` - main function
- [ ] Implement `validateQuizStructure()` - validate quiz
- [ ] Implement `createQuizTransaction()` - DB transaction:
  ```go
  BEGIN TRANSACTION
    1. Create quiz (use quiz/dao)
    2. Create questions (use quiz/dao)
    3. Create answers (use quiz/dao)
    4. Set quiz meta: {"ai_generated": true}
  COMMIT or ROLLBACK
  ```
- [ ] Create `ai-agent/worker/quiz_generator_worker.go`
- [ ] Implement `StartQuizGeneratorWorker()` - start background worker
- [ ] Implement `ProcessPendingJobs()` - fetch and process
- [ ] Implement `ProcessJob()` - full pipeline:
  ```go
  1. Update status = 'processing', progress = 10%
  2. Get file or text content
  3. Parse file if needed, progress = 30%
  4. Call AI service, progress = 60%
  5. Parse AI response, progress = 80%
  6. Create quiz in DB, progress = 90%
  7. Update status = 'completed', progress = 100%, set quiz_id
  ```
- [ ] Add to `main.go`:
  ```go
  // Start background worker
  go aiworker.StartQuizGeneratorWorker(context.Background())
  ```
- [ ] Test end-to-end:
  ```bash
  # Create job
  curl -X POST -H "Authorization: Bearer $TOKEN" \
    -d '{"file_id":"...","options":{...}}' \
    http://localhost:8080/api/v1/ai/generate-quiz-from-file

  # Check status
  curl -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/v1/ai/generation-status/:job_id
  ```

**Acceptance Criteria:**
- Worker starts with server
- Polls jobs table every 5 seconds
- Processes jobs asynchronously
- Updates progress correctly
- Creates quiz successfully
- Handles errors (rolls back transaction)

**Estimated Time:** 6-7 hours

---

### Day 6: API Endpoints (Wednesday, Jan 15)

**Tasks:**
- [ ] Complete `api/ai.go` with all endpoints:
  - `POST /api/v1/ai/generate-quiz-from-text`
  - `POST /api/v1/ai/generate-quiz-from-file`
  - `GET /api/v1/ai/generation-status/:id`
  - `GET /api/v1/ai/jobs`
- [ ] Add routes to `main.go`:
  ```go
  ai := v1.Group("/ai")
  ai.Use(middleware.Auth(jwtManager))
  {
      ai.POST("/generate-quiz-from-text", api.GenerateQuizFromText)
      ai.POST("/generate-quiz-from-file", api.GenerateQuizFromFile)
      ai.GET("/generation-status/:id", api.GetGenerationStatus)
      ai.GET("/jobs", api.ListGenerationJobs)
  }
  ```
- [ ] Test each endpoint with Postman:
  - Generate from text
  - Generate from file
  - Check status
  - List jobs
  - Error cases
- [ ] Add proper error handling
- [ ] Add logging for all endpoints
- [ ] Document request/response formats

**Acceptance Criteria:**
- All endpoints work correctly
- Returns 202 Accepted for async operations
- Status endpoint shows progress
- Jobs endpoint lists user's jobs
- Error responses are clear

**Estimated Time:** 4-5 hours

---

### Day 7: Rate Limiting & Cleanup (Thursday, Jan 16)

**Tasks:**
- [ ] Create `internal/middleware/rate_limiter.go`
- [ ] Implement in-memory rate limiting:
  ```go
  type RateLimiter struct {
      uploadLimits map[string]*TokenBucket  // user_id -> bucket
      aiLimits     map[string]*TokenBucket
  }
  ```
- [ ] Add rate limit middleware:
  - 10 uploads per hour per user
  - 20 AI generations per day per user
- [ ] Create quota tracking table (optional):
  ```sql
  CREATE TABLE user_quotas (
      user_id UUID PRIMARY KEY,
      uploads_today INT DEFAULT 0,
      ai_generations_today INT DEFAULT 0,
      last_reset TIMESTAMP
  );
  ```
- [ ] Create file cleanup cron job:
  ```go
  func StartFileCleanupWorker(ctx context.Context) {
      ticker := time.NewTicker(24 * time.Hour)
      for {
          select {
          case <-ctx.Done():
              return
          case <-ticker.C:
              CleanupExpiredFiles(ctx)
          }
      }
  }
  ```
- [ ] Add to `main.go`:
  ```go
  go aiworker.StartFileCleanupWorker(context.Background())
  ```
- [ ] Create admin endpoint:
  - `GET /api/v1/admin/ai-usage-stats`
- [ ] Log AI API usage to file:
  ```
  logs/ai-usage-YYYY-MM-DD.log
  ```

**Acceptance Criteria:**
- Rate limiting works (returns 429)
- File cleanup runs daily
- Old files are deleted
- Admin can view usage stats
- AI usage is logged

**Estimated Time:** 4-5 hours

---

### Day 8: Testing & Documentation (Friday, Jan 17)

**Tasks:**
- [ ] Write unit tests:
  - `ai-agent/services/file_upload_test.go`
  - `ai-agent/services/file_parser_test.go`
  - `ai-agent/core/ai_service_test.go`
  - `ai-agent/services/quiz_generator_test.go`
- [ ] Write integration tests:
  - Upload file → generate quiz → verify result
  - Generate from text → verify result
  - Test rate limiting
  - Test concurrent jobs
- [ ] Test error scenarios:
  - AI timeout
  - Invalid file format
  - Corrupted PDF
  - Empty content
  - Database transaction failure
  - Worker crash and recovery
- [ ] Manual testing:
  - Upload various file types
  - Generate quizzes with different options
  - Verify quiz quality
  - Test rate limiting
- [ ] Update documentation:
  - Update `docs/CURRENT_STATE.md`
  - Update `docs/API_DOCUMENTATION.md`
  - Update `README.md`
- [ ] Run full test suite:
  ```bash
  go test ./... -v
  ```

**Acceptance Criteria:**
- All unit tests pass
- All integration tests pass
- Manual testing successful
- AI generates good quality quizzes
- Documentation updated

**Estimated Time:** 6-7 hours

---

## 6. Security Considerations

### 6.1 API Key Security

✅ **Implemented:**
- API key stored in `.env` (not committed to git)
- Never logged or exposed in responses
- Server-side only (never sent to client)

⚠️ **To Do:**
- Rotate key periodically
- Set up Google Cloud budget alerts
- Monitor usage for anomalies

### 6.2 File Upload Security

✅ **Implemented:**
- File type whitelist (PDF, DOCX, TXT, MD only)
- File size limit (50MB)
- Filename sanitization (remove special chars)
- User can only access their own files

⚠️ **To Do:**
- Consider virus scanning (ClamAV)
- Add file hash to prevent duplicates
- Rate limit uploads (10/hour)

### 6.3 Input Validation

✅ **Implemented:**
- SQL injection prevention (parameterized queries)
- XSS prevention (escape user input)
- Content length validation (100-50,000 chars)

### 6.4 Authorization

✅ **Implemented:**
- JWT authentication required
- User can only access own files/jobs
- Quiz ownership verified before editing

---

## 7. Testing Strategy

### 7.1 Unit Tests

**Coverage Target:** 70%+

**Test Files:**
- `ai-agent/services/file_upload_test.go`
- `ai-agent/services/file_parser_test.go`
- `ai-agent/dao/file_dao_test.go`
- `ai-agent/dao/job_dao_test.go`

**Test Cases:**
- File validation (valid/invalid types)
- Filename sanitization
- PDF parsing (valid/corrupted/encrypted)
- DOCX parsing
- JSON parsing (valid/malformed)
- Quiz structure validation

### 7.2 Integration Tests

**Test Files:**
- `ai-agent/integration_test.go`

**Test Cases:**
- End-to-end: Upload → Parse → AI → Create Quiz
- Rate limiting enforcement
- Concurrent job processing
- Transaction rollback on error

### 7.3 Manual Tests

**Test Scenarios:**
1. Upload various file types (PDF, DOCX, TXT, MD)
2. Generate quiz from 1-page PDF
3. Generate quiz from 10-page PDF
4. Generate quiz from text (500 words)
5. Generate quiz with different difficulties
6. Test rate limiting (upload 11 files in 1 hour)
7. Test concurrent generations (5 users at once)
8. Verify quiz quality (check questions make sense)

---

## 8. Deployment Checklist

### 8.1 Prerequisites

Before deploying:
- [ ] Google Cloud account created
- [ ] Gemini API enabled
- [ ] API key obtained: `AIzaSyDkTtFFjLXOkrlW6JRfgxXu3kuXMsIjptc`
- [ ] Budget alerts configured in Google Cloud

### 8.2 Environment Setup

**Add to `.env`:**
```bash
# Google AI Configuration
GOOGLE_AI_API_KEY=AIzaSyDkTtFFjLXOkrlW6JRfgxXu3kuXMsIjptc
GOOGLE_AI_MODEL=gemini-1.5-pro
GOOGLE_AI_TIMEOUT=60s
GOOGLE_AI_MAX_TOKENS=8000

# File Upload Configuration
UPLOAD_MAX_FILE_SIZE=52428800  # 50MB
UPLOAD_ALLOWED_TYPES=pdf,docx,txt,md
UPLOAD_STORAGE_PATH=./uploads
UPLOAD_RETENTION_DAYS=7

# Rate Limiting
RATE_LIMIT_UPLOAD_PER_HOUR=10
RATE_LIMIT_AI_PER_DAY=20

# Feature Flags
FEATURE_AI_GENERATION=true
FEATURE_FILE_UPLOAD=true
```

### 8.3 Database Migration

```bash
# Ensure clean migration state
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft \
  -c "SELECT * FROM schema_migrations;"

# If dirty, fix it:
docker exec -it quiz-craft-postgres psql -U postgres -d quiz_craft \
  -c "UPDATE schema_migrations SET dirty = false;"

# Run migrations
go run main.go  # Migrations run automatically
```

### 8.4 File System Setup

```bash
# Create uploads directory
mkdir -p uploads
chmod 755 uploads

# Verify permissions
ls -la uploads/
```

### 8.5 Smoke Tests

After deployment:
```bash
# 1. Test file upload
curl -X POST -F "file=@test.pdf" \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/files/upload

# 2. Test quiz generation
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text":"Sample text","options":{"num_questions":5}}' \
  http://localhost:8080/api/v1/ai/generate-quiz-from-text

# 3. Check worker is running
curl http://localhost:8080/api/v1/health
```

---

## 9. API Documentation

### 9.1 Upload File

```http
POST /api/v1/files/upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: <binary>
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "file_id": "uuid",
    "filename": "user123_1736524800_abc123.pdf",
    "original_filename": "notes.pdf",
    "size": 1048576,
    "mime_type": "application/pdf"
  }
}
```

### 9.2 Generate Quiz from Text

```http
POST /api/v1/ai/generate-quiz-from-text
Authorization: Bearer {token}
Content-Type: application/json

{
  "text": "Your content here...",
  "options": {
    "num_questions": 10,
    "difficulty": "medium",
    "question_type": "multiple_choice",
    "category": "Science",
    "language": "en"
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "message": "Quiz generation started",
  "data": {
    "job_id": "uuid",
    "status": "pending",
    "estimated_time_seconds": 30
  }
}
```

### 9.3 Generate Quiz from File

```http
POST /api/v1/ai/generate-quiz-from-file
Authorization: Bearer {token}
Content-Type: application/json

{
  "file_id": "uuid",
  "options": {
    "num_questions": 15,
    "difficulty": "hard",
    "question_type": "mixed",
    "category": "History"
  }
}
```

**Response (202 Accepted):**
```json
{
  "success": true,
  "message": "Quiz generation started",
  "data": {
    "job_id": "uuid",
    "status": "pending",
    "estimated_time_seconds": 45
  }
}
```

### 9.4 Check Generation Status

```http
GET /api/v1/ai/generation-status/:job_id
Authorization: Bearer {token}
```

**Response (200 OK - Processing):**
```json
{
  "success": true,
  "data": {
    "job_id": "uuid",
    "status": "processing",
    "progress": 60,
    "started_at": "2026-01-10T15:30:00Z"
  }
}
```

**Response (200 OK - Completed):**
```json
{
  "success": true,
  "data": {
    "job_id": "uuid",
    "status": "completed",
    "progress": 100,
    "quiz_id": "uuid",
    "completed_at": "2026-01-10T15:31:00Z"
  }
}
```

**Response (200 OK - Failed):**
```json
{
  "success": true,
  "data": {
    "job_id": "uuid",
    "status": "failed",
    "progress": 40,
    "error_message": "AI timeout: request took longer than 60s"
  }
}
```

### 9.5 List User's Generation Jobs

```http
GET /api/v1/ai/jobs?page=1&limit=20
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "jobs": [
      {
        "job_id": "uuid",
        "status": "completed",
        "progress": 100,
        "quiz_id": "uuid",
        "created_at": "2026-01-10T15:30:00Z",
        "completed_at": "2026-01-10T15:31:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 20
  }
}
```

### 9.6 List User's Files

```http
GET /api/v1/files?page=1&limit=20
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "files": [
      {
        "file_id": "uuid",
        "original_filename": "notes.pdf",
        "size": 1048576,
        "mime_type": "application/pdf",
        "status": "processed",
        "created_at": "2026-01-10T15:00:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "limit": 20
  }
}
```

### 9.7 Delete File

```http
DELETE /api/v1/files/:file_id
Authorization: Bearer {token}
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "File deleted successfully"
}
```

---

## 10. Risk Assessment

### 10.1 High Priority Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **AI generates poor quality quizzes** | High | Medium | Draft status for review, user can edit before publishing |
| **Google AI API cost overrun** | High | Medium | Rate limiting (20/day per user), monitor usage, budget alerts |
| **File parsing fails (encrypted/corrupted)** | Medium | High | Graceful error handling, suggest alternative formats |
| **Worker crashes and jobs stuck** | Medium | Low | Worker restarts automatically, jobs remain in pending state |
| **Database connection lost during transaction** | High | Low | Transaction rollback, retry mechanism |

### 10.2 Medium Priority Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Large file upload timeouts** | Medium | Medium | Increase timeout for large files, chunk processing |
| **Concurrent job processing overload** | Medium | Low | Limit to 5 concurrent jobs, queue others |
| **File storage fills disk** | Medium | Medium | Cleanup cron job (7-day retention), monitor disk space |
| **AI response malformed JSON** | Low | Medium | Validation before parsing, retry with modified prompt |

### 10.3 Contingency Plans

**If AI API is unavailable:**
1. Show user-friendly error: "AI service temporarily unavailable"
2. Jobs remain in `pending` state
3. Worker retries periodically
4. User can cancel and retry later

**If worker crashes:**
1. Worker restarts with server
2. Picks up pending jobs automatically
3. No data loss (jobs in database)

**If file storage full:**
1. Reject new uploads with clear error
2. Manual cleanup of old files
3. Monitor alert to admin

---

## 11. Success Metrics

Phase 4 will be considered successful when:

✅ **Functional:**
- Users can upload PDF/DOCX/TXT/MD files (50MB max)
- AI generates quizzes from uploaded files
- AI generates quizzes from pasted text
- Users can review AI-generated quizzes before publishing
- Background worker processes jobs asynchronously
- Progress tracking works (0-100%)

✅ **Performance:**
- File upload < 5 seconds
- AI generation < 60 seconds (per quiz)
- Worker polls every 5 seconds
- Can handle 5 concurrent jobs

✅ **Quality:**
- 80%+ of AI-generated questions are usable
- Questions are grammatically correct
- Correct answers are actually correct
- No duplicate questions in same quiz

✅ **Reliability:**
- Rate limiting prevents abuse
- File cleanup works (7-day retention)
- Transaction rollback prevents partial data
- Worker recovers from crashes

✅ **Documentation:**
- CURRENT_STATE.md updated
- API documentation complete
- README includes Phase 4 info

---

## 12. Post-Implementation

### 12.1 Monitoring

**Metrics to track:**
- AI API calls per day
- AI API cost per day
- File uploads per day
- Quiz generation success rate
- Quiz generation failure rate
- Average generation time
- Storage usage (disk space)

**Logging:**
- Log all AI requests/responses
- Log file uploads
- Log job status changes
- Log errors with context

### 12.2 Future Enhancements (Phase 6+)

- Migrate to cloud storage (S3/GCS)
- Add Redis for job queue (scale beyond 100 concurrent jobs)
- Add OCR for scanned PDFs
- Support more file types (PPT, HTML, etc.)
- Add quiz quality scoring
- Allow regenerating specific questions
- Add AI feedback mechanism (improve prompts)
- Add quiz templates for common categories

---

## 13. Review Checklist

**Before starting implementation, verify:**

- [ ] I understand the architecture (separate ai-agent module)
- [ ] I have the Google AI API key ready
- [ ] I understand async processing with database queue
- [ ] I know the file storage location (/uploads)
- [ ] I understand the 8-day timeline
- [ ] I have reviewed all code structures in section 4
- [ ] I understand the database schema (section 3)
- [ ] I know what endpoints to create (section 9)
- [ ] I understand the security requirements (section 6)
- [ ] I understand the testing strategy (section 7)

**Questions for reviewer:**

1. Is the 7-8 day timeline realistic?
2. Should we use a different AI model?
3. Should we add more file types (PPT, HTML)?
4. Should rate limits be configurable per user tier?
5. Should we add virus scanning for uploads?
6. Any concerns about the architecture?

---

## 14. References

- **Plan Document:** `C:\Users\affan\.claude\plans\splendid-sauteeing-barto.md`
- **Implementation Plan:** `docs/IMPLEMENTATION_PLAN.md`
- **Current State:** `docs/CURRENT_STATE.md`
- **Google AI Docs:** https://cloud.google.com/vertex-ai/docs/generative-ai
- **Gemini API:** https://ai.google.dev/docs

---

**Ready for Implementation:** ✅ Awaiting approval to start Day 1

**Last Updated:** January 10, 2026, 5:15 PM
