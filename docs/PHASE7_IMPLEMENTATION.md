# Phase 7 Implementation Document: Admin Dashboard

**Date:** January 10, 2026
**Status:** Planning
**Estimated Duration:** 4 working days (20-24 hours)
**Risk Level:** Low

---

## Executive Summary

Phase 7 adds administrative capabilities to the quiz-craft platform, enabling **platform management, content moderation, and system oversight**. This phase implements basic moderation tools and comprehensive analytics for administrators to maintain platform quality and monitor system health.

### Key Features:
- ✅ Role-based access control (User, Admin)
- ✅ User management (view, disable, enable, delete)
- ✅ Quiz moderation (view all, delete inappropriate content)
- ✅ Platform-wide statistics and analytics
- ✅ Activity logging for audit trails
- ✅ System health monitoring

### Out of Scope (Deferred):
- ❌ Advanced reporting system
- ❌ Automated content moderation (AI-based)
- ❌ Advanced user roles (Moderator, SuperAdmin)
- ❌ Email notification system for admins
- ❌ Bulk operations dashboard

---

## User Requirements Summary

Based on user clarification:

| Requirement | Decision |
|-------------|----------|
| **Moderation Level** | Basic - View, disable, delete actions |
| **User Reporting** | No - Simple admin-initiated actions only |
| **Automated Flagging** | No - Manual review only |
| **Activity Logging** | Yes - Audit trail for all admin actions |

---

## System Architecture

### New Database Tables

Phase 7 adds **2 new tables**:

#### Table 1: Add role column to users table
```sql
-- Migration: Extend users table with role
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user' NOT NULL;
ALTER TABLE users ADD CONSTRAINT chk_user_role CHECK (role IN ('user', 'admin'));

CREATE INDEX idx_users_role ON users(role);
```

#### Table 2: activity_logs table
```sql
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id UUID,
    details JSONB DEFAULT '{}'::jsonb,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX idx_activity_logs_action ON activity_logs(action);
CREATE INDEX idx_activity_logs_resource ON activity_logs(resource_type, resource_id);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
```

### Module Structure

```
admin/
├── model/
│   ├── activity_log.go    ← NEW
│   └── admin_stats.go     ← NEW
├── dao/
│   ├── activity_log.go    ← NEW
│   └── admin_stats.go     ← NEW
└── core/
    ├── user_management.go ← NEW
    ├── quiz_moderation.go ← NEW
    ├── statistics.go      ← NEW
    └── activity_logger.go ← NEW

api/
└── admin.go               ← NEW

internal/
└── middleware/
    └── admin.go           ← NEW (admin authorization)
```

---

## Data Models

### Activity Log Model

```go
// ActivityLog represents a system activity for audit trail
type ActivityLog struct {
    ID           string                 `json:"id" db:"id"`
    UserID       *string                `json:"user_id,omitempty" db:"user_id"`
    Action       string                 `json:"action" db:"action"`
    ResourceType *string                `json:"resource_type,omitempty" db:"resource_type"`
    ResourceID   *string                `json:"resource_id,omitempty" db:"resource_id"`
    Details      map[string]interface{} `json:"details" db:"details"`
    IPAddress    *string                `json:"ip_address,omitempty" db:"ip_address"`
    UserAgent    *string                `json:"user_agent,omitempty" db:"user_agent"`
    CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// Activity action constants
const (
    // User actions
    ActionUserCreated      = "user.created"
    ActionUserUpdated      = "user.updated"
    ActionUserDeleted      = "user.deleted"
    ActionUserDisabled     = "user.disabled"
    ActionUserEnabled      = "user.enabled"
    ActionUserRoleChanged  = "user.role_changed"

    // Quiz actions
    ActionQuizCreated      = "quiz.created"
    ActionQuizPublished    = "quiz.published"
    ActionQuizUnpublished  = "quiz.unpublished"
    ActionQuizDeleted      = "quiz.deleted"

    // Attempt actions
    ActionAttemptStarted   = "attempt.started"
    ActionAttemptCompleted = "attempt.completed"

    // Admin actions
    ActionAdminLogin       = "admin.login"
    ActionAdminAccess      = "admin.access"
)
```

