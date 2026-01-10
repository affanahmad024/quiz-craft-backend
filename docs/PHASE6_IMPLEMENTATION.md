# Phase 6 Implementation Document: Advanced Features

**Date:** January 10, 2026
**Status:** Planning
**Estimated Duration:** 6 working days (30-35 hours)
**Risk Level:** Low

---

## Executive Summary

Phase 6 enhances the quiz-craft platform with advanced discovery and organizational features. This phase focuses on making quizzes **easily discoverable** through search, filtering, and categorization, while providing creators with **powerful statistics** and **reusable templates**.

### Key Features:
- ✅ Full-text search across quiz titles and descriptions
- ✅ Multi-criteria filtering (category, difficulty, tags, creator)
- ✅ Flexible sorting (newest, popular, highest rated, most attempted)
- ✅ Category and tag management system
- ✅ Detailed quiz statistics for creators
- ✅ User-level performance analytics
- ✅ Quiz template system for reusability
- ✅ Quiz sharing with unique links

### Out of Scope (Deferred to Future):
- ❌ Leaderboards (removed per user request)
- ❌ Social features (likes, comments, follows)
- ❌ Quiz recommendations/AI suggestions
- ❌ Real-time collaboration on quiz editing
- ❌ Quiz versioning/history

---

## User Requirements Summary

Based on user clarification:

| Requirement | Decision |
|-------------|----------|
| **Leaderboard** | No - Not needed for Phase 6 |
| **Search** | Full-text search + multi-criteria filtering + sorting |
| **Statistics** | Detailed per-quiz and per-user analytics |
| **Templates** | Yes - Save and reuse quiz structures |
| **Sharing** | Yes - Unique shareable links with tracking |

---

## System Architecture

### No New Tables Required

Phase 6 leverages existing tables with minor enhancements:

#### Enhancement 1: Add search indexes to quizzes table
```sql
-- Full-text search index on title and description
CREATE INDEX idx_quizzes_search ON quizzes USING GIN(
    to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(description, ''))
);

-- Additional indexes for filtering
CREATE INDEX idx_quizzes_difficulty ON quizzes(difficulty) WHERE deleted_at IS NULL;
CREATE INDEX idx_quizzes_created_at ON quizzes(created_at DESC) WHERE deleted_at IS NULL;
```

#### Enhancement 2: Add sharing and template support to quizzes table
```sql
-- Add sharing and template columns
ALTER TABLE quizzes ADD COLUMN share_code VARCHAR(20) UNIQUE;
ALTER TABLE quizzes ADD COLUMN share_count INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN view_count INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN is_template BOOLEAN DEFAULT false;
ALTER TABLE quizzes ADD COLUMN template_usage_count INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN based_on_template_id UUID REFERENCES quizzes(id) ON DELETE SET NULL;

-- Indexes
CREATE INDEX idx_quizzes_share_code ON quizzes(share_code) WHERE share_code IS NOT NULL;
CREATE INDEX idx_quizzes_is_template ON quizzes(is_template) WHERE is_template = true;
```

### Module Structure

```
quiz/
├── model/
│   ├── quiz.go           (existing - add new fields)
│   └── statistics.go     ← NEW
├── dao/
│   ├── quiz.go           (existing - add new functions)
│   └── statistics.go     ← NEW
├── core/
│   ├── quiz.go           (existing - add new logic)
│   ├── search.go         ← NEW
│   ├── statistics.go     ← NEW
│   └── template.go       ← NEW
└── worker/
    └── attempt_worker.go (existing)

api/
├── quiz.go               (existing - add new endpoints)
├── search.go             ← NEW
└── template.go           ← NEW
```

---

## Data Models

### Enhanced Quiz Model

```go
// Update existing Quiz struct
type Quiz struct {
    ID                 string     `json:"id" db:"id"`
    Title              string     `json:"title" db:"title"`
    Description        *string    `json:"description,omitempty" db:"description"`
    CreatorID          string     `json:"creator_id" db:"creator_id"`
    Category           *string    `json:"category,omitempty" db:"category"`
    Difficulty         *string    `json:"difficulty,omitempty" db:"difficulty"`
    TimeLimitMinutes   *int       `json:"time_limit_minutes,omitempty" db:"time_limit_minutes"`
    PassingScore       int        `json:"passing_score" db:"passing_score"`
    IsPublished        bool       `json:"is_published" db:"is_published"`
    IsPublic           bool       `json:"is_public" db:"is_public"`
    ThumbnailURL       *string    `json:"thumbnail_url,omitempty" db:"thumbnail_url"`
    Tags               []string   `json:"tags" db:"tags"`

    // Phase 6 additions
    ShareCode          *string    `json:"share_code,omitempty" db:"share_code"`
    ShareCount         int        `json:"share_count" db:"share_count"`
    ViewCount          int        `json:"view_count" db:"view_count"`
    IsTemplate         bool       `json:"is_template" db:"is_template"`
    TemplateUsageCount int        `json:"template_usage_count" db:"template_usage_count"`
    BasedOnTemplateID  *string    `json:"based_on_template_id,omitempty" db:"based_on_template_id"`

    Meta               JSONMap    `json:"meta" db:"meta"`
    CreatedAt          time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
    DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
```

