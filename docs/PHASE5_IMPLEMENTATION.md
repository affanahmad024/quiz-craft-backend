# Phase 5 Implementation Document: Quiz Attempts & Scoring

**Date:** January 10, 2026
**Status:** Planning
**Estimated Duration:** 8 working days (40-45 hours)
**Risk Level:** Low

---

## Executive Summary

Phase 5 enables users to **take quizzes, submit answers, receive scores, and track their attempt history**. This phase transforms the quiz-craft platform from a quiz creation tool into a complete quiz-taking experience with automatic scoring and progress tracking.

### Key Features:
- ✅ Start quiz attempts with optional time limits
- ✅ Submit answers progressively (save as you go)
- ✅ Automatic scoring for multiple choice and true/false questions
- ✅ Time limit enforcement with auto-submit when expired
- ✅ Resume support (pause and continue later)
- ✅ Immediate results with correct answers and explanations
- ✅ Complete attempt history tracking
- ✅ Basic quiz statistics for creators

### Out of Scope (Deferred to Phase 6):
- ❌ Short answer question scoring (manual grading)
- ❌ Leaderboards or competitive features
- ❌ Detailed analytics dashboard
- ❌ Certificate generation
- ❌ Social features (sharing attempts)

---

## User Requirements Summary

Based on user clarification sessions:

| Requirement | Decision |
|-------------|----------|
| **Attempt Model** | Unlimited attempts - users can retake quizzes anytime |
| **Session Model** | Resume allowed - users can pause and continue, auto-submit on time expiry |
| **Submission** | Progressive - save each answer individually as user answers |
| **Results Display** | Immediate feedback via separate API after submission |
| **Leaderboard** | No - only personal attempt history |
| **Short Answers** | Skip for Phase 5 - only multiple choice and true/false |
| **Analytics** | Basic stats only - total attempts and average score |

---

## System Architecture

### Database Schema

Phase 5 adds **2 new tables**:

#### Table 1: `quiz_attempts`
```sql
CREATE TABLE quiz_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'in_progress',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    submitted_at TIMESTAMP WITH TIME ZONE,
    time_limit_seconds INT,
    expires_at TIMESTAMP WITH TIME ZONE,
    score DECIMAL(5,2),
    max_score INT,
    percentage DECIMAL(5,2),
    passed BOOLEAN,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CHECK (status IN ('in_progress', 'submitted', 'expired', 'abandoned')),
    CHECK (score >= 0 AND score <= max_score),
    CHECK (percentage >= 0 AND percentage <= 100)
);

CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_quiz_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_status ON quiz_attempts(status);
CREATE INDEX idx_quiz_attempts_user_quiz ON quiz_attempts(user_id, quiz_id, created_at DESC);
```

**Status Values**:
- `in_progress` - User is actively taking the quiz
- `submitted` - User finished and quiz is scored
- `expired` - Time limit reached, auto-submitted
- `abandoned` - Started but never submitted (cleanup status)

#### Table 2: `user_answers`
```sql
CREATE TABLE user_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,
    answer_text TEXT,
    is_correct BOOLEAN,
    points_earned DECIMAL(5,2) DEFAULT 0,
    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(attempt_id, question_id)
);

CREATE INDEX idx_user_answers_attempt_id ON user_answers(attempt_id);
CREATE INDEX idx_user_answers_question_id ON user_answers(question_id);
```

**Key Design Points**:
- `UNIQUE(attempt_id, question_id)` ensures one answer per question
- `answer_id` for MC/TF questions (references answers table)
- `answer_text` for future short answer support
- `is_correct` and `points_earned` calculated immediately on submission
- Upsert pattern for answer updates before final submission

---

### Module Structure

```
quiz/
├── model/
│   ├── quiz.go           (existing)
│   ├── question.go       (existing)
│   ├── answer.go         (existing)
│   └── attempt.go        ← NEW
├── dao/
│   ├── quiz.go           (existing)
│   ├── question.go       (existing)
│   ├── answer.go         (existing)
│   └── attempt.go        ← NEW
├── core/
│   ├── quiz.go           (existing)
│   └── attempt.go        ← NEW
└── worker/
    └── attempt_worker.go ← NEW

api/
└── attempt.go            ← NEW
```

---

## Data Models

### Core Models