### Admin Statistics Models

```go
// PlatformStats represents overall platform statistics
type PlatformStats struct {
    // User statistics
    TotalUsers          int       `json:"total_users"`
    ActiveUsers         int       `json:"active_users"` // Last 30 days
    NewUsersToday       int       `json:"new_users_today"`
    NewUsersThisWeek    int       `json:"new_users_this_week"`
    NewUsersThisMonth   int       `json:"new_users_this_month"`
    DisabledUsers       int       `json:"disabled_users"`

    // Quiz statistics
    TotalQuizzes        int       `json:"total_quizzes"`
    PublishedQuizzes    int       `json:"published_quizzes"`
    PrivateQuizzes      int       `json:"private_quizzes"`
    QuizzesCreatedToday int       `json:"quizzes_created_today"`
    TotalQuestions      int       `json:"total_questions"`

    // Attempt statistics
    TotalAttempts       int       `json:"total_attempts"`
    CompletedAttempts   int       `json:"completed_attempts"`
    AttemptsToday       int       `json:"attempts_today"`
    AverageScore        float64   `json:"average_score"`

    // Storage statistics
    TotalUploads        int       `json:"total_uploads"`
    StorageUsedMB       float64   `json:"storage_used_mb"`

    // System health
    DatabaseStatus      string    `json:"database_status"`
    APIResponseTimeMs   float64   `json:"api_response_time_ms"`
    LastBackupAt        *time.Time `json:"last_backup_at,omitempty"`

    GeneratedAt         time.Time `json:"generated_at"`
}

// UserActivity represents user activity over time
type UserActivity struct {
    Date          string `json:"date"`
    NewUsers      int    `json:"new_users"`
    ActiveUsers   int    `json:"active_users"`
    QuizzesCreated int   `json:"quizzes_created"`
    AttemptsCount int    `json:"attempts_count"`
}

// AdminUserDetail represents detailed user info for admin
type AdminUserDetail struct {
    // Basic info
    ID              string     `json:"id"`
    Email           string     `json:"email"`
    Name            string     `json:"name"`
    PhoneNumber     *string    `json:"phone_number,omitempty"`
    Role            string     `json:"role"`
    EmailVerified   bool       `json:"email_verified"`
    Enabled         bool       `json:"enabled"`

    // Statistics
    QuizzesCreated  int        `json:"quizzes_created"`
    QuizzesTaken    int        `json:"quizzes_taken"`
    AverageScore    float64    `json:"average_score"`

    // Metadata
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty"`
    LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

// AdminQuizDetail represents detailed quiz info for admin
type AdminQuizDetail struct {
    ID              string     `json:"id"`
    Title           string     `json:"title"`
    Description     *string    `json:"description,omitempty"`
    Creator         UserSummary `json:"creator"`
    Category        *string    `json:"category,omitempty"`
    Difficulty      *string    `json:"difficulty,omitempty"`
    IsPublished     bool       `json:"is_published"`
    IsPublic        bool       `json:"is_public"`

    // Statistics
    QuestionCount   int        `json:"question_count"`
    AttemptCount    int        `json:"attempt_count"`
    ViewCount       int        `json:"view_count"`
    AverageScore    float64    `json:"average_score"`

    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
    DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}
```

---

## API Endpoints

### Admin Middleware

All admin endpoints require:
- Valid JWT token
- User role = "admin"

```go
// Apply to routes: router.Use(middleware.RequireAdmin())
```

---

### User Management Endpoints

#### 1. List All Users
```
GET /api/v1/admin/users
```

**Authorization**: Admin only

**Query Parameters**:
- `page` - Page number (default: 1)
- `limit` - Results per page (default: 20, max: 100)
- `role` - Filter by role (user, admin)
- `enabled` - Filter by status (true, false)
- `email_verified` - Filter by verification status
- `search` - Search by name or email
- `sort_by` - created_at, updated_at, name
- `sort_order` - asc, desc

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "user-uuid",
        "email": "john@example.com",
        "name": "John Doe",
        "role": "user",
        "email_verified": true,
        "enabled": true,
        "quizzes_created": 5,
        "quizzes_taken": 12,
        "created_at": "2025-12-01T00:00:00Z",
        "last_login_at": "2026-01-10T10:00:00Z"
      }
    ],
    "total": 150,
    "page": 1,
    "limit": 20,
    "total_pages": 8
  }
}
```