### New Statistics Models

```go
// QuizStatistics represents detailed stats for a quiz
type QuizStatistics struct {
    QuizID              string    `json:"quiz_id"`
    Title               string    `json:"title"`
    CreatorID           string    `json:"creator_id"`

    // Attempt statistics
    TotalAttempts       int       `json:"total_attempts"`
    UniqueUsers         int       `json:"unique_users"`
    CompletedAttempts   int       `json:"completed_attempts"`
    InProgressAttempts  int       `json:"in_progress_attempts"`
    AbandonedAttempts   int       `json:"abandoned_attempts"`

    // Score statistics
    AverageScore        float64   `json:"average_score"`
    HighestScore        float64   `json:"highest_score"`
    LowestScore         float64   `json:"lowest_score"`
    MedianScore         float64   `json:"median_score"`
    PassRate            float64   `json:"pass_rate"` // Percentage

    // Time statistics
    AverageTimeSeconds  float64   `json:"average_time_seconds"`
    FastestTimeSeconds  *int      `json:"fastest_time_seconds,omitempty"`

    // Engagement statistics
    ViewCount           int       `json:"view_count"`
    ShareCount          int       `json:"share_count"`

    // Question-level stats (optional)
    QuestionStats       []QuestionStat `json:"question_stats,omitempty"`
}

// QuestionStat represents statistics for a single question
type QuestionStat struct {
    QuestionID         string  `json:"question_id"`
    QuestionText       string  `json:"question_text"`
    TimesAnswered      int     `json:"times_answered"`
    CorrectAnswers     int     `json:"correct_answers"`
    IncorrectAnswers   int     `json:"incorrect_answers"`
    SuccessRate        float64 `json:"success_rate"` // Percentage
    AverageTimeSeconds float64 `json:"average_time_seconds"`
}

// UserStatistics represents aggregate stats for a user
type UserStatistics struct {
    UserID              string    `json:"user_id"`
    Name                string    `json:"name"`
    Email               string    `json:"email"`

    // Quiz creation stats
    QuizzesCreated      int       `json:"quizzes_created"`
    PublishedQuizzes    int       `json:"published_quizzes"`
    TotalQuizViews      int       `json:"total_quiz_views"`
    TotalQuizAttempts   int       `json:"total_quiz_attempts"`

    // Quiz attempt stats
    QuizzesTaken        int       `json:"quizzes_taken"`
    AttemptsCompleted   int       `json:"attempts_completed"`
    AverageScore        float64   `json:"average_score"`
    TotalPointsEarned   float64   `json:"total_points_earned"`
    QuizzesPassed       int       `json:"quizzes_passed"`
    PassRate            float64   `json:"pass_rate"` // Percentage

    // Streak and engagement
    CurrentStreak       int       `json:"current_streak"` // Days
    LongestStreak       int       `json:"longest_streak"` // Days
    LastActivityAt      time.Time `json:"last_activity_at"`
    JoinedAt            time.Time `json:"joined_at"`
}

// CategoryStats represents statistics for a category
type CategoryStats struct {
    Category      string `json:"category"`
    QuizCount     int    `json:"quiz_count"`
    TotalAttempts int    `json:"total_attempts"`
    AverageScore  float64 `json:"average_score"`
}
```

### Search and Filter Models

```go
// SearchQuizzesInput defines search/filter parameters
type SearchQuizzesInput struct {
    // Search
    Query          string   `form:"q"`

    // Filters
    Category       string   `form:"category"`
    Difficulty     string   `form:"difficulty"`
    Tags           []string `form:"tags"`
    CreatorID      string   `form:"creator_id"`
    IsPublished    *bool    `form:"is_published"`
    IsPublic       *bool    `form:"is_public"`
    IsTemplate     *bool    `form:"is_template"`

    // Date range
    CreatedAfter   *time.Time `form:"created_after"`
    CreatedBefore  *time.Time `form:"created_before"`

    // Sorting
    SortBy         string `form:"sort_by"` // newest, popular, highest_rated, most_attempted
    SortOrder      string `form:"sort_order"` // asc, desc

    // Pagination
    Page           int    `form:"page"`
    Limit          int    `form:"limit"`
}

// SearchQuizzesResponse represents search results
type SearchQuizzesResponse struct {
    Quizzes      []QuizSummary `json:"quizzes"`
    Total        int           `json:"total"`
    Page         int           `json:"page"`
    Limit        int           `json:"limit"`
    TotalPages   int           `json:"total_pages"`
    Facets       SearchFacets  `json:"facets"` // For filter counts
}

// SearchFacets provides counts for filter options
type SearchFacets struct {
    Categories   []FacetCount `json:"categories"`
    Difficulties []FacetCount `json:"difficulties"`
    Tags         []FacetCount `json:"tags"`
}

type FacetCount struct {
    Value string `json:"value"`
    Count int    `json:"count"`
}

// QuizSummary is a lightweight quiz representation
type QuizSummary struct {
    ID               string    `json:"id"`
    Title            string    `json:"title"`
    Description      *string   `json:"description,omitempty"`
    Category         *string   `json:"category,omitempty"`
    Difficulty       *string   `json:"difficulty,omitempty"`
    QuestionCount    int       `json:"question_count"`
    AttemptCount     int       `json:"attempt_count"`
    AverageScore     float64   `json:"average_score"`
    Creator          UserSummary `json:"creator"`
    Tags             []string  `json:"tags"`
    ThumbnailURL     *string   `json:"thumbnail_url,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
}