```go
// QuizAttempt represents a user's attempt at a quiz
type QuizAttempt struct {
    ID               string     `json:"id" db:"id"`
    QuizID           string     `json:"quiz_id" db:"quiz_id"`
    UserID           string     `json:"user_id" db:"user_id"`
    Status           string     `json:"status" db:"status"`
    StartedAt        time.Time  `json:"started_at" db:"started_at"`
    SubmittedAt      *time.Time `json:"submitted_at,omitempty" db:"submitted_at"`
    TimeLimitSeconds *int       `json:"time_limit_seconds,omitempty" db:"time_limit_seconds"`
    ExpiresAt        *time.Time `json:"expires_at,omitempty" db:"expires_at"`
    Score            *float64   `json:"score,omitempty" db:"score"`
    MaxScore         *int       `json:"max_score,omitempty" db:"max_score"`
    Percentage       *float64   `json:"percentage,omitempty" db:"percentage"`
    Passed           *bool      `json:"passed,omitempty" db:"passed"`
    Meta             JSONMap    `json:"meta" db:"meta"`
    CreatedAt        time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// UserAnswer represents a user's answer to a question
type UserAnswer struct {
    ID           string     `json:"id" db:"id"`
    AttemptID    string     `json:"attempt_id" db:"attempt_id"`
    QuestionID   string     `json:"question_id" db:"question_id"`
    AnswerID     *string    `json:"answer_id,omitempty" db:"answer_id"`
    AnswerText   *string    `json:"answer_text,omitempty" db:"answer_text"`
    IsCorrect    *bool      `json:"is_correct,omitempty" db:"is_correct"`
    PointsEarned *float64   `json:"points_earned,omitempty" db:"points_earned"`
    AnsweredAt   time.Time  `json:"answered_at" db:"answered_at"`
    UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}
```

### API DTOs

```go
// Input Models
type StartAttemptInput struct {
    QuizID string `json:"quiz_id" binding:"required,uuid"`
}

type SubmitAnswerInput struct {
    QuestionID string  `json:"question_id" binding:"required,uuid"`
    AnswerID   *string `json:"answer_id" binding:"omitempty,uuid"`
    AnswerText *string `json:"answer_text"`
}

// Response Models
type AttemptDetailResponse struct {
    Attempt       QuizAttempt   `json:"attempt"`
    Quiz          QuizSummary   `json:"quiz"`
    Questions     []Question    `json:"questions"`
    UserAnswers   []UserAnswer  `json:"user_answers"`
    TimeRemaining *int          `json:"time_remaining_seconds,omitempty"`
}

type AttemptResultResponse struct {
    Attempt         QuizAttempt          `json:"attempt"`
    Quiz            QuizSummary          `json:"quiz"`
    Questions       []QuestionWithAnswer `json:"questions"`
    UserAnswers     []UserAnswer         `json:"user_answers"`
    CorrectCount    int                  `json:"correct_count"`
    TotalQuestions  int                  `json:"total_questions"`
}

type QuizStatsResponse struct {
    QuizID         string  `json:"quiz_id"`
    TotalAttempts  int     `json:"total_attempts"`
    AverageScore   float64 `json:"average_score"`
    PassRate       float64 `json:"pass_rate"`
    HighestScore   float64 `json:"highest_score"`
    LowestScore    float64 `json:"lowest_score"`
}
```

---

## API Endpoints

### 1. Start Quiz Attempt
```
POST /api/v1/quizzes/:id/attempts
```

**Authorization**: Required (JWT)

**Request Body**: None (quiz ID from URL)

**Response (201 Created)**:
```json
{
  "success": true,
  "message": "Quiz attempt started",
  "data": {
    "id": "attempt-uuid",
    "quiz_id": "quiz-uuid",
    "user_id": "user-uuid",
    "status": "in_progress",
    "started_at": "2026-01-10T18:00:00Z",
    "time_limit_seconds": 3600,
    "expires_at": "2026-01-10T19:00:00Z",
    "created_at": "2026-01-10T18:00:00Z"
  }
}
```

**Business Rules**:
- Quiz must exist and be published
- Quiz must be accessible (is_public OR user is creator)
- Time limit calculated if quiz has `time_limit_minutes`
- Creates `in_progress` attempt

**Error Codes**:
- `400` - quiz-not-published
- `403` - quiz-not-accessible
- `404` - quiz-not-found

---

### 2. Get Attempt Details
```
GET /api/v1/attempts/:id
```

**Authorization**: Required (must own attempt)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "attempt": { ... },
    "quiz": {
      "id": "quiz-uuid",
      "title": "Quiz Title",
      "description": "...",
      "time_limit_minutes": 60,
      "passing_score": 70
    },
    "questions": [
      {
        "id": "q1",
        "question_text": "What is 2+2?",
        "question_type": "multiple_choice",
        "points": 5,
        "answers": [
          {"id": "a1", "answer_text": "3"},
          {"id": "a2", "answer_text": "4"},
          {"id": "a3", "answer_text": "5"}
        ]
      }
    ],
    "user_answers": [
      {
        "question_id": "q1",
        "answer_id": "a2",
        "is_correct": true,
        "points_earned": 5.0
      }
    ],
    "time_remaining": 2400
  }
}
```

**Business Rules**:
- Does NOT show which answers are correct (for in-progress attempts)
- Shows time remaining if time limit set
- User must own the attempt

---

### 3. Submit Answer
```
POST /api/v1/attempts/:id/answers
```

**Authorization**: Required (must own attempt)

**Request Body**:
```json
{
  "question_id": "question-uuid",
  "answer_id": "answer-uuid"
}
```

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Answer submitted",
  "data": {
    "id": "user-answer-uuid",
    "attempt_id": "attempt-uuid",
    "question_id": "question-uuid",
    "answer_id": "answer-uuid",
    "is_correct": true,
    "points_earned": 10.0,
    "answered_at": "2026-01-10T18:05:00Z"
  }
}
```