---

#### 2. Get User Details
```
GET /api/v1/admin/users/:id
```

**Authorization**: Admin only

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "id": "user-uuid",
    "email": "john@example.com",
    "name": "John Doe",
    "phone_number": "+1234567890",
    "role": "user",
    "email_verified": true,
    "enabled": true,
    "quizzes_created": 5,
    "quizzes_taken": 12,
    "average_score": 78.5,
    "created_at": "2025-12-01T00:00:00Z",
    "updated_at": "2026-01-10T10:00:00Z",
    "last_login_at": "2026-01-10T10:00:00Z"
  }
}
```

---

#### 3. Change User Role
```
PATCH /api/v1/admin/users/:id/role
```

**Authorization**: Admin only

**Request Body**:
```json
{
  "role": "admin"
}
```

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "User role updated to admin"
}
```

**Business Rules**:
- Only admins can change roles
- Cannot change own role (prevent lockout)
- Logs action to activity_logs

---

#### 4. Disable User Account
```
POST /api/v1/admin/users/:id/disable
```

**Authorization**: Admin only

**Request Body**:
```json
{
  "reason": "Violating terms of service"
}
```

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "User account disabled"
}
```

**Business Rules**:
- Sets `enabled = false`
- User cannot login
- Reason stored in activity log
- Cannot disable own account

---

#### 5. Enable User Account
```
POST /api/v1/admin/users/:id/enable
```

**Authorization**: Admin only

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "User account enabled"
}
```

---

#### 6. Permanently Delete User
```
DELETE /api/v1/admin/users/:id
```

**Authorization**: Admin only

**Query Parameters**:
- `permanent=true` - Required to confirm deletion

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "User permanently deleted"
}
```

**Business Rules**:
- Hard delete (removes from database)
- Cascades to all user data (quizzes, attempts)
- Cannot delete own account
- Cannot be undone - requires confirmation

---

### Quiz Moderation Endpoints

#### 7. List All Quizzes (Admin View)
```
GET /api/v1/admin/quizzes
```

**Authorization**: Admin only

**Query Parameters**:
- `page`, `limit` - Pagination
- `is_published` - Filter by publish status
- `is_public` - Filter by visibility
- `creator_id` - Filter by creator
- `category` - Filter by category
- `search` - Search by title
- `sort_by` - created_at, attempt_count, view_count
- `sort_order` - asc, desc
- `include_deleted` - Include soft-deleted quizzes

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "quizzes": [
      {
        "id": "quiz-uuid",
        "title": "JavaScript Quiz",
        "creator": {
          "id": "user-uuid",
          "name": "John Doe"
        },
        "category": "programming",
        "is_published": true,
        "is_public": true,
        "question_count": 20,
        "attempt_count": 150,
        "view_count": 500,
        "average_score": 78.5,
        "created_at": "2026-01-05T00:00:00Z",
        "deleted_at": null
      }
    ],
    "total": 75,
    "page": 1,
    "limit": 20
  }
}
```

---

#### 8. Get Quiz Details (Admin View)
```
GET /api/v1/admin/quizzes/:id
```