type UserSummary struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

---

## API Endpoints

### 1. Search Quizzes
```
GET /api/v1/quizzes/search
```

**Authorization**: Optional (public quizzes visible to all, private only to creator)

**Query Parameters**:
- `q` - Search query (title, description)
- `category` - Filter by category
- `difficulty` - Filter by difficulty (easy, medium, hard)
- `tags` - Comma-separated tags
- `creator_id` - Filter by creator
- `is_template` - Filter templates
- `sort_by` - newest, popular, highest_rated, most_attempted
- `sort_order` - asc, desc (default: desc)
- `page` - Page number (default: 1)
- `limit` - Results per page (default: 20, max: 100)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "quizzes": [
      {
        "id": "quiz-uuid",
        "title": "JavaScript Fundamentals",
        "description": "Test your JS knowledge",
        "category": "programming",
        "difficulty": "medium",
        "question_count": 20,
        "attempt_count": 150,
        "average_score": 78.5,
        "creator": {
          "id": "user-uuid",
          "name": "John Doe"
        },
        "tags": ["javascript", "programming", "web-dev"],
        "created_at": "2026-01-10T10:00:00Z"
      }
    ],
    "total": 42,
    "page": 1,
    "limit": 20,
    "total_pages": 3,
    "facets": {
      "categories": [
        {"value": "programming", "count": 25},
        {"value": "science", "count": 12}
      ],
      "difficulties": [
        {"value": "easy", "count": 10},
        {"value": "medium", "count": 20},
        {"value": "hard", "count": 12}
      ]
    }
  }
}
```

---

### 2. Get Quiz Statistics (Creator Only)
```
GET /api/v1/quizzes/:id/statistics
```

**Authorization**: Required (must be quiz creator)

**Query Parameters**:
- `include_questions` - Include question-level stats (default: false)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "quiz_id": "quiz-uuid",
    "title": "JavaScript Quiz",
    "creator_id": "user-uuid",
    "total_attempts": 150,
    "unique_users": 120,
    "completed_attempts": 140,
    "in_progress_attempts": 5,
    "abandoned_attempts": 5,
    "average_score": 78.5,
    "highest_score": 98.0,
    "lowest_score": 45.0,
    "median_score": 80.0,
    "pass_rate": 85.5,
    "average_time_seconds": 1200,
    "fastest_time_seconds": 600,
    "view_count": 500,
    "share_count": 25,
    "question_stats": [
      {
        "question_id": "q1",
        "question_text": "What is a closure?",
        "times_answered": 150,
        "correct_answers": 120,
        "incorrect_answers": 30,
        "success_rate": 80.0,
        "average_time_seconds": 45.5
      }
    ]
  }
}
```

---

### 3. Get User Statistics
```
GET /api/v1/users/me/statistics
```

**Authorization**: Required

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "user_id": "user-uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "quizzes_created": 5,
    "published_quizzes": 4,
    "total_quiz_views": 500,
    "total_quiz_attempts": 150,
    "quizzes_taken": 25,
    "attempts_completed": 30,
    "average_score": 82.5,
    "total_points_earned": 2475.0,
    "quizzes_passed": 22,
    "pass_rate": 73.3,
    "current_streak": 5,
    "longest_streak": 12,
    "last_activity_at": "2026-01-10T10:00:00Z",
    "joined_at": "2025-12-01T00:00:00Z"
  }
}
```

---

### 4. Get Categories
```
GET /api/v1/categories
```

**Authorization**: Optional

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "categories": [
      {
        "category": "programming",
        "quiz_count": 45,
        "total_attempts": 1200,
        "average_score": 75.5
      },
      {
        "category": "science",
        "quiz_count": 30,
        "total_attempts": 800,
        "average_score": 70.2
      }
    ]
  }
}
```

---

### 5. Get Popular Tags
```
GET /api/v1/tags
```

**Authorization**: Optional

**Query Parameters**:
- `limit` - Number of tags (default: 50, max: 100)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "tags": [
      {"tag": "javascript", "count": 45},
      {"tag": "python", "count": 38},
      {"tag": "react", "count": 25}
    ]
  }
}
```

---

### 6. Generate Share Link
```
POST /api/v1/quizzes/:id/share
```

**Authorization**: Required (must be quiz creator)

**Request Body**: None

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Share link generated",
  "data": {
    "share_code": "abc123xyz",
    "share_url": "https://quizcraft.com/share/abc123xyz",
    "expires_at": null
  }
}
```

