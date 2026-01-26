# Educator Role Implementation - Architecture & Approach

## Executive Summary

This document outlines the phased implementation plan for adding an **Educator Role** to the QuizCraft platform. Users can self-promote to educators (with basic verification checks), create content, gain followers through a subscription system, and receive detailed analytics about learner performance. The system also includes live quiz sessions for real-time competitive learning experiences.

---

## 📋 Table of Contents

1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Feature Requirements](#feature-requirements)
3. [Architecture Design](#architecture-design)
4. [Phased Implementation Plan](#phased-implementation-plan)
5. [Database Schema Changes](#database-schema-changes)
6. [API Endpoints](#api-endpoints)
7. [Security & Access Control](#security--access-control)
8. [Analytics & Reporting](#analytics--reporting)
9. [Enhanced Features & Ideas](#enhanced-features--ideas)
10. [Testing Strategy](#testing-strategy)

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

#### 3. Live Quiz Sessions
- Educators can schedule and host live quiz sessions
- Real-time competitive gameplay with live leaderboards
- Time-bound challenges with synchronized start times
- Post-session analytics and replay functionality

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

### Phase 7: Live Quiz Sessions (Week 7-8)

#### Goals
- Real-time competitive quiz gameplay
- Synchronized quiz sessions with live leaderboards
- Enhanced engagement through live events

#### Tasks

**7.1 Database Schema**
- Migration: Create `live_sessions` table
- Migration: Create `live_session_participants` table
- Migration: Create `live_session_answers` table (or extend user_answers)

**7.2 Live Session Models**
```go
// File: backend/livesession/model/session.go

type LiveSession struct {
    ID              uuid.UUID
    QuizID          uuid.UUID
    EducatorID      uuid.UUID
    Title           string
    Description     string
    ScheduledAt     time.Time    // When session is scheduled
    StartedAt       *time.Time   // Actual start time
    EndedAt         *time.Time   // Actual end time
    Status          string       // scheduled, live, completed, cancelled
    MaxParticipants *int         // Optional capacity limit
    EntryCode       string       // Optional entry code for private sessions
    IsPublic        bool         // Public vs private session
    Meta            JSONB
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type LiveSessionParticipant struct {
    ID            uuid.UUID
    SessionID     uuid.UUID
    UserID        uuid.UUID
    JoinedAt      time.Time
    CurrentScore  int
    Rank          int
    Status        string  // active, disconnected, completed
}
```

**7.3 Live Session Service**
- Create/schedule live sessions
- Start live session (educator control)
- Participant join/leave management
- Real-time score calculation and ranking
- Question progression control (educator advances questions)
- Session end and final results

**7.4 WebSocket/Real-Time Architecture**
```
┌─────────────────┐         ┌──────────────────┐
│   Participants  │ ◄─────► │  WebSocket Hub   │
│   (Learners)    │         │  (Go Channels)   │
└─────────────────┘         └────────┬─────────┘
                                     │
┌─────────────────┐                  │
│    Educator     │ ◄────────────────┤
│   (Host)        │                  │
└─────────────────┘                  │
                                     ▼
                            ┌─────────────────┐
                            │  Session Manager│
                            │  (Worker)       │
                            └────────┬────────┘
                                     │
                            ┌────────▼────────┐
                            │   PostgreSQL    │
                            │   (State Store) │
                            └─────────────────┘
```

**7.5 Real-Time Events**
- `session_started`: Notify all participants session is live
- `question_revealed`: Show next question to all
- `answer_submitted`: Update leaderboard
- `question_ended`: Show correct answer and stats
- `session_ended`: Final leaderboard and results
- `participant_joined`: Update participant count

**7.6 API Endpoints**
```
# Educator endpoints
POST   /api/v1/educator/live-sessions             # Create/schedule live session
GET    /api/v1/educator/live-sessions             # List my live sessions
GET    /api/v1/educator/live-sessions/:id         # Get session details
PATCH  /api/v1/educator/live-sessions/:id         # Update session
DELETE /api/v1/educator/live-sessions/:id         # Cancel session
POST   /api/v1/educator/live-sessions/:id/start   # Start the session
POST   /api/v1/educator/live-sessions/:id/next    # Move to next question
POST   /api/v1/educator/live-sessions/:id/end     # End the session

# Participant endpoints
GET    /api/v1/live-sessions                      # List upcoming/live sessions
GET    /api/v1/live-sessions/:id                  # Get session info
POST   /api/v1/live-sessions/:id/join             # Join session
POST   /api/v1/live-sessions/:id/leave            # Leave session
GET    /api/v1/live-sessions/:id/leaderboard      # Current leaderboard

# WebSocket endpoint
WS     /api/v1/live-sessions/:id/ws               # WebSocket connection
```

**7.7 Live Session Flow**

**Setup Phase (Educator):**
1. Educator creates live session linked to a quiz
2. Sets scheduled time, optional entry code
3. System sends notifications to followers
4. Session shows in "Upcoming" list

**Pre-Start Phase:**
5. Participants can join waiting room
6. Educator sees participant count in real-time
7. Educator starts session when ready

**Live Phase:**
8. Question displayed to all participants simultaneously
9. Timer starts (configurable per question)
10. Participants submit answers in real-time
11. Points awarded based on correctness + speed
12. Leaderboard updates live
13. Educator advances to next question
14. Repeat until all questions completed

**Post-Session:**
15. Final leaderboard displayed
16. Individual results saved
17. Replay/review available
18. Analytics generated for educator

**7.8 Scoring System**
- Base points for correct answer
- Speed bonus (faster = more points)
- Streak bonus (consecutive correct answers)
- Example: `points = base_points * (1 + speed_factor) * streak_multiplier`

**7.9 WebSocket Message Format**
```json
// Client → Server
{
  "type": "submit_answer",
  "session_id": "uuid",
  "question_id": "uuid",
  "answer_id": "uuid",
  "timestamp": "ISO8601"
}

// Server → Clients
{
  "type": "leaderboard_update",
  "session_id": "uuid",
  "leaderboard": [
    {"user_id": "uuid", "name": "John", "score": 450, "rank": 1},
    ...
  ]
}
```

**7.10 Notification Integration**
- Notify followers when educator schedules live session
- Reminder notification 10 minutes before session starts
- Push notification when session goes live

#### Deliverables
- ✅ Educators can create and host live quiz sessions
- ✅ Real-time synchronized gameplay
- ✅ Live leaderboard with speed-based scoring
- ✅ WebSocket-based real-time communication
- ✅ Post-session analytics and replay
- ✅ Notification system for session events

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

### Migration 7: Create Live Sessions

```sql
-- File: backend/pkg/database/migrations/000027_create_live_sessions.up.sql

BEGIN;

CREATE TABLE IF NOT EXISTS live_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    educator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled',
    max_participants INT CHECK (max_participants > 0),
    entry_code VARCHAR(20), -- Optional code for private sessions
    is_public BOOLEAN DEFAULT true,
    current_question_index INT DEFAULT 0,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CHECK (status IN ('scheduled', 'live', 'completed', 'cancelled')),
    CHECK (scheduled_at > NOW() OR status != 'scheduled')
);

CREATE TABLE IF NOT EXISTS live_session_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    current_score INT DEFAULT 0,
    rank INT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    meta JSONB DEFAULT '{}'::jsonb,

    UNIQUE (session_id, user_id),
    CHECK (status IN ('active', 'disconnected', 'completed'))
);

CREATE TABLE IF NOT EXISTS live_session_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES live_session_participants(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,
    is_correct BOOLEAN,
    points_earned INT DEFAULT 0,
    time_taken_ms INT NOT NULL, -- Milliseconds to answer
    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (session_id, participant_id, question_id)
);

-- Indexes for live sessions
CREATE INDEX idx_live_sessions_educator_id ON live_sessions(educator_id);
CREATE INDEX idx_live_sessions_quiz_id ON live_sessions(quiz_id);
CREATE INDEX idx_live_sessions_status ON live_sessions(status);
CREATE INDEX idx_live_sessions_scheduled_at ON live_sessions(scheduled_at)
  WHERE status = 'scheduled';
CREATE INDEX idx_live_sessions_started_at ON live_sessions(started_at DESC)
  WHERE status = 'live';

-- Indexes for participants
CREATE INDEX idx_live_session_participants_session_id ON live_session_participants(session_id);
CREATE INDEX idx_live_session_participants_user_id ON live_session_participants(user_id);
CREATE INDEX idx_live_session_participants_rank ON live_session_participants(session_id, rank)
  WHERE status = 'active';

-- Indexes for answers
CREATE INDEX idx_live_session_answers_session_id ON live_session_answers(session_id);
CREATE INDEX idx_live_session_answers_participant_id ON live_session_answers(participant_id);
CREATE INDEX idx_live_session_answers_question_id ON live_session_answers(question_id);

COMMENT ON TABLE live_sessions IS 'Real-time competitive quiz sessions hosted by educators';
COMMENT ON TABLE live_session_participants IS 'Participants/players in live sessions';
COMMENT ON TABLE live_session_answers IS 'Answers submitted during live sessions with timing';

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

This comprehensive implementation plan follows the existing architecture patterns of the QuizCraft platform while introducing a powerful new educator role system. The phased approach ensures:

1. **Incremental Value**: Each phase delivers usable features
2. **Risk Management**: Early phases validate core assumptions
3. **Scalability**: Architecture supports growth
4. **Flexibility**: Easy to adjust based on feedback
5. **Quality**: Thorough testing at each phase

The educator role transforms QuizCraft from a quiz-taking platform into a comprehensive learning ecosystem where content creators can:
- **Self-promote** with automated validation
- **Build audiences** through follow/notification system
- **Host live sessions** for real-time competitive learning
- **Track learner progress** with detailed analytics
- **Deliver impactful educational experiences** at scale

Key differentiators:
- ✅ **No approval bottleneck** - Instant educator access with basic verification
- ✅ **In-app notifications** - Native engagement without email dependency
- ✅ **Live competitive gameplay** - Real-time sessions with leaderboards
- ✅ **Open access model** - All content remains public and accessible

---

**Document Version**: 2.0
**Last Updated**: 2026-01-27 (Updated with self-service promotion, live sessions, in-app notifications)
**Author**: AI Assistant
**Status**: Updated Draft for Review