**Business Rules**:
- Attempt must be `in_progress`
- Time limit must not be exceeded
- For MC/TF: `answer_id` required
- Answer scored immediately on submission
- Upserts (can change answer before finishing)

**Error Codes**:
- `400` - attempt-already-finished
- `400` - time-limit-exceeded
- `400` - answer-id-required
- `404` - question-not-in-quiz

---

### 4. Finish Attempt
```
POST /api/v1/attempts/:id/finish
```

**Authorization**: Required (must own attempt)

**Request Body**: None

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Quiz submitted successfully",
  "data": {
    "attempt": {
      "id": "attempt-uuid",
      "status": "submitted",
      "score": 85.0,
      "max_score": 100.0,
      "percentage": 85.0,
      "passed": true,
      "submitted_at": "2026-01-10T18:25:00Z"
    },
    "quiz": { ... },
    "questions": [ ... with correct answers ... ],
    "user_answers": [ ... ],
    "correct_count": 17,
    "total_questions": 20
  }
}
```

**Business Rules**:
- Attempt must be `in_progress`
- Calculates final score from all user answers
- Updates status to `submitted`
- Returns full results immediately

---

### 5. Get Attempt Results
```
GET /api/v1/attempts/:id/results
```

**Authorization**: Required (must own attempt)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "attempt": { ... with score },
    "quiz": { ... },
    "questions": [
      {
        "id": "q1",
        "question_text": "What is 2+2?",
        "question_type": "multiple_choice",
        "points": 5,
        "explanation": "Basic addition",
        "answers": [
          {"id": "a1", "answer_text": "3", "is_correct": false},
          {"id": "a2", "answer_text": "4", "is_correct": true},
          {"id": "a3", "answer_text": "5", "is_correct": false}
        ]
      }
    ],
    "user_answers": [ ... ],
    "correct_count": 17,
    "total_questions": 20
  }
}
```

**Business Rules**:
- Attempt must be `submitted` or `expired` (not `in_progress`)
- Shows correct answers and explanations
- Shows user's answers with correctness flags

---

### 6. List My Attempts
```
GET /api/v1/attempts?quiz_id=<uuid>&page=1&limit=20
```

**Authorization**: Required

**Query Parameters**:
- `quiz_id` (optional) - Filter by specific quiz
- `page` (optional, default: 1)
- `limit` (optional, default: 20, max: 100)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "attempts": [
      {
        "id": "attempt-uuid",
        "quiz": {
          "id": "quiz-uuid",
          "title": "Quiz Title"
        },
        "status": "submitted",
        "score": 85.0,
        "percentage": 85.0,
        "passed": true,
        "started_at": "2026-01-10T18:00:00Z",
        "submitted_at": "2026-01-10T18:25:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 20
  }
}
```

**Business Rules**:
- Returns only current user's attempts
- Sorted by `created_at DESC` (most recent first)

---

### 7. Get Quiz Statistics (Creator Only)
```
GET /api/v1/quizzes/:id/stats
```

**Authorization**: Required (must be quiz creator)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "quiz_id": "quiz-uuid",
    "total_attempts": 150,
    "average_score": 78.5,
    "pass_rate": 82.0,
    "highest_score": 100.0,
    "lowest_score": 35.0
  }
}
```

**Business Rules**:
- User must be quiz creator
- No user-level details (privacy)
- Basic stats only

**Error Codes**:
- `403` - not-quiz-creator

---

## Business Logic (Core Layer)

### StartAttempt Flow

```go
func StartAttempt(ctx context.Context, userID string, input StartAttemptInput) (*QuizAttempt, error) {
    // 1. Get quiz and validate
    quiz := dao.GetQuizByID(input.QuizID)
    if quiz == nil {
        return nil, commons.NotFound("quiz-not-found")
    }

    // 2. Check quiz is published
    if !quiz.IsPublished {
        return nil, commons.BadInput("quiz-not-published")
    }

    // 3. Check user has access
    if !quiz.IsPublic && quiz.CreatorID != userID {
        return nil, commons.Forbidden("quiz-not-accessible")
    }

    // 4. Get all questions to calculate max_score
    questions := dao.GetQuestionsByQuizID(quiz.ID)
    maxScore := calculateMaxScore(questions)

    // 5. Calculate time limit and expiry
    timeLimitSeconds, expiresAt := calculateExpiry(quiz)

    // 6. Create attempt
    attempt := &QuizAttempt{
        ID:               uuid.New().String(),
        QuizID:           quiz.ID,
        UserID:           userID,
        Status:           "in_progress",
        StartedAt:        time.Now(),
        TimeLimitSeconds: timeLimitSeconds,
        ExpiresAt:        expiresAt,
        MaxScore:         &maxScore,
        CreatedAt:        time.Now(),
    }

    dao.CreateAttempt(attempt)
    return attempt, nil
}
```

---

### SubmitAnswer Flow

