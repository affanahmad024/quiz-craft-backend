# Educator Role & Live Quiz Sessions - Architecture & Approach

## Executive Summary

This document outlines the phased implementation plan for adding an **Educator Role** and **Unified Live Quiz Sessions** to the QuizCraft platform.

**Educator Features:**
- Users can self-promote to educators (with basic verification checks)
- Gain followers through a subscription system with in-app notifications
- Receive detailed analytics about learner performance
- Create public live quiz sessions with full host control

**Live Quiz Sessions (Unified Architecture):**
- **Educator Public Sessions**: Scheduled, discoverable events with host control
- **User Private Sessions**: Code-based group quizzes with automatic progression
- Same technical infrastructure supports both use cases
- Real-time competitive gameplay with live leaderboards
- Post-session detailed analysis for all participants

---

## 📋 Table of Contents

1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Feature Requirements](#feature-requirements)
3. [Unified Live Session Architecture](#unified-live-session-architecture)
4. [In-App Notification System](#in-app-notification-system)
5. [Architecture Design](#architecture-design)
6. [Phased Implementation Plan](#phased-implementation-plan)
7. [Database Schema Changes](#database-schema-changes)
8. [API Endpoints](#api-endpoints)
9. [Security & Access Control](#security--access-control)
10. [Analytics & Reporting](#analytics--reporting)
11. [Enhanced Features & Ideas](#enhanced-features--ideas)
12. [Testing Strategy](#testing-strategy)

---

## Current Architecture Analysis

### Existing Role System
- **Roles**: `user`, `admin`
- **Storage**: `users.role` column with CHECK constraint
- **Middleware**: `middleware.RequireAdmin()` for admin-only routes
- **Pattern**: Role-based access control (RBAC) at route level

### Existing Quiz System
- **Ownership**: Each quiz has `creator_id` linking to user
- **Visibility**: Controlled by `is_published` and `is_public` flags
- **Access Pattern**: Owner has full control, others read-only if public

### Existing Analytics
- Quiz stats: Total attempts, average score, completion rate
- User stats: Quizzes created, attempts taken, performance
- **Gap**: No per-learner analytics for quiz creators

---

## Feature Requirements

### Core Requirements

#### 1. Educator Role Management (Self-Service)
- Users can promote themselves to educators instantly
- Basic automated requirements: email verified, phone number added
- Admins can demote educators if they misuse the platform
- Educators can create and manage their quiz portfolio
- Educators have enhanced profile with bio, expertise, credentials

#### 2. Subscription System (Follow/Unfollow)
- Users can follow/unfollow educators
- Followers get **in-app notifications** when educators publish new quizzes
- Educators can see their follower count and list
- All educator quizzes are public (no premium/subscriber-only access)

#### 3. Live Quiz Sessions (Unified System)

**A. Educator Public Sessions:**
- Educators can schedule and host public live quiz sessions
- Manual question progression (host controls when to advance)
- Discoverable in public session listings
- Notification to all followers when session is scheduled
- Real-time competitive gameplay with live leaderboards
- Post-session analytics for educator

**B. User Private Sessions:**
- ANY user can create instant or scheduled private live quiz
- Automatic timer-based question progression
- Access via shareable code
- Direct user invitations with in-app notifications
- Same real-time features (WebSocket sync, leaderboards)
- Post-session detailed analysis visible to all participants

#### 4. Detailed Analytics for Educators
Educators can see for each quiz:
- **Aggregate Stats**: Total attempts, completion rate, average score, pass/fail ratio
- **Per-Learner Analytics**:
  - Individual learner scores and attempt history
  - Time spent per question
  - Common wrong answers (question-level analysis)
  - Learning progression over multiple attempts
- **Question Analytics**:
  - Difficulty analysis (which questions have lowest success rate)
  - Time distribution per question
  - Most commonly missed questions
- **Cohort Analysis**: Compare performance across different time periods or subscriber groups

---

## Unified Live Session Architecture

### Design Philosophy

Instead of building two separate systems (educator sessions vs user group sessions), we build **ONE unified system** that supports both use cases through configuration flags and smart access control.

### Key Differentiators

| Feature | Educator Public Sessions | User Private Sessions |
|---------|-------------------------|----------------------|
| **Creator** | Educators only | ANY user |
| **Discovery** | Public, discoverable | Private, code-based |
| **Access** | Open to all users | Requires access code |
| **Progression** | Manual (host control) | Automatic (timer-based) |
| **Scheduling** | Always scheduled | Instant OR scheduled |
| **Notifications** | Sent to all followers | Sent to invited users only |
| **Post-Session Analytics** | Visible to host (educator) | Visible to ALL participants |

### Unified Database Schema

```sql
CREATE TABLE live_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id),
    creator_id UUID NOT NULL REFERENCES users(id),  -- ANY user, not just educators

    -- Session type determines behavior
    session_type VARCHAR(20) NOT NULL,  -- 'public' or 'private'

    -- Basic info
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Timing (flexible)
    scheduled_at TIMESTAMP,  -- NULL = instant, NOT NULL = scheduled
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'waiting',  -- waiting, live, completed, cancelled

    -- Access control
    access_code VARCHAR(10),  -- Generated for private sessions
    is_discoverable BOOLEAN DEFAULT false,  -- Public listing?
    max_participants INT,

    -- Progression mode
    progression_mode VARCHAR(20) NOT NULL,  -- 'manual' or 'automatic'
    question_timer_seconds INT,  -- For automatic mode
    current_question_index INT DEFAULT 0,

    -- Config (flexible for future features)
    config JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CHECK (session_type IN ('public', 'private')),
    CHECK (progression_mode IN ('manual', 'automatic')),
    CHECK (status IN ('waiting', 'live', 'completed', 'cancelled')),

    -- Business rules
    CONSTRAINT valid_public_session CHECK (
        session_type = 'private' OR
        (session_type = 'public' AND scheduled_at IS NOT NULL)
    ),
    CONSTRAINT valid_private_session CHECK (
        session_type = 'public' OR
        (session_type = 'private' AND access_code IS NOT NULL)
    )
);

-- Session invitations (for private sessions)
CREATE TABLE live_session_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id),
    invitee_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',  -- pending, accepted, declined
    invited_at TIMESTAMP DEFAULT NOW(),
    responded_at TIMESTAMP,

    UNIQUE (session_id, invitee_id),
    CHECK (status IN ('pending', 'accepted', 'declined'))
);
```

### Access Control Logic

```go
// File: backend/internal/services/live_session_service.go

func (s *LiveSessionService) CanCreateSession(userID uuid.UUID, sessionType string) error {
    if sessionType == "public" {
        // Only educators can create public sessions
        if !s.IsEducator(userID) {
            return errors.New("only educators can create public sessions")
        }
    }
    // Anyone can create private sessions
    return nil
}

func (s *LiveSessionService) CanJoinSession(userID uuid.UUID, session *LiveSession) error {
    if session.SessionType == "public" {
        // Public sessions are open to all
        return nil
    }

    // Private sessions require access code OR invitation
    // This check happens at the API layer when user provides code
    return nil
}

func (s *LiveSessionService) CanControlSession(userID uuid.UUID, session *LiveSession) error {
    // Only the creator can control the session
    if session.CreatorID != userID {
        return errors.New("only session creator can control the session")
    }

    // Manual control only available for manual progression mode
    if session.ProgressionMode != "manual" {
        return errors.New("this session uses automatic progression")
    }

    return nil
}
```

### Unified API Endpoints

```
# ========================================
# LIVE SESSION CREATION (UNIFIED)
# ========================================

POST   /api/v1/live-sessions
Body: {
  "quiz_id": "...",
  "title": "...",
  "description": "...",
  "session_type": "public" | "private",
  "scheduled_at": "2024-02-01T19:00:00Z" | null,  // null = instant
  "progression_mode": "manual" | "automatic",
  "question_timer_seconds": 30,  // For automatic mode
  "max_participants": 50,
  "is_discoverable": true | false
}
Response: {
  "session": {...},
  "access_code": "XYZ123"  // Only for private sessions
}

# Validation:
# - If session_type='public', require creator to be educator
# - If session_type='public', require scheduled_at
# - If session_type='private', generate access_code
# - If progression_mode='automatic', require question_timer_seconds

# ========================================
# SESSION INVITATION (PRIVATE SESSIONS)
# ========================================

POST   /api/v1/live-sessions/:id/invite
Body: {
  "user_ids": ["uuid1", "uuid2", ...]
}
Response: {
  "message": "Invitations sent",
  "invited_count": 5
}
# Creates in-app notifications for invited users
# Only private session creators can invite

GET    /api/v1/live-sessions/:id/invitations
Response: {
  "invitations": [
    {
      "user": {...},
      "status": "pending",
      "invited_at": "..."
    }
  ]
}

# ========================================
# SESSION DISCOVERY (PUBLIC vs PRIVATE)
# ========================================

# Public sessions (discoverable)
GET    /api/v1/live-sessions/public
Query: ?status=upcoming,live&page=1
Response: {
  "sessions": [...]  // Only public, discoverable sessions
}

# My sessions (creator view)
GET    /api/v1/live-sessions/mine
Response: {
  "sessions": [...]  // All my created sessions (public + private)
}

# Sessions I'm invited to
GET    /api/v1/live-sessions/invitations
Response: {
  "sessions": [...]  // Private sessions I'm invited to
}

# ========================================
# SESSION JOIN (ACCESS CONTROL)
# ========================================

POST   /api/v1/live-sessions/:id/join
Body: {
  "access_code": "XYZ123"  // Required for private, ignored for public
}
Response: {
  "message": "Joined session",
  "participant": {...}
}

# Validation:
# - Public session: anyone can join
# - Private session: must provide correct access_code OR have invitation

# ========================================
# SESSION CONTROL (HOST ONLY)
# ========================================

# Manual progression (educators with public sessions)
POST   /api/v1/live-sessions/:id/start
POST   /api/v1/live-sessions/:id/next-question
POST   /api/v1/live-sessions/:id/end

# Validation:
# - Only session creator can call these
# - Only works if progression_mode='manual'
# - Automatic sessions start/progress/end automatically

# ========================================
# POST-SESSION ANALYTICS
# ========================================

GET    /api/v1/live-sessions/:id/results
Response: {
  "session": {...},
  "final_leaderboard": [...],
  "my_performance": {...},

  // For public sessions (educator view)
  "detailed_analytics": {
    "question_stats": [...],
    "participant_performance": [...]
  },

  // For private sessions (all participants can see)
  "participant_details": [
    {
      "user": {...},
      "score": 850,
      "rank": 3,
      "answers": [
        {
          "question": {...},
          "answer": {...},
          "is_correct": true,
          "time_taken_ms": 4500
        }
      ]
    }
  ]
}

# Access control:
# - Public sessions: Only creator sees detailed_analytics
# - Private sessions: ALL participants see participant_details
```

### Progression Modes

#### Manual Progression (Educator Public Sessions)

```go
// Educator controls when to show each question
// Flow:
1. Educator calls POST /live-sessions/:id/start
2. Question 1 revealed to all participants
3. Participants submit answers
4. Educator calls POST /live-sessions/:id/next-question
5. Question 2 revealed, and so on...
6. Educator calls POST /live-sessions/:id/end when done

// Benefits:
// - Educator can discuss answers between questions
// - Can adjust pacing based on participant engagement
// - Can skip/repeat questions
```

#### Automatic Progression (User Private Sessions)

```go
// Timer automatically advances questions
// Flow:
1. Creator calls POST /live-sessions/:id/start (or auto-starts at scheduled_at)
2. Question 1 revealed with timer (e.g., 30 seconds)
3. Timer counts down via WebSocket
4. When timer reaches 0, automatically advance to Question 2
5. Repeat until all questions completed
6. Session automatically ends after last question

// Implementation:
// - Background worker monitors 'live' sessions with progression_mode='automatic'
// - Uses question_timer_seconds from config
// - Sends WebSocket events: timer_update, question_auto_advance
```

### WebSocket Events (Unified)

```json
// Common events for both session types

// Session lifecycle
{"type": "session_started", "session_id": "...", "total_questions": 20}
{"type": "question_revealed", "question": {...}, "index": 1, "timer_seconds": 30}
{"type": "participant_joined", "participant": {...}, "total_participants": 45}
{"type": "participant_left", "participant": {...}}

// Real-time scoring
{"type": "answer_submitted", "participant_id": "...", "score": 100}
{"type": "leaderboard_update", "leaderboard": [{...}]}

// Question completion
{"type": "question_ended", "correct_answer": {...}, "stats": {...}}

// Automatic progression specific
{"type": "timer_update", "seconds_remaining": 15}
{"type": "question_auto_advance", "next_question_index": 2}

// Session end
{"type": "session_ended", "final_leaderboard": [...], "session_id": "..."}
```

### Implementation Benefits

✅ **No Code Duplication**: Single codebase handles both use cases
✅ **Flexible Access Control**: Simple checks based on session_type
✅ **Unified WebSocket Infrastructure**: Same real-time system for both
✅ **Easy Testing**: Test once, works for both scenarios
✅ **Future-Proof**: Easy to add new session types (e.g., "semi-private")

---

## In-App Notification System

### Overview

The in-app notification system delivers real-time notifications to users without relying on email. Notifications are stored in the database and delivered via REST API + optional WebSocket for real-time delivery.

### Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Trigger Events                      │
│  - Educator publishes quiz                           │
│  - Educator schedules live session                   │
│  - User invited to private session                   │
│  - Live session starting (10 min reminder)           │
│  - New follower                                      │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│          Notification Service                        │
│  - CreateNotification(userID, type, data)            │
│  - CreateBulkNotifications(userIDs[], type, data)    │
│  - MarkAsRead(notificationID)                        │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│           PostgreSQL (notifications table)           │
│  - Persistent storage                                │
│  - Indexed by user_id, is_read, created_at          │
└────────────────┬────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────┐
│         Delivery Channels                            │
│  1. REST API: GET /api/v1/notifications              │
│  2. WebSocket (optional): Real-time push             │
└─────────────────────────────────────────────────────┘
```

### Database Schema

```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Notification type and content
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,

    -- Context data (JSONB for flexibility)
    data JSONB DEFAULT '{}'::jsonb,

    -- Read status
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP DEFAULT NOW(),

    CHECK (type IN (
        'new_quiz',
        'live_session_scheduled',
        'live_session_starting',  -- 10 min reminder
        'live_session_invitation',
        'new_follower',
        'system'
    ))
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_notifications_created_at ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);
```

### Notification Service Implementation

```go
// File: backend/internal/services/notification_service.go

package services

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

type NotificationService struct {
    db *sql.DB
}

type NotificationData map[string]interface{}

// CreateNotification creates a single notification
func (s *NotificationService) CreateNotification(
    ctx context.Context,
    userID uuid.UUID,
    notifType string,
    title string,
    message string,
    data NotificationData,
) error {
    dataJSON, _ := json.Marshal(data)

    query := `
        INSERT INTO notifications (user_id, type, title, message, data)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

    var notifID uuid.UUID
    err := s.db.QueryRowContext(ctx, query, userID, notifType, title, message, dataJSON).Scan(&notifID)
    if err != nil {
        return err
    }

    // Optional: Trigger real-time WebSocket delivery
    s.sendWebSocketNotification(userID, notifID)

    return nil
}

// CreateBulkNotifications creates notifications for multiple users (e.g., all followers)
func (s *NotificationService) CreateBulkNotifications(
    ctx context.Context,
    userIDs []uuid.UUID,
    notifType string,
    title string,
    message string,
    data NotificationData,
) error {
    dataJSON, _ := json.Marshal(data)

    query := `
        INSERT INTO notifications (user_id, type, title, message, data)
        SELECT unnest($1::uuid[]), $2, $3, $4, $5
    `

    _, err := s.db.ExecContext(ctx, query, pq.Array(userIDs), notifType, title, message, dataJSON)
    return err
}

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(
    ctx context.Context,
    userID uuid.UUID,
    unreadOnly bool,
    limit int,
    offset int,
) ([]Notification, error) {
    query := `
        SELECT id, type, title, message, data, is_read, read_at, created_at
        FROM notifications
        WHERE user_id = $1
    `

    if unreadOnly {
        query += " AND is_read = false"
    }

    query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"

    rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var notifications []Notification
    for rows.Next() {
        var n Notification
        var dataJSON []byte

        err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Message, &dataJSON, &n.IsRead, &n.ReadAt, &n.CreatedAt)
        if err != nil {
            return nil, err
        }

        json.Unmarshal(dataJSON, &n.Data)
        notifications = append(notifications, n)
    }

    return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID, userID uuid.UUID) error {
    query := `
        UPDATE notifications
        SET is_read = true, read_at = NOW()
        WHERE id = $1 AND user_id = $2
    `

    _, err := s.db.ExecContext(ctx, query, notificationID, userID)
    return err
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
    query := `
        UPDATE notifications
        SET is_read = true, read_at = NOW()
        WHERE user_id = $1 AND is_read = false
    `

    _, err := s.db.ExecContext(ctx, query, userID)
    return err
}

// GetUnreadCount gets the count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
    err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)
    return count, err
}
```

### Notification Triggers

#### 1. New Quiz Published (Educator → Followers)

```go
// File: backend/api/quiz.go

func PublishQuiz(c *gin.Context) {
    // ... existing publish logic ...

    // If user is educator, notify all followers
    if isEducator(userID) {
        go func() {
            followerIDs := subscriptionService.GetFollowerIDs(educatorID)

            notificationService.CreateBulkNotifications(
                context.Background(),
                followerIDs,
                "new_quiz",
                fmt.Sprintf("New Quiz from %s", educatorName),
                fmt.Sprintf("%s just published '%s'", educatorName, quizTitle),
                NotificationData{
                    "quiz_id": quizID,
                    "educator_id": educatorID,
                    "educator_name": educatorName,
                    "quiz_title": quizTitle,
                },
            )
        }()
    }

    // ... return response ...
}
```

#### 2. Live Session Scheduled (Educator → Followers)

```go
// When educator creates a public live session
func CreateLiveSession(c *gin.Context) {
    // ... create session logic ...

    if session.SessionType == "public" {
        go func() {
            followerIDs := subscriptionService.GetFollowerIDs(session.CreatorID)

            notificationService.CreateBulkNotifications(
                context.Background(),
                followerIDs,
                "live_session_scheduled",
                fmt.Sprintf("Live Session from %s", educatorName),
                fmt.Sprintf("Join '%s' on %s", session.Title, session.ScheduledAt.Format("Jan 2, 3:04 PM")),
                NotificationData{
                    "session_id": session.ID,
                    "educator_id": session.CreatorID,
                    "scheduled_at": session.ScheduledAt,
                    "title": session.Title,
                },
            )
        }()
    }
}
```

#### 3. Private Session Invitation

```go
// When user invites others to private session
func InviteToSession(c *gin.Context) {
    // ... validation logic ...

    for _, inviteeID := range req.UserIDs {
        // Create invitation record
        invitation := createInvitation(sessionID, userID, inviteeID)

        // Send notification
        notificationService.CreateNotification(
            context.Background(),
            inviteeID,
            "live_session_invitation",
            fmt.Sprintf("%s invited you to a live quiz", inviterName),
            fmt.Sprintf("Join '%s' and compete with your friends!", session.Title),
            NotificationData{
                "session_id": session.ID,
                "inviter_id": userID,
                "inviter_name": inviterName,
                "access_code": session.AccessCode,
                "scheduled_at": session.ScheduledAt,  // May be null for instant
            },
        )
    }
}
```

#### 4. Live Session Starting (Reminder)

```go
// Background worker that runs every minute
func SessionReminderWorker() {
    for {
        time.Sleep(1 * time.Minute)

        // Find sessions starting in 10 minutes
        sessions := getSessionsStartingIn(10 * time.Minute)

        for _, session := range sessions {
            if session.SessionType == "public" {
                // Notify all followers
                followerIDs := subscriptionService.GetFollowerIDs(session.CreatorID)
                notificationService.CreateBulkNotifications(
                    context.Background(),
                    followerIDs,
                    "live_session_starting",
                    "Live Session Starting Soon!",
                    fmt.Sprintf("'%s' starts in 10 minutes", session.Title),
                    NotificationData{
                        "session_id": session.ID,
                        "starts_at": session.ScheduledAt,
                    },
                )
            } else {
                // Private session: notify invited participants
                invitedUserIDs := getInvitedUserIDs(session.ID)
                notificationService.CreateBulkNotifications(
                    context.Background(),
                    invitedUserIDs,
                    "live_session_starting",
                    "Live Quiz Starting Soon!",
                    fmt.Sprintf("'%s' starts in 10 minutes", session.Title),
                    NotificationData{
                        "session_id": session.ID,
                        "starts_at": session.ScheduledAt,
                    },
                )
            }

            // Mark session as "reminder_sent" to avoid duplicate notifications
            markReminderSent(session.ID)
        }
    }
}
```

### API Endpoints

```
# ========================================
# NOTIFICATION ENDPOINTS
# ========================================

GET    /api/v1/notifications
Query: ?unread_only=true&page=1&limit=20
Response: {
  "notifications": [
    {
      "id": "...",
      "type": "new_quiz",
      "title": "New Quiz from John Doe",
      "message": "John Doe just published 'Python Basics'",
      "data": {
        "quiz_id": "...",
        "educator_id": "...",
        "educator_name": "John Doe",
        "quiz_title": "Python Basics"
      },
      "is_read": false,
      "created_at": "2024-01-27T10:30:00Z"
    },
    {
      "id": "...",
      "type": "live_session_invitation",
      "title": "Jane invited you to a live quiz",
      "message": "Join 'Friday Study Session' and compete with your friends!",
      "data": {
        "session_id": "...",
        "inviter_id": "...",
        "inviter_name": "Jane Smith",
        "access_code": "ABC123",
        "scheduled_at": "2024-01-28T19:00:00Z"
      },
      "is_read": false,
      "created_at": "2024-01-27T15:45:00Z"
    }
  ],
  "unread_count": 5,
  "total": 24
}

PATCH  /api/v1/notifications/:id/read
Response: { "message": "Notification marked as read" }

PATCH  /api/v1/notifications/read-all
Response: { "message": "All notifications marked as read", "count": 5 }

DELETE /api/v1/notifications/:id
Response: { "message": "Notification deleted" }

GET    /api/v1/notifications/unread-count
Response: { "unread_count": 5 }
```

### Optional: Real-Time WebSocket Delivery

```go
// WebSocket connection for real-time notifications
// Endpoint: WS /api/v1/notifications/ws

// When user connects
func NotificationWebSocketHandler(c *gin.Context) {
    userID := getUserIDFromToken(c)
    conn := upgradeToWebSocket(c)

    // Register connection in hub
    notificationHub.Register(userID, conn)

    // Listen for disconnect
    defer func() {
        notificationHub.Unregister(userID, conn)
    }()

    // Keep connection alive
    for {
        // Ping/pong to keep connection alive
        time.Sleep(30 * time.Second)
        conn.WriteMessage(websocket.PingMessage, []byte{})
    }
}

// When notification is created
func (s *NotificationService) sendWebSocketNotification(userID uuid.UUID, notificationID uuid.UUID) {
    // Check if user is connected via WebSocket
    if conn := notificationHub.GetConnection(userID); conn != nil {
        // Fetch the notification
        notification := s.GetNotificationByID(notificationID)

        // Send via WebSocket
        conn.WriteJSON(map[string]interface{}{
            "type": "new_notification",
            "notification": notification,
        })
    }
}
```

### Frontend Integration

```typescript
// Example frontend code

// 1. Polling approach (simpler)
setInterval(async () => {
  const response = await fetch('/api/v1/notifications/unread-count');
  const { unread_count } = await response.json();
  updateNotificationBadge(unread_count);
}, 30000); // Poll every 30 seconds

// 2. WebSocket approach (real-time)
const ws = new WebSocket('ws://localhost:8080/api/v1/notifications/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === 'new_notification') {
    // Show toast notification
    showToast(data.notification.title, data.notification.message);

    // Update badge count
    incrementNotificationBadge();

    // Add to notification list
    addNotificationToList(data.notification);
  }
};
```

### Performance Considerations

1. **Bulk Inserts**: Use `unnest()` for creating notifications for multiple users
2. **Indexing**: Proper indexes on `user_id`, `is_read`, `created_at`
3. **Pagination**: Always paginate notification lists
4. **Cleanup**: Periodic job to delete old read notifications (e.g., > 90 days)
5. **Rate Limiting**: Prevent notification spam

```go
// Cleanup worker (runs daily)
func NotificationCleanupWorker() {
    for {
        time.Sleep(24 * time.Hour)

        // Delete read notifications older than 90 days
        query := `
            DELETE FROM notifications
            WHERE is_read = true AND created_at < NOW() - INTERVAL '90 days'
        `
        db.Exec(query)
    }
}
```

---

## Architecture Design

Following the existing layered architecture pattern:

```
┌─────────────────────────────────────────┐
│           API Layer (Gin)               │
│  - Route handlers                       │
│  - Request/response marshalling         │
│  - api/educator_*.go                    │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        Middleware Layer                 │
│  - Auth middleware                      │
│  - RequireEducator() middleware         │
│  - middleware/educator.go               │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Service Layer                   │
│  - Business logic                       │
│  - Data aggregation                     │
│  - services/educator_service.go         │
│  - services/subscription_service.go     │
│  - services/analytics_service.go        │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          DAO Layer                      │
│  - Database operations                  │
│  - educator/dao/                        │
│  - subscription/dao/                    │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Database Layer                  │
│  - PostgreSQL schema                    │
│  - Migrations                           │
└─────────────────────────────────────────┘
```

### Directory Structure

```
backend/
├── educator/
│   ├── dao/
│   │   └── educator.go          # Educator CRUD operations
│   ├── model/
│   │   └── educator.go          # Educator profile model
│   └── worker/
│       └── notification.go      # Background notifications for followers
├── subscription/
│   ├── dao/
│   │   └── subscription.go      # Subscription/follow CRUD
│   └── model/
│       └── subscription.go      # Subscription model
├── livesession/
│   ├── dao/
│   │   └── session.go           # Live session CRUD
│   ├── model/
│   │   └── session.go           # Live session models
│   └── worker/
│       └── session_manager.go   # Live session orchestration
├── notification/
│   ├── dao/
│   │   └── notification.go      # In-app notification storage
│   └── model/
│       └── notification.go      # Notification models
├── analytics/
│   ├── dao/
│   │   └── analytics.go         # Analytics queries
│   └── model/
│       └── analytics.go         # Analytics models
├── api/
│   ├── educator_profile.go      # Educator profile endpoints
│   ├── educator_promotion.go    # Become educator endpoint
│   ├── subscription.go          # Follow/unfollow endpoints
│   ├── educator_analytics.go    # Analytics endpoints
│   ├── live_session.go          # Live session endpoints
│   └── notification.go          # Notification endpoints
└── internal/
    ├── middleware/
    │   └── educator.go          # RequireEducator middleware
    └── services/
        ├── educator_service.go  # Educator business logic
        ├── subscription_service.go
        ├── live_session_service.go
        ├── notification_service.go
        └── analytics_service.go
```

---

## Phased Implementation Plan

### Phase 1: Foundation - Self-Service Educator Promotion (Week 1-2)

#### Goals
- Add educator role to the system
- Create educator profile structure
- Implement self-service promotion with automated checks
- Admin controls for demotion

#### Tasks

**1.1 Database Schema**
- Migration: Add `educator` to role CHECK constraint
- Migration: Create `educator_profiles` table
- Migration: Create `notifications` table (for in-app notifications)

**1.2 Models & DAOs**
- Create `educator/model/profile.go`
- Create `educator/dao/profile.go`
- Create `notification/model/notification.go`
- Create `notification/dao/notification.go`

**1.3 Middleware**
- Create `middleware/educator.go` with `RequireEducator()`
- Create helper function `IsEducator()`

**1.4 Promotion Service**
- Create `services/educator_service.go`
- Implement `PromoteToEducator()` with validation:
  - Check if email is verified
  - Check if phone number is added
  - Check if user is not already an educator
  - Create educator profile automatically
  - Update user role to 'educator'

**1.5 API Endpoints**
```
POST   /api/v1/educator/promote        # Promote self to educator (automated checks)
GET    /api/v1/educator/profile        # Get my educator profile

# Admin endpoints
POST   /api/v1/admin/users/:id/demote  # Demote educator to user
GET    /api/v1/admin/educators         # List all educators
```

**1.6 Update Existing User APIs**
- Include educator profile in user response (if educator)
- Add educator flag to user model

#### Deliverables
- ✅ Users can promote themselves to educator instantly (with basic checks)
- ✅ Automated validation (email verified, phone number added)
- ✅ Admins can demote educators
- ✅ Educator role properly enforced via middleware
- ✅ Educator profiles created automatically upon promotion

---

### Phase 2: Educator Profiles & Discovery (Week 2-3)

#### Goals
- Build educator profile management
- Create educator discovery/listing
- Enhanced profile pages

#### Tasks

**2.1 Educator Profile CRUD**
- Update educator profile (bio, expertise, credentials, avatar)
- Get educator profile (public view)
- List educators (directory/discovery)

**2.2 API Endpoints**
```
# Educator profile management
GET    /api/v1/educators                      # List all educators (public)
GET    /api/v1/educators/:id                  # Get educator profile (public)
GET    /api/v1/educators/:id/quizzes          # List educator's quizzes (public)
PATCH  /api/v1/educator/profile               # Update my educator profile
GET    /api/v1/educator/profile               # Get my educator profile

# Search & discovery
GET    /api/v1/educators/search               # Search educators by expertise/name
GET    /api/v1/educators/featured             # Get featured educators
```

**2.3 Profile Enhancements**
- Bio/About section
- Expertise areas (tags)
- Credentials/certifications
- Social links
- Profile banner image
- Total students count
- Total quizzes count
- Average rating (future: rating system)

#### Deliverables
- ✅ Educators can manage comprehensive profiles
- ✅ Public educator directory with search
- ✅ Educator profile pages show their quiz portfolio

---

### Phase 3: Follow System & In-App Notifications (Week 3-4)

#### Goals
- Implement follow/unfollow functionality
- In-app notification system for new quizzes
- Follower management for educators

#### Tasks

**3.1 Database Schema**
- Migration: Create `subscriptions` table (renamed to `follows` conceptually, but keeping same name for consistency)
- Notifications table already created in Phase 1

**3.2 Follow/Unfollow Logic**
- Follow an educator
- Unfollow an educator
- List my followed educators
- Check if following

**3.3 API Endpoints**
```
# User follow endpoints
POST   /api/v1/educators/:id/follow           # Follow educator
DELETE /api/v1/educators/:id/unfollow         # Unfollow educator
GET    /api/v1/following                      # My followed educators
GET    /api/v1/following/feed                 # Recent quizzes from followed educators

# Educator follower management
GET    /api/v1/educator/followers             # List my followers
GET    /api/v1/educator/followers/stats       # Follower growth stats

# Notification endpoints
GET    /api/v1/notifications                  # Get my in-app notifications
PATCH  /api/v1/notifications/:id/read         # Mark notification as read
PATCH  /api/v1/notifications/read-all         # Mark all as read
DELETE /api/v1/notifications/:id              # Delete notification
```

**3.4 In-App Notification System**
- Background worker to create notifications when educator publishes quiz
- Notifications stored in database with read/unread status
- Real-time notification delivery via WebSocket (optional)
- Notification types: new_quiz, live_session_starting, etc.

**3.5 Update Quiz Publish Logic**
- When educator publishes quiz, trigger notification creation
- Create notification for each follower
- Include quiz details and educator info

#### Deliverables
- ✅ Users can follow/unfollow educators
- ✅ Educators can see follower list and count
- ✅ In-app notifications when educator publishes new quiz
- ✅ Notification management (read, delete)
- ✅ All quizzes remain public (no subscriber-only access)

---

### Phase 4: Quiz Collections & Enhanced Organization (Week 4-5)

#### Goals
- Better organization via collections/courses
- Enhanced quiz metadata for learning paths
- Public collections for structured learning

#### Tasks

**4.1 Educator Quiz Enhancements**
- Tag quizzes by topic/learning objective
- Group quizzes into collections/courses
- Enhanced quiz metadata
- Collection-based learning paths

**4.2 Database Schema**
- Migration: Create `quiz_collections` table
- Migration: Create `collection_quizzes` junction table
- Migration: Add enhanced metadata fields to quizzes (optional)

**4.3 API Endpoints**
```
# Collections
POST   /api/v1/educator/collections           # Create quiz collection
GET    /api/v1/educator/collections           # List my collections
PATCH  /api/v1/educator/collections/:id       # Update collection
DELETE /api/v1/educator/collections/:id       # Delete collection
POST   /api/v1/educator/collections/:id/quizzes/:quiz_id  # Add quiz to collection
DELETE /api/v1/educator/collections/:id/quizzes/:quiz_id  # Remove quiz

# Public collection access
GET    /api/v1/educators/:id/collections      # List educator's collections
GET    /api/v1/collections/:id                # Get collection with quizzes
```

**4.4 Enhanced Quiz Metadata**
- Learning objectives
- Prerequisites (reference to other quizzes)
- Estimated completion time
- Difficulty level refinement
- Resource attachments/links

**4.5 Collection Analytics**
- Track learner progression through collections
- Completion rate for entire collection
- Average time to complete collection

#### Deliverables
- ✅ Educators can organize quizzes into collections/courses
- ✅ Public collections for structured learning paths
- ✅ Enhanced quiz metadata
- ✅ Collection-level analytics

---

### Phase 5: Core Analytics - Aggregate Stats (Week 5-6)

#### Goals
- Implement aggregate analytics for educators
- Dashboard overview
- Quiz performance metrics

#### Tasks

**5.1 Analytics Service Layer**
- Create `services/analytics_service.go`
- Implement aggregate stat calculations
- Caching layer for expensive queries

**5.2 Aggregate Analytics**
For each quiz:
- Total attempts
- Unique learners
- Completion rate
- Average score
- Pass/fail ratio
- Average time spent
- Trend over time (daily/weekly/monthly)

**5.3 API Endpoints**
```
# Educator analytics
GET    /api/v1/educator/dashboard             # Overview stats for all quizzes
GET    /api/v1/educator/quizzes/:id/analytics # Detailed analytics for one quiz

# Response includes:
# {
#   "total_attempts": 156,
#   "unique_learners": 98,
#   "completion_rate": 87.5,
#   "average_score": 78.3,
#   "pass_rate": 82.1,
#   "avg_time_minutes": 23.5,
#   "trend": [...],
#   "top_performers": [...],
#   "needs_attention": [...]
# }
```

**5.4 Dashboard Components**
- Overview cards (total students, total attempts, avg performance)
- Recent activity timeline
- Quiz performance comparison
- Trend charts data endpoints

#### Deliverables
- ✅ Educator dashboard with aggregate analytics
- ✅ Per-quiz detailed statistics
- ✅ Performance trends over time
- ✅ Visual-ready data for frontend charts

---

### Phase 6: Advanced Analytics - Learner-Level Insights (Week 6-7)

#### Goals
- Per-learner performance tracking
- Detailed attempt analysis
- Question-level analytics

#### Tasks

**6.1 Per-Learner Analytics**
```
GET /api/v1/educator/quizzes/:id/learners
# Returns list of learners who attempted quiz with:
# - User info (name, email if permitted)
# - Number of attempts
# - Best score
# - Latest score
# - Completion status
# - Time spent
# - Last attempt date

GET /api/v1/educator/quizzes/:id/learners/:learner_id
# Detailed view of one learner's performance:
# - All attempt history
# - Score progression
# - Time per attempt
# - Answers given (per question breakdown)
# - Strengths and weaknesses
```

**6.2 Question-Level Analytics**
```
GET /api/v1/educator/quizzes/:id/questions/analytics
# Returns for each question:
# {
#   "question_id": "...",
#   "question_text": "...",
#   "total_attempts": 150,
#   "correct_rate": 67.3,
#   "avg_time_seconds": 45,
#   "common_wrong_answers": [
#     {"answer": "Option B", "selected_count": 35},
#     ...
#   ],
#   "difficulty_rating": "medium" // calculated
# }
```

**6.3 Database Optimizations**
- Add indexes for analytics queries
- Create materialized views for heavy aggregations (optional)
- Consider read replicas for analytics (production optimization)

**6.4 Enhanced User Answer Tracking**
- Migration: Add `time_spent_seconds` to user_answers table
- Track time per question (not just per quiz)
- Store attempt metadata (device, browser - optional)

**6.5 Privacy & Permissions**
- Educators can only see analytics for their own quizzes
- Learner identity obfuscation option (show performance, hide name)
- GDPR compliance: learners can opt-out of detailed tracking
- Admin override to view any analytics

#### Deliverables
- ✅ Per-learner performance dashboard
- ✅ Detailed attempt history and analysis
- ✅ Question difficulty and effectiveness metrics
- ✅ Privacy controls and compliance

---

### Phase 7: Unified Live Quiz Sessions (Week 7-8)

#### Goals
- Build unified live session system supporting BOTH use cases
- Educator public sessions with manual host control
- User private sessions with automatic progression
- Real-time competitive gameplay with live leaderboards
- Post-session detailed analysis for all participants

#### Tasks

**7.1 Database Schema**
- Migration: Create `live_sessions` table (unified schema)
- Migration: Create `live_session_participants` table
- Migration: Create `live_session_answers` table
- Migration: Create `live_session_invitations` table (for private sessions)

**7.2 Live Session Models**
```go
// File: backend/livesession/model/session.go

type LiveSession struct {
    ID              uuid.UUID
    QuizID          uuid.UUID
    CreatorID       uuid.UUID    // ANY user, not just educators

    // Session type determines behavior
    SessionType     string       // 'public' or 'private'

    // Basic info
    Title           string
    Description     string

    // Timing (flexible)
    ScheduledAt     *time.Time   // NULL = instant, NOT NULL = scheduled
    StartedAt       *time.Time
    EndedAt         *time.Time
    Status          string       // waiting, live, completed, cancelled

    // Access control
    AccessCode      string       // Generated for private sessions
    IsDiscoverable  bool         // Public listing?
    MaxParticipants *int

    // Progression mode
    ProgressionMode string       // 'manual' or 'automatic'
    QuestionTimerSeconds int     // For automatic mode
    CurrentQuestionIndex int

    // Config (flexible for future features)
    Config          JSONB

    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type LiveSessionParticipant struct {
    ID            uuid.UUID
    SessionID     uuid.UUID
    UserID        uuid.UUID
    JoinedAt      time.Time
    LeftAt        *time.Time
    CurrentScore  int
    Rank          int
    Status        string  // active, disconnected, completed
}

type LiveSessionInvitation struct {
    ID          uuid.UUID
    SessionID   uuid.UUID
    InviterID   uuid.UUID
    InviteeID   uuid.UUID
    Status      string    // pending, accepted, declined
    InvitedAt   time.Time
    RespondedAt *time.Time
}
```

**7.3 Live Session Service**
- Create/schedule live sessions (unified)
- Access control based on session_type
- Participant join/leave management
- Real-time score calculation and ranking
- Manual progression (for educators)
- Automatic progression (for user private sessions)
- Session invitation system
- Session end and final results

**7.4 WebSocket/Real-Time Architecture**
```
┌─────────────────┐         ┌──────────────────┐
│   Participants  │ ◄─────► │  WebSocket Hub   │
│   (All Users)   │         │  (Go Channels)   │
└─────────────────┘         └────────┬─────────┘
                                     │
┌─────────────────┐                  │
│  Session Host   │ ◄────────────────┤
│ (Educator/User) │                  │
└─────────────────┘                  │
                                     ▼
                            ┌─────────────────┐
                            │  Session Manager│
                            │  (Worker)       │
                            │  - Manual mode  │
                            │  - Auto mode    │
                            └────────┬────────┘
                                     │
                            ┌────────▼────────┐
                            │   PostgreSQL    │
                            │   (State Store) │
                            └─────────────────┘
```

**7.5 Real-Time Events (Unified)**
- `session_started`: Notify all participants session is live
- `question_revealed`: Show next question to all
- `timer_update`: Countdown updates (automatic mode only)
- `question_auto_advance`: Auto-progression event (automatic mode)
- `answer_submitted`: Update leaderboard
- `question_ended`: Show correct answer and stats
- `session_ended`: Final leaderboard and results
- `participant_joined`: Update participant count
- `participant_left`: Update participant count

**7.6 API Endpoints (Unified)**
```
# ========================================
# SESSION CREATION (UNIFIED)
# ========================================

POST   /api/v1/live-sessions
Body: {
  "quiz_id": "...",
  "title": "...",
  "description": "...",
  "session_type": "public" | "private",
  "scheduled_at": "2024-02-01T19:00:00Z" | null,  // null = instant
  "progression_mode": "manual" | "automatic",
  "question_timer_seconds": 30,  // For automatic mode
  "max_participants": 50,
  "is_discoverable": true | false
}
# Validation:
# - If session_type='public', creator must be educator
# - If session_type='public', scheduled_at required
# - If session_type='private', generates access_code
Response: {
  "session": {...},
  "access_code": "XYZ123"  // Only for private sessions
}

# ========================================
# SESSION INVITATION (PRIVATE ONLY)
# ========================================

POST   /api/v1/live-sessions/:id/invite
Body: { "user_ids": ["uuid1", "uuid2", ...] }
Response: { "message": "Invitations sent", "invited_count": 5 }
# Creates in-app notifications for invited users

GET    /api/v1/live-sessions/:id/invitations
Response: { "invitations": [...] }

# ========================================
# SESSION DISCOVERY
# ========================================

# Public discoverable sessions
GET    /api/v1/live-sessions/public
Query: ?status=upcoming,live&page=1
Response: { "sessions": [...] }

# My created sessions
GET    /api/v1/live-sessions/mine
Response: { "sessions": [...] }  // All my sessions (public + private)

# Sessions I'm invited to
GET    /api/v1/live-sessions/invitations
Response: { "sessions": [...] }  // Private sessions I'm invited to

# ========================================
# SESSION JOIN/LEAVE
# ========================================

GET    /api/v1/live-sessions/:id
Response: { "session": {...}, "can_join": true, "is_joined": false }

POST   /api/v1/live-sessions/:id/join
Body: { "access_code": "XYZ123" }  // Required for private, ignored for public
Response: { "message": "Joined session", "participant": {...} }

POST   /api/v1/live-sessions/:id/leave
Response: { "message": "Left session" }

# ========================================
# SESSION CONTROL (MANUAL MODE ONLY)
# ========================================

POST   /api/v1/live-sessions/:id/start
POST   /api/v1/live-sessions/:id/next-question
POST   /api/v1/live-sessions/:id/end
# Only works if progression_mode='manual' and user is creator

# ========================================
# REAL-TIME
# ========================================

WS     /api/v1/live-sessions/:id/ws               # WebSocket connection
GET    /api/v1/live-sessions/:id/leaderboard      # Current leaderboard

# ========================================
# POST-SESSION ANALYTICS
# ========================================

GET    /api/v1/live-sessions/:id/results
Response: {
  "session": {...},
  "final_leaderboard": [...],
  "my_performance": {...},

  // For public sessions (educator view)
  "detailed_analytics": {...},

  // For private sessions (all participants can see)
  "participant_details": [
    {
      "user": {...},
      "score": 850,
      "rank": 3,
      "answers": [
        {
          "question": {...},
          "answer": {...},
          "is_correct": true,
          "time_taken_ms": 4500
        }
      ]
    }
  ]
}
```

**7.7 Live Session Flows**

**Flow A: Educator Public Session (Manual Progression)**
1. Educator creates public session with scheduled_at, progression_mode='manual'
2. System notifies all followers
3. Session appears in public discovery
4. Participants join waiting room
5. **Educator manually starts** session
6. Question 1 revealed to all
7. Participants submit answers
8. **Educator manually advances** to Question 2
9. Repeat until all questions completed
10. **Educator manually ends** session or auto-ends after last question
11. Final leaderboard shown
12. Educator sees detailed analytics

**Flow B: User Private Session (Automatic Progression)**
1. User creates private session with scheduled_at (or null for instant), progression_mode='automatic'
2. System generates access_code
3. User invites friends via in-app notifications
4. Invited users join via access code
5. **Session auto-starts** at scheduled_at (or when host manually starts)
6. Question 1 revealed with timer (e.g., 30 seconds)
7. Timer counts down via WebSocket
8. Participants submit answers
9. **Timer auto-advances** to Question 2 when it reaches 0
10. Repeat until all questions completed
11. **Session auto-ends** after last question
12. Final leaderboard and ALL participant details visible to everyone

**7.8 Scoring System (Same for Both)**
- Base points for correct answer
- Speed bonus (faster = more points)
- Streak bonus (consecutive correct answers)
- Example: `points = base_points * (1 + speed_factor) * streak_multiplier`

**7.9 Automatic Progression Worker**
```go
// File: backend/livesession/worker/auto_progression.go

// Background worker that handles automatic progression
func AutoProgressionWorker(ctx context.Context) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Find all live sessions with automatic progression
            sessions := getLiveAutomaticSessions()

            for _, session := range sessions {
                // Check if current question timer expired
                if session.CurrentQuestionStartedAt.Add(
                    time.Duration(session.QuestionTimerSeconds) * time.Second,
                ).Before(time.Now()) {
                    // Time's up! Advance to next question
                    if session.CurrentQuestionIndex < session.TotalQuestions-1 {
                        advanceToNextQuestion(session.ID)
                        broadcastWebSocketEvent(session.ID, "question_auto_advance", ...)
                    } else {
                        // Last question completed, end session
                        endSession(session.ID)
                        broadcastWebSocketEvent(session.ID, "session_ended", ...)
                    }
                }

                // Send timer updates every second
                remainingSeconds := calculateRemainingSeconds(session)
                broadcastWebSocketEvent(session.ID, "timer_update", map[string]int{
                    "seconds_remaining": remainingSeconds,
                })
            }

        case <-ctx.Done():
            return
        }
    }
}
```

**7.10 WebSocket Message Format**
```json
// Timer update (automatic progression only)
{
  "type": "timer_update",
  "seconds_remaining": 15,
  "question_index": 2
}

// Auto-advance event
{
  "type": "question_auto_advance",
  "next_question": {...},
  "next_question_index": 3
}

// Common events (both modes)
{
  "type": "session_started",
  "session_id": "...",
  "total_questions": 20
}

{
  "type": "question_revealed",
  "question": {...},
  "index": 1,
  "timer_seconds": 30  // null for manual mode
}

{
  "type": "leaderboard_update",
  "leaderboard": [
    {"user_id": "...", "name": "John", "score": 450, "rank": 1},
    ...
  ]
}

{
  "type": "session_ended",
  "final_leaderboard": [...],
  "session_id": "..."
}
```

**7.11 Notification Integration**
- Public sessions: Notify all followers when scheduled
- Private sessions: Notify invited users
- 10-minute reminders for all scheduled sessions
- Session starting notifications

#### Deliverables
- ✅ Unified system supporting both educator public and user private sessions
- ✅ Educators can create public sessions with manual control
- ✅ ANY user can create private sessions with automatic progression
- ✅ Real-time synchronized gameplay via WebSocket
- ✅ Live leaderboard with speed-based scoring
- ✅ User invitation system with in-app notifications
- ✅ Post-session analytics (detailed for educators, transparent for private groups)
- ✅ Access code system for private sessions
- ✅ Automatic timer-based progression for private sessions
- ✅ No code duplication between public and private session logic

---

### Phase 8: Advanced Features & Enhancements (Week 9+)

#### 8.1 Cohort Analysis
```
GET /api/v1/educator/quizzes/:id/cohorts
# Compare performance across:
# - Time periods (this week vs last week)
# - Subscriber status (subscribers vs public)
# - Attempt number (1st attempt vs retakes)
# - Demographics (if collected)
```

#### 7.2 Learning Path Analytics
- Track learner progression through collections
- Completion rate of learning paths
- Drop-off points in multi-quiz sequences
- Prerequisite effectiveness

#### 7.3 Comparative Analytics
- Compare quiz performance across similar topics
- Benchmark against platform averages
- Identify overperforming/underperforming content

#### 8.2 Real-time Analytics
- Live attempt tracking (who's taking quiz right now)
- Real-time notifications of completions
- Live leaderboard updates

#### 8.3 Export & Reporting
```
GET /api/v1/educator/quizzes/:id/export?format=csv
GET /api/v1/educator/quizzes/:id/report?type=pdf
# Generate downloadable reports:
# - CSV exports of learner data
# - PDF performance reports
# - Excel spreadsheets with charts
```

#### 8.4 Engagement Metrics
- Quiz view count (how many opened but didn't start)
- Bounce rate (started but didn't finish first question)
- Return rate (users who take multiple quizzes from educator)
- Social shares and referrals

#### 8.5 Feedback System
- Learners can rate quizzes
- Learners can leave comments/feedback for educators
- Educators respond to feedback
- Report inappropriate content

#### 8.6 Gamification for Educators
- Badges for milestones (100 students, 1000 attempts, etc.)
- Leaderboard of top educators
- Impact score (reach × quality)

---

## Database Schema Changes

### Migration 1: Add Educator Role

```sql
-- File: backend/pkg/database/migrations/000020_add_educator_role.up.sql

BEGIN;

-- Update role constraint to include 'educator'
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_role;
ALTER TABLE users ADD CONSTRAINT chk_user_role
  CHECK (role IN ('user', 'admin', 'educator'));

COMMIT;
```

### Migration 2: Create Educator Profiles

```sql
-- File: backend/pkg/database/migrations/000021_create_educator_profiles.up.sql

BEGIN;

CREATE TABLE IF NOT EXISTS educator_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    tagline VARCHAR(255),
    expertise_areas TEXT[], -- Array of expertise tags
    credentials TEXT, -- JSON array of credentials/certifications
    social_links JSONB DEFAULT '{}'::jsonb, -- LinkedIn, Twitter, website, etc.
    profile_banner_url TEXT,
    is_featured BOOLEAN DEFAULT false,
    is_verified BOOLEAN DEFAULT false, -- Platform verified badge
    total_subscribers INT DEFAULT 0,
    total_quizzes INT DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.0, -- Future: rating system
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_educator_profiles_user_id ON educator_profiles(user_id);
CREATE INDEX idx_educator_profiles_featured ON educator_profiles(is_featured) WHERE is_featured = true;
CREATE INDEX idx_educator_profiles_expertise ON educator_profiles USING GIN (expertise_areas);

COMMIT;
```

### Migration 3: Create In-App Notifications

```sql
-- File: backend/pkg/database/migrations/000022_create_notifications.up.sql

BEGIN;

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- new_quiz, live_session, follower, etc.
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}'::jsonb, -- Additional context (quiz_id, educator_id, etc.)
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,

    CHECK (type IN ('new_quiz', 'live_session_scheduled', 'live_session_starting',
                    'live_session_started', 'follower', 'system'))
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX idx_notifications_created_at ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);

COMMENT ON TABLE notifications IS 'In-app notifications for users';
COMMENT ON COLUMN notifications.data IS 'Additional context like quiz_id, educator_id, session_id';

COMMIT;
```

### Migration 4: Create Subscriptions (Follows)

```sql
-- File: backend/pkg/database/migrations/000023_create_subscriptions.up.sql

BEGIN;

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    educator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    notification_enabled BOOLEAN DEFAULT true,

    UNIQUE (subscriber_id, educator_id),
    CHECK (subscriber_id != educator_id) -- Can't subscribe to yourself
);

CREATE INDEX idx_subscriptions_subscriber_id ON subscriptions(subscriber_id);
CREATE INDEX idx_subscriptions_educator_id ON subscriptions(educator_id);
CREATE INDEX idx_subscriptions_subscribed_at ON subscriptions(subscribed_at DESC);

-- Trigger to update educator total_subscribers count
CREATE OR REPLACE FUNCTION update_educator_subscriber_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE educator_profiles
        SET total_subscribers = total_subscribers + 1,
            updated_at = NOW()
        WHERE user_id = NEW.educator_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE educator_profiles
        SET total_subscribers = GREATEST(total_subscribers - 1, 0),
            updated_at = NOW()
        WHERE user_id = OLD.educator_id;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_educator_subscriber_count
AFTER INSERT OR DELETE ON subscriptions
FOR EACH ROW EXECUTE FUNCTION update_educator_subscriber_count();

COMMIT;
```

### Migration 5: Create Quiz Collections

```sql
-- File: backend/pkg/database/migrations/000025_create_quiz_collections.up.sql

BEGIN;

CREATE TABLE IF NOT EXISTS quiz_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    educator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    is_published BOOLEAN DEFAULT false,
    display_order INT DEFAULT 0,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS collection_quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    collection_id UUID NOT NULL REFERENCES quiz_collections(id) ON DELETE CASCADE,
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    display_order INT DEFAULT 0,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (collection_id, quiz_id)
);

CREATE INDEX idx_quiz_collections_educator_id ON quiz_collections(educator_id)
  WHERE deleted_at IS NULL;
CREATE INDEX idx_quiz_collections_published ON quiz_collections(is_published)
  WHERE is_published = true AND deleted_at IS NULL;
CREATE INDEX idx_collection_quizzes_collection_id ON collection_quizzes(collection_id);
CREATE INDEX idx_collection_quizzes_quiz_id ON collection_quizzes(quiz_id);

COMMIT;
```

### Migration 6: Enhanced User Answers for Analytics

```sql
-- File: backend/pkg/database/migrations/000026_enhance_user_answers_analytics.up.sql

BEGIN;

ALTER TABLE user_answers ADD COLUMN time_spent_seconds INT;
ALTER TABLE user_answers ADD COLUMN attempt_number INT DEFAULT 1;

CREATE INDEX idx_user_answers_time_spent ON user_answers(time_spent_seconds)
  WHERE time_spent_seconds IS NOT NULL;

COMMENT ON COLUMN user_answers.time_spent_seconds IS 'Time spent on this specific question';
COMMENT ON COLUMN user_answers.attempt_number IS 'Which attempt of the quiz this answer belongs to';

COMMIT;
```

### Migration 7: Create Live Sessions (Unified Schema)

```sql
-- File: backend/pkg/database/migrations/000027_create_live_sessions.up.sql

BEGIN;

-- ========================================
-- Live Sessions (Unified for Public & Private)
-- ========================================

CREATE TABLE IF NOT EXISTS live_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- ANY user, not just educators

    -- Session type determines behavior
    session_type VARCHAR(20) NOT NULL,  -- 'public' or 'private'

    -- Basic info
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Timing (flexible)
    scheduled_at TIMESTAMP WITH TIME ZONE,  -- NULL = instant, NOT NULL = scheduled
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'waiting',  -- waiting, live, completed, cancelled

    -- Access control
    access_code VARCHAR(10),  -- Generated for private sessions
    is_discoverable BOOLEAN DEFAULT false,  -- Public listing?
    max_participants INT CHECK (max_participants > 0),

    -- Progression mode
    progression_mode VARCHAR(20) NOT NULL,  -- 'manual' or 'automatic'
    question_timer_seconds INT,  -- For automatic mode (e.g., 30 seconds per question)
    current_question_index INT DEFAULT 0,
    current_question_started_at TIMESTAMP WITH TIME ZONE,  -- For automatic timer

    -- Config (flexible for future features)
    config JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CHECK (session_type IN ('public', 'private')),
    CHECK (progression_mode IN ('manual', 'automatic')),
    CHECK (status IN ('waiting', 'live', 'completed', 'cancelled')),

    -- Business rules
    CONSTRAINT valid_public_session CHECK (
        session_type = 'private' OR
        (session_type = 'public' AND scheduled_at IS NOT NULL)
    ),
    CONSTRAINT valid_private_session CHECK (
        session_type = 'public' OR
        (session_type = 'private' AND access_code IS NOT NULL)
    ),
    CONSTRAINT valid_automatic_mode CHECK (
        progression_mode = 'manual' OR
        (progression_mode = 'automatic' AND question_timer_seconds IS NOT NULL)
    )
);

-- ========================================
-- Live Session Participants
-- ========================================

CREATE TABLE IF NOT EXISTS live_session_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    current_score INT DEFAULT 0,
    rank INT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    config JSONB DEFAULT '{}'::jsonb,  -- For storing participant-specific config

    UNIQUE (session_id, user_id),
    CHECK (status IN ('active', 'disconnected', 'completed'))
);

-- ========================================
-- Live Session Answers
-- ========================================

CREATE TABLE IF NOT EXISTS live_session_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES live_session_participants(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,
    is_correct BOOLEAN,
    points_earned INT DEFAULT 0,
    time_taken_ms INT NOT NULL,  -- Milliseconds to answer
    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (session_id, participant_id, question_id)
);

-- ========================================
-- Live Session Invitations (for private sessions)
-- ========================================

CREATE TABLE IF NOT EXISTS live_session_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, accepted, declined
    invited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,

    UNIQUE (session_id, invitee_id),
    CHECK (status IN ('pending', 'accepted', 'declined'))
);

-- ========================================
-- Indexes for live_sessions
-- ========================================

CREATE INDEX idx_live_sessions_creator_id ON live_sessions(creator_id);
CREATE INDEX idx_live_sessions_quiz_id ON live_sessions(quiz_id);
CREATE INDEX idx_live_sessions_status ON live_sessions(status);
CREATE INDEX idx_live_sessions_session_type ON live_sessions(session_type);
CREATE INDEX idx_live_sessions_progression_mode ON live_sessions(progression_mode);
CREATE INDEX idx_live_sessions_access_code ON live_sessions(access_code) WHERE access_code IS NOT NULL;
CREATE INDEX idx_live_sessions_discoverable ON live_sessions(is_discoverable, status)
  WHERE is_discoverable = true AND status IN ('waiting', 'live');
CREATE INDEX idx_live_sessions_scheduled_at ON live_sessions(scheduled_at)
  WHERE scheduled_at IS NOT NULL AND status = 'waiting';
CREATE INDEX idx_live_sessions_started_at ON live_sessions(started_at DESC)
  WHERE status = 'live';

-- ========================================
-- Indexes for participants
-- ========================================

CREATE INDEX idx_live_session_participants_session_id ON live_session_participants(session_id);
CREATE INDEX idx_live_session_participants_user_id ON live_session_participants(user_id);
CREATE INDEX idx_live_session_participants_rank ON live_session_participants(session_id, rank)
  WHERE status = 'active';

-- ========================================
-- Indexes for answers
-- ========================================

CREATE INDEX idx_live_session_answers_session_id ON live_session_answers(session_id);
CREATE INDEX idx_live_session_answers_participant_id ON live_session_answers(participant_id);
CREATE INDEX idx_live_session_answers_question_id ON live_session_answers(question_id);

-- ========================================
-- Indexes for invitations
-- ========================================

CREATE INDEX idx_live_session_invitations_session_id ON live_session_invitations(session_id);
CREATE INDEX idx_live_session_invitations_invitee_id ON live_session_invitations(invitee_id, status)
  WHERE status = 'pending';
CREATE INDEX idx_live_session_invitations_inviter_id ON live_session_invitations(inviter_id);

-- ========================================
-- Comments
-- ========================================

COMMENT ON TABLE live_sessions IS 'Unified live quiz sessions - supports both educator public sessions and user private sessions';
COMMENT ON COLUMN live_sessions.session_type IS 'public (educator only) or private (any user)';
COMMENT ON COLUMN live_sessions.progression_mode IS 'manual (host controls) or automatic (timer-based)';
COMMENT ON COLUMN live_sessions.scheduled_at IS 'NULL for instant sessions, NOT NULL for scheduled sessions';
COMMENT ON COLUMN live_sessions.access_code IS 'Required for private sessions, used for joining';
COMMENT ON COLUMN live_sessions.is_discoverable IS 'If true, session appears in public listings';

COMMENT ON TABLE live_session_participants IS 'Participants/players in live sessions';
COMMENT ON TABLE live_session_answers IS 'Answers submitted during live sessions with timing';
COMMENT ON TABLE live_session_invitations IS 'Invitations for private sessions';

COMMIT;
```

### Migration 8: Analytics Materialized Views (Optional - for performance)

```sql
-- File: backend/pkg/database/migrations/000027_create_analytics_views.up.sql

BEGIN;

-- Materialized view for quiz analytics (refreshed periodically)
CREATE MATERIALIZED VIEW IF NOT EXISTS quiz_analytics_summary AS
SELECT
    q.id AS quiz_id,
    q.creator_id,
    COUNT(DISTINCT qa.id) AS total_attempts,
    COUNT(DISTINCT qa.user_id) AS unique_learners,
    COUNT(DISTINCT qa.id) FILTER (WHERE qa.status = 'submitted') AS completed_attempts,
    AVG(qa.percentage) FILTER (WHERE qa.status = 'submitted') AS avg_score,
    COUNT(DISTINCT qa.id) FILTER (WHERE qa.passed = true) AS passed_attempts,
    AVG(EXTRACT(EPOCH FROM (qa.submitted_at - qa.started_at))/60) FILTER (WHERE qa.submitted_at IS NOT NULL) AS avg_time_minutes,
    MAX(qa.submitted_at) AS last_attempt_at
FROM quizzes q
LEFT JOIN quiz_attempts qa ON q.id = qa.quiz_id
WHERE q.deleted_at IS NULL
GROUP BY q.id, q.creator_id;

CREATE UNIQUE INDEX idx_quiz_analytics_summary_quiz_id ON quiz_analytics_summary(quiz_id);
CREATE INDEX idx_quiz_analytics_summary_creator_id ON quiz_analytics_summary(creator_id);

-- Question difficulty analysis view
CREATE MATERIALIZED VIEW IF NOT EXISTS question_analytics_summary AS
SELECT
    qst.id AS question_id,
    qst.quiz_id,
    COUNT(ua.id) AS total_answers,
    COUNT(ua.id) FILTER (WHERE ua.is_correct = true) AS correct_answers,
    ROUND(100.0 * COUNT(ua.id) FILTER (WHERE ua.is_correct = true) / NULLIF(COUNT(ua.id), 0), 2) AS correct_percentage,
    AVG(ua.time_spent_seconds) AS avg_time_seconds,
    CASE
        WHEN ROUND(100.0 * COUNT(ua.id) FILTER (WHERE ua.is_correct = true) / NULLIF(COUNT(ua.id), 0), 2) >= 80 THEN 'easy'
        WHEN ROUND(100.0 * COUNT(ua.id) FILTER (WHERE ua.is_correct = true) / NULLIF(COUNT(ua.id), 0), 2) >= 50 THEN 'medium'
        ELSE 'hard'
    END AS calculated_difficulty
FROM questions qst
LEFT JOIN user_answers ua ON qst.id = ua.question_id
WHERE qst.deleted_at IS NULL
GROUP BY qst.id, qst.quiz_id;

CREATE UNIQUE INDEX idx_question_analytics_summary_question_id ON question_analytics_summary(question_id);
CREATE INDEX idx_question_analytics_summary_quiz_id ON question_analytics_summary(quiz_id);

-- Function to refresh analytics views
CREATE OR REPLACE FUNCTION refresh_analytics_views()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY quiz_analytics_summary;
    REFRESH MATERIALIZED VIEW CONCURRENTLY question_analytics_summary;
END;
$$ LANGUAGE plpgsql;

COMMIT;
```

---

## API Endpoints

### Complete API Reference

```
# ========================================
# EDUCATOR PROMOTION (SELF-SERVICE)
# ========================================

POST   /api/v1/educator/promote
Body: {} # No body required, validation is automatic
Response: {
  "message": "Successfully promoted to educator",
  "educator_profile": {...}
}
# Returns 400 if email not verified or phone not added
# Returns 409 if already an educator

GET    /api/v1/educator/can-promote
Response: {
  "can_promote": true,
  "requirements": {
    "email_verified": true,
    "phone_added": true
  }
}

# ========================================
# EDUCATOR PROFILE (PUBLIC & PRIVATE)
# ========================================

GET    /api/v1/educators
Query: ?page=1&limit=20&expertise=programming&featured=true
Response: { "educators": [...], "pagination": {...} }

GET    /api/v1/educators/:id
Response: {
  "id": "...",
  "name": "John Doe",
  "email": "john@example.com",
  "educator_profile": {
    "bio": "...",
    "tagline": "Teaching the world to code",
    "expertise_areas": ["Python", "Machine Learning"],
    "credentials": [...],
    "social_links": {...},
    "is_featured": true,
    "is_verified": true,
    "total_subscribers": 1250,
    "total_quizzes": 45,
    "average_rating": 4.8
  }
}

GET    /api/v1/educators/:id/quizzes
Query: ?page=1&limit=20&published=true
Response: { "quizzes": [...], "pagination": {...} }

PATCH  /api/v1/educator/profile
Body: {
  "bio": "Updated bio",
  "tagline": "New tagline",
  "expertise_areas": ["Python", "Django", "FastAPI"],
  "social_links": {
    "linkedin": "https://...",
    "twitter": "https://..."
  }
}
Response: { "educator_profile": {...} }

GET    /api/v1/educator/profile
Response: { "educator_profile": {...} }

# ========================================
# FOLLOW SYSTEM (SUBSCRIPTIONS)
# ========================================

POST   /api/v1/educators/:id/follow
Response: { "message": "Following educator", "educator": {...} }

DELETE /api/v1/educators/:id/unfollow
Response: { "message": "Unfollowed educator" }

GET    /api/v1/following
Response: {
  "following": [
    {
      "educator": {...},
      "followed_at": "...",
      "latest_quiz": {...}
    }
  ]
}

GET    /api/v1/following/feed
Query: ?page=1&limit=20&since=2024-01-01
Response: {
  "quizzes": [
    {
      "quiz": {...},
      "educator": {...},
      "published_at": "..."
    }
  ]
}

# Educator endpoints
GET    /api/v1/educator/followers
Query: ?page=1&limit=50&sort=recent
Response: {
  "followers": [
    {
      "user": {...},
      "followed_at": "...",
      "quizzes_taken": 5
    }
  ],
  "total": 1250
}

GET    /api/v1/educator/followers/stats
Response: {
  "total_followers": 1250,
  "growth_this_month": 45,
  "growth_percentage": 3.6,
  "timeline": [
    { "date": "2024-01-01", "count": 1205 },
    { "date": "2024-01-08", "count": 1220 },
    ...
  ]
}

# ========================================
# IN-APP NOTIFICATIONS
# ========================================

GET    /api/v1/notifications
Query: ?page=1&limit=20&unread=true
Response: {
  "notifications": [
    {
      "id": "...",
      "type": "new_quiz",
      "title": "New Quiz from John Doe",
      "message": "Python Fundamentals Quiz is now available",
      "data": {
        "quiz_id": "...",
        "educator_id": "...",
        "educator_name": "John Doe"
      },
      "is_read": false,
      "created_at": "..."
    }
  ],
  "unread_count": 5
}

PATCH  /api/v1/notifications/:id/read
Response: { "message": "Notification marked as read" }

PATCH  /api/v1/notifications/read-all
Response: { "message": "All notifications marked as read" }

DELETE /api/v1/notifications/:id
Response: { "message": "Notification deleted" }

GET    /api/v1/notifications/unread-count
Response: { "unread_count": 5 }

# ========================================
# QUIZ COLLECTIONS
# ========================================

POST   /api/v1/educator/collections
Body: {
  "title": "Python Fundamentals Course",
  "description": "Complete beginner to intermediate Python",
  "is_published": false
}
Response: { "collection": {...} }

GET    /api/v1/educator/collections
Response: { "collections": [...] }

GET    /api/v1/collections/:id
Response: {
  "id": "...",
  "title": "...",
  "description": "...",
  "quizzes": [
    { "quiz": {...}, "order": 1 },
    { "quiz": {...}, "order": 2 }
  ]
}

PATCH  /api/v1/educator/collections/:id
Body: { "title": "Updated title", "is_published": true }
Response: { "collection": {...} }

DELETE /api/v1/educator/collections/:id
Response: { "message": "Collection deleted" }

POST   /api/v1/educator/collections/:id/quizzes/:quiz_id
Body: { "display_order": 1 }
Response: { "message": "Quiz added to collection" }

DELETE /api/v1/educator/collections/:id/quizzes/:quiz_id
Response: { "message": "Quiz removed from collection" }

# ========================================
# EDUCATOR ANALYTICS - DASHBOARD
# ========================================

GET    /api/v1/educator/dashboard
Response: {
  "overview": {
    "total_quizzes": 45,
    "total_students": 1250,
    "total_attempts": 5678,
    "avg_completion_rate": 87.5,
    "avg_quiz_score": 78.3
  },
  "recent_activity": [
    {
      "type": "new_subscriber",
      "user": {...},
      "timestamp": "..."
    },
    {
      "type": "quiz_completed",
      "user": {...},
      "quiz": {...},
      "score": 85,
      "timestamp": "..."
    }
  ],
  "top_quizzes": [
    {
      "quiz": {...},
      "attempts": 234,
      "avg_score": 82.5,
      "completion_rate": 91.2
    }
  ],
  "needs_attention": [
    {
      "quiz": {...},
      "issue": "Low completion rate (45%)",
      "suggestion": "Consider reducing difficulty"
    }
  ]
}

# ========================================
# EDUCATOR ANALYTICS - QUIZ LEVEL
# ========================================

GET    /api/v1/educator/quizzes/:id/analytics
Response: {
  "quiz": {...},
  "aggregate_stats": {
    "total_attempts": 234,
    "unique_learners": 189,
    "completion_rate": 87.5,
    "average_score": 78.3,
    "pass_rate": 82.1,
    "avg_time_minutes": 23.5
  },
  "score_distribution": [
    { "range": "0-20", "count": 5 },
    { "range": "21-40", "count": 12 },
    { "range": "41-60", "count": 28 },
    { "range": "61-80", "count": 89 },
    { "range": "81-100", "count": 100 }
  ],
  "trend": [
    {
      "date": "2024-01-01",
      "attempts": 15,
      "avg_score": 76.5,
      "completion_rate": 85.0
    },
    ...
  ],
  "top_performers": [
    { "user": {...}, "score": 98, "time_minutes": 18 }
  ]
}

# ========================================
# EDUCATOR ANALYTICS - LEARNER LEVEL
# ========================================

GET    /api/v1/educator/quizzes/:id/learners
Query: ?page=1&limit=50&sort=score_desc&status=completed
Response: {
  "learners": [
    {
      "user_id": "...",
      "name": "Jane Smith",
      "email": "jane@example.com",
      "total_attempts": 3,
      "best_score": 92.5,
      "latest_score": 92.5,
      "avg_score": 85.0,
      "completion_status": "completed",
      "total_time_minutes": 68,
      "first_attempt_at": "...",
      "last_attempt_at": "...",
      "improvement": "+15.5" // Compared to first attempt
    }
  ],
  "pagination": {...}
}

GET    /api/v1/educator/quizzes/:id/learners/:learner_id
Response: {
  "learner": {
    "user": {...},
    "attempts": [
      {
        "attempt_id": "...",
        "started_at": "...",
        "submitted_at": "...",
        "score": 85,
        "percentage": 85.0,
        "passed": true,
        "time_minutes": 22,
        "correct_answers": 17,
        "total_questions": 20
      }
    ],
    "score_progression": [75, 80, 85],
    "strengths": ["Multiple Choice", "Short Answer"],
    "weaknesses": ["Complex Problem Solving"],
    "question_performance": [
      {
        "question": {...},
        "correct_count": 2,
        "total_attempts": 3,
        "latest_answer": {...}
      }
    ]
  }
}

# ========================================
# EDUCATOR ANALYTICS - QUESTION LEVEL
# ========================================

GET    /api/v1/educator/quizzes/:id/questions/analytics
Response: {
  "questions": [
    {
      "question_id": "...",
      "question_text": "What is polymorphism?",
      "question_type": "multiple_choice",
      "points": 5,
      "total_attempts": 189,
      "correct_attempts": 127,
      "correct_rate": 67.2,
      "avg_time_seconds": 45,
      "calculated_difficulty": "medium",
      "common_wrong_answers": [
        {
          "answer_text": "Inheritance only",
          "selected_count": 38,
          "percentage": 20.1
        }
      ],
      "avg_points_earned": 3.36
    }
  ],
  "summary": {
    "easiest_questions": [...],
    "hardest_questions": [...],
    "avg_question_difficulty": "medium"
  }
}

# ========================================
# EDUCATOR ANALYTICS - EXPORTS
# ========================================

GET    /api/v1/educator/quizzes/:id/export
Query: ?format=csv&include=learners,attempts,scores
Response: [CSV file download]

GET    /api/v1/educator/quizzes/:id/report
Query: ?type=pdf&date_range=last_30_days
Response: [PDF file download]

# ========================================
# LIVE QUIZ SESSIONS
# ========================================

# Educator endpoints
POST   /api/v1/educator/live-sessions
Body: {
  "quiz_id": "...",
  "title": "Friday Night Quiz Challenge",
  "description": "Join me for a live competitive quiz!",
  "scheduled_at": "2024-02-01T19:00:00Z",
  "max_participants": 100,
  "is_public": true,
  "entry_code": null
}
Response: { "session": {...} }

GET    /api/v1/educator/live-sessions
Query: ?status=scheduled&page=1&limit=20
Response: { "sessions": [...] }

GET    /api/v1/educator/live-sessions/:id
Response: {
  "session": {...},
  "participants": [...],
  "current_question": {...}
}

PATCH  /api/v1/educator/live-sessions/:id
Body: { "title": "Updated title", "scheduled_at": "..." }
Response: { "session": {...} }

DELETE /api/v1/educator/live-sessions/:id
Response: { "message": "Session cancelled" }

POST   /api/v1/educator/live-sessions/:id/start
Response: { "message": "Session started", "session": {...} }

POST   /api/v1/educator/live-sessions/:id/next-question
Response: { "question": {...}, "question_index": 2 }

POST   /api/v1/educator/live-sessions/:id/end
Response: {
  "message": "Session ended",
  "final_leaderboard": [...]
}

GET    /api/v1/educator/live-sessions/:id/analytics
Response: {
  "total_participants": 87,
  "completion_rate": 92.3,
  "avg_score": 750,
  "question_stats": [...],
  "leaderboard": [...]
}

# Participant endpoints
GET    /api/v1/live-sessions
Query: ?status=live,scheduled&page=1&limit=20
Response: {
  "sessions": [
    {
      "session": {...},
      "educator": {...},
      "participant_count": 45,
      "max_participants": 100
    }
  ]
}

GET    /api/v1/live-sessions/:id
Response: {
  "session": {...},
  "educator": {...},
  "can_join": true,
  "is_joined": false
}

POST   /api/v1/live-sessions/:id/join
Body: { "entry_code": "ABC123" } # Optional, only if required
Response: {
  "message": "Joined session",
  "participant": {...}
}

POST   /api/v1/live-sessions/:id/leave
Response: { "message": "Left session" }

GET    /api/v1/live-sessions/:id/leaderboard
Response: {
  "leaderboard": [
    {
      "rank": 1,
      "user": {...},
      "score": 950,
      "correct_answers": 18,
      "total_questions": 20
    },
    ...
  ],
  "my_rank": 15,
  "my_score": 720
}

# WebSocket endpoint for real-time updates
WS     /api/v1/live-sessions/:id/ws
# Events: session_started, question_revealed, leaderboard_update,
#         question_ended, session_ended, participant_joined, participant_left

# ========================================
# ADMIN - EDUCATOR MANAGEMENT
# ========================================

GET    /api/v1/admin/educators
Query: ?page=1&limit=50&sort=followers_desc
Response: {
  "educators": [
    {
      "user": {...},
      "educator_profile": {...},
      "total_followers": 1250,
      "total_quizzes": 45,
      "created_at": "..."
    }
  ]
}

POST   /api/v1/admin/users/:id/demote
Body: { "reason": "Policy violation" }
Response: {
  "message": "User demoted from educator to user",
  "user": {...}
}

POST   /api/v1/admin/educators/:id/feature
Response: { "message": "Educator featured" }

POST   /api/v1/admin/educators/:id/verify
Response: { "message": "Educator verified" }

GET    /api/v1/admin/educators/stats
Response: {
  "total_educators": 450,
  "new_this_month": 23,
  "featured_educators": 25,
  "verified_educators": 180,
  "total_followers": 125000,
  "top_educators": [...],
  "avg_followers_per_educator": 277.8
}
```

---

## Security & Access Control

### Access Control Matrix

| Role     | Educator Application | View Educator Profile | Subscribe | Manage Own Profile | View Analytics | Admin Functions |
|----------|---------------------|----------------------|-----------|-------------------|----------------|----------------|
| Guest    | ❌                  | ✅                   | ❌        | ❌                | ❌             | ❌             |
| User     | ✅                  | ✅                   | ✅        | ❌                | ❌             | ❌             |
| Educator | ✅ (if not already) | ✅                   | ✅        | ✅                | ✅ (own only)  | ❌             |
| Admin    | ✅                  | ✅                   | ✅        | ✅ (all)          | ✅ (all)       | ✅             |

### Middleware Implementation

```go
// File: backend/internal/middleware/educator.go

package middleware

import (
    "github.com/affanahmad024/quiz-craft/backend/commons"
    "github.com/affanahmad024/quiz-craft/backend/user/dao"
    "github.com/gin-gonic/gin"
)

// RequireEducator middleware ensures the user has educator role
// Must be used after Auth() middleware
func RequireEducator() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := commons.Ctx(c)

        userID, exists := commons.UsrID(ctx)
        if !exists {
            commons.Resp401(c, "authentication-required")
            c.Abort()
            return
        }

        user, err := userdao.FindByID(ctx, userID)
        if err != nil {
            commons.Resp404(c, "user-not-found")
            c.Abort()
            return
        }

        if user.Role != "educator" && user.Role != "admin" {
            commons.Resp403(c, "educator-access-required")
            c.Abort()
            return
        }

        c.Set("user_role", user.Role)
        c.Next()
    }
}

// IsEducator checks if user has educator role
func IsEducator(c *gin.Context, userID uuid.UUID) bool {
    ctx := commons.Ctx(c)
    user, err := userdao.FindByID(ctx, userID)
    if err != nil {
        return false
    }
    return user.Role == "educator" || user.Role == "admin"
}

// CanAccessEducatorAnalytics checks if user can access analytics for a quiz
func CanAccessEducatorAnalytics(c *gin.Context, quizID uuid.UUID) bool {
    ctx := commons.Ctx(c)

    userID, exists := commons.UsrID(ctx)
    if !exists {
        return false
    }

    // Check if admin
    if IsAdmin(c, userID) {
        return true
    }

    // Check if educator and owns the quiz
    if !IsEducator(c, userID) {
        return false
    }

    quiz, err := quizdao.FindByID(ctx, quizID.String())
    if err != nil {
        return false
    }

    return quiz.CreatorID == userID.String()
}
```

### Route Protection Examples

```go
// In main.go

// Educator-only endpoints
educator := v1.Group("/educator")
educator.Use(middleware.Auth(jwtManager))
educator.Use(middleware.RequireEducator())
{
    educator.GET("/profile", api.GetMyEducatorProfile)
    educator.PATCH("/profile", api.UpdateEducatorProfile)
    educator.GET("/dashboard", api.GetEducatorDashboard)
    educator.GET("/subscribers", api.GetMySubscribers)
    educator.GET("/quizzes/:id/analytics", api.GetQuizAnalytics)
    // ... more educator endpoints
}

// Public educator discovery
v1.GET("/educators", api.ListEducators)
v1.GET("/educators/:id", api.GetEducatorProfile)
v1.GET("/educators/:id/quizzes", api.GetEducatorQuizzes)

// User subscription endpoints (auth required)
educators := v1.Group("/educators")
educators.Use(middleware.Auth(jwtManager))
{
    educators.POST("/:id/subscribe", api.SubscribeToEducator)
    educators.DELETE("/:id/unsubscribe", api.UnsubscribeFromEducator)
}
```

### Privacy Considerations

1. **Learner Identity Protection**
   - Option to anonymize learner data in analytics
   - Show performance without revealing identity
   - Aggregate stats default, drill-down optional

2. **Data Access Restrictions**
   - Educators can only access analytics for their own quizzes
   - Admins have override capability
   - Audit log for sensitive data access

3. **GDPR Compliance**
   - Right to be forgotten: delete learner data
   - Data export for learners
   - Opt-out of detailed analytics tracking
   - Clear privacy policy

4. **Rate Limiting**
   - Analytics API rate limits to prevent scraping
   - Export limits per time period
   - Real-time analytics throttling

---

## Analytics & Reporting

### Analytics Architecture

```
┌─────────────────────────────────────────┐
│        Frontend Dashboard               │
│  (Charts, Tables, Insights)             │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│     Analytics API Endpoints             │
│  - Aggregate stats                      │
│  - Learner-level analytics              │
│  - Question analytics                   │
│  - Trend analysis                       │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│    Analytics Service Layer              │
│  - Query building                       │
│  - Data aggregation                     │
│  - Caching layer (Redis)                │
│  - Complex calculations                 │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│      Analytics DAO                      │
│  - Optimized SQL queries                │
│  - Materialized views                   │
│  - Index utilization                    │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        PostgreSQL Database              │
│  - quiz_attempts table                  │
│  - user_answers table                   │
│  - Materialized views                   │
│  - Specialized indexes                  │
└─────────────────────────────────────────┘
```

### Key Analytics Queries

#### 1. Quiz Aggregate Stats

```go
// File: backend/analytics/dao/quiz_analytics.go

func GetQuizAggregateStats(ctx context.Context, quizID uuid.UUID) (*model.QuizAggregateStats, error) {
    qry := `
        SELECT
            COUNT(DISTINCT qa.id) AS total_attempts,
            COUNT(DISTINCT qa.user_id) AS unique_learners,
            COUNT(DISTINCT qa.id) FILTER (WHERE qa.status = 'submitted') AS completed_attempts,
            ROUND(AVG(qa.percentage) FILTER (WHERE qa.status = 'submitted'), 2) AS avg_score,
            COUNT(DISTINCT qa.id) FILTER (WHERE qa.passed = true) AS passed_attempts,
            ROUND(AVG(EXTRACT(EPOCH FROM (qa.submitted_at - qa.started_at))/60)
                  FILTER (WHERE qa.submitted_at IS NOT NULL), 2) AS avg_time_minutes
        FROM quiz_attempts qa
        WHERE qa.quiz_id = $1
    `
    // ... execute and return
}
```

#### 2. Per-Learner Performance

```go
func GetLearnerPerformance(ctx context.Context, quizID uuid.UUID, learnerID uuid.UUID) (*model.LearnerPerformance, error) {
    qry := `
        SELECT
            qa.id,
            qa.started_at,
            qa.submitted_at,
            qa.score,
            qa.percentage,
            qa.passed,
            EXTRACT(EPOCH FROM (qa.submitted_at - qa.started_at))/60 AS time_minutes,
            (
                SELECT COUNT(*)
                FROM user_answers ua
                WHERE ua.attempt_id = qa.id AND ua.is_correct = true
            ) AS correct_answers,
            (
                SELECT COUNT(*)
                FROM user_answers ua
                WHERE ua.attempt_id = qa.id
            ) AS total_answers
        FROM quiz_attempts qa
        WHERE qa.quiz_id = $1 AND qa.user_id = $2
        ORDER BY qa.started_at ASC
    `
    // ... execute and return all attempts
}
```

#### 3. Question Difficulty Analysis

```go
func GetQuestionAnalytics(ctx context.Context, quizID uuid.UUID) ([]*model.QuestionAnalytics, error) {
    qry := `
        SELECT
            q.id,
            q.question_text,
            q.question_type,
            q.points,
            COUNT(ua.id) AS total_attempts,
            COUNT(ua.id) FILTER (WHERE ua.is_correct = true) AS correct_attempts,
            ROUND(100.0 * COUNT(ua.id) FILTER (WHERE ua.is_correct = true) /
                  NULLIF(COUNT(ua.id), 0), 2) AS correct_rate,
            ROUND(AVG(ua.time_spent_seconds), 2) AS avg_time_seconds,
            CASE
                WHEN ROUND(100.0 * COUNT(ua.id) FILTER (WHERE ua.is_correct = true) /
                     NULLIF(COUNT(ua.id), 0), 2) >= 80 THEN 'easy'
                WHEN ROUND(100.0 * COUNT(ua.id) FILTER (WHERE ua.is_correct = true) /
                     NULLIF(COUNT(ua.id), 0), 2) >= 50 THEN 'medium'
                ELSE 'hard'
            END AS calculated_difficulty
        FROM questions q
        LEFT JOIN user_answers ua ON q.id = ua.question_id
        WHERE q.quiz_id = $1 AND q.deleted_at IS NULL
        GROUP BY q.id
        ORDER BY correct_rate ASC
    `
    // ... execute and return
}
```

#### 4. Common Wrong Answers

```go
func GetCommonWrongAnswers(ctx context.Context, questionID uuid.UUID) ([]*model.WrongAnswerStats, error) {
    qry := `
        SELECT
            a.id,
            a.answer_text,
            COUNT(ua.id) AS selection_count,
            ROUND(100.0 * COUNT(ua.id) / (
                SELECT COUNT(*) FROM user_answers WHERE question_id = $1
            ), 2) AS selection_percentage
        FROM answers a
        INNER JOIN user_answers ua ON a.id = ua.answer_id
        WHERE ua.question_id = $1
          AND ua.is_correct = false
        GROUP BY a.id, a.answer_text
        ORDER BY selection_count DESC
        LIMIT 5
    `
    // ... execute and return
}
```

### Caching Strategy

```go
// File: backend/internal/services/analytics_service.go

type AnalyticsService struct {
    cache *redis.Client
    dao   *analyticsdao.AnalyticsDAO
}

func (s *AnalyticsService) GetQuizAggregateStats(ctx context.Context, quizID uuid.UUID) (*model.QuizAggregateStats, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("analytics:quiz:%s:aggregate", quizID)

    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var stats model.QuizAggregateStats
        json.Unmarshal([]byte(cached), &stats)
        return &stats, nil
    }

    // Cache miss, fetch from DB
    stats, err := s.dao.GetQuizAggregateStats(ctx, quizID)
    if err != nil {
        return nil, err
    }

    // Cache for 5 minutes
    data, _ := json.Marshal(stats)
    s.cache.Set(ctx, cacheKey, data, 5*time.Minute)

    return stats, nil
}
```

---

## Enhanced Features & Ideas

### 1. Adaptive Learning
- AI-powered difficulty adjustment based on performance
- Personalized quiz recommendations
- Remedial content suggestions
- Learning path optimization

### 2. Peer Learning
- Learners can create study groups
- Group challenges and competitions
- Peer-to-peer quiz recommendations
- Discussion forums per quiz

### 3. Monetization (Future)
- Premium educator subscriptions
- Pay-per-quiz model
- Course bundles
- Platform revenue sharing

### 4. Gamification
- Streak tracking for regular learners
- Badges and achievements
- Leaderboards (global, per educator, per quiz)
- Points and rewards system

### 5. Social Features
- Share quiz results on social media
- Challenge friends to beat your score
- Quiz review and rating system
- Comments and discussions

### 6. Advanced Content Types
- Video explanations for answers
- Interactive diagrams
- Code execution for programming quizzes
- Simulation-based questions

### 7. Mobile App Features
- Offline quiz taking
- Push notifications for new content
- Mobile-optimized analytics
- Quick quiz creation tools

### 8. Integration Features
- LMS integration (Canvas, Moodle, Blackboard)
- Google Classroom integration
- Microsoft Teams integration
- Zoom integration for live sessions
- Calendar integration for scheduled quizzes

---

## Testing Strategy

### Unit Tests

```go
// File: backend/educator/dao/educator_test.go

func TestCreateEducatorProfile(t *testing.T) {
    // Test creating educator profile
    // Test validation
    // Test constraints
}

func TestUpdateEducatorProfile(t *testing.T) {
    // Test profile updates
    // Test field validation
}

// File: backend/subscription/dao/subscription_test.go

func TestSubscribeToEducator(t *testing.T) {
    // Test subscription creation
    // Test duplicate prevention
    // Test self-subscription prevention
}

func TestUnsubscribeFromEducator(t *testing.T) {
    // Test unsubscription
    // Test subscriber count updates
}

// File: backend/analytics/dao/analytics_test.go

func TestGetQuizAggregateStats(t *testing.T) {
    // Test aggregate calculations
    // Test with various data scenarios
}

func TestGetQuestionAnalytics(t *testing.T) {
    // Test question-level analytics
    // Test difficulty calculations
}
```

### Integration Tests

```go
// File: backend/api/educator_integration_test.go

func TestEducatorApplicationFlow(t *testing.T) {
    // 1. User applies to become educator
    // 2. Admin approves application
    // 3. Educator profile created
    // 4. User role updated to educator
}

func TestSubscriptionFlow(t *testing.T) {
    // 1. User subscribes to educator
    // 2. Subscription count updated
    // 3. Educator publishes new quiz
    // 4. Subscriber receives notification
}

func TestAnalyticsPermissions(t *testing.T) {
    // 1. Educator can access own quiz analytics
    // 2. Educator cannot access other educator's analytics
    // 3. Admin can access all analytics
}
```

### Performance Tests

```go
// File: backend/analytics/performance_test.go

func BenchmarkQuizAggregateStats(b *testing.B) {
    // Benchmark analytics query with 10k attempts
    // Target: < 200ms response time
}

func BenchmarkLearnerAnalytics(b *testing.B) {
    // Benchmark learner-level queries
    // Target: < 500ms for detailed view
}
```

### E2E Tests

```
# Test educator onboarding
- Apply as educator
- Admin approves
- Complete profile
- Create first quiz
- Publish quiz

# Test subscription workflow
- Subscribe to educator
- Receive notification for new quiz
- Take quiz
- Educator sees analytics

# Test analytics access
- Educator views dashboard
- Educator views per-quiz analytics
- Educator exports data
- Educator views individual learner performance
```

---

## Implementation Checklist

### Phase 1: Foundation
- [ ] Database migration: Add educator role
- [ ] Database migration: Create educator_profiles table
- [ ] Database migration: Create notifications table
- [ ] Model: educator/model/profile.go
- [ ] Model: notification/model/notification.go
- [ ] DAO: educator/dao/profile.go
- [ ] DAO: notification/dao/notification.go
- [ ] Middleware: middleware/educator.go
- [ ] Service: educator_service.go (with promotion logic)
- [ ] API: Promote to educator (with validation)
- [ ] API: Check promotion eligibility
- [ ] API: Admin - Demote educator
- [ ] API: Admin - List educators
- [ ] Tests: Unit tests for educator DAO
- [ ] Tests: Unit tests for promotion service
- [ ] Tests: Integration tests for promotion flow

### Phase 2: Profiles
- [ ] API: Update educator profile
- [ ] API: Get my educator profile
- [ ] API: List all educators (public)
- [ ] API: Get educator profile (public)
- [ ] API: List educator's quizzes (public)
- [ ] API: Search educators
- [ ] Tests: Profile CRUD tests
- [ ] Tests: Profile validation tests

### Phase 3: Follow System & Notifications
- [ ] Database migration: Create subscriptions (follows) table
- [ ] Model: subscription/model/subscription.go
- [ ] DAO: subscription/dao/subscription.go
- [ ] Service: subscription_service.go
- [ ] Service: notification_service.go
- [ ] API: Follow educator
- [ ] API: Unfollow educator
- [ ] API: List my followed educators
- [ ] API: Following feed (recent quizzes)
- [ ] API: Educator - List followers
- [ ] API: Educator - Follower stats
- [ ] API: Get notifications
- [ ] API: Mark notification as read
- [ ] API: Mark all as read
- [ ] API: Delete notification
- [ ] Worker: Notification creation worker for new quizzes
- [ ] Tests: Follow/unfollow tests
- [ ] Tests: Notification tests
- [ ] Tests: Integration tests

### Phase 4: Quiz Collections
- [ ] Database migration: Create quiz_collections table
- [ ] Database migration: Create collection_quizzes table
- [ ] Model: collection models
- [ ] DAO: collection DAO
- [ ] API: Collection CRUD endpoints
- [ ] API: Add/remove quiz from collection
- [ ] API: Public collection access endpoints
- [ ] Collection analytics
- [ ] Tests: Collection tests
- [ ] Tests: Collection analytics tests

### Phase 5: Core Analytics
- [ ] Service: analytics_service.go
- [ ] DAO: analytics/dao/analytics.go
- [ ] Model: analytics models
- [ ] API: Educator dashboard
- [ ] API: Quiz aggregate analytics
- [ ] Cache layer for analytics
- [ ] Tests: Analytics calculation tests
- [ ] Tests: Performance tests

### Phase 6: Advanced Analytics
- [ ] Database migration: Enhance user_answers table
- [ ] Database migration: Create analytics views
- [ ] DAO: Per-learner analytics queries
- [ ] DAO: Question-level analytics queries
- [ ] API: Learner list endpoint
- [ ] API: Individual learner analytics
- [ ] API: Question analytics
- [ ] Privacy controls
- [ ] Tests: Detailed analytics tests
- [ ] Tests: Privacy tests

### Phase 7: Live Quiz Sessions
- [ ] Database migration: Create live_sessions table
- [ ] Database migration: Create live_session_participants table
- [ ] Database migration: Create live_session_answers table
- [ ] Model: livesession/model/session.go
- [ ] DAO: livesession/dao/session.go
- [ ] Service: live_session_service.go
- [ ] Worker: session_manager.go (real-time orchestration)
- [ ] WebSocket: Hub implementation
- [ ] API: Create/schedule live session
- [ ] API: Start/end session
- [ ] API: Join/leave session
- [ ] API: Next question control
- [ ] API: Get leaderboard
- [ ] API: Session analytics
- [ ] WebSocket: Real-time events
- [ ] Notification integration (session alerts)
- [ ] Scoring system implementation
- [ ] Tests: Session creation tests
- [ ] Tests: WebSocket tests
- [ ] Tests: Real-time gameplay tests
- [ ] Tests: Leaderboard tests

### Phase 8: Enhancements
- [ ] Export functionality (CSV/PDF)
- [ ] Real-time analytics (optional)
- [ ] Cohort analysis
- [ ] Comparative analytics
- [ ] Feedback system
- [ ] Rating system
- [ ] Tests: Export tests
- [ ] Tests: E2E tests

---

## Rollout Plan

### Development Environment (Week 1-2)
1. Set up feature branch: `feature/educator-role`
2. Implement Phase 1: Foundation
3. Internal testing with dev team
4. Code review and refactoring

### Staging Environment (Week 3-4)
1. Deploy Phase 1-3 to staging
2. QA testing
3. Performance testing with sample data
4. Security audit

### Production Rollout (Phased)

**Week 5: Soft Launch (Phase 1)**
- Enable self-service educator promotion
- Monitor metrics: promotion rate, profile completion
- Gather feedback from early educators

**Week 6: Phase 2 Public Launch**
- Educator profiles and discovery live
- Monitor: profile views, educator growth

**Week 7: Phase 3 Launch (Follow System)**
- Enable follow system and in-app notifications
- Announcement campaign to all users
- Monitor: follow rate, notification engagement

**Week 8: Phase 4-5 Launch**
- Launch collections and core analytics
- Educator onboarding guide and tutorials
- Monitor: collection usage, analytics access rate

**Week 9: Phase 6 Launch (Advanced Analytics)**
- Advanced analytics rollout
- Detailed educator training session
- Monitor: analytics query performance, feature usage

**Week 10: Phase 7 Launch (Live Sessions)**
- Live quiz sessions beta with select educators
- WebSocket infrastructure testing
- Monitor: session participation, real-time performance

**Week 11: Live Sessions Public**
- Open live sessions to all educators
- Marketing campaign for live events
- Monitor: session creation, participant engagement

**Week 12+: Continuous Improvement**
- Based on metrics and feedback
- Phase 8 enhancements
- Platform optimization
- New feature iterations

---

## Success Metrics

### Promotion & Onboarding
- Educator promotion rate (promotions / eligible users)
- Profile completion rate after promotion
- First quiz creation rate (within 7 days of promotion)
- Time to first quiz publish (avg days)
- Educator retention rate (active after 30 days)

### Engagement
- Active educators (created/published quiz in last 30 days)
- Quizzes per educator (avg)
- Follower growth rate
- Follow-back rate (educators following other educators)
- Quiz attempt rate (attempts / published quiz)

### Live Sessions
- Live sessions hosted per week
- Avg participants per session
- Session completion rate
- Repeat participation rate
- Live session satisfaction score

### Analytics Usage
- Educators viewing analytics (%)
- Avg sessions per educator per week
- Most used analytics features
- Export usage rate

### Quality
- Average quiz rating (from learners)
- Completion rate (educator quizzes vs regular)
- Learner satisfaction score
- Return rate (learners taking multiple quizzes from same educator)

### Business
- Total educators on platform
- Total follows across all educators
- Total quiz attempts on educator quizzes
- Live session participation rate
- Platform engagement increase (time spent)
- User retention improvement
- Notification engagement rate

---

## Conclusion

This comprehensive implementation plan follows the existing architecture patterns of the QuizCraft platform while introducing a powerful new **Educator Role** and **Unified Live Quiz Session** system. The phased approach ensures:

1. **Incremental Value**: Each phase delivers usable features
2. **Risk Management**: Early phases validate core assumptions
3. **Scalability**: Architecture supports growth
4. **Flexibility**: Easy to adjust based on feedback
5. **Quality**: Thorough testing at each phase

### Educator Role Features

The educator role transforms QuizCraft from a quiz-taking platform into a comprehensive learning ecosystem where content creators can:
- **Self-promote** with automated validation (email verified + phone number)
- **Build audiences** through follow/subscription system
- **Receive in-app notifications** for engagement tracking
- **Host public live sessions** with manual control
- **Track learner progress** with detailed per-learner analytics
- **Deliver impactful educational experiences** at scale

### Unified Live Session System

The unified architecture supports TWO distinct use cases with ONE codebase:

**Educator Public Sessions:**
- Scheduled, discoverable events
- Manual host control (educator advances questions)
- Notifications sent to all followers
- Post-session analytics for educator

**User Private Sessions:**
- Instant or scheduled
- Automatic timer-based progression
- Code-based access + user invitations
- Post-session detailed analysis visible to ALL participants
- Perfect for study groups and peer learning

### Key Technical Achievements

✅ **No Code Duplication**: Single unified system with `session_type` flag
✅ **Smart Access Control**: Middleware validates based on session configuration
✅ **Flexible Progression**: Manual (host) vs Automatic (timer) modes
✅ **Unified WebSocket**: Same real-time infrastructure for both use cases
✅ **Comprehensive Notifications**: In-app notification system with WebSocket delivery
✅ **Invitation System**: Direct user invitations for private sessions
✅ **Transparent Analytics**: Private session participants see everyone's detailed results

### Platform Differentiators

- ✅ **No approval bottleneck** - Instant educator access with basic verification
- ✅ **In-app notifications** - Native engagement without email dependency
- ✅ **Unified live sessions** - Both public events and private study groups
- ✅ **Live competitive gameplay** - Real-time sessions with leaderboards
- ✅ **Open access model** - All educator content remains public and accessible
- ✅ **User empowerment** - ANY user can create private live quiz sessions

---

**Document Version**: 3.0
**Last Updated**: 2026-01-27 (Added unified live session architecture, private user sessions, invitation system, comprehensive in-app notifications)
**Author**: AI Assistant
**Status**: Ready for Implementation
