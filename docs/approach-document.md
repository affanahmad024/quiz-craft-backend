# Quiz-Craft Backend - Production-Ready Approach Document

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Technology Stack](#technology-stack)
3. [Database Design](#database-design)
4. [API Design](#api-design)
5. [Authentication & Authorization](#authentication--authorization)
6. [AI Integration](#ai-integration)
7. [Security Implementation](#security-implementation)
8. [File Upload & Processing](#file-upload--processing)
9. [Error Handling](#error-handling)
10. [Logging & Monitoring](#logging--monitoring)
11. [Testing Strategy](#testing-strategy)
12. [Deployment & DevOps](#deployment--devops)
13. [Performance Optimization](#performance-optimization)
14. [Project Structure](#project-structure)

---

## Architecture Overview

### High-Level Architecture
```
┌─────────────┐      ┌─────────────────┐      ┌──────────────┐
│   Client    │────▶│   API Gateway   │────▶│   Services   │
│  (Frontend) │      │  (Gin Router)   │      │    Layer     │
└─────────────┘      └─────────────────┘      └──────────────┘
                              │                       │
                              ▼                       ▼
                     ┌─────────────────┐    ┌──────────────┐
                     │  Middleware     │    │  PostgreSQL  │
                     │  - Auth         │    │   Database   │
                     │  - CORS         │    └──────────────┘
                     │  - Rate Limit   │            │
                     │  - Logging      │            ▼
                     └─────────────────┘    ┌──────────────┐
                              │              │   Redis      │
                              ▼              │   Cache      │
                     ┌─────────────────┐    └──────────────┘
                     │  External APIs  │
                     │  - Google AI    │
                     │  - Email (SMTP) │
                     │  - File Storage │
                     └─────────────────┘
```

### Design Patterns
- **Clean Architecture**: Separation of concerns (handlers → services → repositories)
- **Repository Pattern**: Database abstraction layer
- **Dependency Injection**: For testability and modularity
- **Middleware Chain**: For cross-cutting concerns
- **Factory Pattern**: For creating complex objects (AI clients, email senders)

### Key Principles
1. **Stateless API**: All session data in JWT tokens or Redis
2. **Idempotent Operations**: Safe retry mechanisms
3. **Fail-Fast**: Early validation and error detection
4. **Graceful Degradation**: System remains functional if non-critical services fail
5. **12-Factor App**: Environment-based configuration

---

## Technology Stack

### Core Framework
- **Language**: Go 1.21+
- **Web Framework**: Gin v1.10.0
- **Reason**: High performance, built-in middleware, excellent routing

### Database & ORM
- **Primary Database**: PostgreSQL 15+
- **ORM**: GORM v2 (gorm.io/gorm)
- **Migrations**: golang-migrate/migrate v4
- **Connection Pooling**: pgxpool
- **Why PostgreSQL**:
  - JSONB support for flexible meta fields
  - ACID compliance for data integrity
  - Advanced indexing (GIN, GiST for JSONB)
  - Row-level security
  - Better complex queries than MongoDB

### Caching & Session Store
- **Redis 7+**: For sessions, rate limiting, temporary data
- **Library**: go-redis/redis v9

### Authentication
- **JWT**: golang-jwt/jwt v5
- **Password Hashing**: bcrypt (x/crypto/bcrypt)
- **2FA**: TOTP (pquerna/otp)

### Email Service
- **Library**: gomail.v2
- **SMTP**: Support for Gmail, SendGrid, AWS SES

### AI Integration
- **Google Generative AI**: Official Go SDK
- **Model**: Gemini 2.0 Flash
- **Alternative**: OpenAI Go SDK (for fallback)

### File Storage
- **Local**: For development
- **Production**: AWS S3 / Google Cloud Storage
- **Library**: aws-sdk-go-v2 / google.golang.org/api

### Validation
- **Library**: go-playground/validator v10
- **Custom validators**: For business logic

### Testing
- **Framework**: testify/suite, testify/mock
- **HTTP Testing**: httptest package
- **Database Testing**: dockertest (ephemeral Postgres containers)

### Monitoring & Logging
- **Logging**: zerolog / logrus (structured logging)
- **Metrics**: Prometheus client
- **Tracing**: OpenTelemetry
- **APM**: Sentry for error tracking

### Security
- **Rate Limiting**: ulule/limiter
- **CORS**: gin-contrib/cors
- **Helmet Equivalent**: secure middleware (gin-contrib/secure)
- **Input Sanitization**: bluemonday (HTML sanitizer)

### Documentation
- **API Docs**: Swag (Swagger/OpenAPI 3.0 generator)
- **Code Docs**: godoc

### DevOps
- **Containerization**: Docker
- **Orchestration**: Docker Compose (dev), Kubernetes (prod)
- **CI/CD**: GitHub Actions
- **Configuration**: viper (12-factor compliant)

---

## Database Design

### PostgreSQL Schema

#### 1. Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    email_verified BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE, -- Soft delete
    meta JSONB DEFAULT '{}'::jsonb, -- Flexible metadata

    -- Indexes
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_meta ON users USING GIN(meta);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

**Meta JSON Structure**:
```json
{
  "avatar_url": "string",
  "timezone": "string",
  "language": "string",
  "preferences": {
    "notifications": "boolean",
    "theme": "dark|light"
  },
  "last_login_ip": "string",
  "last_login_at": "timestamp",
  "login_count": "number"
}
```

#### 2. User Passwords Table
```sql
CREATE TABLE user_passwords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(255) NOT NULL,

    -- Token management
    email_verification_token VARCHAR(255),
    email_token_expires_at TIMESTAMP WITH TIME ZONE,
    password_reset_token VARCHAR(255),
    reset_token_expires_at TIMESTAMP WITH TIME ZONE,

    -- 2FA
    totp_secret VARCHAR(255),
    totp_enabled BOOLEAN DEFAULT FALSE,
    backup_codes TEXT[], -- Array of backup codes

    -- Session management
    logged_sessions TEXT[] DEFAULT '{}', -- Array of JWT token IDs

    -- Password history (prevent reuse)
    password_history JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    meta JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_user_passwords_user_id ON user_passwords(user_id);
CREATE INDEX idx_user_passwords_email_token ON user_passwords(email_verification_token)
    WHERE email_verification_token IS NOT NULL;
CREATE INDEX idx_user_passwords_reset_token ON user_passwords(password_reset_token)
    WHERE password_reset_token IS NOT NULL;
```

**Meta JSON Structure**:
```json
{
  "last_password_change": "timestamp",
  "failed_login_attempts": "number",
  "last_failed_login": "timestamp",
  "account_locked_until": "timestamp",
  "password_expires_at": "timestamp"
}
```

#### 3. Quizzes Table
```sql
CREATE TABLE quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    test_name VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Quiz settings
    time_limit_minutes INTEGER NOT NULL CHECK (time_limit_minutes > 0),
    passing_score_percentage INTEGER DEFAULT 60 CHECK (passing_score_percentage BETWEEN 0 AND 100),
    shuffle_questions BOOLEAN DEFAULT FALSE,
    shuffle_options BOOLEAN DEFAULT FALSE,
    show_results_immediately BOOLEAN DEFAULT TRUE,
    allow_review BOOLEAN DEFAULT TRUE,
    max_attempts INTEGER, -- NULL = unlimited

    -- Status
    status VARCHAR(50) DEFAULT 'draft', -- draft, published, archived
    is_public BOOLEAN DEFAULT FALSE,

    -- Analytics
    total_attempts INTEGER DEFAULT 0,
    average_score DECIMAL(5,2),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    meta JSONB DEFAULT '{}'::jsonb,

    UNIQUE(user_id, test_name)
);

CREATE INDEX idx_quizzes_user_id ON quizzes(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_quizzes_status ON quizzes(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_quizzes_public ON quizzes(is_public) WHERE deleted_at IS NULL AND status = 'published';
CREATE INDEX idx_quizzes_meta ON quizzes USING GIN(meta);
CREATE INDEX idx_quizzes_test_name ON quizzes(user_id, test_name);
```

**Meta JSON Structure**:
```json
{
  "category": "string",
  "tags": ["string"],
  "difficulty": "easy|medium|hard",
  "source_file": {
    "original_name": "string",
    "size": "number",
    "mime_type": "string",
    "upload_timestamp": "timestamp"
  },
  "ai_extraction": {
    "model": "string",
    "confidence_score": "number",
    "processing_time_ms": "number"
  },
  "custom_fields": {}
}
```

#### 4. Questions Table
```sql
CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,

    -- Question content
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) DEFAULT 'multiple_choice', -- multiple_choice, true_false, multi_select
    options JSONB NOT NULL, -- Array of options with metadata
    correct_answer JSONB NOT NULL, -- Can be single or multiple answers

    -- Settings
    points INTEGER DEFAULT 1 CHECK (points > 0),
    order_index INTEGER NOT NULL DEFAULT 0,
    time_limit_seconds INTEGER, -- Per-question time limit (optional)

    -- Explanation
    explanation TEXT,
    reference_links JSONB DEFAULT '[]'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    meta JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_questions_quiz_id ON questions(quiz_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_questions_order ON questions(quiz_id, order_index);
CREATE INDEX idx_questions_options ON questions USING GIN(options);
CREATE INDEX idx_questions_meta ON questions USING GIN(meta);
```

**Options JSONB Structure**:
```json
[
  {
    "id": "uuid",
    "text": "string",
    "order": "number",
    "metadata": {
      "image_url": "string",
      "is_correct": "boolean"
    }
  }
]
```

**Correct Answer JSONB Structure**:
```json
{
  "type": "single|multiple",
  "values": ["option_id"],
  "case_sensitive": "boolean",
  "exact_match": "boolean"
}
```

**Meta JSON Structure**:
```json
{
  "difficulty": "easy|medium|hard",
  "tags": ["string"],
  "source_page": "number",
  "ai_confidence": "number",
  "manual_review": "boolean",
  "media": {
    "image_url": "string",
    "audio_url": "string"
  }
}
```

#### 5. Quiz Attempts Table
```sql
CREATE TABLE quiz_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Attempt details
    attempt_number INTEGER NOT NULL,
    status VARCHAR(50) DEFAULT 'in_progress', -- in_progress, completed, abandoned, timed_out

    -- Scoring
    total_questions INTEGER NOT NULL,
    correct_answers INTEGER DEFAULT 0,
    score_percentage DECIMAL(5,2),
    total_points_earned INTEGER DEFAULT 0,
    total_points_possible INTEGER NOT NULL,
    passed BOOLEAN,

    -- Timing
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    submitted_at TIMESTAMP WITH TIME ZONE,
    time_taken_seconds INTEGER,

    -- Session info
    ip_address INET,
    user_agent TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    meta JSONB DEFAULT '{}'::jsonb,

    UNIQUE(quiz_id, user_id, attempt_number)
);

CREATE INDEX idx_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX idx_attempts_status ON quiz_attempts(status);
CREATE INDEX idx_attempts_score ON quiz_attempts(score_percentage DESC);
CREATE INDEX idx_attempts_started_at ON quiz_attempts(started_at DESC);
CREATE INDEX idx_attempts_meta ON quiz_attempts USING GIN(meta);
```

**Meta JSON Structure**:
```json
{
  "browser": "string",
  "device": "desktop|mobile|tablet",
  "location": {
    "country": "string",
    "city": "string"
  },
  "interruptions": "number",
  "tab_switches": "number",
  "questions_flagged": ["uuid"]
}
```

#### 6. User Answers Table
```sql
CREATE TABLE user_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,

    -- Answer data
    selected_answer JSONB NOT NULL, -- User's answer(s)
    is_correct BOOLEAN NOT NULL,
    points_earned INTEGER DEFAULT 0,

    -- Timing
    time_spent_seconds INTEGER,
    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Flags
    is_flagged BOOLEAN DEFAULT FALSE,
    is_skipped BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    meta JSONB DEFAULT '{}'::jsonb,

    UNIQUE(attempt_id, question_id)
);

CREATE INDEX idx_user_answers_attempt_id ON user_answers(attempt_id);
CREATE INDEX idx_user_answers_question_id ON user_answers(question_id);
CREATE INDEX idx_user_answers_correct ON user_answers(is_correct);
CREATE INDEX idx_user_answers_selected ON user_answers USING GIN(selected_answer);
```

**Selected Answer JSONB Structure**:
```json
{
  "type": "single|multiple",
  "option_ids": ["uuid"],
  "confidence": "number",
  "changed_count": "number"
}
```

#### 7. Audit Logs Table
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    entity_type VARCHAR(100) NOT NULL, -- users, quizzes, questions, etc.
    entity_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL, -- create, update, delete, login, logout, etc.

    -- Change tracking
    old_values JSONB,
    new_values JSONB,

    -- Request context
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    meta JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
```

#### 8. File Uploads Table
```sql
CREATE TABLE file_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id UUID REFERENCES quizzes(id) ON DELETE SET NULL,

    -- File info
    original_filename VARCHAR(500) NOT NULL,
    stored_filename VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,

    -- Storage
    storage_type VARCHAR(50) DEFAULT 'local', -- local, s3, gcs
    storage_path TEXT NOT NULL,
    storage_url TEXT,

    -- Processing
    processing_status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    processing_error TEXT,

    -- AI extraction results
    extraction_results JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    meta JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_file_uploads_user_id ON file_uploads(user_id);
CREATE INDEX idx_file_uploads_quiz_id ON file_uploads(quiz_id);
CREATE INDEX idx_file_uploads_status ON file_uploads(processing_status);
CREATE INDEX idx_file_uploads_created_at ON file_uploads(created_at DESC);
```

### Database Functions & Triggers

#### 1. Updated_at Trigger
```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_passwords_updated_at BEFORE UPDATE ON user_passwords
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_quizzes_updated_at BEFORE UPDATE ON quizzes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_questions_updated_at BEFORE UPDATE ON questions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_quiz_attempts_updated_at BEFORE UPDATE ON quiz_attempts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

#### 2. Quiz Statistics Update Function
```sql
CREATE OR REPLACE FUNCTION update_quiz_statistics()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        UPDATE quizzes
        SET
            total_attempts = total_attempts + 1,
            average_score = (
                SELECT AVG(score_percentage)
                FROM quiz_attempts
                WHERE quiz_id = NEW.quiz_id AND status = 'completed'
            )
        WHERE id = NEW.quiz_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_quiz_stats AFTER INSERT OR UPDATE ON quiz_attempts
    FOR EACH ROW EXECUTE FUNCTION update_quiz_statistics();
```

#### 3. Soft Delete Function
```sql
CREATE OR REPLACE FUNCTION soft_delete()
RETURNS TRIGGER AS $$
BEGIN
    NEW.deleted_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### Views for Common Queries

#### 1. Quiz Summary View
```sql
CREATE VIEW quiz_summaries AS
SELECT
    q.id,
    q.user_id,
    q.test_name,
    q.title,
    q.status,
    q.time_limit_minutes,
    COUNT(DISTINCT qu.id) as question_count,
    q.total_attempts,
    q.average_score,
    q.created_at,
    q.updated_at,
    q.meta
FROM quizzes q
LEFT JOIN questions qu ON q.id = qu.quiz_id AND qu.deleted_at IS NULL
WHERE q.deleted_at IS NULL
GROUP BY q.id;
```

#### 2. User Quiz Performance View
```sql
CREATE VIEW user_quiz_performance AS
SELECT
    qa.user_id,
    qa.quiz_id,
    q.test_name,
    q.title,
    COUNT(qa.id) as total_attempts,
    MAX(qa.score_percentage) as best_score,
    AVG(qa.score_percentage) as average_score,
    MIN(qa.time_taken_seconds) as fastest_time,
    MAX(qa.submitted_at) as last_attempt_at
FROM quiz_attempts qa
JOIN quizzes q ON qa.quiz_id = q.id
WHERE qa.status = 'completed'
GROUP BY qa.user_id, qa.quiz_id, q.test_name, q.title;
```

---

## API Design

### RESTful API Principles
- **Resource-based URLs**: `/api/v1/quizzes`, not `/api/v1/getQuiz`
- **HTTP Methods**: GET (read), POST (create), PUT (full update), PATCH (partial update), DELETE
- **Status Codes**: Proper use of 2xx, 4xx, 5xx
- **Versioning**: URL-based (`/api/v1/`)
- **Pagination**: Cursor-based for performance
- **Filtering**: Query parameters (`?status=published&sort=-created_at`)

### API Structure

#### Base URL
```
Development: http://localhost:8080/api/v1
Production: https://api.quizcraft.com/api/v1
```

### Endpoints

#### Authentication Endpoints

**POST /auth/signup**
```json
Request:
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe",
  "phone_number": "1234567890"
}

Response (201):
{
  "success": true,
  "data": {
    "user_id": "uuid",
    "email": "user@example.com",
    "email_verification_sent": true
  },
  "message": "Signup successful. Please verify your email."
}
```

**POST /auth/verify-email**
```json
Request:
{
  "user_id": "uuid",
  "token": "verification-token"
}

Response (200):
{
  "success": true,
  "message": "Email verified successfully"
}
```

**POST /auth/resend-verification**
```json
Request:
{
  "email": "user@example.com"
}

Response (200):
{
  "success": true,
  "message": "Verification email sent"
}
```

**POST /auth/login**
```json
Request:
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "remember_me": false,
  "totp_code": "123456" // If 2FA enabled
}

Response (200):
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "name": "John Doe",
      "email_verified": true
    },
    "access_token": "jwt-token",
    "refresh_token": "refresh-token",
    "expires_in": 86400
  },
  "message": "Login successful"
}

Set-Cookie: access_token=jwt; HttpOnly; Secure; SameSite=Strict; Max-Age=86400
Set-Cookie: refresh_token=jwt; HttpOnly; Secure; SameSite=Strict; Max-Age=604800
```

**POST /auth/refresh**
```json
Request:
{
  "refresh_token": "refresh-token"
}

Response (200):
{
  "success": true,
  "data": {
    "access_token": "new-jwt-token",
    "expires_in": 86400
  }
}
```

**POST /auth/logout**
```json
Response (200):
{
  "success": true,
  "message": "Logged out successfully"
}
```

**POST /auth/forgot-password**
```json
Request:
{
  "email": "user@example.com"
}

Response (200):
{
  "success": true,
  "message": "Password reset email sent"
}
```

**POST /auth/reset-password**
```json
Request:
{
  "user_id": "uuid",
  "token": "reset-token",
  "new_password": "NewSecurePass123!"
}

Response (200):
{
  "success": true,
  "message": "Password reset successful"
}
```

**POST /auth/change-password** (Authenticated)
```json
Request:
{
  "current_password": "OldPass123!",
  "new_password": "NewPass123!"
}

Response (200):
{
  "success": true,
  "message": "Password changed successfully"
}
```

#### User Endpoints

**GET /users/me** (Authenticated)
```json
Response (200):
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe",
    "phone_number": "1234567890",
    "email_verified": true,
    "enabled": true,
    "created_at": "2025-01-01T00:00:00Z",
    "meta": {
      "avatar_url": "string",
      "timezone": "UTC",
      "preferences": {}
    }
  }
}
```

**PATCH /users/me** (Authenticated)
```json
Request:
{
  "name": "Jane Doe",
  "phone_number": "0987654321",
  "meta": {
    "timezone": "America/New_York",
    "preferences": {
      "theme": "dark"
    }
  }
}

Response (200):
{
  "success": true,
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Jane Doe",
    "phone_number": "0987654321",
    "updated_at": "2025-01-01T00:00:00Z"
  }
}
```

**DELETE /users/me** (Authenticated)
```json
Response (200):
{
  "success": true,
  "message": "Account deleted successfully"
}
```

#### Quiz Endpoints

**POST /quizzes** (Authenticated)
```json
Request:
{
  "test_name": "javascript-basics",
  "title": "JavaScript Fundamentals Quiz",
  "description": "Test your knowledge of JS basics",
  "time_limit_minutes": 30,
  "passing_score_percentage": 70,
  "shuffle_questions": true,
  "shuffle_options": true,
  "is_public": false,
  "meta": {
    "category": "programming",
    "difficulty": "medium",
    "tags": ["javascript", "web-development"]
  }
}

Response (201):
{
  "success": true,
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "test_name": "javascript-basics",
    "title": "JavaScript Fundamentals Quiz",
    "status": "draft",
    "created_at": "2025-01-01T00:00:00Z"
  }
}
```

**POST /quizzes/upload** (Authenticated, Multipart)
```
Form Data:
- file: image/pdf file
- test_name: string
- title: string
- time_limit_minutes: number
- other quiz settings...

Response (202 Accepted):
{
  "success": true,
  "data": {
    "upload_id": "uuid",
    "status": "processing",
    "message": "File uploaded successfully. Processing in background."
  }
}
```

**GET /quizzes/upload/{upload_id}/status** (Authenticated)
```json
Response (200):
{
  "success": true,
  "data": {
    "upload_id": "uuid",
    "status": "completed", // pending, processing, completed, failed
    "quiz_id": "uuid",
    "questions_extracted": 25,
    "processing_time_ms": 5432
  }
}
```

**GET /quizzes** (Authenticated)
```
Query Params:
- page: 1
- limit: 20
- status: published|draft|archived
- sort: -created_at, title, average_score
- search: keyword

Response (200):
{
  "success": true,
  "data": {
    "quizzes": [
      {
        "id": "uuid",
        "test_name": "javascript-basics",
        "title": "JavaScript Fundamentals Quiz",
        "status": "published",
        "question_count": 25,
        "total_attempts": 150,
        "average_score": 75.5,
        "created_at": "2025-01-01T00:00:00Z",
        "meta": {}
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_count": 100,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

**GET /quizzes/public** (Public)
```
Response: Same as GET /quizzes but only shows public published quizzes
```

**GET /quizzes/{quiz_id}** (Authenticated/Owner or Public)
```json
Response (200):
{
  "success": true,
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "test_name": "javascript-basics",
    "title": "JavaScript Fundamentals Quiz",
    "description": "Test your knowledge of JS basics",
    "status": "published",
    "time_limit_minutes": 30,
    "passing_score_percentage": 70,
    "shuffle_questions": true,
    "shuffle_options": true,
    "is_public": false,
    "question_count": 25,
    "total_attempts": 150,
    "average_score": 75.5,
    "created_at": "2025-01-01T00:00:00Z",
    "meta": {}
  }
}
```

**PUT /quizzes/{quiz_id}** (Authenticated/Owner)
```json
Request: Same as POST /quizzes

Response (200):
{
  "success": true,
  "data": {
    // Updated quiz object
  }
}
```

**PATCH /quizzes/{quiz_id}** (Authenticated/Owner)
```json
Request:
{
  "title": "Updated Title",
  "status": "published"
}

Response (200):
{
  "success": true,
  "data": {
    // Updated quiz object
  }
}
```

**DELETE /quizzes/{quiz_id}** (Authenticated/Owner)
```json
Response (200):
{
  "success": true,
  "message": "Quiz deleted successfully"
}
```

**POST /quizzes/{quiz_id}/publish** (Authenticated/Owner)
```json
Response (200):
{
  "success": true,
  "data": {
    "id": "uuid",
    "status": "published",
    "published_at": "2025-01-01T00:00:00Z"
  }
}
```

**POST /quizzes/{quiz_id}/duplicate** (Authenticated)
```json
Request:
{
  "new_test_name": "javascript-basics-copy",
  "new_title": "JavaScript Basics Copy"
}

Response (201):
{
  "success": true,
  "data": {
    // New quiz object
  }
}
```

#### Question Endpoints

**GET /quizzes/{quiz_id}/questions** (Authenticated/Owner or Public Quiz)
```json
Response (200):
{
  "success": true,
  "data": {
    "questions": [
      {
        "id": "uuid",
        "quiz_id": "uuid",
        "question_text": "What is a closure in JavaScript?",
        "question_type": "multiple_choice",
        "options": [
          {
            "id": "uuid",
            "text": "A function with access to outer scope",
            "order": 1
          }
        ],
        "correct_answer": {
          "type": "single",
          "values": ["uuid"]
        },
        "points": 1,
        "order_index": 0,
        "explanation": "Closures are functions that have access...",
        "meta": {}
      }
    ]
  }
}
```

**POST /quizzes/{quiz_id}/questions** (Authenticated/Owner)
```json
Request:
{
  "question_text": "What is a closure in JavaScript?",
  "question_type": "multiple_choice",
  "options": [
    {"text": "A function with access to outer scope", "order": 1},
    {"text": "A loop structure", "order": 2},
    {"text": "A type of variable", "order": 3},
    {"text": "None of the above", "order": 4}
  ],
  "correct_answer": {
    "type": "single",
    "values": [0] // Index of correct option
  },
  "points": 1,
  "explanation": "Closures are...",
  "meta": {
    "difficulty": "medium"
  }
}

Response (201):
{
  "success": true,
  "data": {
    // Created question object
  }
}
```

**PUT /questions/{question_id}** (Authenticated/Owner)
**PATCH /questions/{question_id}** (Authenticated/Owner)
**DELETE /questions/{question_id}** (Authenticated/Owner)

#### Quiz Attempt Endpoints

**POST /quizzes/{quiz_id}/attempts/start** (Authenticated)
```json
Response (201):
{
  "success": true,
  "data": {
    "attempt_id": "uuid",
    "quiz_id": "uuid",
    "attempt_number": 1,
    "status": "in_progress",
    "started_at": "2025-01-01T00:00:00Z",
    "time_limit_minutes": 30,
    "questions": [
      {
        "id": "uuid",
        "question_text": "What is a closure?",
        "question_type": "multiple_choice",
        "options": [
          {"id": "uuid", "text": "Option 1"}
        ],
        "points": 1
        // Note: correct_answer NOT included during attempt
      }
    ]
  }
}
```

**POST /attempts/{attempt_id}/answers** (Authenticated)
```json
Request:
{
  "question_id": "uuid",
  "selected_answer": {
    "type": "single",
    "option_ids": ["uuid"]
  },
  "time_spent_seconds": 45,
  "is_flagged": false
}

Response (200):
{
  "success": true,
  "message": "Answer saved"
}
```

**POST /attempts/{attempt_id}/submit** (Authenticated)
```json
Response (200):
{
  "success": true,
  "data": {
    "attempt_id": "uuid",
    "status": "completed",
    "submitted_at": "2025-01-01T00:30:00Z",
    "time_taken_seconds": 1800,
    "total_questions": 25,
    "correct_answers": 20,
    "score_percentage": 80.00,
    "total_points_earned": 20,
    "total_points_possible": 25,
    "passed": true
  }
}
```

**GET /attempts/{attempt_id}** (Authenticated/Owner)
```json
Response (200):
{
  "success": true,
  "data": {
    "attempt_id": "uuid",
    "quiz": {
      "id": "uuid",
      "title": "JavaScript Fundamentals"
    },
    "user": {
      "id": "uuid",
      "name": "John Doe"
    },
    "status": "completed",
    "score_percentage": 80.00,
    "started_at": "2025-01-01T00:00:00Z",
    "submitted_at": "2025-01-01T00:30:00Z",
    "answers": [
      {
        "question_id": "uuid",
        "question_text": "What is a closure?",
        "selected_answer": {"type": "single", "option_ids": ["uuid"]},
        "correct_answer": {"type": "single", "values": ["uuid"]},
        "is_correct": true,
        "points_earned": 1,
        "explanation": "Closures are..."
      }
    ]
  }
}
```

**GET /quizzes/{quiz_id}/attempts** (Authenticated/Owner)
```json
Response (200):
{
  "success": true,
  "data": {
    "attempts": [
      {
        "id": "uuid",
        "attempt_number": 1,
        "status": "completed",
        "score_percentage": 80.00,
        "started_at": "2025-01-01T00:00:00Z",
        "submitted_at": "2025-01-01T00:30:00Z"
      }
    ],
    "statistics": {
      "total_attempts": 3,
      "best_score": 85.00,
      "average_score": 78.33,
      "attempts_remaining": 2 // If max_attempts set
    }
  }
}
```

**GET /users/me/attempts** (Authenticated)
```
Query: ?status=completed&sort=-submitted_at&page=1&limit=20

Response: List of user's attempts across all quizzes
```

#### Analytics Endpoints

**GET /quizzes/{quiz_id}/analytics** (Authenticated/Owner)
```json
Response (200):
{
  "success": true,
  "data": {
    "quiz_id": "uuid",
    "total_attempts": 150,
    "unique_users": 120,
    "completion_rate": 85.5,
    "average_score": 75.5,
    "average_time_seconds": 1650,
    "score_distribution": {
      "0-20": 5,
      "21-40": 10,
      "41-60": 25,
      "61-80": 60,
      "81-100": 50
    },
    "question_analytics": [
      {
        "question_id": "uuid",
        "question_text": "What is a closure?",
        "total_attempts": 150,
        "correct_count": 120,
        "accuracy_percentage": 80.00,
        "average_time_seconds": 45
      }
    ]
  }
}
```

**GET /users/me/statistics** (Authenticated)
```json
Response (200):
{
  "success": true,
  "data": {
    "quizzes_created": 10,
    "quizzes_attempted": 25,
    "quizzes_completed": 20,
    "average_score": 78.5,
    "total_time_spent_hours": 15.5,
    "achievements": [
      {
        "type": "streak",
        "name": "7 Day Streak",
        "earned_at": "2025-01-01T00:00:00Z"
      }
    ]
  }
}
```

### Response Format Standards

#### Success Response
```json
{
  "success": true,
  "data": {},
  "message": "Optional success message",
  "meta": {
    "request_id": "uuid",
    "timestamp": "2025-01-01T00:00:00Z"
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format",
        "value": "invalid-email"
      }
    ]
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2025-01-01T00:00:00Z"
  }
}
```

### HTTP Status Codes

- **200 OK**: Successful GET, PUT, PATCH, DELETE
- **201 Created**: Successful POST
- **202 Accepted**: Request accepted, processing async
- **204 No Content**: Successful DELETE with no response body
- **400 Bad Request**: Invalid request data
- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Authenticated but not authorized
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource conflict (duplicate)
- **422 Unprocessable Entity**: Validation error
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Server error
- **503 Service Unavailable**: Service temporarily unavailable

---

## Authentication & Authorization

### JWT Token Structure

#### Access Token (Short-lived: 24 hours)
```json
Header:
{
  "alg": "HS256",
  "typ": "JWT"
}

Payload:
{
  "sub": "user_id",
  "email": "user@example.com",
  "name": "John Doe",
  "type": "access",
  "jti": "token_id", // Unique token ID for blacklist
  "iat": 1234567890,
  "exp": 1234654290
}
```

#### Refresh Token (Long-lived: 7 days)
```json
Payload:
{
  "sub": "user_id",
  "type": "refresh",
  "jti": "token_id",
  "iat": 1234567890,
  "exp": 1235172690
}
```

### Token Storage
- **Access Token**: HTTP-only cookie (primary), can also be in header
- **Refresh Token**: HTTP-only cookie only
- **Cookie Attributes**:
  - `HttpOnly`: Prevents XSS
  - `Secure`: HTTPS only in production
  - `SameSite=Strict`: CSRF protection
  - `Path=/api`: Limit scope

### Authentication Flow

1. **Login**:
   - User submits credentials
   - Server validates and generates tokens
   - Tokens stored in HTTP-only cookies
   - Token ID added to Redis logged_sessions set

2. **Request**:
   - Client sends request with cookie
   - Middleware extracts and validates JWT
   - Checks token ID exists in Redis
   - Attaches user object to context

3. **Refresh**:
   - Access token expires
   - Client sends refresh token
   - Server validates and issues new access token
   - Old access token ID added to blacklist in Redis

4. **Logout**:
   - Remove token ID from logged_sessions in Redis
   - Add token ID to blacklist in Redis (TTL = remaining token life)
   - Clear cookies

### Authorization Middleware

```go
// Authentication middleware
func AuthRequired() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract token from cookie or header
        // Validate JWT signature
        // Check if token ID exists in Redis logged_sessions
        // Check if token ID is blacklisted
        // Attach user to context
    }
}

// Owner authorization
func RequireOwner(resourceType string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get resource ID from URL
        // Get user from context
        // Check if user owns resource
        // Allow admin bypass
    }
}

// Admin authorization
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if user has admin role
    }
}

// Rate limiting
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check Redis for request count
        // Increment counter
        // Return 429 if exceeded
    }
}
```

### Password Security

#### Password Requirements
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character
- Not in common password list
- Not same as previous 5 passwords

#### Hashing Strategy
```go
// Generate salt
salt := generateRandomString(32)

// Combine salt + password
combined := salt + password

// Hash with bcrypt (cost factor 12)
hash, err := bcrypt.GenerateFromPassword([]byte(combined), 12)

// Store both salt and hash separately
```

### Two-Factor Authentication (2FA)

#### Setup Flow
1. User enables 2FA in settings
2. Server generates TOTP secret
3. Return QR code (otpauth:// URI)
4. User scans with authenticator app
5. User submits verification code
6. Server validates and enables 2FA
7. Generate and show backup codes (10 codes)

#### Login with 2FA
1. User submits email + password
2. Server validates credentials
3. Server returns `requires_2fa: true` (no tokens yet)
4. User submits TOTP code
5. Server validates TOTP code or backup code
6. If backup code used, mark as consumed
7. Issue access + refresh tokens

### Session Management

#### Redis Session Structure
```
Key: user_sessions:{user_id}
Type: Set
Value: [token_id_1, token_id_2, ...]
TTL: 7 days (refresh token expiry)

Key: token_blacklist:{token_id}
Type: String
Value: "1"
TTL: Remaining token lifetime
```

#### Multi-Device Support
- Each login creates new token pair
- User can see active sessions
- User can revoke specific sessions
- Global logout revokes all sessions

---

## AI Integration

### Google Generative AI (Gemini)

#### Configuration
```go
import "google.golang.org/genai"

client := genai.NewClient(ctx, option.WithAPIKey(apiKey))
model := client.GenerativeModel("gemini-2.0-flash-exp")

// Configure model
model.SetTemperature(0.2) // Lower for consistent output
model.SetTopP(0.8)
model.SetTopK(40)
model.SetMaxOutputTokens(4096)
```

#### Question Extraction Prompt
```
You are an expert at extracting quiz questions from images and PDFs.

Extract all multiple-choice questions from the provided document. For each question:

1. Extract the complete question text (remove question numbers like "Q.1", "1.", etc.)
2. Extract all answer options (remove option letters like "a)", "b)", etc.)
3. Identify the correct answer (the actual answer text, not the letter)
4. If no correct answer is marked, leave the answer field empty

Return ONLY a valid JSON array with this exact structure:
[
  {
    "question": "question text here",
    "options": ["option 1", "option 2", "option 3", "option 4"],
    "answer": "correct option text",
    "explanation": "optional explanation if provided",
    "metadata": {
      "confidence": 0.95,
      "page": 1,
      "question_type": "multiple_choice"
    }
  }
]

Rules:
- Remove all formatting artifacts (numbers, letters, bullets)
- Preserve mathematical notation and special characters
- If an image contains non-question content, ignore it
- If confidence is below 0.7 for any question, mark it in metadata
- Handle multi-part questions as separate entries
```

#### Request Structure
```go
// Convert file to base64
fileData, _ := ioutil.ReadFile(filePath)
base64Data := base64.StdEncoding.EncodeToString(fileData)

// Create image part
imagePart := genai.ImageData{
    MIMEType: mimeType,
    Data:     fileData,
}

// Send request with schema
resp, err := model.GenerateContent(ctx,
    imagePart,
    genai.Text(prompt),
    &genai.GenerateContentConfig{
        ResponseMIMEType: "application/json",
        ResponseSchema: &genai.Schema{
            Type: genai.TypeArray,
            Items: &genai.Schema{
                Type: genai.TypeObject,
                Properties: map[string]*genai.Schema{
                    "question": {Type: genai.TypeString},
                    "options": {
                        Type: genai.TypeArray,
                        Items: &genai.Schema{Type: genai.TypeString},
                    },
                    "answer": {Type: genai.TypeString},
                    "explanation": {Type: genai.TypeString},
                    "metadata": {Type: genai.TypeObject},
                },
                Required: []string{"question", "options", "answer"},
            },
        },
    },
)
```

#### Response Parsing
```go
type ExtractedQuestion struct {
    Question    string                 `json:"question"`
    Options     []string               `json:"options"`
    Answer      string                 `json:"answer"`
    Explanation string                 `json:"explanation"`
    Metadata    map[string]interface{} `json:"metadata"`
}

// Parse JSON response
var questions []ExtractedQuestion
err := json.Unmarshal([]byte(resp.Text()), &questions)

// Validate and sanitize
for i, q := range questions {
    // Trim whitespace
    q.Question = strings.TrimSpace(q.Question)

    // Validate minimum requirements
    if len(q.Question) < 5 {
        continue // Skip invalid questions
    }

    if len(q.Options) < 2 {
        continue // Need at least 2 options
    }

    // Sanitize HTML (prevent XSS)
    q.Question = sanitizeHTML(q.Question)
    for j := range q.Options {
        q.Options[j] = sanitizeHTML(q.Options[j])
    }

    // Check confidence score
    if confidence, ok := q.Metadata["confidence"].(float64); ok {
        if confidence < 0.7 {
            q.Metadata["manual_review"] = true
        }
    }
}
```

#### Error Handling
```go
// Retry logic with exponential backoff
maxRetries := 3
backoff := time.Second

for attempt := 0; attempt < maxRetries; attempt++ {
    resp, err := model.GenerateContent(ctx, ...)
    if err == nil {
        break
    }

    // Check error type
    if isRateLimitError(err) {
        time.Sleep(backoff)
        backoff *= 2
        continue
    }

    if isInvalidContentError(err) {
        return nil, fmt.Errorf("invalid content: %w", err)
    }

    // Other errors
    return nil, fmt.Errorf("AI extraction failed: %w", err)
}
```

#### Background Processing
```go
// Process file upload asynchronously
func ProcessFileUpload(uploadID string) {
    // Update status to processing
    updateUploadStatus(uploadID, "processing")

    // Extract questions
    questions, err := extractQuestionsFromFile(uploadID)
    if err != nil {
        updateUploadStatus(uploadID, "failed")
        saveProcessingError(uploadID, err)
        sendErrorNotification(uploadID)
        return
    }

    // Create quiz and questions
    quizID, err := createQuizFromExtraction(uploadID, questions)
    if err != nil {
        updateUploadStatus(uploadID, "failed")
        saveProcessingError(uploadID, err)
        return
    }

    // Update status to completed
    updateUploadStatus(uploadID, "completed")
    updateUploadWithQuizID(uploadID, quizID)
    sendSuccessNotification(uploadID, quizID)
}
```

### Alternative AI Provider (OpenAI)

```go
// Fallback to OpenAI if Gemini fails
import "github.com/sashabaranov/go-openai"

client := openai.NewClient(apiKey)

resp, err := client.CreateChatCompletion(
    ctx,
    openai.ChatCompletionRequest{
        Model: openai.GPT4VisionPreview,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: prompt,
            },
            {
                Role: openai.ChatMessageRoleUser,
                MultiContent: []openai.ChatMessagePart{
                    {
                        Type: openai.ChatMessagePartTypeImageURL,
                        ImageURL: &openai.ChatMessageImageURL{
                            URL: base64ImageURL,
                        },
                    },
                },
            },
        },
        ResponseFormat: &openai.ChatCompletionResponseFormat{
            Type: openai.ChatCompletionResponseFormatTypeJSONObject,
        },
    },
)
```

---

## Security Implementation

### Input Validation

```go
// Use go-playground/validator
type SignupRequest struct {
    Email       string `json:"email" validate:"required,email,max=255"`
    Password    string `json:"password" validate:"required,min=8,max=128,password"`
    Name        string `json:"name" validate:"required,min=2,max=255"`
    PhoneNumber string `json:"phone_number" validate:"omitempty,e164"`
}

// Custom password validator
func validatePassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()

    // Check length
    if len(password) < 8 || len(password) > 128 {
        return false
    }

    // Check complexity
    hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
    hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
    hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
    hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)

    return hasUpper && hasLower && hasNumber && hasSpecial
}

// Register custom validator
validate := validator.New()
validate.RegisterValidation("password", validatePassword)
```

### SQL Injection Prevention

```go
// GORM automatically escapes SQL
// But for raw queries:
db.Raw("SELECT * FROM users WHERE email = ?", email).Scan(&user)

// NEVER do this:
db.Raw(fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email))
```

### XSS Prevention

```go
import "github.com/microcosm-cc/bluemonday"

// Sanitize HTML input
policy := bluemonday.StrictPolicy()
cleanHTML := policy.Sanitize(userInput)

// For rich text (if needed later)
ugcPolicy := bluemonday.UGCPolicy()
cleanHTML = ugcPolicy.Sanitize(userInput)
```

### CSRF Protection

```go
// Use gin CSRF middleware
import "github.com/utrack/gin-csrf"

router.Use(csrf.Middleware(csrf.Options{
    Secret: os.Getenv("CSRF_SECRET"),
    ErrorFunc: func(c *gin.Context) {
        c.JSON(403, gin.H{"error": "CSRF token invalid"})
        c.Abort()
    },
}))

// For API with JWT, CSRF not needed if:
// 1. Tokens in HTTP-only cookies with SameSite=Strict
// 2. Custom header required (X-Requested-With)
```

### Rate Limiting

```go
import (
    "github.com/ulule/limiter/v3"
    "github.com/ulule/limiter/v3/drivers/store/redis"
)

// Create rate limiter
rate := limiter.Rate{
    Period: 1 * time.Hour,
    Limit:  100, // 100 requests per hour
}

store := redis.NewStore(redisClient)
instance := limiter.New(store, rate)

// Middleware
func RateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Use user ID if authenticated, IP otherwise
        key := getUserID(c)
        if key == "" {
            key = c.ClientIP()
        }

        context, err := instance.Get(c, key)
        if err != nil {
            c.JSON(500, gin.H{"error": "Rate limit check failed"})
            c.Abort()
            return
        }

        // Set rate limit headers
        c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
        c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
        c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

        if context.Reached {
            c.JSON(429, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": context.Reset,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// Different limits for different endpoints
authLimiter := limiter.New(store, limiter.Rate{Period: 15 * time.Minute, Limit: 5})  // Login
apiLimiter := limiter.New(store, limiter.Rate{Period: 1 * time.Hour, Limit: 1000})   // General API
uploadLimiter := limiter.New(store, limiter.Rate{Period: 1 * time.Hour, Limit: 10})  // File uploads
```

### CORS Configuration

```go
import "github.com/gin-contrib/cors"

config := cors.Config{
    AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
    ExposeHeaders:    []string{"Content-Length", "X-RateLimit-Limit", "X-RateLimit-Remaining"},
    AllowCredentials: true,
    MaxAge:           12 * time.Hour,
}

router.Use(cors.New(config))
```

### Security Headers

```go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevent clickjacking
        c.Header("X-Frame-Options", "DENY")

        // Prevent MIME sniffing
        c.Header("X-Content-Type-Options", "nosniff")

        // XSS protection (legacy but doesn't hurt)
        c.Header("X-XSS-Protection", "1; mode=block")

        // Content Security Policy
        c.Header("Content-Security-Policy", "default-src 'self'")

        // Strict Transport Security (HTTPS only)
        if c.Request.TLS != nil {
            c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        }

        // Referrer Policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

        // Permissions Policy
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

        c.Next()
    }
}
```

### File Upload Security

```go
func ValidateFileUpload(file *multipart.FileHeader) error {
    // Check file size (10MB max)
    maxSize := int64(10 * 1024 * 1024)
    if file.Size > maxSize {
        return errors.New("file too large (max 10MB)")
    }

    // Check MIME type
    allowedMIMETypes := map[string]bool{
        "image/jpeg":      true,
        "image/png":       true,
        "application/pdf": true,
    }

    if !allowedMIMETypes[file.Header.Get("Content-Type")] {
        return errors.New("invalid file type")
    }

    // Read file and verify actual MIME type (not just header)
    fileReader, _ := file.Open()
    defer fileReader.Close()

    buffer := make([]byte, 512)
    _, err := fileReader.Read(buffer)
    if err != nil {
        return err
    }

    actualMIME := http.DetectContentType(buffer)
    if !allowedMIMETypes[actualMIME] {
        return errors.New("file content doesn't match extension")
    }

    // Check for malicious content (scan with ClamAV in production)
    // scanResult := antivirusScan(file)

    return nil
}

// Generate safe filename
func GenerateSafeFilename(originalName string) string {
    ext := filepath.Ext(originalName)
    name := strings.TrimSuffix(originalName, ext)

    // Remove dangerous characters
    name = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(name, "_")

    // Generate unique filename
    timestamp := time.Now().Unix()
    random := generateRandomString(8)

    return fmt.Sprintf("%d_%s_%s%s", timestamp, random, name, ext)
}
```

### Secrets Management

```go
// Use environment variables with viper
import "github.com/spf13/viper"

func LoadConfig() error {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("/etc/quizcraft/")
    viper.AddConfigPath("$HOME/.quizcraft")
    viper.AddConfigPath(".")

    // Environment variables take precedence
    viper.AutomaticEnv()
    viper.SetEnvPrefix("QUIZCRAFT")

    // Required configs
    viper.SetDefault("PORT", "8080")
    viper.SetDefault("ENV", "development")

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found; using env vars and defaults
        } else {
            return err
        }
    }

    // Validate required secrets
    required := []string{
        "DATABASE_URL",
        "JWT_SECRET",
        "REDIS_URL",
        "GOOGLE_AI_API_KEY",
    }

    for _, key := range required {
        if !viper.IsSet(key) {
            return fmt.Errorf("required config missing: %s", key)
        }
    }

    return nil
}

// In production, use secret managers:
// - AWS Secrets Manager
// - Google Secret Manager
// - HashiCorp Vault
// - Kubernetes Secrets
```

---

## File Upload & Processing

### File Upload Flow

```
1. Client uploads file → multipart/form-data
2. Server validates file (size, type, content)
3. Save file to temporary location
4. Create file_uploads record (status: pending)
5. Return upload_id immediately (202 Accepted)
6. Process file asynchronously in background worker
7. Update file_uploads status to processing
8. Extract questions with AI
9. Create quiz and questions
10. Update file_uploads status to completed
11. Move file to permanent storage (or delete if configured)
12. Send notification/webhook to client
```

### Storage Strategy

#### Development
```go
// Local filesystem
uploadDir := "./uploads"
filePath := filepath.Join(uploadDir, safeFilename)
err := c.SaveUploadedFile(file, filePath)
```

#### Production (AWS S3)
```go
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

// Initialize S3 client
cfg, _ := config.LoadDefaultConfig(ctx)
s3Client := s3.NewFromConfig(cfg)

// Upload file
fileReader, _ := file.Open()
defer fileReader.Close()

_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: aws.String("quizcraft-uploads"),
    Key:    aws.String(safeFilename),
    Body:   fileReader,
    ContentType: aws.String(file.Header.Get("Content-Type")),
    ServerSideEncryption: types.ServerSideEncryptionAes256,
    Metadata: map[string]string{
        "user-id":    userID,
        "upload-id":  uploadID,
        "original":   file.Filename,
    },
})

// Generate presigned URL for retrieval (expires in 1 hour)
presignClient := s3.NewPresignClient(s3Client)
presignedURL, _ := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
    Bucket: aws.String("quizcraft-uploads"),
    Key:    aws.String(safeFilename),
}, s3.WithPresignExpires(1*time.Hour))
```

### Background Worker

```go
// Use a work queue (Redis + worker pool)
type Job struct {
    Type    string                 `json:"type"`
    Payload map[string]interface{} `json:"payload"`
}

// Worker pool
func StartWorkerPool(numWorkers int) {
    jobQueue := make(chan Job, 100)

    for i := 0; i < numWorkers; i++ {
        go worker(i, jobQueue)
    }

    // Redis subscriber for new jobs
    go subscribeToJobs(jobQueue)
}

func worker(id int, jobs <-chan Job) {
    for job := range jobs {
        log.Printf("Worker %d processing job: %s", id, job.Type)

        switch job.Type {
        case "file_upload":
            processFileUpload(job.Payload)
        case "send_email":
            sendEmail(job.Payload)
        default:
            log.Printf("Unknown job type: %s", job.Type)
        }
    }
}

// Enqueue job
func EnqueueJob(jobType string, payload map[string]interface{}) error {
    job := Job{
        Type:    jobType,
        Payload: payload,
    }

    jobJSON, _ := json.Marshal(job)

    // Push to Redis list
    return redisClient.LPush(ctx, "job_queue", jobJSON).Err()
}

// Subscribe to jobs
func subscribeToJobs(jobQueue chan<- Job) {
    for {
        // Blocking pop from Redis list
        result, err := redisClient.BRPop(ctx, 0, "job_queue").Result()
        if err != nil {
            log.Printf("Error popping job: %v", err)
            time.Sleep(1 * time.Second)
            continue
        }

        var job Job
        if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
            log.Printf("Error unmarshaling job: %v", err)
            continue
        }

        jobQueue <- job
    }
}
```

### File Processing

```go
func processFileUpload(payload map[string]interface{}) {
    uploadID := payload["upload_id"].(string)

    // Get upload record
    var upload FileUpload
    db.First(&upload, "id = ?", uploadID)

    // Update status
    db.Model(&upload).Update("processing_status", "processing")

    // Extract questions from file
    questions, err := extractQuestionsFromFile(upload.StoragePath)
    if err != nil {
        db.Model(&upload).Updates(map[string]interface{}{
            "processing_status": "failed",
            "processing_error":  err.Error(),
        })

        // Send error notification
        sendNotification(upload.UserID, "upload_failed", map[string]interface{}{
            "upload_id": uploadID,
            "error":     err.Error(),
        })

        return
    }

    // Create quiz
    quiz := Quiz{
        UserID:            upload.UserID,
        TestName:          payload["test_name"].(string),
        Title:             payload["title"].(string),
        TimeLimitMinutes:  int(payload["time_limit_minutes"].(float64)),
        Status:            "draft",
    }
    db.Create(&quiz)

    // Create questions
    for i, q := range questions {
        // Convert options to JSONB
        optionsJSON, _ := json.Marshal(q.Options)
        correctAnswerJSON, _ := json.Marshal(q.CorrectAnswer)

        question := Question{
            QuizID:        quiz.ID,
            QuestionText:  q.Question,
            Options:       optionsJSON,
            CorrectAnswer: correctAnswerJSON,
            OrderIndex:    i,
            Points:        1,
        }
        db.Create(&question)
    }

    // Update upload record
    extractionResults, _ := json.Marshal(map[string]interface{}{
        "questions_count": len(questions),
        "processing_time": time.Since(upload.CreatedAt).Seconds(),
    })

    db.Model(&upload).Updates(map[string]interface{}{
        "processing_status":   "completed",
        "quiz_id":             quiz.ID,
        "extraction_results":  extractionResults,
    })

    // Send success notification
    sendNotification(upload.UserID, "upload_completed", map[string]interface{}{
        "upload_id":       uploadID,
        "quiz_id":         quiz.ID,
        "questions_count": len(questions),
    })
}
```

---

## Error Handling

### Error Types

```go
// Custom error types
type AppError struct {
    Code       string                 `json:"code"`
    Message    string                 `json:"message"`
    Details    interface{}            `json:"details,omitempty"`
    HTTPStatus int                    `json:"-"`
    Internal   error                  `json:"-"` // Internal error (not exposed)
}

func (e *AppError) Error() string {
    return e.Message
}

// Common errors
var (
    ErrValidation = &AppError{
        Code:       "VALIDATION_ERROR",
        Message:    "Validation failed",
        HTTPStatus: 422,
    }

    ErrUnauthorized = &AppError{
        Code:       "UNAUTHORIZED",
        Message:    "Authentication required",
        HTTPStatus: 401,
    }

    ErrForbidden = &AppError{
        Code:       "FORBIDDEN",
        Message:    "You don't have permission to access this resource",
        HTTPStatus: 403,
    }

    ErrNotFound = &AppError{
        Code:       "NOT_FOUND",
        Message:    "Resource not found",
        HTTPStatus: 404,
    }

    ErrConflict = &AppError{
        Code:       "CONFLICT",
        Message:    "Resource already exists",
        HTTPStatus: 409,
    }

    ErrInternal = &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "An internal error occurred",
        HTTPStatus: 500,
    }

    ErrServiceUnavailable = &AppError{
        Code:       "SERVICE_UNAVAILABLE",
        Message:    "Service temporarily unavailable",
        HTTPStatus: 503,
    }
)

// Create specific error
func NewValidationError(details interface{}) *AppError {
    return &AppError{
        Code:       ErrValidation.Code,
        Message:    ErrValidation.Message,
        Details:    details,
        HTTPStatus: ErrValidation.HTTPStatus,
    }
}
```

### Error Handler Middleware

```go
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // Check if there are errors
        if len(c.Errors) == 0 {
            return
        }

        // Get last error
        err := c.Errors.Last()

        // Generate request ID if not exists
        requestID := c.GetString("request_id")
        if requestID == "" {
            requestID = generateRequestID()
        }

        // Log error
        logger.Error("Request error",
            "request_id", requestID,
            "path", c.Request.URL.Path,
            "method", c.Request.Method,
            "error", err.Error(),
        )

        // Handle AppError
        if appErr, ok := err.Err.(*AppError); ok {
            c.JSON(appErr.HTTPStatus, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    appErr.Code,
                    "message": appErr.Message,
                    "details": appErr.Details,
                },
                "meta": gin.H{
                    "request_id": requestID,
                    "timestamp":  time.Now().UTC(),
                },
            })
            return
        }

        // Handle validation errors
        if validationErrs, ok := err.Err.(validator.ValidationErrors); ok {
            details := make([]map[string]interface{}, 0)
            for _, e := range validationErrs {
                details = append(details, map[string]interface{}{
                    "field":   e.Field(),
                    "message": getValidationMessage(e),
                    "value":   e.Value(),
                })
            }

            c.JSON(422, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "VALIDATION_ERROR",
                    "message": "Validation failed",
                    "details": details,
                },
                "meta": gin.H{
                    "request_id": requestID,
                    "timestamp":  time.Now().UTC(),
                },
            })
            return
        }

        // Handle GORM errors
        if errors.Is(err.Err, gorm.ErrRecordNotFound) {
            c.JSON(404, gin.H{
                "success": false,
                "error": gin.H{
                    "code":    "NOT_FOUND",
                    "message": "Resource not found",
                },
                "meta": gin.H{
                    "request_id": requestID,
                    "timestamp":  time.Now().UTC(),
                },
            })
            return
        }

        // Handle unknown errors
        // Don't expose internal error details in production
        message := "An internal error occurred"
        if viper.GetString("ENV") == "development" {
            message = err.Error()
        }

        c.JSON(500, gin.H{
            "success": false,
            "error": gin.H{
                "code":    "INTERNAL_ERROR",
                "message": message,
            },
            "meta": gin.H{
                "request_id": requestID,
                "timestamp":  time.Now().UTC(),
            },
        })
    }
}

func getValidationMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "This field is required"
    case "email":
        return "Invalid email format"
    case "min":
        return fmt.Sprintf("Minimum length is %s", e.Param())
    case "max":
        return fmt.Sprintf("Maximum length is %s", e.Param())
    case "password":
        return "Password must contain uppercase, lowercase, number, and special character"
    default:
        return "Invalid value"
    }
}
```

### Panic Recovery

```go
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // Get stack trace
                stack := make([]byte, 4096)
                length := runtime.Stack(stack, false)

                requestID := c.GetString("request_id")

                // Log panic
                logger.Error("Panic recovered",
                    "request_id", requestID,
                    "error", err,
                    "stack", string(stack[:length]),
                )

                // Report to Sentry
                if viper.GetString("ENV") == "production" {
                    sentry.CurrentHub().Recover(err)
                    sentry.Flush(2 * time.Second)
                }

                // Return error response
                c.JSON(500, gin.H{
                    "success": false,
                    "error": gin.H{
                        "code":    "INTERNAL_ERROR",
                        "message": "An unexpected error occurred",
                    },
                    "meta": gin.H{
                        "request_id": requestID,
                        "timestamp":  time.Now().UTC(),
                    },
                })

                c.Abort()
            }
        }()

        c.Next()
    }
}
```

---

## Logging & Monitoring

### Structured Logging

```go
import "github.com/rs/zerolog"

// Initialize logger
func InitLogger() zerolog.Logger {
    // Pretty logging for development
    if viper.GetString("ENV") == "development" {
        output := zerolog.ConsoleWriter{
            Out:        os.Stdout,
            TimeFormat: time.RFC3339,
        }
        return zerolog.New(output).With().Timestamp().Logger()
    }

    // JSON logging for production
    return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// Request logging middleware
func RequestLogger(logger zerolog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery

        // Generate request ID
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID()
        }
        c.Set("request_id", requestID)

        // Process request
        c.Next()

        // Calculate latency
        latency := time.Since(start)

        // Log request
        logEvent := logger.Info()

        if c.Writer.Status() >= 500 {
            logEvent = logger.Error()
        } else if c.Writer.Status() >= 400 {
            logEvent = logger.Warn()
        }

        if raw != "" {
            path = path + "?" + raw
        }

        logEvent.
            Str("request_id", requestID).
            Str("method", c.Request.Method).
            Str("path", path).
            Int("status", c.Writer.Status()).
            Dur("latency", latency).
            Str("ip", c.ClientIP()).
            Str("user_agent", c.Request.UserAgent()).
            Int("body_size", c.Writer.Size()).
            Msg("Request completed")
    }
}
```

### Metrics (Prometheus)

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    quizzesCreated = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "quizzes_created_total",
            Help: "Total number of quizzes created",
        },
    )

    quizAttempts = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "quiz_attempts_total",
            Help: "Total number of quiz attempts",
        },
    )

    aiExtractionDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "ai_extraction_duration_seconds",
            Help:    "AI extraction duration in seconds",
            Buckets: []float64{1, 2, 5, 10, 20, 30, 60},
        },
    )

    activeUsers = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_users",
            Help: "Number of currently active users",
        },
    )
)