---

### 7. Access Shared Quiz
```
GET /share/:code
```

**Authorization**: Optional

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "id": "quiz-uuid",
    "title": "JavaScript Quiz",
    "description": "Test your knowledge",
    "question_count": 20,
    "time_limit_minutes": 30,
    "passing_score": 70,
    "creator": {
      "name": "John Doe"
    }
  }
}
```

**Business Rules**:
- Increments view_count
- Works even if quiz is not public
- Redirects to full quiz details

---

### 8. Save Quiz as Template
```
POST /api/v1/quizzes/:id/save-as-template
```

**Authorization**: Required (must be quiz creator)

**Request Body**:
```json
{
  "template_name": "JavaScript Quiz Template",
  "template_description": "Template for JS quizzes",
  "make_public": true
}
```

**Response (201 Created)**:
```json
{
  "success": true,
  "message": "Quiz saved as template",
  "data": {
    "template_id": "template-uuid"
  }
}
```

**Business Rules**:
- Creates a copy of the quiz with `is_template = true`
- Copies all questions and answers
- Original quiz remains unchanged

---

### 9. List Templates
```
GET /api/v1/templates
```

**Authorization**: Optional (public templates visible to all)

**Query Parameters**:
- `page` - Page number
- `limit` - Results per page
- `category` - Filter by category
- `my_templates` - Show only user's templates (requires auth)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "templates": [
      {
        "id": "template-uuid",
        "title": "JavaScript Quiz Template",
        "description": "Template for JS quizzes",
        "question_count": 20,
        "usage_count": 15,
        "creator": {
          "id": "user-uuid",
          "name": "John Doe"
        },
        "created_at": "2026-01-05T00:00:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "limit": 20
  }
}
```

---

### 10. Create Quiz from Template
```
POST /api/v1/templates/:id/use
```

**Authorization**: Required

**Request Body**:
```json
{
  "title": "My JavaScript Quiz",
  "customize": true
}
```

**Response (201 Created)**:
```json
{
  "success": true,
  "message": "Quiz created from template",
  "data": {
    "quiz_id": "new-quiz-uuid"
  }
}
```

**Business Rules**:
- Creates a new quiz copying structure from template
- Copies all questions and answers
- Sets `based_on_template_id` to reference template
- Increments template's `template_usage_count`
- User becomes creator of new quiz

---

## Business Logic (Core Layer)

### Search and Filtering Logic

```go
func SearchQuizzes(ctx context.Context, input SearchQuizzesInput, userID *string) (*SearchQuizzesResponse, error) {
    // 1. Build base query
    query := buildBaseSearchQuery(userID)

    // 2. Apply full-text search if query provided
    if input.Query != "" {
        query = applyFullTextSearch(query, input.Query)
    }

    // 3. Apply filters
    query = applyFilters(query, input)

    // 4. Apply sorting
    query = applySorting(query, input.SortBy, input.SortOrder)

    // 5. Execute paginated query
    quizzes, total := dao.SearchQuizzes(query, input.Page, input.Limit)

    // 6. Get facets for filters
    facets := dao.GetSearchFacets(query)

    // 7. Build response
    return &SearchQuizzesResponse{
        Quizzes:    quizzes,
        Total:      total,
        Page:       input.Page,
        Limit:      input.Limit,
        TotalPages: (total + input.Limit - 1) / input.Limit,
        Facets:     facets,
    }, nil
}

func buildBaseSearchQuery(userID *string) string {
    // Only show published + public quizzes, OR user's own quizzes
    if userID != nil {
        return `deleted_at IS NULL AND (
            (is_published = true AND is_public = true) OR
            creator_id = $1
        )`
    }
    return `deleted_at IS NULL AND is_published = true AND is_public = true`
}

func applyFullTextSearch(query string, searchTerm string) string {
    // Use PostgreSQL full-text search
    return query + ` AND to_tsvector('english', title || ' ' || COALESCE(description, ''))
                     @@ plainto_tsquery('english', $searchTerm)`
}

func applySorting(query string, sortBy string, sortOrder string) string {
    switch sortBy {
    case "newest":
        return query + ` ORDER BY created_at ` + sortOrder
    case "popular":
        return query + ` ORDER BY view_count ` + sortOrder
    case "most_attempted":
        // Join with quiz_attempts count
        return query + ` ORDER BY attempt_count ` + sortOrder
    case "highest_rated":
        // Join with average score from quiz_attempts
        return query + ` ORDER BY average_score ` + sortOrder
    default:
        return query + ` ORDER BY created_at DESC`
    }
}
```

---

### Quiz Statistics Logic

