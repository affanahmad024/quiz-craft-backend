# Architecture

**Analysis Date:** 2026-03-29

## Overall Pattern

**Monolithic service** with domain-driven internal module organization. The application follows a layered architecture (API Handler → Core/Business Logic → DAO → Database) with no microservice separation. All domains (auth, quiz, user, AI, live session, notification, educator, admin) run in a single Go binary.

**Key Characteristics:**
- Single Go binary serving two HTTP servers (external :8080, internal :8081)
- Domain modules each own their `core/`, `dao/`, and `model/` sub-packages
- Background goroutine workers for async processing
- WebSocket hubs running as in-process goroutines for real-time features
- No message broker or external queue; polling-based job queue stored in PostgreSQL

## Core Modules & Responsibilities

| Module | Path | Responsibility |
|---|---|---|
| `api` | `api/` | HTTP handler layer — binds requests, calls core, returns responses |
| `auth/core` | `auth/core/auth.go` | Authentication business logic (signup, login, JWT, password flows) |
| `user/core` | `user/core/user.go` | User profile management |
| `user/dao` | `user/dao/` | Raw SQL queries for users and passwords |
| `quiz/core` | `quiz/core/` | Quiz CRUD, attempts, search, statistics, templates |
| `quiz/dao` | `quiz/dao/` | Raw SQL for quizzes, questions, answers, attempts, statistics |
| `quiz/worker` | `quiz/worker/attempt_worker.go` | Background goroutine auto-submitting timed-out attempts |
| `ai-agent/core` | `ai-agent/core/ai_service.go` | Google Gemini AI client wrapper |
| `ai-agent/services` | `ai-agent/services/` | File parsing (PDF, DOCX, images), quiz generation, rate limiting |
| `ai-agent/worker` | `ai-agent/worker/quiz_generator_worker.go` | Polls DB for pending AI jobs, processes up to 5 concurrently |
| `ai-agent/dao` | `ai-agent/dao/` | Persists AI generation jobs and uploaded files |
| `livesession/core` | `livesession/core/session.go` | Live quiz session lifecycle, scoring, leaderboard |
| `livesession/dao` | `livesession/dao/` | Raw SQL for sessions, participants, answers, invitations |
| `livesession/websocket` | `livesession/websocket/` | WebSocket hub broadcasting session events to participants |
| `livesession/worker` | `livesession/worker/session_worker.go` | Background goroutine for auto-starting scheduled sessions and advancing questions |
| `notification/core` | `notification/core/notification.go` | Notification creation and real-time delivery |
| `notification/websocket` | `notification/websocket/` | WebSocket hub broadcasting per-user notifications |
| `notification/worker` | `notification/worker/cleanup_worker.go` | Periodic cleanup of old notifications |
| `educator/core` | `educator/core/` | Educator profile management, follower/subscription logic |
| `admin/core` | `admin/core/` | Admin user management, quiz moderation, platform statistics, activity logging |
| `internal/middleware` | `internal/middleware/` | Gin middleware: Auth (JWT), RequireAdmin, RequireEducator, CORS, error handler |
| `internal/utils` | `internal/utils/` | JWT manager, password hashing, token generation, validation |
| `internal/config` | `internal/config/config.go` | Centralized config loading via Viper (env vars + optional YAML) |
| `commons` | `commons/` | Shared context helpers, structured error types, standardized HTTP responses |
| `pkg/database` | `pkg/database/` | GORM + raw sql.DB initialization, connection pooling, migration runner |
| `pkg/logger` | `pkg/logger/` | Logrus-based logger initialization |
| `constants` | `constants/` | Application-wide constants (token expiry, lockout durations, etc.) |

## Data Flow

### Standard REST Request Flow

```
HTTP Request
    → Gin Router (main.go)
    → middleware.Auth (JWT validation, sets user_id in context)
    → api/{domain}.go handler
        → commons.Ctx(c)          — extract trace ID + user ID into context
        → ShouldBindJSON          — bind and validate request body
        → {domain}/core/*.go      — business logic
            → {domain}/dao/*.go   — raw SQL via *sql.DB
                → PostgreSQL
        → commons.RespOKWithMsg   — standardized JSON response
```

### AI Quiz Generation Flow (Async)

```
POST /api/v1/ai/generate-quiz-from-*
    → api/ai.go handler
    → Creates GenerationJob record in DB (status=pending)
    → Returns job_id immediately (202 Accepted pattern)

[Background: ai-agent/worker polls every 5s]
    → dao.GetPendingJobs
    → goroutine per job (max 5 concurrent via semaphore)
        → ai-agent/services.FileParserService (PDF/DOCX/image → text)
        → ai-agent/core.AIService (text → Gemini API → quiz JSON)
        → ai-agent/services.QuizGeneratorService (create quiz in DB)
        → dao.CompleteJob (status=completed, quiz_id=...)

GET /api/v1/ai/generation-status/:id
    → returns current job status and progress (0–100%)
```

### Live Session WebSocket Flow

```
POST /api/v1/live-sessions          — create session
POST /api/v1/live-sessions/:id/join — join session
GET  /api/v1/live-sessions/:id/ws   — upgrade to WebSocket

WebSocket connected:
    → livesession/websocket/handler.go registers Client with Hub
    → Hub.Run() goroutine manages channel routing per session

Events broadcast by worker or core:
    session_started | question_revealed | answer_submitted
    leaderboard_update | question_ended | session_ended
    timer_update | whole_quiz_timer_update | question_auto_advance

POST /api/v1/live-sessions/:id/answer
    → core.SubmitAnswer → calculates speed-bonus points → saves to DB
    → Hub.BroadcastToSession(sessionID, leaderboard_update event)
```