// Metrics middleware
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.FullPath() // Use route pattern, not actual path
        if path == "" {
            path = "not_found"
        }

        c.Next()

        duration := time.Since(start).Seconds()
        status := fmt.Sprintf("%d", c.Writer.Status())

        httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
    }
}

// Expose metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### Health Checks

```go
type HealthStatus struct {
    Status    string                 `json:"status"`
    Timestamp time.Time              `json:"timestamp"`
    Services  map[string]ServiceInfo `json:"services"`
}

type ServiceInfo struct {
    Status   string  `json:"status"`
    Latency  float64 `json:"latency_ms,omitempty"`
    Message  string  `json:"message,omitempty"`
}

func HealthCheckHandler(c *gin.Context) {
    health := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now().UTC(),
        Services:  make(map[string]ServiceInfo),
    }

    // Check database
    dbStart := time.Now()
    if err := db.Exec("SELECT 1").Error; err != nil {
        health.Status = "unhealthy"
        health.Services["database"] = ServiceInfo{
            Status:  "down",
            Message: "Database connection failed",
        }
    } else {
        health.Services["database"] = ServiceInfo{
            Status:  "up",
            Latency: float64(time.Since(dbStart).Milliseconds()),
        }
    }

    // Check Redis
    redisStart := time.Now()
    if err := redisClient.Ping(ctx).Err(); err != nil {
        health.Status = "unhealthy"
        health.Services["redis"] = ServiceInfo{
            Status:  "down",
            Message: "Redis connection failed",
        }
    } else {
        health.Services["redis"] = ServiceInfo{
            Status:  "up",
            Latency: float64(time.Since(redisStart).Milliseconds()),
        }
    }

    // Check AI service (lightweight check)
    health.Services["ai_service"] = ServiceInfo{
        Status: "up",
    }

    statusCode := 200
    if health.Status == "unhealthy" {
        statusCode = 503
    }

    c.JSON(statusCode, health)
}

// Liveness probe (Kubernetes)
func LivenessHandler(c *gin.Context) {
    c.JSON(200, gin.H{"status": "alive"})
}

// Readiness probe (Kubernetes)
func ReadinessHandler(c *gin.Context) {
    // Check if service is ready to handle requests
    if err := db.Exec("SELECT 1").Error; err != nil {
        c.JSON(503, gin.H{"status": "not ready"})
        return
    }

    c.JSON(200, gin.H{"status": "ready"})
}
```