```go
func SubmitAnswer(ctx context.Context, userID string, attemptID string, input SubmitAnswerInput) (*UserAnswer, error) {
    // 1. Get attempt and validate ownership
    attempt := dao.GetAttemptByID(attemptID)
    if attempt.UserID != userID {
        return nil, commons.Forbidden("attempt-not-owned")
    }

    // 2. Check attempt is in_progress
    if attempt.Status != "in_progress" {
        return nil, commons.BadInput("attempt-already-finished")
    }

    // 3. Check time limit
    if err := checkTimeLimit(attempt); err != nil {
        return nil, err
    }

    // 4. Get question and validate belongs to quiz
    question := dao.GetQuestionByID(input.QuestionID)
    if question.QuizID != attempt.QuizID {
        return nil, commons.BadInput("question-not-in-quiz")
    }

    // 5. Validate answer format
    if err := validateAnswer(question, input); err != nil {
        return nil, err
    }

    // 6. Score answer immediately
    isCorrect, pointsEarned := scoreAnswer(question, input.AnswerID)

    // 7. Upsert user_answer
    userAnswer := &UserAnswer{
        ID:           uuid.New().String(),
        AttemptID:    attemptID,
        QuestionID:   input.QuestionID,
        AnswerID:     input.AnswerID,
        IsCorrect:    &isCorrect,
        PointsEarned: &pointsEarned,
        AnsweredAt:   time.Now(),
    }

    dao.UpsertUserAnswer(userAnswer)
    return userAnswer, nil
}
```

---

### FinishAttempt Flow

```go
func FinishAttempt(ctx context.Context, userID string, attemptID string) (*AttemptResultResponse, error) {
    // 1. Get attempt and validate
    attempt := dao.GetAttemptByID(attemptID)
    if attempt.UserID != userID {
        return nil, commons.Forbidden("attempt-not-owned")
    }
    if attempt.Status != "in_progress" {
        return nil, commons.BadInput("attempt-already-finished")
    }

    // 2. Get all questions for quiz
    questions := dao.GetQuestionsByQuizID(attempt.QuizID)

    // 3. Get all user answers
    userAnswers := dao.GetUserAnswersByAttempt(attemptID)

    // 4. Calculate final score
    score, maxScore, percentage := calculateFinalScore(questions, userAnswers)

    // 5. Check if passed
    quiz := dao.GetQuizByID(attempt.QuizID)
    passed := percentage >= float64(quiz.PassingScore)

    // 6. Update attempt
    attempt.Score = &score
    attempt.MaxScore = &maxScore
    attempt.Percentage = &percentage
    attempt.Passed = &passed
    attempt.Status = "submitted"
    attempt.SubmittedAt = time.Now()

    dao.UpdateAttemptScore(attempt)

    // 7. Build and return full results
    return buildResultResponse(attempt, quiz, questions, userAnswers), nil
}
```

---

## Scoring Logic

### Immediate Scoring (On Answer Submission)

```go
func scoreAnswer(question *Question, selectedAnswerID *string) (isCorrect bool, pointsEarned float64) {
    // Skip short answer (not supported in Phase 5)
    if question.QuestionType == "short_answer" {
        return false, 0
    }

    // Get all answers for question
    answers := dao.GetAnswersByQuestionID(question.ID)

    // Find selected answer
    var selectedAnswer *Answer
    for _, ans := range answers {
        if ans.ID == *selectedAnswerID {
            selectedAnswer = &ans
            break
        }
    }

    if selectedAnswer == nil {
        return false, 0
    }

    // Check correctness
    isCorrect = selectedAnswer.IsCorrect
    if isCorrect {
        pointsEarned = float64(question.Points)
    } else {
        pointsEarned = 0
    }

    return isCorrect, pointsEarned
}
```

### Final Score Calculation

```go
func calculateFinalScore(questions []Question, userAnswers []UserAnswer) (score, maxScore, percentage float64) {
    score = 0.0
    maxScore = 0.0

    // Calculate max possible score (only MC/TF questions)
    for _, q := range questions {
        if q.QuestionType != "short_answer" {
            maxScore += float64(q.Points)
        }
    }

    // Sum earned points
    for _, ans := range userAnswers {
        if ans.PointsEarned != nil {
            score += *ans.PointsEarned
        }
    }

    // Calculate percentage
    if maxScore > 0 {
        percentage = (score / maxScore) * 100
    } else {
        percentage = 0
    }

    return score, maxScore, percentage
}
```

---

## Time Limit Enforcement

### Calculate Expiry on Start

```go
func calculateExpiry(quiz *Quiz) (timeLimitSeconds *int, expiresAt *time.Time) {
    if quiz.TimeLimitMinutes == nil {
        return nil, nil // No time limit
    }

    seconds := *quiz.TimeLimitMinutes * 60
    expiry := time.Now().Add(time.Duration(seconds) * time.Second)

    return &seconds, &expiry
}
```

### Check Time Limit