```go
func GetQuizStatistics(ctx context.Context, quizID string, includeQuestions bool) (*QuizStatistics, error) {
    // 1. Get quiz
    quiz := dao.GetQuizByID(quizID)
    if quiz == nil {
        return nil, commons.NotFound("quiz-not-found")
    }

    // 2. Get attempt aggregates
    stats := dao.GetQuizAttemptStats(quizID)

    // 3. Calculate pass rate
    passRate := 0.0
    if stats.CompletedAttempts > 0 {
        passRate = (float64(stats.PassedCount) / float64(stats.CompletedAttempts)) * 100
    }

    // 4. Get question-level stats if requested
    var questionStats []QuestionStat
    if includeQuestions {
        questionStats = dao.GetQuestionStats(quizID)
    }

    // 5. Build response
    return &QuizStatistics{
        QuizID:            quizID,
        Title:             quiz.Title,
        CreatorID:         quiz.CreatorID,
        TotalAttempts:     stats.TotalAttempts,
        UniqueUsers:       stats.UniqueUsers,
        CompletedAttempts: stats.CompletedAttempts,
        AverageScore:      stats.AverageScore,
        HighestScore:      stats.HighestScore,
        LowestScore:       stats.LowestScore,
        PassRate:          passRate,
        ViewCount:         quiz.ViewCount,
        ShareCount:        quiz.ShareCount,
        QuestionStats:     questionStats,
    }, nil
}
```

---

### User Statistics Logic

```go
func GetUserStatistics(ctx context.Context, userID string) (*UserStatistics, error) {
    // 1. Get user
    user := dao.GetUserByID(userID)
    if user == nil {
        return nil, commons.NotFound("user-not-found")
    }

    // 2. Get quiz creation stats
    creatorStats := dao.GetUserCreatorStats(userID)

    // 3. Get quiz attempt stats
    attemptStats := dao.GetUserAttemptStats(userID)

    // 4. Calculate streak
    currentStreak := calculateStreak(userID)

    // 5. Build response
    return &UserStatistics{
        UserID:            userID,
        Name:              user.Name,
        Email:             user.Email,
        QuizzesCreated:    creatorStats.TotalQuizzes,
        PublishedQuizzes:  creatorStats.PublishedQuizzes,
        TotalQuizViews:    creatorStats.TotalViews,
        QuizzesTaken:      attemptStats.UniqueQuizzes,
        AttemptsCompleted: attemptStats.CompletedAttempts,
        AverageScore:      attemptStats.AverageScore,
        PassRate:          attemptStats.PassRate,
        CurrentStreak:     currentStreak,
    }, nil
}

func calculateStreak(userID string) int {
    // Get attempts grouped by date
    attempts := dao.GetUserAttemptsByDate(userID)

    streak := 0
    currentDate := time.Now().Truncate(24 * time.Hour)

    for _, attemptDate := range attempts {
        if attemptDate.Equal(currentDate) || attemptDate.Equal(currentDate.AddDate(0, 0, -streak)) {
            streak++
        } else {
            break
        }
    }

    return streak
}
```

---

### Template Logic

```go
func SaveQuizAsTemplate(ctx context.Context, quizID string, userID string, input SaveAsTemplateInput) (string, error) {
    // 1. Get original quiz
    quiz := dao.GetQuizByID(quizID)
    if quiz == nil {
        return "", commons.NotFound("quiz-not-found")
    }

    // 2. Check ownership
    if quiz.CreatorID != userID {
        return "", commons.Forbidden("not-quiz-creator")
    }

    // 3. Get questions and answers
    questions := dao.GetQuestionsByQuizID(quizID)

    // 4. Create template quiz (copy)
    template := &Quiz{
        ID:           uuid.New().String(),
        Title:        input.TemplateName,
        Description:  &input.TemplateDescription,
        CreatorID:    userID,
        Category:     quiz.Category,
        Difficulty:   quiz.Difficulty,
        PassingScore: quiz.PassingScore,
        IsPublished:  input.MakePublic,
        IsPublic:     input.MakePublic,
        IsTemplate:   true,
        Tags:         quiz.Tags,
        CreatedAt:    time.Now(),
    }

    dao.CreateQuiz(template)

    // 5. Copy questions
    for _, q := range questions {
        newQuestion := copyQuestion(q, template.ID)
        dao.CreateQuestion(newQuestion)

        // Copy answers
        answers := dao.GetAnswersByQuestionID(q.ID)
        for _, a := range answers {
            newAnswer := copyAnswer(a, newQuestion.ID)
            dao.CreateAnswer(newAnswer)
        }
    }

    return template.ID, nil
}

func CreateQuizFromTemplate(ctx context.Context, templateID string, userID string, input UseTemplateInput) (string, error) {
    // Similar logic to SaveQuizAsTemplate
    // Copy template structure to new quiz
    // Set based_on_template_id
    // Increment template_usage_count
}
```

---

### Share Link Logic