### APM (Sentry)

```go
import "github.com/getsentry/sentry-go"

func InitSentry() {
    err := sentry.Init(sentry.ClientOptions{
        Dsn:              viper.GetString("SENTRY_DSN"),
        Environment:      viper.GetString("ENV"),
        Release:          "quizcraft@" + viper.GetString("VERSION"),
        TracesSampleRate: 0.2, // 20% of transactions
        BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
            // Filter out sensitive data
            if event.Request != nil {
                event.Request.Cookies = ""
                event.Request.Headers = filterHeaders(event.Request.Headers)
            }
            return event
        },
    })

    if err != nil {
        log.Printf("Sentry initialization failed: %v", err)
    }
}

// Sentry middleware
func SentryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        hub := sentry.CurrentHub().Clone()
        hub.Scope().SetRequest(c.Request)
        hub.Scope().SetUser(sentry.User{
            ID:        c.GetString("user_id"),
            IPAddress: c.ClientIP(),
        })
        hub.Scope().SetTag("request_id", c.GetString("request_id"))

        c.Set("sentry_hub", hub)
        c.Next()

        // Capture errors
        if len(c.Errors) > 0 {
            for _, err := range c.Errors {
                hub.CaptureException(err.Err)
            }
        }
    }
}
```

---

## Testing Strategy