**Authorization**: Admin only

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "id": "quiz-uuid",
    "title": "JavaScript Quiz",
    "description": "Test your knowledge",
    "creator": {
      "id": "user-uuid",
      "name": "John Doe",
      "email": "john@example.com"
    },
    "category": "programming",
    "difficulty": "medium",
    "is_published": true,
    "is_public": true,
    "question_count": 20,
    "attempt_count": 150,
    "view_count": 500,
    "average_score": 78.5,
    "created_at": "2026-01-05T00:00:00Z",
    "questions": [
      {
        "id": "q1",
        "question_text": "What is a closure?",
        "question_type": "multiple_choice",
        "points": 5
      }
    ]
  }
}
```

---

#### 9. Delete Quiz (Admin)
```
DELETE /api/v1/admin/quizzes/:id
```

**Authorization**: Admin only

**Query Parameters**:
- `permanent=true` - Hard delete (optional)

**Request Body**:
```json
{
  "reason": "Inappropriate content"
}
```

**Response (200 OK)**:
```json
{
  "success": true,
  "message": "Quiz deleted"
}
```

**Business Rules**:
- Default: Soft delete (sets deleted_at)
- With `permanent=true`: Hard delete (removes from DB)
- Reason logged to activity_logs

---

### Statistics Endpoints

#### 10. Get Platform Statistics
```
GET /api/v1/admin/stats
```

**Authorization**: Admin only

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "total_users": 1250,
    "active_users": 450,
    "new_users_today": 12,
    "new_users_this_week": 78,
    "new_users_this_month": 320,
    "disabled_users": 5,
    "total_quizzes": 450,
    "published_quizzes": 380,
    "private_quizzes": 70,
    "quizzes_created_today": 8,
    "total_questions": 9000,
    "total_attempts": 15000,
    "completed_attempts": 14200,
    "attempts_today": 125,
    "average_score": 75.5,
    "total_uploads": 250,
    "storage_used_mb": 850.5,
    "database_status": "healthy",
    "api_response_time_ms": 45.2,
    "generated_at": "2026-01-10T10:00:00Z"
  }
}
```

---

#### 11. Get User Activity Timeline
```
GET /api/v1/admin/stats/activity
```

**Authorization**: Admin only

**Query Parameters**:
- `days` - Number of days (default: 30, max: 90)

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "activity": [
      {
        "date": "2026-01-10",
        "new_users": 12,
        "active_users": 450,
        "quizzes_created": 8,
        "attempts_count": 125
      },
      {
        "date": "2026-01-09",
        "new_users": 15,
        "active_users": 430,
        "quizzes_created": 10,
        "attempts_count": 110
      }
    ]
  }
}
```

---

#### 12. Get Activity Logs
```
GET /api/v1/admin/logs
```

**Authorization**: Admin only

**Query Parameters**:
- `page`, `limit` - Pagination
- `user_id` - Filter by user
- `action` - Filter by action type
- `resource_type` - Filter by resource (quiz, user, attempt)
- `start_date`, `end_date` - Date range

**Response (200 OK)**:
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log-uuid",
        "user_id": "user-uuid",
        "action": "user.disabled",
        "resource_type": "user",
        "resource_id": "target-user-uuid",
        "details": {
          "reason": "Violating terms",
          "admin_name": "Admin User"
        },
        "ip_address": "192.168.1.1",
        "created_at": "2026-01-10T10:00:00Z"
      }
    ],
    "total": 500,
    "page": 1,
    "limit": 50
  }
}
```

---

## Business Logic (Core Layer)

### Admin User Management

```go
func DisableUser(ctx context.Context, adminID string, targetUserID string, reason string) error {
    // 1. Get admin user
    admin := dao.GetUserByID(adminID)
    if admin.Role != "admin" {
        return commons.Forbidden("admin-access-required")
    }

    // 2. Prevent self-disable
    if adminID == targetUserID {
        return commons.BadInput("cannot-disable-self")
    }

    // 3. Get target user
    targetUser := dao.GetUserByID(targetUserID)
    if targetUser == nil {
        return commons.NotFound("user-not-found")
    }

    // 4. Disable user
    err := dao.UpdateUserEnabled(targetUserID, false)
    if err != nil {
        return err
    }

    // 5. Log activity
    LogActivity(ctx, adminID, ActionUserDisabled, "user", targetUserID, map[string]interface{}{
        "reason": reason,
        "admin_name": admin.Name,
    })

    return nil
}

func ChangeUserRole(ctx context.Context, adminID string, targetUserID string, newRole string) error {
    // 1. Validate admin
    admin := dao.GetUserByID(adminID)
    if admin.Role != "admin" {
        return commons.Forbidden("admin-access-required")
    }

    // 2. Prevent self role change
    if adminID == targetUserID {
        return commons.BadInput("cannot-change-own-role")
    }

    // 3. Validate new role
    if newRole != "user" && newRole != "admin" {
        return commons.BadInput("invalid-role")
    }

    // 4. Update role
    err := dao.UpdateUserRole(targetUserID, newRole)
    if err != nil {
        return err
    }

    // 5. Log activity
    LogActivity(ctx, adminID, ActionUserRoleChanged, "user", targetUserID, map[string]interface{}{
        "new_role": newRole,
        "admin_name": admin.Name,
    })

    return nil
}
```