```go
func checkTimeLimit(attempt *QuizAttempt) error {
    if attempt.ExpiresAt == nil {
        return nil // No time limit
    }

    if time.Now().After(*attempt.ExpiresAt) {
        return commons.BadInput("time-limit-exceeded")
    }

    return nil
}
```

### Calculate Time Remaining

```go
func calculateTimeRemaining(attempt *QuizAttempt) *int {
    if attempt.ExpiresAt == nil {
        return nil
    }

    remaining := int(time.Until(*attempt.ExpiresAt).Seconds())
    if remaining < 0 {
        remaining = 0
    }

    return &remaining
}
```

---

## Background Worker

### Auto-Submit Expired Attempts

```go
type AttemptWorker struct {
    db       *sql.DB
    stopChan chan struct{}
}

func (w *AttemptWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-w.stopChan:
            return
        case <-ticker.C:
            w.autoSubmitExpired(ctx)
        }
    }
}

func (w *AttemptWorker) autoSubmitExpired(ctx context.Context) {
    // Get expired attempts
    expiredAttempts := dao.GetExpiredAttempts(ctx)

    for _, attempt := range expiredAttempts {
        // Get user answers
        userAnswers := dao.GetUserAnswersByAttempt(attempt.ID)

        // Get questions
        questions := dao.GetQuestionsByQuizID(attempt.QuizID)

        // Calculate score
        score, maxScore, percentage := calculateFinalScore(questions, userAnswers)

        // Get quiz to check passing score
        quiz := dao.GetQuizByID(attempt.QuizID)
        passed := percentage >= float64(quiz.PassingScore)

        // Update attempt
        dao.UpdateAttemptScore(attempt.ID, score, maxScore, percentage, passed)
        dao.UpdateAttemptStatus(attempt.ID, "expired")
    }

    log.Infof("Auto-submitted %d expired attempts", len(expiredAttempts))
}
```

**Worker Initialization** (in main.go):
```go
// Start attempt worker for auto-submitting expired attempts
attemptWorker := quizworker.NewAttemptWorker(sqlDB)
go attemptWorker.Start(context.Background())
log.Info("Attempt worker started for auto-submission")
```

---

## Validation Rules

### Quiz Access Validation

```go
func (q *Quiz) CanBeAttempted() error {
    if !q.IsPublished {
        return commons.BadInput("quiz-not-published")
    }
    if q.DeletedAt != nil {
        return commons.NotFound("quiz-deleted")
    }
    return nil
}

func (q *Quiz) IsAccessibleBy(userID string) bool {
    return q.IsPublic || q.CreatorID == userID
}
```

### Attempt Status Validation

```go
func (a *QuizAttempt) CanSubmitAnswers() error {
    if a.Status != "in_progress" {
        return commons.BadInput("attempt-already-finished")
    }
    if a.ExpiresAt != nil && time.Now().After(*a.ExpiresAt) {
        return commons.BadInput("time-limit-exceeded")
    }
    return nil
}

func (a *QuizAttempt) CanViewResults() error {
    if a.Status == "in_progress" {
        return commons.BadInput("attempt-not-finished")
    }
    return nil
}
```

### Answer Validation

```go
func validateAnswer(question *Question, input SubmitAnswerInput) error {
    switch question.QuestionType {
    case "multiple_choice", "true_false":
        if input.AnswerID == nil {
            return commons.BadInput("answer-id-required")
        }
        // Verify answer belongs to question
        answer := dao.GetAnswerByID(*input.AnswerID)
        if answer == nil || answer.QuestionID != question.ID {
            return commons.BadInput("answer-not-in-question")
        }

    case "short_answer":
        return commons.BadInput("short-answer-not-supported")

    default:
        return commons.BadInput("invalid-question-type")
    }
    return nil
}
```

---

## Error Codes

### New Error Codes for Phase 5

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `quiz-not-published` | 400 | Cannot attempt unpublished quiz |
| `quiz-not-accessible` | 403 | Quiz is private and user is not creator |
| `attempt-not-found` | 404 | Attempt ID doesn't exist |
| `attempt-not-owned` | 403 | User doesn't own this attempt |
| `attempt-already-finished` | 400 | Cannot modify submitted/expired attempt |
| `attempt-not-finished` | 400 | Cannot view results of in-progress attempt |
| `time-limit-exceeded` | 400 | Attempt expired, auto-submitted |
| `question-not-in-quiz` | 400 | Question doesn't belong to this quiz |
| `answer-not-in-question` | 400 | Answer doesn't belong to this question |
| `answer-id-required` | 400 | MC/TF questions require answer_id |
| `short-answer-not-supported` | 400 | Short answer questions not in Phase 5 |
| `not-quiz-creator` | 403 | User is not the quiz creator |

---

## Implementation Timeline

### Day 1: Database & Models (4-5 hours)
**Tasks**:
- [ ] Create migration `000008_create_quiz_attempts_table.up.sql`
- [ ] Create migration `000008_create_quiz_attempts_table.down.sql`
- [ ] Create migration `000009_create_user_answers_table.up.sql`
- [ ] Create migration `000009_create_user_answers_table.down.sql`
- [ ] Run migrations on dev database
- [ ] Create `/quiz/model/attempt.go` with all structs
- [ ] Test migrations (up/down)