### Unit Tests

```go
// services/quiz_service_test.go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock repository
type MockQuizRepository struct {
    mock.Mock
}

func (m *MockQuizRepository) Create(quiz *Quiz) error {
    args := m.Called(quiz)
    return args.Error(0)
}

func (m *MockQuizRepository) FindByID(id string) (*Quiz, error) {
    args := m.Called(id)
    return args.Get(0).(*Quiz), args.Error(1)
}

// Test
func TestCreateQuiz(t *testing.T) {
    // Arrange
    mockRepo := new(MockQuizRepository)
    service := NewQuizService(mockRepo)

    quiz := &Quiz{
        TestName: "test-quiz",
        Title:    "Test Quiz",
        UserID:   "user-id",
    }

    mockRepo.On("Create", quiz).Return(nil)

    // Act
    err := service.CreateQuiz(quiz)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

func TestCreateQuiz_DuplicateName(t *testing.T) {
    mockRepo := new(MockQuizRepository)
    service := NewQuizService(mockRepo)

    quiz := &Quiz{
        TestName: "existing-quiz",
        Title:    "Test Quiz",
        UserID:   "user-id",
    }

    mockRepo.On("Create", quiz).Return(ErrConflict)

    err := service.CreateQuiz(quiz)

    assert.Error(t, err)
    assert.Equal(t, ErrConflict, err)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

```go
// tests/integration/quiz_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/suite"
)

