# QuizCraft New Features - Frontend Integration Guide

## Complete API Reference for Notifications, Live Sessions & Educator Features

**Version**: 2.0
**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: Bearer Token (JWT)
**Last Updated**: January 28, 2026

---

## Table of Contents

1. [Overview](#overview)
2. [Notification System](#notification-system)
   - [Notification APIs](#notification-apis)
   - [Notification WebSocket](#notification-websocket)
3. [Live Quiz Sessions](#live-quiz-sessions)
   - [Session Management APIs](#session-management-apis)
   - [Session Participation APIs](#session-participation-apis)
   - [Live Session WebSocket](#live-session-websocket)
4. [Educator Role & Follow System](#educator-role--follow-system)
   - [Educator Management APIs](#educator-management-apis)
   - [Follow System APIs](#follow-system-apis)
   - [Educator Discovery APIs](#educator-discovery-apis)
5. [Complete Flow Examples](#complete-flow-examples)
6. [Error Codes Reference](#error-codes-reference)
7. [WebSocket Message Types](#websocket-message-types)

---

## Overview

This guide covers three major feature implementations:

### Phase 1: Notification System
- In-app notifications for users
- Real-time WebSocket delivery
- Notification types: new_quiz, live_session_scheduled, live_session_starting, live_session_invitation, new_follower, system
- Automatic cleanup of read notifications after 90 days

### Phase 2: Live Quiz Sessions
- Real-time competitive quiz sessions
- Support for both public (educator-hosted) and private (user-created) sessions
- WebSocket-based real-time updates
- Automatic session progression with timers
- Live leaderboards with speed-based scoring
- Post-session detailed analytics

### Phase 3: Educator Role & Follow System
- Self-service educator promotion
- Educator profile management
- Follow/unfollow educators
- Automatic notifications to followers when educator publishes quiz
- Educator discovery and search

---

## Notification System

### Notification APIs

#### 1. Get Notifications

**Endpoint**: `GET /notifications`
**Auth**: Required
**Description**: Get paginated list of user's notifications

**Query Parameters**:
```
limit: number (default: 20, max: 100)
offset: number (default: 0)
type: string (optional) - filter by notification type
unread_only: boolean (optional) - only unread notifications
```

**Request Example**:
```javascript
GET /api/v1/notifications?limit=20&offset=0&unread_only=true
Authorization: Bearer {access_token}
```

**Response (200 OK)**:
```json
{
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "type": "new_quiz",
        "title": "New Quiz Published",
        "message": "John Doe published a new quiz: JavaScript Fundamentals",
        "data": {
          "quiz_id": "uuid",
          "educator_id": "uuid"
        },
        "is_read": false,
        "read_at": null,
        "created_at": "2026-01-28T10:30:00Z"
      },
      {
        "id": "uuid",
        "type": "live_session_starting",
        "title": "Live Session Starting Soon",
        "message": "The live quiz session will start in 5 minutes",
        "data": {
          "session_id": "uuid",
          "quiz_id": "uuid"
        },
        "is_read": false,
        "created_at": "2026-01-28T09:55:00Z"
      }
    ],
    "total": 45
  }
}
```

#### 2. Get Unread Count

**Endpoint**: `GET /notifications/unread-count`
**Auth**: Required
**Description**: Get count of unread notifications

**Response (200 OK)**:
```json
{
  "data": {
    "count": 5
  }
}
```

#### 3. Mark Notification as Read

**Endpoint**: `PATCH /notifications/:id/read`
**Auth**: Required
**Description**: Mark a specific notification as read

**Response (200 OK)**:
```json
{
  "message": "Notification marked as read"
}
```

#### 4. Mark All Notifications as Read

**Endpoint**: `PATCH /notifications/read-all`
**Auth**: Required
**Description**: Mark all user's notifications as read

**Response (200 OK)**:
```json
{
  "message": "All notifications marked as read"
}
```

#### 5. Delete Notification

**Endpoint**: `DELETE /notifications/:id`
**Auth**: Required
**Description**: Delete a notification

**Response (200 OK)**:
```json
{
  "message": "Notification deleted successfully"
}
```

### Notification WebSocket

**Endpoint**: `WS /notifications/ws`
**Auth**: Required (pass token as query parameter)
**Description**: Real-time notification delivery

**Connection**:
```javascript
const token = localStorage.getItem('access_token');
const ws = new WebSocket(`ws://localhost:8080/api/v1/notifications/ws?token=${token}`);

ws.onopen = () => {
  console.log('Connected to notification stream');
};

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('New notification:', notification);

  // Update UI with new notification
  displayNotification(notification);
  updateNotificationBadge();
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from notification stream');
  // Implement reconnection logic
};
```

**Message Format**:
```json
{
  "type": "notification",
  "data": {
    "id": "uuid",
    "type": "new_quiz",
    "title": "New Quiz Published",
    "message": "John Doe published a new quiz: JavaScript Fundamentals",
    "data": {
      "quiz_id": "uuid",
      "educator_id": "uuid"
    },
    "created_at": "2026-01-28T10:30:00Z"
  }
}
```

### Notification Types

```typescript
type NotificationType =
  | 'new_quiz'                    // Educator published new quiz
  | 'live_session_scheduled'      // New live session scheduled
  | 'live_session_starting'       // Live session starting soon (5 min)
  | 'live_session_invitation'     // Invited to private session
  | 'new_follower'                // Someone followed you
  | 'system';                     // System announcement
```

---

## Live Quiz Sessions

### Session Management APIs

#### 1. Create Live Session

**Endpoint**: `POST /live-sessions`
**Auth**: Required
**Description**: Create a new live quiz session

**Request Body**:
```json
{
  "quiz_id": "uuid",
  "title": "JavaScript Fundamentals Live Challenge",
  "description": "Test your JavaScript knowledge in real-time!",
  "session_type": "public",
  "progression_mode": "automatic",
  "question_timer_seconds": 30,
  "scheduled_at": "2026-01-28T15:00:00Z",
  "max_participants": 100,
  "is_discoverable": true,
  "config": {
    "allow_late_join": false,
    "show_answers_after_question": true
  }
}
```

**Field Descriptions**:
- `session_type`: "public" (discoverable) or "private" (access code required)
- `progression_mode`: "manual" (host advances) or "automatic" (timer-based)
- `question_timer_seconds`: Time per question (5-300 seconds)
- `scheduled_at`: Optional - if provided, session auto-starts at this time
- `max_participants`: Optional - limit number of participants
- `is_discoverable`: For public sessions, show in discovery listings

**Response (200 OK)**:
```json
{
  "data": {
    "id": "uuid",
    "quiz_id": "uuid",
    "creator_id": "uuid",
    "session_type": "public",
    "title": "JavaScript Fundamentals Live Challenge",
    "description": "Test your JavaScript knowledge in real-time!",
    "scheduled_at": "2026-01-28T15:00:00Z",
    "status": "waiting",
    "access_code": "ABC123",
    "progression_mode": "automatic",
    "question_timer_seconds": 30,
    "current_question_index": 0,
    "is_discoverable": true,
    "max_participants": 100,
    "participant_count": 0,
    "created_at": "2026-01-28T10:00:00Z"
  }
}
```

#### 2. Get Live Session

**Endpoint**: `GET /live-sessions/:id`
**Auth**: Required
**Description**: Get session details

**Response (200 OK)**:
```json
{
  "data": {
    "id": "uuid",
    "quiz_id": "uuid",
    "creator_id": "uuid",
    "session_type": "public",
    "title": "JavaScript Fundamentals Live Challenge",
    "status": "live",
    "current_question_index": 3,
    "current_question_started_at": "2026-01-28T15:05:30Z",
    "participant_count": 42,
    "progression_mode": "automatic",
    "question_timer_seconds": 30
  }
}
```

#### 3. List My Sessions

**Endpoint**: `GET /live-sessions`
**Auth**: Required
**Description**: List sessions created by the authenticated user

**Query Parameters**:
```
limit: number (default: 20, max: 100)
offset: number (default: 0)
status: string (optional) - "waiting", "live", "completed"
```

**Response (200 OK)**:
```json
{
  "data": {
    "sessions": [
      {
        "id": "uuid",
        "title": "JavaScript Fundamentals Live Challenge",
        "status": "live",
        "participant_count": 42,
        "scheduled_at": "2026-01-28T15:00:00Z",
        "started_at": "2026-01-28T15:00:05Z",
        "created_at": "2026-01-28T10:00:00Z"
      }
    ],
    "total": 15
  }
}
```

#### 4. List Public Sessions

**Endpoint**: `GET /live-sessions/public`
**Auth**: Required
**Description**: List discoverable public sessions

**Query Parameters**:
```
limit: number (default: 20)
offset: number (default: 0)
status: string (optional) - "waiting", "live"
```

**Response (200 OK)**:
```json
{
  "data": {
    "sessions": [
      {
        "id": "uuid",
        "title": "JavaScript Fundamentals Live Challenge",
        "creator_name": "John Doe",
        "status": "waiting",
        "participant_count": 12,
        "max_participants": 100,
        "scheduled_at": "2026-01-28T15:00:00Z"
      }
    ],
    "total": 25
  }
}
```

#### 5. Start Live Session

**Endpoint**: `POST /live-sessions/:id/start`
**Auth**: Required (Creator only)
**Description**: Manually start a live session

**Response (200 OK)**:
```json
{
  "message": "Session started successfully"
}
```

#### 6. Advance to Next Question

**Endpoint**: `POST /live-sessions/:id/advance`
**Auth**: Required (Creator only)
**Description**: Advance to next question (manual progression mode only)

**Response (200 OK)**:
```json
{
  "message": "Advanced to next question",
  "current_question_index": 4
}
```

#### 7. End Live Session

**Endpoint**: `POST /live-sessions/:id/end`
**Auth**: Required (Creator only)
**Description**: End the live session

**Response (200 OK)**:
```json
{
  "message": "Session ended successfully"
}
```

### Session Participation APIs

#### 8. Join Live Session

**Endpoint**: `POST /live-sessions/:id/join`
**Auth**: Required
**Description**: Join a live session

**Request Body** (for private sessions):
```json
{
  "access_code": "ABC123"
}
```

**Response (200 OK)**:
```json
{
  "message": "Successfully joined session",
  "participant": {
    "id": "uuid",
    "session_id": "uuid",
    "user_id": "uuid",
    "score": 0,
    "rank": 1,
    "status": "active",
    "joined_at": "2026-01-28T15:00:10Z"
  }
}
```

#### 9. Submit Answer

**Endpoint**: `POST /live-sessions/:id/answer`
**Auth**: Required
**Description**: Submit answer for current question

**Request Body**:
```json
{
  "question_id": "uuid",
  "answer_id": "uuid",
  "time_taken_ms": 12500
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "is_correct": true,
    "points_earned": 125,
    "total_score": 425,
    "current_rank": 3,
    "correct_answer": "Option A"
  }
}
```

**Scoring Formula**:
```
Base Points: 100
Speed Bonus: Up to 50% based on answer speed

points = 100 * (1 + speedBonus)
speedBonus = (1 - timeTaken/maxTime) * 0.5

Example:
- Answer in 10s out of 30s: 100 * (1 + 0.67*0.5) = 133 points
- Answer in 30s out of 30s: 100 * (1 + 0*0.5) = 100 points
```

#### 10. Get Live Leaderboard

**Endpoint**: `GET /live-sessions/:id/leaderboard`
**Auth**: Required
**Description**: Get current session leaderboard

**Query Parameters**:
```
limit: number (default: 100)
```

**Response (200 OK)**:
```json
{
  "data": {
    "leaderboard": [
      {
        "rank": 1,
        "user_id": "uuid",
        "user_name": "Alice",
        "score": 950,
        "correct_answers": 10,
        "total_answers": 10
      },
      {
        "rank": 2,
        "user_id": "uuid",
        "user_name": "Bob",
        "score": 925,
        "correct_answers": 10,
        "total_answers": 10
      }
    ],
    "my_rank": 5,
    "my_score": 850
  }
}
```

#### 11. Get Session Results

**Endpoint**: `GET /live-sessions/:id/results`
**Auth**: Required
**Description**: Get detailed session results (after session ends)

**Response (200 OK)**:
```json
{
  "data": {
    "session": {
      "id": "uuid",
      "title": "JavaScript Fundamentals Live Challenge",
      "status": "completed",
      "total_participants": 42,
      "total_questions": 10,
      "started_at": "2026-01-28T15:00:00Z",
      "ended_at": "2026-01-28T15:15:00Z"
    },
    "my_performance": {
      "rank": 5,
      "score": 850,
      "correct_answers": 9,
      "total_answers": 10,
      "accuracy": 90.0,
      "average_time_ms": 15000
    },
    "answers": [
      {
        "question_id": "uuid",
        "question_text": "What is a closure in JavaScript?",
        "my_answer": "A function with access to outer scope",
        "correct_answer": "A function with access to outer scope",
        "is_correct": true,
        "points_earned": 125,
        "time_taken_ms": 12500
      }
    ],
    "leaderboard": [
      {
        "rank": 1,
        "user_name": "Alice",
        "score": 950,
        "correct_answers": 10
      }
    ],
    "question_stats": [
      {
        "question_id": "uuid",
        "question_text": "What is a closure?",
        "total_answers": 42,
        "correct_answers": 38,
        "correct_rate": 90.48,
        "average_time_ms": 14200
      }
    ]
  }
}
```

#### 12. Invite Users to Session

**Endpoint**: `POST /live-sessions/:id/invite`
**Auth**: Required (Creator only)
**Description**: Invite specific users to a private session

**Request Body**:
```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**Response (200 OK)**:
```json
{
  "message": "Invitations sent successfully",
  "invited_count": 3
}
```

#### 13. List My Session Invitations

**Endpoint**: `GET /live-sessions/invitations`
**Auth**: Required
**Description**: List sessions user has been invited to

**Response (200 OK)**:
```json
{
  "data": {
    "invitations": [
      {
        "id": "uuid",
        "session_id": "uuid",
        "session_title": "Team Quiz Challenge",
        "inviter_name": "John Doe",
        "status": "pending",
        "invited_at": "2026-01-28T10:00:00Z"
      }
    ],
    "total": 3
  }
}
```

### Live Session WebSocket

**Endpoint**: `WS /live-sessions/:id/ws`
**Auth**: Required (pass token as query parameter)
**Description**: Real-time session updates

**Connection**:
```javascript
const sessionId = 'uuid';
const token = localStorage.getItem('access_token');
const ws = new WebSocket(
  `ws://localhost:8080/api/v1/live-sessions/${sessionId}/ws?token=${token}`
);

ws.onopen = () => {
  console.log('Connected to live session');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  switch(message.type) {
    case 'session_started':
      handleSessionStarted(message.data);
      break;
    case 'question_revealed':
      handleQuestionRevealed(message.data);
      break;
    case 'timer_update':
      handleTimerUpdate(message.data);
      break;
    case 'question_auto_advance':
      handleQuestionAdvance(message.data);
      break;
    case 'leaderboard_update':
      handleLeaderboardUpdate(message.data);
      break;
    case 'session_ended':
      handleSessionEnded(message.data);
      break;
    case 'participant_joined':
      handleParticipantJoined(message.data);
      break;
  }
};
```

**WebSocket Message Types**:

1. **session_started**:
```json
{
  "type": "session_started",
  "data": {
    "session_id": "uuid",
    "total_questions": 10,
    "first_question": {
      "id": "uuid",
      "question_text": "What is a closure in JavaScript?",
      "question_type": "multiple_choice",
      "points": 100,
      "answers": [
        {"id": "uuid", "answer_text": "Option A"},
        {"id": "uuid", "answer_text": "Option B"}
      ]
    }
  }
}
```

2. **question_revealed**:
```json
{
  "type": "question_revealed",
  "data": {
    "question_index": 3,
    "question": {
      "id": "uuid",
      "question_text": "What is async/await?",
      "answers": [...]
    },
    "timer_seconds": 30
  }
}
```

3. **timer_update**:
```json
{
  "type": "timer_update",
  "data": {
    "seconds_remaining": 25,
    "current_question_index": 3
  }
}
```

4. **question_auto_advance**:
```json
{
  "type": "question_auto_advance",
  "data": {
    "next_question_index": 4,
    "next_question": {
      "id": "uuid",
      "question_text": "What is a Promise?",
      "answers": [...]
    }
  }
}
```

5. **leaderboard_update**:
```json
{
  "type": "leaderboard_update",
  "data": {
    "leaderboard": [
      {"rank": 1, "user_name": "Alice", "score": 950}
    ]
  }
}
```

6. **session_ended**:
```json
{
  "type": "session_ended",
  "data": {
    "session_id": "uuid",
    "final_leaderboard": [...]
  }
}
```

---

## Educator Role & Follow System

### Educator Management APIs

#### 1. Promote to Educator

**Endpoint**: `POST /educator/promote`
**Auth**: Required
**Description**: Self-promote to educator role (automated validation)

**Requirements**:
- Email must be verified
- Phone number must be added to account

**Request Body**:
```json
{
  "bio": "Experienced JavaScript developer with 10 years of teaching experience",
  "tagline": "Making coding fun and accessible",
  "expertise_areas": ["JavaScript", "React", "Node.js"]
}
```

**Response (200 OK)**:
```json
{
  "message": "Successfully promoted to educator",
  "profile": {
    "id": "uuid",
    "user_id": "uuid",
    "bio": "Experienced JavaScript developer...",
    "tagline": "Making coding fun and accessible",
    "expertise_areas": ["JavaScript", "React", "Node.js"],
    "total_subscribers": 0,
    "total_quizzes": 0,
    "is_featured": false,
    "is_verified": false,
    "created_at": "2026-01-28T10:00:00Z"
  }
}
```

**Error Responses**:
```json
// Email not verified
{
  "error": {
    "code": "email-not-verified",
    "message": "Email must be verified to become an educator"
  }
}

// Phone number required
{
  "error": {
    "code": "phone-number-required",
    "message": "Phone number is required to become an educator"
  }
}

// Already educator
{
  "error": {
    "code": "already-educator",
    "message": "User is already an educator"
  }
}
```

#### 2. Get My Educator Profile

**Endpoint**: `GET /educator/profile`
**Auth**: Required (Educator only)
**Description**: Get own educator profile

**Response (200 OK)**:
```json
{
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "bio": "Experienced JavaScript developer...",
    "tagline": "Making coding fun and accessible",
    "expertise_areas": ["JavaScript", "React", "Node.js"],
    "credentials": ["Certified JavaScript Developer", "10+ years experience"],
    "social_links": {
      "twitter": "https://twitter.com/johndoe",
      "linkedin": "https://linkedin.com/in/johndoe",
      "website": "https://johndoe.dev"
    },
    "profile_banner_url": "https://example.com/banner.jpg",
    "total_subscribers": 1250,
    "total_quizzes": 45,
    "average_rating": 4.8,
    "is_featured": false,
    "is_verified": true,
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-01-28T10:00:00Z"
  }
}
```

#### 3. Update My Educator Profile

**Endpoint**: `PATCH /educator/profile`
**Auth**: Required (Educator only)
**Description**: Update educator profile

**Request Body** (all fields optional):
```json
{
  "bio": "Updated bio with more details about teaching philosophy",
  "tagline": "New tagline",
  "expertise_areas": ["JavaScript", "React", "Node.js", "TypeScript"],
  "credentials": ["Certified JavaScript Developer", "Google Developer Expert"],
  "social_links": {
    "twitter": "https://twitter.com/johndoe",
    "linkedin": "https://linkedin.com/in/johndoe",
    "github": "https://github.com/johndoe"
  },
  "profile_banner_url": "https://example.com/new-banner.jpg"
}
```

**Response (200 OK)**:
```json
{
  "message": "Profile updated successfully"
}
```

### Follow System APIs

#### 4. Follow Educator

**Endpoint**: `POST /educators/:id/follow`
**Auth**: Required
**Description**: Follow an educator

**Response (200 OK)**:
```json
{
  "message": "Successfully followed educator",
  "followed": true
}
```

**Error Responses**:
```json
// Cannot follow yourself
{
  "error": {
    "code": "cannot-follow-yourself",
    "message": "You cannot follow yourself"
  }
}

// User not an educator
{
  "error": {
    "code": "user-not-educator",
    "message": "This user is not an educator"
  }
}

// Already following
{
  "error": {
    "code": "already-following",
    "message": "You are already following this educator"
  }
}
```

#### 5. Unfollow Educator

**Endpoint**: `DELETE /educators/:id/unfollow`
**Auth**: Required
**Description**: Unfollow an educator

**Response (200 OK)**:
```json
{
  "message": "Successfully unfollowed educator",
  "followed": false
}
```

#### 6. Get My Following

**Endpoint**: `GET /following`
**Auth**: Required
**Description**: List educators I'm following

**Query Parameters**:
```
limit: number (default: 20, max: 100)
offset: number (default: 0)
```

**Response (200 OK)**:
```json
{
  "data": {
    "following": [
      {
        "user_id": "uuid",
        "username": "John Doe",
        "email": "john@example.com",
        "profile": {
          "bio": "Experienced JavaScript developer...",
          "tagline": "Making coding fun",
          "expertise_areas": ["JavaScript", "React"],
          "total_subscribers": 1250,
          "total_quizzes": 45,
          "is_featured": false,
          "is_verified": true
        }
      }
    ],
    "total": 15
  }
}
```

#### 7. Get My Followers

**Endpoint**: `GET /educator/followers`
**Auth**: Required (Educator only)
**Description**: List my followers

**Query Parameters**:
```
limit: number (default: 20, max: 100)
offset: number (default: 0)
```

**Response (200 OK)**:
```json
{
  "data": {
    "followers": [
      {
        "user_id": "uuid",
        "username": "Alice Smith",
        "full_name": "Alice Smith",
        "avatar_url": "https://example.com/avatar.jpg",
        "subscribed_at": "2026-01-20T10:00:00Z"
      }
    ],
    "total": 1250
  }
}
```

#### 8. Get Follower Statistics

**Endpoint**: `GET /educator/followers/stats`
**Auth**: Required (Educator only)
**Description**: Get detailed follower statistics

**Response (200 OK)**:
```json
{
  "data": {
    "total_followers": 1250,
    "new_followers_today": 15,
    "new_followers_week": 87,
    "new_followers_month": 320,
    "growth_trend": [
      {
        "date": "2026-01-28",
        "count": 15
      },
      {
        "date": "2026-01-27",
        "count": 23
      },
      {
        "date": "2026-01-26",
        "count": 18
      }
    ]
  }
}
```

### Educator Discovery APIs

#### 9. List All Educators

**Endpoint**: `GET /educators`
**Auth**: Not Required (Public)
**Description**: List all educators (sorted by followers)

**Query Parameters**:
```
limit: number (default: 20, max: 100)
offset: number (default: 0)
```

**Response (200 OK)**:
```json
{
  "data": {
    "educators": [
      {
        "user_id": "uuid",
        "username": "John Doe",
        "email": "john@example.com",
        "profile": {
          "bio": "Experienced JavaScript developer...",
          "tagline": "Making coding fun",
          "expertise_areas": ["JavaScript", "React", "Node.js"],
          "total_subscribers": 1250,
          "total_quizzes": 45,
          "average_rating": 4.8,
          "is_featured": false,
          "is_verified": true
        }
      }
    ],
    "total": 152
  }
}
```

#### 10. Search Educators

**Endpoint**: `GET /educators/search`
**Auth**: Not Required (Public)
**Description**: Search educators by name or expertise

**Query Parameters**:
```
q: string (required) - search query
limit: number (default: 20, max: 100)
offset: number (default: 0)
```

**Request Example**:
```
GET /api/v1/educators/search?q=JavaScript&limit=20
```

**Response (200 OK)**:
```json
{
  "data": {
    "educators": [
      {
        "user_id": "uuid",
        "username": "John Doe",
        "profile": {
          "tagline": "JavaScript expert",
          "expertise_areas": ["JavaScript", "React"],
          "total_subscribers": 1250
        }
      }
    ],
    "total": 25
  }
}
```

#### 11. Get Featured Educators

**Endpoint**: `GET /educators/featured`
**Auth**: Not Required (Public)
**Description**: Get platform-featured educators

**Query Parameters**:
```
limit: number (default: 10, max: 50)
```

**Response (200 OK)**:
```json
{
  "data": {
    "educators": [
      {
        "user_id": "uuid",
        "username": "Jane Doe",
        "profile": {
          "tagline": "Master educator",
          "expertise_areas": ["React", "TypeScript"],
          "total_subscribers": 5000,
          "is_featured": true,
          "is_verified": true
        }
      }
    ],
    "total": 10
  }
}
```

#### 12. Get Educator Profile

**Endpoint**: `GET /educators/:id`
**Auth**: Not Required (Public)
**Description**: Get specific educator's profile

**Response (200 OK)**:
```json
{
  "data": {
    "user_id": "uuid",
    "username": "John Doe",
    "email": "john@example.com",
    "profile": {
      "bio": "Experienced JavaScript developer...",
      "tagline": "Making coding fun",
      "expertise_areas": ["JavaScript", "React", "Node.js"],
      "credentials": ["Certified JavaScript Developer"],
      "social_links": {
        "twitter": "https://twitter.com/johndoe",
        "linkedin": "https://linkedin.com/in/johndoe"
      },
      "profile_banner_url": "https://example.com/banner.jpg",
      "total_subscribers": 1250,
      "total_quizzes": 45,
      "average_rating": 4.8,
      "is_featured": false,
      "is_verified": true
    }
  }
}
```

#### 13. Get Educator's Quizzes

**Endpoint**: `GET /educators/:id/quizzes`
**Auth**: Not Required (Public)
**Description**: List quizzes created by an educator

**Query Parameters**:
```
limit: number (default: 20, max: 100)
offset: number (default: 0)
```

**Response (200 OK)**:
```json
{
  "data": {
    "quizzes": [],
    "total": 0,
    "message": "Quiz listing integration pending"
  }
}
```
*Note: This endpoint will be fully integrated with the existing quiz listing system in a future update.*

### Admin APIs

#### 14. Demote Educator

**Endpoint**: `POST /admin/users/:id/demote`
**Auth**: Required (Admin only)
**Description**: Demote educator to regular user

**Response (200 OK)**:
```json
{
  "message": "Educator demoted to user successfully"
}
```

---

## Complete Flow Examples

### Example 1: Creating and Hosting a Public Live Session

```javascript
// 1. Create a public live session
const createSession = async () => {
  const response = await fetch('http://localhost:8080/api/v1/live-sessions', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      quiz_id: 'quiz-uuid',
      title: 'JavaScript Live Challenge',
      description: 'Test your skills!',
      session_type: 'public',
      progression_mode: 'automatic',
      question_timer_seconds: 30,
      scheduled_at: '2026-01-28T15:00:00Z',
      is_discoverable: true
    })
  });

  const data = await response.json();
  return data.data.id; // session ID
};

// 2. Connect to WebSocket for real-time updates
const connectToSession = (sessionId) => {
  const ws = new WebSocket(
    `ws://localhost:8080/api/v1/live-sessions/${sessionId}/ws?token=${accessToken}`
  );

  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Session update:', message);
  };

  return ws;
};

// 3. Start the session (if not auto-starting)
const startSession = async (sessionId) => {
  await fetch(`http://localhost:8080/api/v1/live-sessions/${sessionId}/start`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });
};

// Complete flow
const hostLiveSession = async () => {
  const sessionId = await createSession();
  const ws = connectToSession(sessionId);

  // Wait for scheduled time or manually start
  await startSession(sessionId);
};
```

### Example 2: Joining and Participating in Live Session

```javascript
// 1. Get list of public sessions
const getPublicSessions = async () => {
  const response = await fetch(
    'http://localhost:8080/api/v1/live-sessions/public?status=waiting',
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );

  const data = await response.json();
  return data.data.sessions;
};

// 2. Join a session
const joinSession = async (sessionId) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/live-sessions/${sessionId}/join`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      }
    }
  );

  return await response.json();
};

// 3. Connect to session WebSocket
const connectAndPlay = (sessionId) => {
  const ws = new WebSocket(
    `ws://localhost:8080/api/v1/live-sessions/${sessionId}/ws?token=${accessToken}`
  );

  ws.onmessage = (event) => {
    const message = JSON.parse(event.data);

    if (message.type === 'question_revealed') {
      displayQuestion(message.data.question);
      startTimer(message.data.timer_seconds);
    }

    if (message.type === 'leaderboard_update') {
      updateLeaderboard(message.data.leaderboard);
    }
  };

  return ws;
};

// 4. Submit answer
const submitAnswer = async (sessionId, questionId, answerId, timeTakenMs) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/live-sessions/${sessionId}/answer`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        question_id: questionId,
        answer_id: answerId,
        time_taken_ms: timeTakenMs
      })
    }
  );

  const result = await response.json();
  console.log('Points earned:', result.data.points_earned);
  console.log('Current rank:', result.data.current_rank);
};

// Complete flow
const participateInSession = async (sessionId) => {
  await joinSession(sessionId);
  const ws = connectAndPlay(sessionId);

  // When question is revealed
  ws.onmessage = async (event) => {
    const message = JSON.parse(event.data);

    if (message.type === 'question_revealed') {
      const { question } = message.data;

      // User selects answer
      const selectedAnswerId = await showQuestionAndGetAnswer(question);
      const timeTaken = getTimeTaken(); // in milliseconds

      // Submit answer
      await submitAnswer(sessionId, question.id, selectedAnswerId, timeTaken);
    }
  };
};
```

### Example 3: Becoming an Educator and Managing Followers

```javascript
// 1. Check requirements and promote to educator
const becomeEducator = async () => {
  try {
    const response = await fetch(
      'http://localhost:8080/api/v1/educator/promote',
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          bio: 'Passionate about teaching web development',
          tagline: 'Making coding accessible to everyone',
          expertise_areas: ['JavaScript', 'React', 'Node.js']
        })
      }
    );

    if (!response.ok) {
      const error = await response.json();
      if (error.error.code === 'email-not-verified') {
        console.log('Please verify your email first');
      }
      if (error.error.code === 'phone-number-required') {
        console.log('Please add a phone number to your profile');
      }
      return;
    }

    const data = await response.json();
    console.log('Successfully promoted to educator!', data.profile);
  } catch (error) {
    console.error('Failed to promote:', error);
  }
};

// 2. Update educator profile
const updateProfile = async () => {
  await fetch('http://localhost:8080/api/v1/educator/profile', {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      bio: 'Updated bio with more information',
      credentials: ['10+ years experience', 'Google Developer Expert'],
      social_links: {
        twitter: 'https://twitter.com/johndoe',
        linkedin: 'https://linkedin.com/in/johndoe'
      }
    })
  });
};

// 3. Get follower statistics
const getFollowerStats = async () => {
  const response = await fetch(
    'http://localhost:8080/api/v1/educator/followers/stats',
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );

  const data = await response.json();
  console.log('Total followers:', data.data.total_followers);
  console.log('New today:', data.data.new_followers_today);
  console.log('Growth trend:', data.data.growth_trend);
};

// 4. List followers
const listFollowers = async (limit = 20, offset = 0) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/educator/followers?limit=${limit}&offset=${offset}`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );

  const data = await response.json();
  return data.data.followers;
};
```

### Example 4: Following Educators and Receiving Notifications

```javascript
// 1. Search for educators
const searchEducators = async (query) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/educators/search?q=${encodeURIComponent(query)}`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );

  const data = await response.json();
  return data.data.educators;
};

// 2. Follow an educator
const followEducator = async (educatorId) => {
  const response = await fetch(
    `http://localhost:8080/api/v1/educators/${educatorId}/follow`,
    {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );

  const data = await response.json();
  console.log(data.message); // "Successfully followed educator"
};

// 3. Connect to notification WebSocket
const connectToNotifications = () => {
  const ws = new WebSocket(
    `ws://localhost:8080/api/v1/notifications/ws?token=${accessToken}`
  );

  ws.onmessage = (event) => {
    const notification = JSON.parse(event.data);

    if (notification.data.type === 'new_quiz') {
      showNotification({
        title: notification.data.title,
        message: notification.data.message,
        data: notification.data.data // contains quiz_id, educator_id
      });

      // Update notification badge
      updateNotificationBadge();
    }
  };

  return ws;
};

// 4. Get notifications
const getNotifications = async (unreadOnly = false) => {
  const url = unreadOnly
    ? 'http://localhost:8080/api/v1/notifications?unread_only=true'
    : 'http://localhost:8080/api/v1/notifications';

  const response = await fetch(url, {
    headers: {
      'Authorization': `Bearer ${accessToken}`
    }
  });

  const data = await response.json();
  return data.data.notifications;
};

// 5. Mark notification as read
const markAsRead = async (notificationId) => {
  await fetch(
    `http://localhost:8080/api/v1/notifications/${notificationId}/read`,
    {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${accessToken}`
      }
    }
  );
};

// Complete flow
const setupEducatorFollowing = async () => {
  // Connect to real-time notifications
  const notificationWs = connectToNotifications();

  // Search and follow educators
  const educators = await searchEducators('JavaScript');
  for (const educator of educators) {
    await followEducator(educator.user_id);
  }

  // When notification arrives about new quiz
  notificationWs.onmessage = async (event) => {
    const notification = JSON.parse(event.data);

    if (notification.data.type === 'new_quiz') {
      // Show notification to user
      showToast({
        title: notification.data.title,
        message: notification.data.message
      });

      // Mark as read
      await markAsRead(notification.data.id);
    }
  };
};
```

---

## Error Codes Reference

### Notification Errors
- `notification-not-found` - Notification doesn't exist or doesn't belong to user

### Live Session Errors
- `session-not-found` - Session doesn't exist
- `session-not-waiting` - Session already started or completed
- `session-full` - Session reached max participants
- `invalid-access-code` - Wrong access code for private session
- `already-participant` - User already joined session
- `session-not-live` - Session not in live state
- `not-session-creator` - Only creator can perform this action
- `question-not-found` - Question doesn't exist
- `already-answered` - Already answered this question
- `session-completed` - Session has ended

### Educator Errors
- `email-not-verified` - Email must be verified to become educator
- `phone-number-required` - Phone number required to become educator
- `already-educator` - User is already an educator
- `educator-profile-not-found` - User is not an educator
- `cannot-follow-yourself` - Cannot follow your own account
- `user-not-educator` - Target user is not an educator
- `already-following` - Already following this educator
- `subscription-not-found` - Not following this educator

---

## WebSocket Message Types

### Notification WebSocket Messages

```typescript
interface NotificationMessage {
  type: 'notification';
  data: {
    id: string;
    type: 'new_quiz' | 'live_session_scheduled' | 'live_session_starting' |
          'live_session_invitation' | 'new_follower' | 'system';
    title: string;
    message: string;
    data: Record<string, any>;
    created_at: string;
  };
}
```

### Live Session WebSocket Messages

```typescript
interface SessionStartedMessage {
  type: 'session_started';
  data: {
    session_id: string;
    total_questions: number;
    first_question: Question;
  };
}

interface QuestionRevealedMessage {
  type: 'question_revealed';
  data: {
    question_index: number;
    question: Question;
    timer_seconds: number;
  };
}

interface TimerUpdateMessage {
  type: 'timer_update';
  data: {
    seconds_remaining: number;
    current_question_index: number;
  };
}

interface QuestionAutoAdvanceMessage {
  type: 'question_auto_advance';
  data: {
    next_question_index: number;
    next_question: Question;
  };
}

interface LeaderboardUpdateMessage {
  type: 'leaderboard_update';
  data: {
    leaderboard: LeaderboardEntry[];
  };
}

interface SessionEndedMessage {
  type: 'session_ended';
  data: {
    session_id: string;
    final_leaderboard: LeaderboardEntry[];
  };
}

interface ParticipantJoinedMessage {
  type: 'participant_joined';
  data: {
    participant_count: number;
    user_name: string;
  };
}
```

---

## Best Practices

### WebSocket Connection Management

```javascript
class WebSocketManager {
  constructor(url, token) {
    this.url = url;
    this.token = token;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
  }

  connect() {
    this.ws = new WebSocket(`${this.url}?token=${this.token}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
    };

    this.ws.onclose = () => {
      console.log('WebSocket closed');
      this.reconnect();
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return this.ws;
  }

  reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);

    setTimeout(() => {
      console.log(`Reconnecting... (attempt ${this.reconnectAttempts})`);
      this.connect();
    }, delay);
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// Usage
const notificationWs = new WebSocketManager(
  'ws://localhost:8080/api/v1/notifications/ws',
  accessToken
);
notificationWs.connect();
```

### Error Handling

```javascript
const handleApiError = (response, defaultMessage = 'An error occurred') => {
  if (!response.ok) {
    return response.json().then(error => {
      throw new Error(error.error?.message || defaultMessage);
    });
  }
  return response.json();
};

// Usage
const followEducator = async (educatorId) => {
  try {
    const response = await fetch(
      `http://localhost:8080/api/v1/educators/${educatorId}/follow`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${accessToken}`
        }
      }
    );

    const data = await handleApiError(response, 'Failed to follow educator');
    return data;
  } catch (error) {
    console.error('Error following educator:', error.message);
    showErrorToast(error.message);
  }
};
```

### Pagination Helper

```javascript
class PaginatedFetcher {
  constructor(baseUrl, token) {
    this.baseUrl = baseUrl;
    this.token = token;
  }

  async fetchAll(limit = 20) {
    let allItems = [];
    let offset = 0;
    let hasMore = true;

    while (hasMore) {
      const response = await fetch(
        `${this.baseUrl}?limit=${limit}&offset=${offset}`,
        {
          headers: {
            'Authorization': `Bearer ${this.token}`
          }
        }
      );

      const data = await response.json();
      const items = data.data.notifications || data.data.sessions || data.data.educators || [];

      allItems = [...allItems, ...items];
      offset += limit;
      hasMore = items.length === limit;
    }

    return allItems;
  }
}

// Usage
const fetcher = new PaginatedFetcher(
  'http://localhost:8080/api/v1/notifications',
  accessToken
);
const allNotifications = await fetcher.fetchAll();
```

---

## TypeScript Type Definitions

```typescript
// Notification Types
interface Notification {
  id: string;
  user_id: string;
  type: NotificationType;
  title: string;
  message: string;
  data: Record<string, any>;
  is_read: boolean;
  read_at: string | null;
  created_at: string;
}

type NotificationType =
  | 'new_quiz'
  | 'live_session_scheduled'
  | 'live_session_starting'
  | 'live_session_invitation'
  | 'new_follower'
  | 'system';

// Live Session Types
interface LiveSession {
  id: string;
  quiz_id: string;
  creator_id: string;
  session_type: 'public' | 'private';
  title: string;
  description?: string;
  scheduled_at?: string;
  started_at?: string;
  ended_at?: string;
  status: 'waiting' | 'live' | 'completed' | 'cancelled';
  access_code?: string;
  is_discoverable: boolean;
  max_participants?: number;
  progression_mode: 'manual' | 'automatic';
  question_timer_seconds: number;
  current_question_index: number;
  current_question_started_at?: string;
  config: Record<string, any>;
  created_at: string;
  updated_at: string;
  participant_count?: number;
}

interface Participant {
  id: string;
  session_id: string;
  user_id: string;
  user_name: string;
  score: number;
  rank: number;
  correct_answers: number;
  total_answers: number;
  status: 'active' | 'completed' | 'left';
  joined_at: string;
}

interface LeaderboardEntry {
  rank: number;
  user_id: string;
  user_name: string;
  score: number;
  correct_answers: number;
  total_answers: number;
}

interface AnswerSubmissionResponse {
  is_correct: boolean;
  points_earned: number;
  total_score: number;
  current_rank: number;
  correct_answer: string;
}

// Educator Types
interface EducatorProfile {
  id: string;
  user_id: string;
  bio?: string;
  tagline?: string;
  expertise_areas: string[];
  credentials: string[];
  social_links: Record<string, string>;
  profile_banner_url?: string;
  is_featured: boolean;
  is_verified: boolean;
  total_subscribers: number;
  total_quizzes: number;
  average_rating: number;
  created_at: string;
  updated_at: string;
}

interface EducatorWithUserInfo {
  user_id: string;
  username: string;
  email: string;
  full_name?: string;
  avatar_url?: string;
  profile: EducatorProfile;
}

interface FollowerInfo {
  user_id: string;
  username: string;
  full_name?: string;
  avatar_url?: string;
  subscribed_at: string;
}

interface FollowerStats {
  total_followers: number;
  new_followers_today: number;
  new_followers_week: number;
  new_followers_month: number;
  growth_trend: Array<{
    date: string;
    count: number;
  }>;
}
```

---

## Testing Checklist

### Notification System
- [ ] Connect to notification WebSocket
- [ ] Receive real-time notifications
- [ ] Get notification list with pagination
- [ ] Mark individual notification as read
- [ ] Mark all notifications as read
- [ ] Delete notifications
- [ ] Get unread count
- [ ] Handle WebSocket reconnection

### Live Quiz Sessions
- [ ] Create public live session
- [ ] Create private live session
- [ ] Join live session
- [ ] Connect to session WebSocket
- [ ] Receive real-time session updates
- [ ] Submit answers during session
- [ ] View live leaderboard
- [ ] Get session results after completion
- [ ] Host manual progression session
- [ ] Host automatic progression session
- [ ] Invite users to private session
- [ ] Handle session timer updates

### Educator System
- [ ] Promote to educator
- [ ] Handle promotion validation errors
- [ ] Update educator profile
- [ ] Follow educators
- [ ] Unfollow educators
- [ ] View follower list
- [ ] View following list
- [ ] Get follower statistics
- [ ] Search educators
- [ ] View educator profile
- [ ] Receive notifications when followed educator publishes quiz
- [ ] List educator's quizzes

---

## Support & Resources

### Documentation
- Main API Guide: `/docs/FRONTEND_INTEGRATION_GUIDE.md`
- Implementation Plan: `/docs/EDUCATOR_ROLE_IMPLEMENTATION.md`

### Swagger/OpenAPI
- Available at: `http://localhost:8080/swagger/index.html`

### WebSocket Testing
- Use tools like Postman, wscat, or browser extensions for WebSocket testing

### Contact
For questions or issues:
- Create an issue in the repository
- Contact the backend team

---

**Last Updated**: January 28, 2026
**Document Version**: 2.0
**API Version**: 1.0