**Acceptance Criteria**:
- ✅ Both tables created successfully
- ✅ Indexes and constraints in place
- ✅ Models compile without errors
- ✅ Can manually insert test records

---

### Day 2: DAO Layer (5-6 hours)
**Tasks**:
- [ ] Create `/quiz/dao/attempt.go`
- [ ] Implement attempt CRUD functions
  - CreateAttempt, GetAttemptByID, ListUserAttempts
  - UpdateAttemptStatus, UpdateAttemptScore
  - GetExpiredAttempts
- [ ] Implement answer functions
  - UpsertUserAnswer, GetUserAnswersByAttempt
  - GetUserAnswer, CountAnsweredQuestions
- [ ] Implement statistics functions
  - GetQuizStats, GetUserQuizStats
- [ ] Add DAO initialization to main.go
- [ ] Test build

**Acceptance Criteria**:
- ✅ All DAO functions compile
- ✅ Can create/retrieve attempts and answers
- ✅ Upsert works correctly (ON CONFLICT UPDATE)
- ✅ Statistics queries return correct aggregations

---

### Day 3: Core Logic - Part 1 (5-6 hours)
**Tasks**:
- [ ] Create `/quiz/core/attempt.go`
- [ ] Implement `StartAttempt()` with all validations
  - Quiz existence check
  - Published status check
  - Access control check
  - Max score calculation
  - Time limit calculation
- [ ] Implement `GetAttemptWithDetails()`
- [ ] Implement helper functions
  - calculateExpiry, checkTimeLimit, calculateTimeRemaining
- [ ] Test build

**Acceptance Criteria**:
- ✅ Can start attempts for published quizzes
- ✅ Time limits calculated correctly
- ✅ Access control enforced
- ✅ Unpublished/private quizzes blocked

---

### Day 4: Core Logic - Part 2 (5-6 hours)
**Tasks**:
- [ ] Implement `SubmitAnswer()` with validation and scoring
  - Ownership check
  - Status check
  - Time limit check
  - Question validation
  - Answer validation
  - Immediate scoring
- [ ] Implement `FinishAttempt()` with final score calculation
- [ ] Implement `GetAttemptResults()`
- [ ] Implement scoring functions
  - scoreAnswer, calculateFinalScore
- [ ] Test scoring calculations manually

**Acceptance Criteria**:
- ✅ Answers scored correctly on submission
- ✅ Final scores calculated accurately
- ✅ Pass/fail logic works
- ✅ Percentage calculations correct

---

### Day 5: API Endpoints (4-5 hours)
**Tasks**:
- [ ] Create `/api/attempt.go`
- [ ] Implement all 7 endpoints:
  - POST /quizzes/:id/attempts (StartAttempt)
  - GET /attempts/:id (GetAttemptWithDetails)
  - POST /attempts/:id/answers (SubmitAnswer)
  - POST /attempts/:id/finish (FinishAttempt)
  - GET /attempts/:id/results (GetAttemptResults)
  - GET /attempts (ListMyAttempts)
  - GET /quizzes/:id/stats (GetQuizStatistics)
- [ ] Add routes to main.go
- [ ] Test build and server startup

**Acceptance Criteria**:
- ✅ All endpoints compile
- ✅ Routes registered correctly
- ✅ Server starts without errors
- ✅ All endpoints return proper JSON

---

### Day 6: Background Worker & Statistics (4-5 hours)
**Tasks**:
- [ ] Create `/quiz/worker/attempt_worker.go`
- [ ] Implement auto-submit logic
  - GetExpiredAttempts query
  - Score calculation for expired attempts
  - Status update to 'expired'
- [ ] Start worker in main.go
- [ ] Implement `GetQuizStatistics()` in core
- [ ] Test worker (create expired attempt manually)

**Acceptance Criteria**:
- ✅ Worker starts and polls every minute
- ✅ Expired attempts auto-submitted correctly
- ✅ Statistics calculations accurate
- ✅ Worker logs properly

---

### Day 7: Testing & Bug Fixes (5-6 hours)
**Tasks**:
- [ ] Test complete user flow with curl/Postman
  - Start attempt → Submit answers → Finish → View results
- [ ] Test time limit enforcement
  - Start timed quiz → wait for expiry → verify auto-submit
- [ ] Test resume flow
  - Start → submit some → close → reopen → continue → finish
- [ ] Test edge cases
  - Invalid UUIDs
  - Unauthorized access
  - Expired time limits
  - Invalid answer IDs
- [ ] Fix any bugs found
- [ ] Test pagination for attempt list

**Acceptance Criteria**:
- ✅ Complete flow works end-to-end
- ✅ All validations work correctly
- ✅ No crashes or data corruption
- ✅ Errors return proper codes/messages

---

### Day 8: Documentation & Cleanup (3-4 hours)
**Tasks**:
- [ ] Update API documentation with Phase 5 endpoints
- [ ] Add code comments to complex functions
- [ ] Create example curl requests for testing
- [ ] Update README with Phase 5 features
- [ ] Final testing of all endpoints
- [ ] Code cleanup (remove debug logs, etc.)