type QuizTestSuite struct {
    suite.Suite
    router *gin.Engine
    db     *gorm.DB
    token  string
}

func (suite *QuizTestSuite) SetupSuite() {
    // Setup test database
    suite.db = setupTestDB()

    // Setup router
    suite.router = setupRouter(suite.db)

    // Create test user and get token
    suite.token = createTestUserAndGetToken(suite.db)
}

func (suite *QuizTestSuite) TearDownSuite() {
    // Cleanup
    cleanupTestDB(suite.db)
}

func (suite *QuizTestSuite) SetupTest() {
    // Clean database before each test
    suite.db.Exec("TRUNCATE quizzes CASCADE")
}

func (suite *QuizTestSuite) TestCreateQuiz() {
    payload := map[string]interface{}{
        "test_name":           "test-quiz",
        "title":               "Test Quiz",
        "time_limit_minutes":  30,
    }

    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/api/v1/quizzes", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+suite.token)

    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)

    suite.Equal(201, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    suite.True(response["success"].(bool))
    suite.NotNil(response["data"])
}

func (suite *QuizTestSuite) TestGetQuiz() {
    // Create quiz
    quiz := createTestQuiz(suite.db, "user-id", "test-quiz")

    req := httptest.NewRequest("GET", "/api/v1/quizzes/"+quiz.ID, nil)
    req.Header.Set("Authorization", "Bearer "+suite.token)

    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)

    suite.Equal(200, w.Code)
}