---

### Quiz Moderation

```go
func DeleteQuizAsAdmin(ctx context.Context, adminID string, quizID string, reason string, permanent bool) error {
    // 1. Validate admin
    admin := dao.GetUserByID(adminID)
    if admin.Role != "admin" {
        return commons.Forbidden("admin-access-required")
    }

    // 2. Get quiz
    quiz := dao.GetQuizByID(quizID)
    if quiz == nil {
        return commons.NotFound("quiz-not-found")
    }

    // 3. Delete quiz
    if permanent {
        err := dao.HardDeleteQuiz(quizID)
        if err != nil {
            return err
        }
    } else {
        err := dao.SoftDeleteQuiz(quizID)
        if err != nil {
            return err
        }
    }

    // 4. Log activity
    LogActivity(ctx, adminID, ActionQuizDeleted, "quiz", quizID, map[string]interface{}{
        "reason": reason,
        "permanent": permanent,
        "admin_name": admin.Name,
        "quiz_title": quiz.Title,
    })

    return nil
}
```

---

### Activity Logging

```go
func LogActivity(ctx context.Context, userID string, action string, resourceType string, resourceID string, details map[string]interface{}) {
    // Extract IP and User-Agent from context
    ipAddress := getIPFromContext(ctx)
    userAgent := getUserAgentFromContext(ctx)

    log := &ActivityLog{
        ID:           uuid.New().String(),
        UserID:       &userID,
        Action:       action,
        ResourceType: &resourceType,
        ResourceID:   &resourceID,
        Details:      details,
        IPAddress:    &ipAddress,
        UserAgent:    &userAgent,
        CreatedAt:    time.Now(),
    }

    dao.CreateActivityLog(log)
}

// Automatically log certain actions via middleware or hooks
func LogUserLogin(userID string, ctx *gin.Context) {
    LogActivity(ctx, userID, ActionAdminLogin, "", "", map[string]interface{}{
        "login_time": time.Now(),
    })
}
```

---

### Platform Statistics

```go
func GetPlatformStats(ctx context.Context) (*PlatformStats, error) {
    stats := &PlatformStats{
        GeneratedAt: time.Now(),
    }

    // User statistics
    stats.TotalUsers = dao.CountUsers()
    stats.ActiveUsers = dao.CountActiveUsers(30) // Last 30 days
    stats.NewUsersToday = dao.CountUsersCreatedAfter(time.Now().Truncate(24 * time.Hour))
    stats.NewUsersThisWeek = dao.CountUsersCreatedAfter(time.Now().AddDate(0, 0, -7))
    stats.NewUsersThisMonth = dao.CountUsersCreatedAfter(time.Now().AddDate(0, -1, 0))
    stats.DisabledUsers = dao.CountDisabledUsers()

    // Quiz statistics
    stats.TotalQuizzes = dao.CountQuizzes()
    stats.PublishedQuizzes = dao.CountQuizzesWhere("is_published = true")
    stats.PrivateQuizzes = dao.CountQuizzesWhere("is_public = false")
    stats.QuizzesCreatedToday = dao.CountQuizzesCreatedAfter(time.Now().Truncate(24 * time.Hour))
    stats.TotalQuestions = dao.CountQuestions()

    // Attempt statistics
    stats.TotalAttempts = dao.CountAttempts()
    stats.CompletedAttempts = dao.CountAttemptsWhere("status = 'submitted'")
    stats.AttemptsToday = dao.CountAttemptsCreatedAfter(time.Now().Truncate(24 * time.Hour))
    stats.AverageScore = dao.GetAverageAttemptScore()

    // System health
    stats.DatabaseStatus = checkDatabaseHealth()
    stats.APIResponseTimeMs = getAverageResponseTime()

    return stats, nil
}

func GetUserActivityTimeline(ctx context.Context, days int) ([]UserActivity, error) {
    if days > 90 {
        days = 90
    }

    startDate := time.Now().AddDate(0, 0, -days).Truncate(24 * time.Hour)

    // Query daily aggregates
    activities := dao.GetDailyUserActivity(startDate)

    return activities, nil
}
```