```go
func GenerateShareLink(ctx context.Context, quizID string, userID string) (*ShareLinkResponse, error) {
    // 1. Get quiz and validate ownership
    quiz := dao.GetQuizByID(quizID)
    if quiz == nil {
        return nil, commons.NotFound("quiz-not-found")
    }
    if quiz.CreatorID != userID {
        return nil, commons.Forbidden("not-quiz-creator")
    }

    // 2. Generate unique share code if doesn't exist
    if quiz.ShareCode == nil {
        shareCode := generateRandomCode(10) // e.g., "abc123xyz"
        dao.UpdateQuizShareCode(quizID, shareCode)
        quiz.ShareCode = &shareCode
    }

    // 3. Build share URL
    shareURL := fmt.Sprintf("https://quizcraft.com/share/%s", *quiz.ShareCode)

    return &ShareLinkResponse{
        ShareCode: *quiz.ShareCode,
        ShareURL:  shareURL,
    }, nil
}

func GetQuizByShareCode(ctx context.Context, shareCode string) (*Quiz, error) {
    // 1. Get quiz by share code
    quiz := dao.GetQuizByShareCode(shareCode)
    if quiz == nil {
        return nil, commons.NotFound("quiz-not-found")
    }

    // 2. Increment view count
    dao.IncrementQuizViewCount(quiz.ID)

    return quiz, nil
}
```

---

## Database Queries

### Search Query with Full-Text Search

```sql
-- Search with full-text search and filters
SELECT q.*, u.name as creator_name,
       COUNT(DISTINCT qa.id) as attempt_count,
       AVG(qa.percentage) as average_score
FROM quizzes q
LEFT JOIN users u ON q.creator_id = u.id
LEFT JOIN quiz_attempts qa ON q.id = qa.quiz_id AND qa.status = 'submitted'
WHERE q.deleted_at IS NULL
  AND q.is_published = true
  AND q.is_public = true
  AND to_tsvector('english', q.title || ' ' || COALESCE(q.description, ''))
      @@ plainto_tsquery('english', $1)
  AND ($2::text IS NULL OR q.category = $2)
  AND ($3::text IS NULL OR q.difficulty = $3)
  AND ($4::text[] IS NULL OR q.tags && $4)
GROUP BY q.id, u.name
ORDER BY q.created_at DESC
LIMIT $5 OFFSET $6;
```

### Quiz Statistics Query

```sql
-- Aggregate quiz attempt statistics
SELECT
    COUNT(*) as total_attempts,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(*) FILTER (WHERE status = 'submitted') as completed_attempts,
    COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress_attempts,
    COUNT(*) FILTER (WHERE status = 'abandoned') as abandoned_attempts,
    AVG(percentage) FILTER (WHERE status = 'submitted') as average_score,
    MAX(percentage) FILTER (WHERE status = 'submitted') as highest_score,
    MIN(percentage) FILTER (WHERE status = 'submitted') as lowest_score,
    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY percentage)
        FILTER (WHERE status = 'submitted') as median_score,
    COUNT(*) FILTER (WHERE passed = true) as passed_count,
    AVG(EXTRACT(EPOCH FROM (submitted_at - started_at)))
        FILTER (WHERE status = 'submitted') as average_time_seconds
FROM quiz_attempts
WHERE quiz_id = $1;
```

### Question Statistics Query

```sql
-- Question-level statistics
SELECT
    q.id as question_id,
    q.question_text,
    COUNT(ua.id) as times_answered,
    COUNT(*) FILTER (WHERE ua.is_correct = true) as correct_answers,
    COUNT(*) FILTER (WHERE ua.is_correct = false) as incorrect_answers,
    AVG(CASE WHEN ua.is_correct THEN 100.0 ELSE 0.0 END) as success_rate
FROM questions q
LEFT JOIN user_answers ua ON q.id = ua.question_id
WHERE q.quiz_id = $1
GROUP BY q.id, q.question_text
ORDER BY q.order_index;
```

### User Creator Statistics Query

```sql
-- User's quiz creation statistics
SELECT
    COUNT(*) as total_quizzes,
    COUNT(*) FILTER (WHERE is_published = true) as published_quizzes,
    SUM(view_count) as total_views,
    SUM((SELECT COUNT(*) FROM quiz_attempts qa WHERE qa.quiz_id = q.id)) as total_attempts
FROM quizzes q
WHERE q.creator_id = $1 AND q.deleted_at IS NULL;
```

### User Attempt Statistics Query

```sql
-- User's quiz attempt statistics
SELECT
    COUNT(DISTINCT quiz_id) as unique_quizzes,
    COUNT(*) FILTER (WHERE status = 'submitted') as completed_attempts,
    AVG(percentage) FILTER (WHERE status = 'submitted') as average_score,
    SUM(score) FILTER (WHERE status = 'submitted') as total_points_earned,
    COUNT(*) FILTER (WHERE passed = true) as quizzes_passed,
    AVG(CASE WHEN passed THEN 100.0 ELSE 0.0 END) as pass_rate
FROM quiz_attempts
WHERE user_id = $1;
```

---

## Implementation Timeline

### Day 1: Database Enhancements & Search Foundation (5-6 hours)
**Tasks**:
- [ ] Create migration `000010_enhance_quizzes_for_phase6.up.sql`
  - Add share_code, view_count, share_count columns
  - Add is_template, template_usage_count, based_on_template_id columns
  - Create full-text search index
  - Create additional indexes for filtering
- [ ] Create migration `000010_enhance_quizzes_for_phase6.down.sql`
- [ ] Run migrations on dev database
- [ ] Update `quiz/model/quiz.go` with new fields
- [ ] Test migrations (up/down)