func TestQuizTestSuite(t *testing.T) {
    suite.Run(t, new(QuizTestSuite))
}
```

### E2E Tests

```go
// tests/e2e/quiz_flow_test.go
package e2e

func TestCompleteQuizFlow(t *testing.T) {
    // 1. Signup
    signupResp := signup("test@example.com", "Password123!")
    assert.Equal(t, 201, signupResp.StatusCode)

    // 2. Verify email (mock)
    verifyEmail(signupResp.UserID, signupResp.Token)

    // 3. Login
    loginResp := login("test@example.com", "Password123!")
    token := loginResp.Token

    // 4. Create quiz
    quizResp := createQuiz(token, map[string]interface{}{
        "test_name": "e2e-test-quiz",
        "title":     "E2E Test Quiz",
    })
    quizID := quizResp.QuizID

    // 5. Add questions
    addQuestion(token, quizID, questionData)
    addQuestion(token, quizID, questionData)

    // 6. Publish quiz
    publishQuiz(token, quizID)

    // 7. Start attempt
    attemptResp := startAttempt(token, quizID)
    attemptID := attemptResp.AttemptID

    // 8. Submit answers
    submitAnswer(token, attemptID, question1ID, answer1)
    submitAnswer(token, attemptID, question2ID, answer2)

    // 9. Submit quiz
    resultResp := submitQuiz(token, attemptID)

    // 10. Verify results
    assert.Equal(t, 2, resultResp.TotalQuestions)
    assert.True(t, resultResp.ScorePercentage > 0)
}
```

### Test Database Setup

```go
import "github.com/ory/dockertest/v3"