---

## Database Queries

### Count Active Users
```sql
-- Count users with activity in last N days
SELECT COUNT(DISTINCT user_id)
FROM quiz_attempts
WHERE created_at >= NOW() - INTERVAL '30 days';
```

### Daily Activity Aggregates
```sql
-- Get daily activity statistics
SELECT
    DATE(created_at) as date,
    COUNT(DISTINCT user_id) as active_users,
    COUNT(*) as attempts_count
FROM quiz_attempts
WHERE created_at >= $1
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Combine with user signups
SELECT
    DATE(u.created_at) as date,
    COUNT(*) as new_users
FROM users u
WHERE u.created_at >= $1
GROUP BY DATE(u.created_at)
ORDER BY date DESC;

-- Combine with quiz creations
SELECT
    DATE(q.created_at) as date,
    COUNT(*) as quizzes_created
FROM quizzes q
WHERE q.created_at >= $1
GROUP BY DATE(q.created_at)
ORDER BY date DESC;
```

### Admin User List Query
```sql
SELECT
    u.id,
    u.email,
    u.name,
    u.role,
    u.email_verified,
    u.enabled,
    u.created_at,
    u.updated_at,
    COUNT(DISTINCT q.id) as quizzes_created,
    COUNT(DISTINCT qa.id) as quizzes_taken
FROM users u
LEFT JOIN quizzes q ON u.id = q.creator_id AND q.deleted_at IS NULL
LEFT JOIN quiz_attempts qa ON u.id = qa.user_id
WHERE u.deleted_at IS NULL
  AND ($1::text IS NULL OR u.email ILIKE '%' || $1 || '%' OR u.name ILIKE '%' || $1 || '%')
  AND ($2::text IS NULL OR u.role = $2)
  AND ($3::boolean IS NULL OR u.enabled = $3)
GROUP BY u.id
ORDER BY u.created_at DESC
LIMIT $4 OFFSET $5;
```

---

## Implementation Timeline

### Day 1: Database & Models (5-6 hours)
**Tasks**:
- [ ] Create migration `000011_add_user_roles.up.sql`
  - Add role column to users table
  - Add check constraint
  - Create index
- [ ] Create migration `000012_create_activity_logs_table.up.sql`
  - Create activity_logs table
  - Create indexes
- [ ] Create down migrations
- [ ] Run migrations on dev database
- [ ] Create `admin/model/activity_log.go`
- [ ] Create `admin/model/admin_stats.go`
- [ ] Test migrations

**Acceptance Criteria**:
- ✅ Role column added successfully
- ✅ Activity logs table created
- ✅ Models compile without errors

---

### Day 2: DAO Layer & Activity Logging (6-7 hours)
**Tasks**:
- [ ] Create `admin/dao/activity_log.go`
  - CreateActivityLog
  - GetActivityLogs (with filters)
  - CountActivityLogs
- [ ] Create `admin/dao/admin_stats.go`
  - Count functions (users, quizzes, attempts)
  - GetDailyUserActivity
  - GetPlatformAggregates
- [ ] Create `admin/core/activity_logger.go`
  - LogActivity helper
  - Context extraction functions
- [ ] Add activity logging to existing endpoints
  - User login
  - Quiz creation/deletion
  - Admin actions
- [ ] Test DAO functions

**Acceptance Criteria**:
- ✅ Activity logs saved correctly
- ✅ Queries return accurate data
- ✅ Context extraction works

---

### Day 3: Core Logic & Statistics (6-7 hours)
**Tasks**:
- [ ] Create `admin/core/user_management.go`
  - DisableUser
  - EnableUser
  - ChangeUserRole
  - PermanentlyDeleteUser
  - ListUsers (admin view)
- [ ] Create `admin/core/quiz_moderation.go`
  - ListQuizzesAdmin
  - GetQuizDetailsAdmin
  - DeleteQuizAsAdmin