**Acceptance Criteria**:
- ✅ Code well-documented
- ✅ API docs complete with examples
- ✅ README includes Phase 5 info
- ✅ Example requests provided

---

## Testing Strategy

### Unit Tests (Optional for MVP)
- DAO functions (CRUD operations)
- Scoring calculations (scoreAnswer, calculateFinalScore)
- Time limit calculations (calculateExpiry, checkTimeLimit)
- Validation functions (quiz access, answer validation)

### Manual Testing (Required)

**Test Case 1: Happy Path**
```bash
# 1. Start attempt
POST /api/v1/quizzes/{quiz_id}/attempts
→ Should return attempt with status "in_progress"

# 2. Submit answers progressively
POST /api/v1/attempts/{attempt_id}/answers
Body: {"question_id": "q1", "answer_id": "a2"}
→ Should return is_correct and points_earned

# 3. Finish attempt
POST /api/v1/attempts/{attempt_id}/finish
→ Should return full results with score

# 4. View results again
GET /api/v1/attempts/{attempt_id}/results
→ Should show same results
```

**Test Case 2: Time Limit**
```bash
# 1. Start timed quiz (e.g., 2 minutes)
POST /api/v1/quizzes/{quiz_id}/attempts
→ Note expires_at timestamp

# 2. Wait for expiry (or manually update in DB)

# 3. Try to submit answer
POST /api/v1/attempts/{attempt_id}/answers
→ Should return 400 "time-limit-exceeded"

# 4. Check attempt status
GET /api/v1/attempts/{attempt_id}
→ Should show status "expired" with auto-calculated score
```

**Test Case 3: Resume Flow**
```bash
# 1. Start attempt
POST /api/v1/quizzes/{quiz_id}/attempts

# 2. Submit some answers
POST /api/v1/attempts/{attempt_id}/answers (x3)

# 3. Close/Leave (simulate browser close)

# 4. Later, get attempt details
GET /api/v1/attempts/{attempt_id}
→ Should show 3 submitted answers

# 5. Continue submitting
POST /api/v1/attempts/{attempt_id}/answers (x more)

# 6. Finish
POST /api/v1/attempts/{attempt_id}/finish
→ Should calculate score from all answers
```

**Test Case 4: Validation Errors**
```bash
# Try starting unpublished quiz
POST /api/v1/quizzes/{unpublished_quiz_id}/attempts
→ 400 "quiz-not-published"

# Try submitting to expired attempt
POST /api/v1/attempts/{expired_attempt_id}/answers
→ 400 "attempt-already-finished"

# Try viewing results of in-progress attempt
GET /api/v1/attempts/{in_progress_attempt_id}/results
→ 400 "attempt-not-finished"

# Try accessing other user's attempt
GET /api/v1/attempts/{other_user_attempt_id}
→ 403 "attempt-not-owned"
```

**Test Case 5: Multiple Attempts**
```bash
# 1. Take quiz once
POST /quizzes/{id}/attempts → submit → finish

# 2. Take same quiz again
POST /quizzes/{id}/attempts → submit → finish

# 3. List attempts
GET /api/v1/attempts?quiz_id={id}
→ Should show both attempts

# 4. Check creator stats
GET /api/v1/quizzes/{id}/stats
→ Should show average of both attempts
```

---

## Security Considerations

### Authorization
- ✅ All endpoints require JWT authentication
- ✅ Users can only access their own attempts
- ✅ Quiz creators can only see stats for their quizzes
- ✅ Verify quiz access before starting attempt
- ✅ Verify attempt ownership before any modification

### Validation
- ✅ Validate all UUIDs (attempt_id, quiz_id, question_id, answer_id)
- ✅ Verify question belongs to quiz
- ✅ Verify answer belongs to question
- ✅ Check time limits before any modification
- ✅ Prevent answer changes after submission
- ✅ Validate score ranges with CHECK constraints

### Data Integrity
- ✅ Use database transactions for score calculation
- ✅ UNIQUE constraint on (attempt_id, question_id)
- ✅ ON DELETE CASCADE for data consistency
- ✅ Validate percentage is 0-100
- ✅ Validate score <= max_score

### Rate Limiting (Optional)
- Consider limiting attempts per user per hour (e.g., 50 attempts/hour)
- Prevents spam/abuse of auto-submission

---

## Performance Considerations

### Database Indexes
```sql
-- Query: List user attempts for specific quiz
CREATE INDEX idx_quiz_attempts_user_quiz ON quiz_attempts(user_id, quiz_id, created_at DESC);

-- Query: Get pending/expired attempts
CREATE INDEX idx_quiz_attempts_status ON quiz_attempts(status);

-- Query: Get user answers for attempt
CREATE INDEX idx_user_answers_attempt_id ON user_answers(attempt_id);
```

### Query Optimization
- Use pagination for attempt listing (default 20, max 100)
- Fetch only necessary columns for list views
- Use JOINs efficiently for quiz + attempt details
- Cache quiz statistics (optional)

