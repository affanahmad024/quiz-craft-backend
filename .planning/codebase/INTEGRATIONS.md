# External Integrations

**Analysis Date:** 2026-03-29

## Databases

**PostgreSQL (primary datastore):**
- Version: PostgreSQL 15 (specified in `docker-compose.prod.yml`)
- Driver: `github.com/lib/pq v1.10.9` (raw `database/sql`) used by all DAO layers
- ORM: `gorm.io/gorm v1.31.1` used only for connection initialization and health check in `pkg/database/postgres.go`
- Pattern: GORM opens the connection, then `GetSQLDB()` extracts the underlying `*sql.DB` which all DAOs use for raw SQL queries
- Connection pool: `MaxOpenConns=25`, `MaxIdleConns=5`, `ConnMaxLifetime=5m` (configurable via env)
- Connection string built by `DatabaseConfig.GetDSN()` in `internal/config/config.go`
- Connection env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- Migrations: `github.com/golang-migrate/migrate/v4` runs 19 sequential `.sql` files from `pkg/database/migrations/` on startup

**Schema tables (from migrations):**
- `users`, `user_passwords` — auth and identity
- `quizzes`, `questions`, `answers` — quiz content
- `quiz_attempts`, `user_answers` — attempt tracking
- `uploaded_files`, `generation_jobs` — AI file pipeline
- `notifications` — in-app notifications
- `live_sessions` (and related) — real-time quiz sessions
- `educator_profiles`, `subscriptions` — educator role
- `activity_logs` — admin audit trail