- [ ] Create `admin/core/statistics.go`
  - GetPlatformStats
  - GetUserActivityTimeline
- [ ] Test all core functions
- [ ] Test authorization logic

**Acceptance Criteria**:
- ✅ User management functions work
- ✅ Quiz moderation functions work
- ✅ Statistics calculated correctly
- ✅ All actions logged

---

### Day 4: API, Middleware & Testing (6-7 hours)
**Tasks**:
- [ ] Create `internal/middleware/admin.go`
  - RequireAdmin middleware
  - Check user role
  - Return 403 if not admin
- [ ] Create `api/admin.go`
- [ ] Implement all admin endpoints
  - User management (6 endpoints)
  - Quiz moderation (3 endpoints)
  - Statistics (3 endpoints)
- [ ] Add admin routes to main.go
  - Protected with RequireAdmin middleware
- [ ] Test all endpoints manually
  - As admin user
  - As regular user (should fail)
  - Authorization checks
  - Activity logging
- [ ] Fix any bugs found
- [ ] Update API documentation

**Acceptance Criteria**:
- ✅ All endpoints functional
- ✅ Admin authorization enforced
- ✅ Regular users blocked
- ✅ Activity logs created for all actions
- ✅ Statistics accurate

---

## Middleware Implementation

### Admin Authorization Middleware

```go
// internal/middleware/admin.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "quiz-craft/commons"
    "quiz-craft/user/dao"
)

// RequireAdmin middleware ensures the user has admin role
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get user ID from JWT (set by RequireAuth middleware)
        userID, exists := c.Get("user_id")
        if !exists {
            commons.SendError(c, commons.Unauthorized("authentication-required"))
            c.Abort()
            return
        }

        // Get user from database
        user, err := dao.UserDAO.GetByID(userID.(string))
        if err != nil {
            commons.SendError(c, commons.NotFound("user-not-found"))
            c.Abort()
            return
        }

        // Check role
        if user.Role != "admin" {
            commons.SendError(c, commons.Forbidden("admin-access-required"))
            c.Abort()
            return
        }

        // Set role in context for later use
        c.Set("user_role", user.Role)
        c.Next()
    }
}
```

### Apply to Routes

```go
// main.go
admin := v1.Group("/admin")
admin.Use(middleware.RequireAuth())
admin.Use(middleware.RequireAdmin())
{
    // User management
    admin.GET("/users", adminapi.ListUsers)
    admin.GET("/users/:id", adminapi.GetUserDetails)
    admin.PATCH("/users/:id/role", adminapi.ChangeUserRole)
    admin.POST("/users/:id/disable", adminapi.DisableUser)
    admin.POST("/users/:id/enable", adminapi.EnableUser)
    admin.DELETE("/users/:id", adminapi.DeleteUser)

    // Quiz moderation
    admin.GET("/quizzes", adminapi.ListQuizzesAdmin)
    admin.GET("/quizzes/:id", adminapi.GetQuizDetailsAdmin)
    admin.DELETE("/quizzes/:id", adminapi.DeleteQuizAdmin)

    // Statistics
    admin.GET("/stats", adminapi.GetPlatformStats)
    admin.GET("/stats/activity", adminapi.GetUserActivity)
    admin.GET("/logs", adminapi.GetActivityLogs)
}
```

---

## Validation Rules

### Role Change Validation
```go
func validateRoleChange(adminID, targetUserID, newRole string) error {
    if adminID == targetUserID {
        return commons.BadInput("cannot-change-own-role")
    }

    validRoles := []string{"user", "admin"}
    if !contains(validRoles, newRole) {
        return commons.BadInput("invalid-role")
    }

    return nil
}
```

### User Disable Validation
```go
func validateUserDisable(adminID, targetUserID string) error {
    if adminID == targetUserID {
        return commons.BadInput("cannot-disable-self")
    }

    return nil
}
```

---

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `admin-access-required` | 403 | User is not an admin |
| `cannot-disable-self` | 400 | Admin cannot disable own account |
| `cannot-change-own-role` | 400 | Admin cannot change own role |
| `invalid-role` | 400 | Invalid role value |
| `user-not-found` | 404 | User doesn't exist |
| `quiz-not-found` | 404 | Quiz doesn't exist |