**Acceptance Criteria**:
- ✅ All columns added successfully
- ✅ Full-text search index created
- ✅ Models compile with new fields

---

### Day 2: Search & Filter Logic (6-7 hours)
**Tasks**:
- [ ] Create `quiz/model/statistics.go` with all stat models
- [ ] Create `quiz/core/search.go`
- [ ] Implement search and filter logic
  - Full-text search integration
  - Multi-criteria filtering
  - Dynamic sorting
  - Pagination
- [ ] Create `quiz/dao/search.go`
- [ ] Implement DAO functions for search
  - SearchQuizzes with filters
  - GetSearchFacets for filter counts
  - GetCategoryStats
  - GetPopularTags
- [ ] Test search functionality

**Acceptance Criteria**:
- ✅ Full-text search works correctly
- ✅ Filters combine properly
- ✅ Sorting works for all options
- ✅ Facets return correct counts

---

### Day 3: Statistics Logic - Part 1 (6-7 hours)
**Tasks**:
- [ ] Create `quiz/dao/statistics.go`
- [ ] Implement quiz statistics functions
  - GetQuizAttemptStats (aggregate query)
  - GetQuestionStats (question-level)
  - GetUserCreatorStats (quiz creation stats)
  - GetUserAttemptStats (quiz taking stats)
- [ ] Create `quiz/core/statistics.go`
- [ ] Implement GetQuizStatistics with all calculations
- [ ] Test quiz statistics calculations

**Acceptance Criteria**:
- ✅ Quiz stats calculated correctly
- ✅ Question-level stats accurate
- ✅ Performance acceptable with large datasets

---

### Day 4: Statistics Logic - Part 2 (5-6 hours)
**Tasks**:
- [ ] Implement GetUserStatistics in core
- [ ] Implement streak calculation logic
  - Current streak
  - Longest streak
- [ ] Implement GetCategories endpoint logic
- [ ] Implement GetPopularTags endpoint logic
- [ ] Test user statistics calculations
- [ ] Test streak calculations with various scenarios

**Acceptance Criteria**:
- ✅ User stats calculated correctly
- ✅ Streak logic works for edge cases
- ✅ Category and tag aggregations accurate

---

### Day 5: Template System & Sharing (6-7 hours)
**Tasks**:
- [ ] Create `quiz/core/template.go`
- [ ] Implement SaveQuizAsTemplate
  - Copy quiz structure
  - Copy questions and answers
  - Set template flags
- [ ] Implement CreateQuizFromTemplate
  - Copy template to new quiz
  - Update template usage count
- [ ] Implement share link generation
  - Generate unique codes
  - Update view counts
- [ ] Add DAO functions for templates
- [ ] Test template creation and usage
- [ ] Test share link generation

**Acceptance Criteria**:
- ✅ Templates created correctly
- ✅ Quiz creation from template works
- ✅ Share links functional
- ✅ Usage counts updated properly

---

### Day 6: API Endpoints & Testing (6-7 hours)
**Tasks**:
- [ ] Create `api/search.go`
- [ ] Implement search endpoint
- [ ] Update `api/quiz.go` with new endpoints
  - GET /quizzes/:id/statistics
  - POST /quizzes/:id/share
  - GET /share/:code
- [ ] Create `api/template.go`
- [ ] Implement template endpoints
  - POST /quizzes/:id/save-as-template
  - GET /templates
  - POST /templates/:id/use
- [ ] Implement statistics endpoints
  - GET /users/me/statistics
  - GET /categories
  - GET /tags
- [ ] Add routes to main.go
- [ ] Test all endpoints manually
- [ ] Fix any bugs found

**Acceptance Criteria**:
- ✅ All endpoints functional
- ✅ Search returns correct results
- ✅ Statistics accurate
- ✅ Templates work end-to-end
- ✅ Share links functional

---

## Validation Rules

### Search Input Validation
```go
func validateSearchInput(input *SearchQuizzesInput) error {
    // Pagination
    if input.Page < 1 {
        input.Page = 1
    }
    if input.Limit < 1 || input.Limit > 100 {
        input.Limit = 20
    }

    // Difficulty
    if input.Difficulty != "" {
        validDifficulties := []string{"easy", "medium", "hard"}
        if !contains(validDifficulties, input.Difficulty) {
            return commons.BadInput("invalid-difficulty")
        }
    }

    // Sort by
    validSortOptions := []string{"newest", "popular", "highest_rated", "most_attempted"}
    if input.SortBy != "" && !contains(validSortOptions, input.SortBy) {
        return commons.BadInput("invalid-sort-by")
    }

    // Sort order
    if input.SortOrder != "" && input.SortOrder != "asc" && input.SortOrder != "desc" {
        return commons.BadInput("invalid-sort-order")
    }

    return nil
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid-difficulty` | 400 | Invalid difficulty value |
| `invalid-sort-by` | 400 | Invalid sort option |
| `invalid-sort-order` | 400 | Invalid sort order |
| `template-not-found` | 404 | Template doesn't exist |
| `share-code-not-found` | 404 | Share code invalid |
| `not-quiz-creator` | 403 | User is not quiz creator |
| `quiz-not-published` | 400 | Cannot share unpublished quiz |