### Notification WebSocket Flow

```
GET /api/v1/notifications/ws
    → Upgrade to WebSocket
    → notification/websocket/Hub.BroadcastToUser per user ID

Triggered by:
    - Live session invitations (livesession/core → notification/core)
    - Other in-process notification calls
```

**State Management:**
- No in-memory session/application state beyond the two WebSocket Hubs
- All durable state stored in PostgreSQL
- JWT tokens are stateless (no server-side token revocation store)
- WebSocket Hubs hold in-process connection state (non-persistent across restarts)

## API Design

**Protocol:** REST over HTTP/1.1 (WebSocket upgrade for real-time endpoints)

**Versioning:** URL path prefix — all routes under `/api/v1/`

**Authentication:**
- Bearer token via `Authorization: Bearer <token>` header, OR `access_token` HttpOnly cookie
- Refresh token accepted as JSON body `{"refresh_token": "..."}` or `refresh_token` cookie
- `middleware.Auth` enforces auth; `middleware.RequireAdmin` and `middleware.RequireEducator` add role checks (DB lookup per request)

**Response Envelope:**
```json
{
  "success": true,
  "message": "Human-readable message",
  "data": { ... }
}
```
Error responses replace `data` with `"error": "<machine-readable-code>"`.

**Error Codes:** Machine-readable kebab-case strings (e.g., `"email-already-exists"`, `"invalid-verification-token"`, `"account-locked"`). HTTP status codes: 400, 401, 403, 404, 409, 500.

**Tracing:** Every request tagged with `X-Trace-ID` header (UUID), propagated through `context.Context` and echoed in all log lines.

**Route Groups (main.go `initExt`):**
- `POST /api/v1/auth/*` — public auth (signup, login, verify, reset)
- `GET/PATCH/DELETE /api/v1/users/*` — protected user management
- `*/api/v1/quizzes/*` — quiz CRUD, questions, answers, attempts
- `POST /api/v1/ai/*` — AI generation and job status
- `POST /api/v1/files/*` — file upload and serving
- `*/api/v1/notifications/*` — notifications + WebSocket
- `*/api/v1/live-sessions/*` — live sessions + WebSocket
- `*/api/v1/educators/*` and `/api/v1/educator/*` — educator profiles and discovery
- `*/api/v1/admin/*` — admin-only management (RequireAdmin middleware)
- `GET /api/v1/quizzes/search`, `/categories`, `/tags`, `/share/:code` — public discovery

**Two Servers:**
- External `:8080` — all public and authenticated routes
- Internal `:8081` — service-to-service routes (currently only health check at `/api/v1/health`)

## Key Design Decisions

1. **Raw SQL over full ORM in DAO layer.** GORM is used only for initialization and migrations. All DAO functions use `database/sql` directly with `db.QueryRowContext` / `db.QueryContext`. This gives full SQL control at the cost of more boilerplate.

2. **Domain-per-package, not package-per-layer.** Each domain (`auth`, `quiz`, `livesession`, etc.) owns its own `core/`, `dao/`, and `model/` sub-packages co-located under a top-level domain directory. Cross-domain calls happen via direct Go function imports (e.g., `livesession/core` imports `quiz/dao` and `notification/core`).

3. **Dependency injection via `Init()` functions.** Each DAO package exposes an `Init(db *sql.DB)` function called in `main.go` at startup rather than passing DB references through constructors everywhere. Same pattern for `authcore.Init(jwtManager, emailService)`.

4. **Typed domain errors with machine-readable codes.** `commons.DomainError` wraps sentinel errors (`ErrNotFound`, `ErrBadInput`, etc.) with a `Code string`. Handlers check error type with `commons.IsNotFound(err)` etc. and map to HTTP status codes.

5. **Polling-based async job queue.** AI generation jobs are stored in PostgreSQL and polled every 5 seconds by a goroutine worker with a semaphore-capped concurrency of 5. No Redis or external queue is used.

6. **In-process WebSocket hubs.** Both notification and live-session WebSocket hubs are goroutines inside the same process, using channel-based message routing. This is simple but non-scalable beyond a single instance.

7. **Versioned SQL migrations.** Database schema is managed with `golang-migrate` using numbered `*.up.sql`/`*.down.sql` files in `pkg/database/migrations/`. Migrations run automatically on startup.

8. **Password security features.** Bcrypt hashing, password history (last N tracked in JSONB), failed login attempt counting, account lockout with configurable duration, and email-verified-before-login enforcement — all in `auth/core`.

## State Management

**HTTP Request State:**
- Trace ID generated per request via `commons.Ctx(c)` from `X-Trace-ID` header or new UUID
- User ID extracted from JWT claims and stored in `gin.Context` by `middleware.Auth`
- Accessed in handlers via `commons.UsrID(ctx)`

**Session / Authentication State:**
- Stateless JWT access tokens (configurable expiry, default 24h)
- Stateless JWT refresh tokens (default 7 days); no server-side revocation
- Access and refresh tokens set as HttpOnly cookies on login and cleared on logout

**Application State (in-process):**
- `notificationws.Hub` — map of userID → active WebSocket connections
- `livesessionws.Hub` — map of sessionID → active WebSocket connections
- Both hubs are non-persistent; connections are lost on process restart

**Database State:**
- PostgreSQL is the sole persistent store
- AI generation job queue stored in `generation_jobs` table
- Live session state (status, current question index, scores) stored in `live_sessions` and `live_session_participants` tables
- Notification read/unread state persisted in `notifications` table

---

*Architecture analysis: 2026-03-29*