---

## Security Considerations

### Admin Access Control
- ✅ All admin endpoints require `role = 'admin'`
- ✅ Prevent admins from modifying themselves (disable, delete, role change)
- ✅ All admin actions logged for audit trail
- ✅ IP address and user agent tracked

### Activity Logging
- ✅ All sensitive actions logged
- ✅ Cannot be deleted by regular users
- ✅ Includes reason for moderation actions
- ✅ Tracks who performed the action

### Data Privacy
- ✅ Admin can view user emails but not passwords
- ✅ Activity logs don't expose sensitive data
- ✅ Hard delete requires explicit confirmation

---

## Testing Strategy

### Manual Testing

**Test Case 1: Admin Authorization**
```bash
# Try accessing as regular user
GET /api/v1/admin/users
Authorization: Bearer <regular_user_token>
→ Should return 403 Forbidden

# Try accessing as admin
GET /api/v1/admin/users
Authorization: Bearer <admin_token>
→ Should return 200 OK with user list
```

**Test Case 2: User Management**
```bash
# Disable user
POST /api/v1/admin/users/{user_id}/disable
Body: {"reason": "Spam"}
→ Should disable user and log action

# Try to disable self
POST /api/v1/admin/users/{admin_own_id}/disable
→ Should return 400 "cannot-disable-self"
```

**Test Case 3: Activity Logging**
```bash
# Perform admin action
POST /api/v1/admin/users/{user_id}/disable

# Check logs
GET /api/v1/admin/logs?action=user.disabled
→ Should show log entry with reason and admin info
```

**Test Case 4: Platform Statistics**
```bash
# Get platform stats
GET /api/v1/admin/stats
→ Should return all statistics

# Get activity timeline
GET /api/v1/admin/stats/activity?days=30
→ Should return 30 days of activity data
```

---

## Success Metrics

Phase 7 is successful when:

✅ **Access Control**:
- Admin middleware enforces role correctly
- Regular users blocked from admin endpoints
- Admins cannot modify themselves

✅ **User Management**:
- Can view all users with filters
- Can disable/enable users
- Can change user roles
- Can delete users (soft and hard)

✅ **Quiz Moderation**:
- Can view all quizzes (including private)
- Can delete inappropriate quizzes
- Deletion reasons tracked

✅ **Statistics**:
- Platform stats accurate and up-to-date
- Activity timeline shows trends
- Activity logs complete and searchable

✅ **Audit Trail**:
- All admin actions logged
- Logs include necessary context
- Logs cannot be tampered with

---

## Critical Files Summary

### New Files (7 files):
1. `pkg/database/migrations/000011_add_user_roles.up.sql`
2. `pkg/database/migrations/000011_add_user_roles.down.sql`
3. `pkg/database/migrations/000012_create_activity_logs_table.up.sql`
4. `pkg/database/migrations/000012_create_activity_logs_table.down.sql`
5. `admin/model/activity_log.go`
6. `admin/model/admin_stats.go`
7. `admin/dao/activity_log.go`
8. `admin/dao/admin_stats.go`
9. `admin/core/user_management.go`
10. `admin/core/quiz_moderation.go`
11. `admin/core/statistics.go`
12. `admin/core/activity_logger.go`
13. `internal/middleware/admin.go`
14. `api/admin.go`

### Modified Files (2 files):
1. `user/model/user.go` - Add Role field
2. `main.go` - Add admin routes

---

## Dependencies

**No New Dependencies Required** ✅

---

## Total Estimated Effort

**Implementation**: 20-24 hours (4 working days at 5-6 hours/day)

**Breakdown**:
- Day 1: Database & Models (5-6 hours)
- Day 2: DAO & Logging (6-7 hours)
- Day 3: Core Logic (6-7 hours)
- Day 4: API & Testing (6-7 hours)

**Complexity**: Low-Medium (straightforward CRUD, statistics aggregation)

**Risk Level**: Low (no external dependencies, standard patterns)

---

**Ready to implement!** 🚀