func setupTestDB() *gorm.DB {
    pool, err := dockertest.NewPool("")
    if err != nil {
        log.Fatalf("Could not connect to docker: %s", err)
    }

    // Start PostgreSQL container
    resource, err := pool.Run("postgres", "15-alpine", []string{
        "POSTGRES_PASSWORD=test",
        "POSTGRES_DB=quizcraft_test",
    })
    if err != nil {
        log.Fatalf("Could not start resource: %s", err)
    }

    // Wait for database to be ready
    var db *gorm.DB
    if err = pool.Retry(func() error {
        dsn := fmt.Sprintf(
            "host=localhost port=%s user=postgres password=test dbname=quizcraft_test sslmode=disable",
            resource.GetPort("5432/tcp"),
        )

        var err error
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
        return err
    }); err != nil {
        log.Fatalf("Could not connect to database: %s", err)
    }

    // Run migrations
    db.AutoMigrate(&User{}, &Quiz{}, &Question{}, ...)

    return db
}
```

---

## Deployment & DevOps

### Docker Configuration

#### Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./main"]
```

#### docker-compose.yml (Development)
```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - PORT=8080
      - DATABASE_URL=postgres://postgres:postgres@db:5432/quizcraft?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - JWT_SECRET=${JWT_SECRET}
      - GOOGLE_AI_API_KEY=${GOOGLE_AI_API_KEY}
    depends_on:
      - db
      - redis
    volumes:
      - ./uploads:/root/uploads
    networks:
      - quizcraft-network

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=quizcraft
    ports:
      - "5432:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - quizcraft-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - quizcraft-network

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    networks:
      - quizcraft-network

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana
    networks:
      - quizcraft-network

volumes:
  postgres-data:
  redis-data:
  prometheus-data:
  grafana-data:

networks:
  quizcraft-network:
    driver: bridge
```

### Kubernetes Deployment

#### deployment.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quizcraft-api
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quizcraft-api
  template:
    metadata:
      labels:
        app: quizcraft-api
    spec:
      containers:
      - name: api
        image: quizcraft/api:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: quizcraft-secrets
              key: database-url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: quizcraft-secrets
              key: redis-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: quizcraft-secrets
              key: jwt-secret
        - name: GOOGLE_AI_API_KEY
          valueFrom:
            secretKeyRef:
              name: quizcraft-secrets
              key: google-ai-api-key
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: quizcraft-api
  namespace: production
spec:
  selector:
    app: quizcraft-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

### CI/CD (GitHub Actions)