### Concurrent Access
- Database handles concurrent answer submissions via UNIQUE constraint
- Upsert pattern prevents duplicate answers
- Transaction isolation for score calculation

---

## Configuration

No new environment variables needed for Phase 5. Uses existing:
- `DB_*` - Database connection
- `JWT_*` - Authentication
- `LOG_LEVEL` - Logging

Optional additions for future:
```bash
# Attempt Configuration (Optional)
MAX_ATTEMPTS_PER_HOUR=50
ATTEMPT_CLEANUP_DAYS=90  # Delete old attempts after 90 days
AUTO_SUBMIT_CHECK_INTERVAL=60  # Seconds between checks
```

---

## Database Migration Strategy

### Development
```bash
# Run migrations
go run main.go  # Auto-migrates on startup

# Verify tables
psql quiz_craft -c "\d quiz_attempts"
psql quiz_craft -c "\d user_answers"
```

### Rollback (if needed)
```sql
-- Down migrations
DROP TABLE user_answers;
DROP TABLE quiz_attempts;
UPDATE schema_migrations SET version = 7;
```

---

## Success Metrics

Phase 5 is successful when:

✅ **Functional**:
- Users can start quiz attempts
- Users can submit answers progressively
- Answers scored immediately on submission
- Users can finish attempts and get results
- Time limits enforced with auto-submit
- Users can view past attempts
- Quiz creators can see basic statistics

✅ **Technical**:
- All validations work correctly
- System handles concurrent attempts
- No data corruption
- Worker auto-submits expired attempts
- All endpoints return proper status codes

✅ **Quality**:
- Scoring calculations are accurate
- Pass/fail logic works correctly
- Time limit enforcement is precise
- Resume flow works seamlessly

---

## Future Enhancements (Phase 6+)

Deferred features:
- ❌ Short answer manual grading
- ❌ Detailed analytics dashboard for creators
- ❌ Leaderboards and competitive features
- ❌ Certificate generation
- ❌ Attempt sharing
- ❌ Question hints during attempts
- ❌ Question/answer shuffling
- ❌ Partial credit scoring
- ❌ Review mode (review before final submit)

---

## Critical Files Summary

### New Files (9 files):
1. `pkg/database/migrations/000008_create_quiz_attempts_table.up.sql`
2. `pkg/database/migrations/000008_create_quiz_attempts_table.down.sql`
3. `pkg/database/migrations/000009_create_user_answers_table.up.sql`
4. `pkg/database/migrations/000009_create_user_answers_table.down.sql`
5. `quiz/model/attempt.go`
6. `quiz/dao/attempt.go`
7. `quiz/core/attempt.go`
8. `api/attempt.go`
9. `quiz/worker/attempt_worker.go`

### Modified Files (1 file):
1. `main.go` - Add routes and worker initialization

---

## Dependencies

**Go Packages** (Already Installed):
- `database/sql` - Database operations
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/google/uuid` - UUID generation
- `github.com/lib/pq` - PostgreSQL driver

**No New Dependencies Required** ✅

---

## Rollback Plan

If Phase 5 needs rollback:

1. **Code Rollback**:
   - Remove attempt routes from main.go
   - Revert to previous commit
   - Restart server

2. **Database Rollback**:
   ```sql
   DROP TABLE user_answers;
   DROP TABLE quiz_attempts;
   UPDATE schema_migrations SET version = 7, dirty = false;
   ```

3. **Data Preservation**:
   - Backup attempts data before rollback
   - Can restore later if needed

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Time limit edge cases | Medium | Medium | Thorough testing, precise time checks |
| Concurrent answer submission | Low | High | Database UNIQUE constraint |
| Worker not running | Low | High | Health check, logging |
| Score calculation bugs | Medium | High | Unit tests, manual verification |
| Performance with many attempts | Low | Medium | Indexes, pagination |

---

## Total Estimated Effort

**Implementation**: 40-45 hours (8 working days at 5-6 hours/day)

**Breakdown**:
- Day 1: Database & Models (4-5 hours)
- Day 2: DAO Layer (5-6 hours)
- Day 3: Core Logic Part 1 (5-6 hours)
- Day 4: Core Logic Part 2 (5-6 hours)
- Day 5: API Endpoints (4-5 hours)
- Day 6: Worker & Stats (4-5 hours)
- Day 7: Testing & Fixes (5-6 hours)
- Day 8: Documentation (3-4 hours)

**Complexity**: Medium (familiar patterns, database-heavy)

**Risk Level**: Low (well-defined scope, no external dependencies)

---

## Questions Answered

✅ **Attempt model**: Unlimited attempts
✅ **Session model**: Resume with auto-submit on expiry
✅ **Submission**: Progressive (save each answer individually)
✅ **Results**: Immediate via separate API
✅ **Leaderboard**: No
✅ **Short answer**: Skip for Phase 5
✅ **Analytics**: Basic stats only (total attempts, avg score)

---

**Ready to implement!** 🚀