**Redis (optional/planned):**
- Listed in `docker-compose.prod.yml` (redis:7-alpine) and `.env.production.example`
- Env vars: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`
- No Redis client library present in `go.mod` — not yet integrated in application code
- Intended for caching and rate limiting per production example config

**File Storage (local filesystem):**
- Uploaded files stored under `./uploads/{userID}/{filename}` on local disk
- Volume mounted as `file_uploads:/app/uploads` in Docker
- No cloud object storage client (S3, GCS) integrated in application code
- `.env.production.example` references `BACKUP_S3_BUCKET` / `BACKUP_S3_ACCESS_KEY` for backups — not implemented

## External APIs

**Google Gemini AI (google/generative-ai-go):**
- SDK: `github.com/google/generative-ai-go v0.20.1`
- Client initialized per-request in `ai-agent/core/ai_service.go` via `genai.NewClient(ctx, option.WithAPIKey(s.APIKey))`
- Default model: `gemini-1.5-pro` (constant `DefaultModel` in `ai-agent/core/ai_service.go`)
- Used for:
  - Text-to-quiz generation (`GenerateQuizFromText`)
  - Image/PDF quiz extraction via Vision API (`ExtractQuizFromImage`, `callGeminiVisionAPI`)
  - Multi-image quiz generation (`callGeminiVisionAPIWithMultipleImages`)
- Config: `GOOGLE_AI_API_KEY` env var → `GoogleAIConfig.APIKey` → passed to `AIService`
- Timeout: 60 seconds per request (`DefaultTimeout` constant)
- Retry: 1 retry with 2-second delay on failure (`MaxRetries = 1`)
- Generation config: Temperature=0.7, TopP=0.95, TopK=40, MaxOutputTokens=8000

**Mailjet Email API:**
- SDK: `github.com/mailjet/mailjet-apiv3-go/v4 v4.0.8`
- Client initialized once at startup in `internal/services/email_service.go`
- Sends transactional emails via HTTPS API (avoids SMTP port 25/587 blocks on cloud providers)
- Config env vars: `MAILJET_API_KEY`, `MAILJET_SECRET_KEY`, `SMTP_FROM_EMAIL`, `SMTP_FROM_NAME`, `FRONTEND_URL`
- Emails sent:
  - Email verification (`SendVerificationEmail`) — link: `{FRONTEND_URL}/verify-email?token=...&user_id=...`
  - Password reset (`SendPasswordResetEmail`) — link: `{FRONTEND_URL}/reset-password?token=...&user_id=...`
  - Welcome after verification (`SendWelcomeEmail`)
  - Password changed notification (`SendPasswordChangedEmail`)
- All email bodies are rendered from inline Go `html/template` strings; no external template service

## Authentication & Authorization

**JWT (self-issued, no external provider):**
- Library: `github.com/golang-jwt/jwt/v5 v5.3.0`
- Implementation: `internal/utils/jwt.go` — `JWTManager` struct
- Algorithm: HMAC-SHA256 (HS256)
- Token types: `access` (24h default) and `refresh` (168h / 7 days default)
- Claims: `UserID` (UUID), `Email`, `Type`, standard JWT registered claims
- Secret: `JWT_SECRET` env var (required, no default)
- Middleware: `internal/middleware/auth.go` — validates `Authorization: Bearer <token>` header
- Role-based access: `internal/middleware/admin.go` and `internal/middleware/educator.go` check role claim after JWT validation

**Password hashing:**
- Library: `golang.org/x/crypto/bcrypt`
- Cost: 12 (`DefaultCost` in `internal/utils/password.go`)
- Strength requirements enforced: min 8 chars, max 128, must contain uppercase, lowercase, digit, special char

**No OAuth/SSO/third-party identity provider** is integrated (no Google OAuth, GitHub, Auth0, etc.)

## Message Queues / Event Streams

No external message broker (Kafka, RabbitMQ, SQS, etc.) is used. Concurrency is handled via internal Go goroutines and channels.

**Background workers (in-process goroutines):**
- `ai-agent/worker/quiz_generator_worker.go` — polls pending AI generation jobs in PostgreSQL and processes them
- `quiz/worker/attempt_worker.go` — auto-submits timed-out quiz attempts
- `notification/worker/cleanup_worker.go` — cleans up old notifications
- `livesession/worker/` — manages live session auto-start, question progression, and end

**WebSocket hubs (in-process):**
- `notification/websocket/` — `Hub` struct manages connected clients for real-time notifications; served at `/api/v1/notifications/ws`
- `livesession/websocket/` — `Hub` struct manages live session participants; served at `/api/v1/live-sessions/{id}/ws`
- Both use `github.com/gorilla/websocket v1.5.3`

## Cloud Services

**Hosting / Deployment:**
- Docker-based deployment defined in `docker-compose.prod.yml`
- Reverse proxy: Nginx 1.25-alpine (SSL termination, port 80/443)
- No cloud-provider-specific SDK is integrated (no AWS SDK, GCP client beyond Gemini, Azure SDK)
- Platform-agnostic: supports `PORT` env var (Heroku/Render/Railway convention handled in `internal/config/config.go`)

**Object Storage:**
- Not integrated; files stored on local disk under `uploads/`
- `.env.production.example` documents optional S3 backup vars (`BACKUP_S3_BUCKET`, `BACKUP_S3_REGION`, `BACKUP_S3_ACCESS_KEY`, `BACKUP_S3_SECRET_KEY`) — no implementation present

## Other External Services

**Error Tracking:**
- Sentry DSN referenced in `.env.production.example` as `SENTRY_DSN=` (empty)
- No Sentry SDK in `go.mod` — not implemented

**Logging (structured, stdout only):**
- `github.com/sirupsen/logrus v1.9.3`
- Dev: colored text to stdout
- Production (`APP_ENV=production`): JSON lines to stdout (suitable for log aggregation services like Datadog, Cloudwatch, etc.)
- No remote log shipping configured

**Monitoring / APM:**
- No APM SDK integrated (no Datadog, New Relic, Prometheus exporter, etc.)
- Health check endpoint: `GET /api/v1/health` — checks DB ping, returns JSON status

**Analytics:**
- None integrated

---

*Integration audit: 2026-03-29*