```.github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: quizcraft_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Install dependencies
      run: go mod download

    - name: Run linter
      run: |
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        golangci-lint run

    - name: Run tests
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/quizcraft_test?sslmode=disable
        REDIS_URL: redis://localhost:6379/0
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v3

    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2

    - name: Login to DockerHub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKERHUB_USERNAME }}
        password: ${{ secrets.DOCKERHUB_TOKEN }}

    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: quizcraft/api:latest,quizcraft/api:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - name: Deploy to Kubernetes
      uses: azure/k8s-deploy@v4
      with:
        manifests: |
          k8s/deployment.yaml
          k8s/service.yaml
        images: |
          quizcraft/api:${{ github.sha }}
        kubeconfig: ${{ secrets.KUBE_CONFIG }}
```

---

## Performance Optimization

### Database Optimization

#### Indexing Strategy
```sql
-- Already covered in schema, but key indexes:

-- User lookups
CREATE INDEX CONCURRENTLY idx_users_email_active
ON users(email) WHERE deleted_at IS NULL;

-- Quiz queries
CREATE INDEX CONCURRENTLY idx_quizzes_user_status
ON quizzes(user_id, status) WHERE deleted_at IS NULL;

CREATE INDEX CONCURRENTLY idx_quizzes_public_published
ON quizzes(is_public, status) WHERE deleted_at IS NULL AND status = 'published';

-- Question lookups
CREATE INDEX CONCURRENTLY idx_questions_quiz_order
ON questions(quiz_id, order_index) WHERE deleted_at IS NULL;

-- Attempt queries
CREATE INDEX CONCURRENTLY idx_attempts_user_quiz
ON quiz_attempts(user_id, quiz_id, attempt_number);

CREATE INDEX CONCURRENTLY idx_attempts_submitted
ON quiz_attempts(submitted_at DESC) WHERE status = 'completed';

-- Answer lookups
CREATE INDEX CONCURRENTLY idx_answers_attempt
ON user_answers(attempt_id, question_id);

-- JSONB indexes
CREATE INDEX CONCURRENTLY idx_quizzes_meta_gin
ON quizzes USING GIN(meta jsonb_path_ops);

CREATE INDEX CONCURRENTLY idx_questions_options_gin
ON questions USING GIN(options jsonb_path_ops);

-- Text search (for quiz title/description search)
CREATE INDEX CONCURRENTLY idx_quizzes_title_trgm
ON quizzes USING gin(title gin_trgm_ops);
```

#### Query Optimization
```go
// Preload related data
db.Preload("Questions").Find(&quizzes)

// Select specific fields
db.Select("id", "title", "status").Find(&quizzes)

// Use joins instead of N+1 queries
db.Joins("LEFT JOIN questions ON quizzes.id = questions.quiz_id").
   Group("quizzes.id").
   Select("quizzes.*, COUNT(questions.id) as question_count").
   Find(&quizzes)

// Pagination with cursor (better performance than offset)
db.Where("created_at < ?", cursor).
   Order("created_at DESC").
   Limit(limit).
   Find(&quizzes)

// Use database aggregations instead of Go loops
db.Model(&QuizAttempt{}).
   Where("quiz_id = ? AND status = ?", quizID, "completed").
   Select("AVG(score_percentage) as avg_score, COUNT(*) as total").
   Scan(&stats)
```

### Caching Strategy

```go
// Cache hot data in Redis
func GetQuizByID(quizID string) (*Quiz, error) {
    cacheKey := fmt.Sprintf("quiz:%s", quizID)

    // Try cache first
    cached, err := redisClient.Get(ctx, cacheKey).Result()
    if err == nil {
        var quiz Quiz
        json.Unmarshal([]byte(cached), &quiz)
        return &quiz, nil
    }

    // Cache miss - fetch from database
    var quiz Quiz
    if err := db.First(&quiz, "id = ?", quizID).Error; err != nil {
        return nil, err
    }

    // Store in cache (1 hour TTL)
    quizJSON, _ := json.Marshal(quiz)
    redisClient.Set(ctx, cacheKey, quizJSON, 1*time.Hour)

    return &quiz, nil
}

// Cache invalidation on update
func UpdateQuiz(quiz *Quiz) error {
    if err := db.Save(quiz).Error; err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("quiz:%s", quiz.ID)
    redisClient.Del(ctx, cacheKey)

    return nil
}

// Cache quiz list with pagination
func GetQuizzes(userID string, page int, limit int) ([]Quiz, error) {
    cacheKey := fmt.Sprintf("quizzes:user:%s:page:%d:limit:%d", userID, page, limit)

    // Try cache
    cached, err := redisClient.Get(ctx, cacheKey).Result()
    if err == nil {
        var quizzes []Quiz
        json.Unmarshal([]byte(cached), &quizzes)
        return quizzes, nil
    }

    // Fetch from database
    var quizzes []Quiz
    offset := (page - 1) * limit
    db.Where("user_id = ?", userID).
       Offset(offset).
       Limit(limit).
       Find(&quizzes)

    // Cache for 5 minutes
    quizzesJSON, _ := json.Marshal(quizzes)
    redisClient.Set(ctx, cacheKey, quizzesJSON, 5*time.Minute)

    return quizzes, nil
}
```

### Connection Pooling

```go
// PostgreSQL connection pool
sqlDB, _ := db.DB()

// Set connection pool settings
sqlDB.SetMaxIdleConns(10)           // Idle connections
sqlDB.SetMaxOpenConns(100)          // Max open connections
sqlDB.SetConnMaxLifetime(time.Hour) // Max connection lifetime

// Redis connection pool
redisClient := redis.NewClient(&redis.Options{
    Addr:         os.Getenv("REDIS_URL"),
    PoolSize:     100,
    MinIdleConns: 10,
    MaxRetries:   3,
})
```

### Response Compression

```go
import "github.com/gin-contrib/gzip"

// Enable gzip compression
router.Use(gzip.Gzip(gzip.DefaultCompression))
```

### API Response Optimization

```go
// Paginated responses
type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
    CurrentPage int  `json:"current_page"`
    TotalPages  int  `json:"total_pages"`
    TotalCount  int  `json:"total_count"`
    Limit       int  `json:"limit"`
    HasNext     bool `json:"has_next"`
    HasPrev     bool `json:"has_prev"`
}

// Lazy loading for nested resources
// Don't include full nested data by default
// Provide separate endpoints or query params to expand
type QuizListResponse struct {
    ID            string `json:"id"`
    Title         string `json:"title"`
    QuestionCount int    `json:"question_count"` // Count only
    // Don't include full Questions array
}

// Use query param to expand
// GET /quizzes?expand=questions
if c.Query("expand") == "questions" {
    db.Preload("Questions").Find(&quizzes)
} else {
    db.Find(&quizzes)
}
```

---

## Project Structure

```
quiz-craft/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration loader
│   ├── models/
│   │   ├── user.go
│   │   ├── quiz.go
│   │   ├── question.go
│   │   ├── attempt.go
│   │   └── audit.go
│   ├── repositories/
│   │   ├── user_repository.go
│   │   ├── quiz_repository.go
│   │   ├── question_repository.go
│   │   └── attempt_repository.go
│   ├── services/
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── quiz_service.go
│   │   ├── question_service.go
│   │   ├── attempt_service.go
│   │   ├── ai_service.go
│   │   ├── email_service.go
│   │   └── file_service.go
│   ├── handlers/
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── quiz_handler.go
│   │   ├── question_handler.go
│   │   ├── attempt_handler.go
│   │   └── health_handler.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── rate_limit.go
│   │   ├── logger.go
│   │   ├── error_handler.go
│   │   ├── security.go
│   │   └── metrics.go
│   ├── utils/
│   │   ├── validator.go
│   │   ├── jwt.go
│   │   ├── password.go
│   │   ├── random.go
│   │   └── sanitizer.go
│   └── workers/
│       ├── file_processor.go
│       └── email_sender.go
├── pkg/
│   ├── database/
│   │   ├── postgres.go
│   │   └── migrations/
│   │       ├── 001_create_users.sql
│   │       ├── 002_create_quizzes.sql
│   │       └── ...
│   ├── cache/
│   │   └── redis.go
│   └── logger/
│       └── logger.go
├── api/
│   └── routes/
│       └── router.go              # Route definitions
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── scripts/
│   ├── migrate.sh
│   └── seed.sh
├── docs/
│   ├── api/
│   │   └── openapi.yaml
│   └── architecture.md
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── ingress.yaml
├── .github/
│   └── workflows/
│       └── ci-cd.yml
├── config/
│   ├── config.yaml
│   └── config.example.yaml
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Makefile

```makefile
.PHONY: help build run test clean migrate-up migrate-down docker-build docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o bin/server cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

migrate-up: ## Run database migrations
	migrate -path pkg/database/migrations -database "${DATABASE_URL}" up

migrate-down: ## Rollback database migrations
	migrate -path pkg/database/migrations -database "${DATABASE_URL}" down

migrate-create: ## Create new migration (usage: make migrate-create name=create_users)
	migrate create -ext sql -dir pkg/database/migrations -seq $(name)

docker-build: ## Build Docker image
	docker build -t quizcraft/api:latest .

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

seed: ## Seed database with test data
	go run scripts/seed/main.go

.DEFAULT_GOAL := help
```

---

## Summary

This backend approach document provides a comprehensive, production-ready architecture for the quiz-craft application using:

- **Go + Gin** for high-performance API
- **PostgreSQL** with JSONB for flexible yet structured data
- **Redis** for caching and session management
- **Google Gemini AI** for intelligent question extraction
- **JWT** authentication with secure session management
- **Comprehensive security** measures (rate limiting, input validation, XSS prevention)
- **Observability** with structured logging, metrics, and APM
- **Scalable architecture** with clean separation of concerns
- **Production-ready** deployment with Docker and Kubernetes
- **Full CI/CD pipeline** with automated testing and deployment

The system is designed to handle production traffic with proper error handling, monitoring, and performance optimizations while maintaining code quality and security best practices.