---

## Performance Considerations

### Database Optimization
- ✅ Full-text search index for fast queries
- ✅ GIN index on tags for array searches
- ✅ Indexes on commonly filtered columns
- ✅ Use of aggregate queries to minimize round-trips
- ✅ Pagination to limit result set size

### Caching Opportunities (Optional Future)
- Category stats (changes infrequently)
- Popular tags (changes infrequently)
- Quiz statistics (can be cached for 5-10 minutes)

### Query Optimization
```sql
-- Use EXPLAIN ANALYZE to verify indexes are used
EXPLAIN ANALYZE
SELECT * FROM quizzes
WHERE to_tsvector('english', title) @@ plainto_tsquery('english', 'javascript');

-- Expected: Should use idx_quizzes_search index
```

---

## Security Considerations

### Authorization
- ✅ Search respects quiz visibility (public/private)
- ✅ Statistics only visible to quiz creators
- ✅ User can only see their own stats
- ✅ Template creation requires ownership
- ✅ Share link generation requires ownership

### Data Privacy
- ✅ User statistics don't expose personal info
- ✅ Question stats don't identify individual users
- ✅ Share links don't bypass access control (increment view only)

### Input Validation
- ✅ Sanitize search queries to prevent injection
- ✅ Validate all filter parameters
- ✅ Limit result set sizes to prevent DoS

---

## Testing Strategy

### Manual Testing

**Test Case 1: Full-Text Search**
```bash
# Search for "javascript"
GET /api/v1/quizzes/search?q=javascript
→ Should return quizzes with "javascript" in title or description

# Search with filters
GET /api/v1/quizzes/search?q=javascript&difficulty=medium&category=programming
→ Should return only medium difficulty programming quizzes about javascript
```

**Test Case 2: Quiz Statistics**
```bash
# Get quiz statistics
GET /api/v1/quizzes/{quiz_id}/statistics?include_questions=true
→ Should return detailed stats with question-level breakdown

# Try as non-creator
GET /api/v1/quizzes/{other_user_quiz_id}/statistics
→ Should return 403 Forbidden
```

**Test Case 3: Template System**
```bash
# Save as template
POST /api/v1/quizzes/{quiz_id}/save-as-template
Body: {"template_name": "Test Template", "make_public": true}
→ Should create template

# Use template
POST /api/v1/templates/{template_id}/use
Body: {"title": "My Quiz"}
→ Should create new quiz with same structure
```

**Test Case 4: Share Links**
```bash
# Generate share link
POST /api/v1/quizzes/{quiz_id}/share
→ Should return share code and URL

# Access via share link
GET /share/{share_code}
→ Should return quiz details and increment view count
```

---

## Configuration

No new environment variables needed. Uses existing database and auth configuration.

---

## Success Metrics

Phase 6 is successful when:

✅ **Search & Discovery**:
- Full-text search returns relevant results
- Filters combine correctly
- Sorting works for all options
- Pagination handles large result sets

✅ **Statistics**:
- Quiz statistics accurate and performant
- User statistics calculated correctly
- Streak calculations work reliably
- Question-level stats provide insights

✅ **Templates**:
- Templates save quiz structure completely
- Quiz creation from templates works seamlessly
- Usage counts tracked accurately

✅ **Sharing**:
- Share links generated uniquely
- View counts incremented correctly
- Shared quizzes accessible

---

## Critical Files Summary

### New Files (5 files):
1. `pkg/database/migrations/000010_enhance_quizzes_for_phase6.up.sql`
2. `pkg/database/migrations/000010_enhance_quizzes_for_phase6.down.sql`
3. `quiz/model/statistics.go`
4. `quiz/core/search.go`
5. `quiz/core/statistics.go`
6. `quiz/core/template.go`
7. `quiz/dao/statistics.go`
8. `api/search.go`
9. `api/template.go`

### Modified Files (3 files):
1. `quiz/model/quiz.go` - Add new fields
2. `quiz/dao/quiz.go` - Add search/stats functions
3. `api/quiz.go` - Add statistics and share endpoints
4. `main.go` - Add new routes

---

## Dependencies

**No New Dependencies Required** ✅

All features use existing Go standard library and PostgreSQL capabilities.

---

## Total Estimated Effort

**Implementation**: 30-35 hours (6 working days at 5-6 hours/day)

**Breakdown**:
- Day 1: Database & Models (5-6 hours)
- Day 2: Search & Filter (6-7 hours)
- Day 3: Statistics Part 1 (6-7 hours)
- Day 4: Statistics Part 2 (5-6 hours)
- Day 5: Templates & Sharing (6-7 hours)
- Day 6: API & Testing (6-7 hours)

**Complexity**: Medium (complex queries, statistics calculations)

**Risk Level**: Low (no external dependencies, familiar patterns)

---

**Ready to implement!** 🚀
