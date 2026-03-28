# Directory Structure

**Analysis Date:** 2026-03-29

## Top-Level Layout

```
backend/
├── main.go                      # Single entry point; wires all modules, starts two HTTP servers
├── go.mod                       # Module: github.com/affanahmad024/quiz-craft/backend (Go 1.24.1)
├── go.sum
├── Dockerfile
├── docker-compose.prod.yml
├── .env                         # Local secrets (not committed)
├── .env.example
├── .env.production.example
├── CHANGELOG.md
├── README.md
│
├── api/                         # HTTP handler layer (one file per domain)
│   ├── auth.go
│   ├── user.go
│   ├── quiz.go
│   ├── attempt.go
│   ├── ai.go
│   ├── notification.go
│   ├── livesession.go
│   ├── educator.go
│   ├── admin.go
│   ├── search.go
│   └── template.go
│
├── auth/                        # Auth domain (no dao/ or model/ — reuses user/)
│   └── core/
│       └── auth.go              # Signup, login, JWT token management, password flows
│
├── user/                        # User domain
│   ├── core/
│   │   └── user.go              # Profile CRUD, user lookup
│   ├── dao/
│   │   ├── user.go              # Raw SQL: find, create, update, delete users
│   │   └── password.go          # Raw SQL: passwords, tokens, lockout, history
│   └── model/
│       ├── user.go              # UserDAO, UserAPI, request/response types
│       └── password.go          # PasswordDAO, PasswordHistory, LoginMeta
│
├── quiz/                        # Quiz domain
│   ├── core/
│   │   ├── quiz.go              # Quiz CRUD, publish/unpublish, share links
│   │   ├── attempt.go           # Attempt lifecycle, answer submission, scoring
│   │   ├── search.go            # Full-text search, category/tag queries
│   │   ├── statistics.go        # Quiz and user statistics aggregation
│   │   └── template.go          # Template save and instantiation
│   ├── dao/
│   │   ├── quiz.go
│   │   ├── question.go
│   │   ├── answer.go
│   │   ├── attempt.go
│   │   └── statistics.go
│   ├── model/
│   │   ├── quiz.go
│   │   ├── question.go
│   │   ├── answer.go
│   │   ├── attempt.go
│   │   └── statistics.go
│   └── worker/
│       └── attempt_worker.go    # Polls for timed-out attempts and auto-submits them
│
├── ai-agent/                    # AI quiz generation domain
│   ├── core/
│   │   └── ai_service.go        # Google Gemini API client (GenerateQuizFromText, ExtractQuizFromImage)
│   ├── services/
│   │   ├── file_parser.go       # PDF, DOCX, plain text → string content
│   │   ├── file_upload.go       # Multipart file handling and local storage
│   │   ├── image_processor.go   # Image preparation for Gemini Vision
│   │   ├── quiz_generator.go    # AI response → DB quiz/question/answer creation
│   │   ├── rate_limiter.go      # Per-user rate limiting for AI endpoints
│   │   └── file_cleanup.go      # Removes orphaned uploaded files
│   ├── dao/
│   │   ├── job.go               # GenerationJob CRUD (pending/processing/completed/failed)
│   │   └── file.go              # UploadedFile CRUD
│   ├── model/
│   │   ├── job.go
│   │   └── file.go
│   └── worker/
│       └── quiz_generator_worker.go  # Polls DB every 5s, runs up to 5 concurrent jobs
│
├── livesession/                 # Live quiz session domain
│   ├── core/
│   │   └── session.go           # Session lifecycle, join, start, advance, end, scoring
│   ├── dao/
│   │   └── (session, participant, answer, invitation SQL)
│   ├── model/
│   │   └── (LiveSessionDAO/API, ParticipantDAO/API, etc.)
│   ├── websocket/
│   │   ├── hub.go               # Channel-based broadcast hub keyed by session UUID
│   │   ├── client.go            # Per-connection WebSocket client goroutines
│   │   └── handler.go           # Gin handler that upgrades HTTP → WebSocket
│   └── worker/
│       └── session_worker.go    # Auto-starts scheduled sessions, advances timed questions
│
├── notification/                # Notification domain
│   ├── core/
│   │   └── notification.go      # Create, bulk-create, mark-read, real-time dispatch
│   ├── dao/
│   ├── model/
│   ├── websocket/
│   │   ├── hub.go               # Per-user broadcast hub
│   │   ├── client.go
│   │   └── handler.go
│   └── worker/
│       └── cleanup_worker.go    # Periodically deletes old/read notifications
│
├── educator/                    # Educator profile and follower domain
│   ├── core/
│   │   ├── educator.go
│   │   ├── profile.go
│   │   └── subscription.go
│   ├── dao/
│   └── model/
│
├── admin/                       # Admin dashboard domain
│   ├── core/
│   │   ├── user_management.go
│   │   ├── quiz_moderation.go
│   │   ├── statistics.go
│   │   └── activity_logger.go
│   ├── dao/
│   └── model/
│
├── internal/                    # Private infrastructure shared across domains
│   ├── config/
│   │   └── config.go            # Viper-based config struct + Load() + env binding
│   ├── middleware/
│   │   ├── auth.go              # JWT Bearer/cookie auth middleware + OptionalAuth
│   │   ├── admin.go             # RequireAdmin middleware (DB role check)
│   │   ├── educator.go          # RequireEducator middleware
│   │   ├── cors.go              # gin-contrib/cors configuration
│   │   ├── logger.go            # Request logging middleware
│   │   └── error_handler.go
│   ├── services/
│   │   └── (email service via Mailjet)
│   ├── utils/
│   │   ├── jwt.go               # JWTManager: generate/validate access+refresh tokens
│   │   ├── password.go          # HashPassword, ComparePassword, ValidatePasswordStrength
│   │   ├── validator.go
│   │   └── migrate.go
│   ├── models/                  # (legacy/unused — domain models live in domain/model/)
│   └── repositories/            # (legacy/unused — queries live in domain/dao/)
│
├── commons/                     # Cross-cutting utilities (imported by all layers)
│   ├── ctx.go                   # Ctx(), Tid(), UsrID(), SetUserID() — trace + user context
│   ├── errors.go                # DomainError, sentinel errors, IsNotFound/IsBadInput/etc.
│   ├── response.go              # RespOK, RespBadReq, Resp401, Resp404, RespISE, etc.
│   └── log.go                   # Log, LogW, LogE, LogC helpers (structured with trace ID)
│
├── constants/                   # Application-wide constants
│   └── (token expiry, lockout durations, password history count, etc.)
│
├── pkg/                         # Reusable infrastructure packages
│   ├── database/
│   │   ├── postgres.go          # GORM init, connection pool, GetSQLDB(), Close()
│   │   ├── migrate.go           # golang-migrate runner
│   │   └── migrations/          # 19 numbered *.up.sql / *.down.sql files
│   ├── cache/                   # (directory exists, not yet implemented)
│   └── logger/
│       └── (logrus logger init and GetLogger())
│
├── config/                      # Optional YAML config files
├── deployments/
│   ├── docker/
│   └── k8s/
├── nginx/
│   ├── conf.d/
│   └── ssl/
├── docs/                        # API documentation
├── uploads/                     # Local file upload storage (per-user UUID subdirs)
├── bin/                         # Build output
└── .planning/
    └── codebase/                # GSD planning documents
```

## Module Organization

Each business domain follows an identical three-layer internal structure:

```
{domain}/
├── core/       # Business logic functions — no HTTP, no SQL
├── dao/        # Raw SQL database access (uses package-level var db *sql.DB)
└── model/      # Data types: DAO structs (sql.Null*), API structs (JSON tags), request/response types
```

Domains with async or real-time features add:
```
{domain}/
├── websocket/  # hub.go, client.go, handler.go
└── worker/     # polling goroutine started in main.go
```

The `api/` package is a flat collection of Gin handler functions grouped by domain filename. It is the only layer that imports `github.com/gin-gonic/gin` directly.

## Entry Points

**`main.go`** — sole binary entry point.

Startup sequence:
1. `config.Load()` — load config from `.env` + env vars via Viper
2. `logger.Init(cfg.App.LogLevel)` — configure Logrus
3. `database.InitPostgres(&cfg.Database)` — open GORM connection, set pool
4. `database.RunMigrations()` — apply pending SQL migrations
5. `utils.NewJWTManager(...)` — instantiate JWT manager
6. `{domain}dao.Init(sqlDB)` — inject `*sql.DB` into each DAO package
7. `services.NewEmailService(&cfg.Email)` + `authcore.Init(jwtManager, emailService)` — wire auth dependencies
8. `go aiWorker.Start(ctx)` — start AI job worker goroutine
9. `go attemptWorker.Start(ctx)` — start attempt auto-submit worker goroutine
10. `wsHub = notificationws.NewHub(); go wsHub.Run()` — start notification WebSocket hub
11. `api.InitNotificationService(wsHub)` — inject hub into API layer
12. `go notificationCleanupWorker.Start(ctx)` — start notification cleanup worker
13. `liveSessionWsHub = livesessionws.NewHub(); go liveSessionWsHub.Run()` — start live session WebSocket hub
14. `go sessionWorker.Start(ctx)` — start session progression worker
15. `go initInt()` — start internal server on `:8081` (background)
16. `initExt()` — start external server on `:8080` (blocks, handles shutdown signal)

## Key Files

| File | Role |
|---|---|
| `main.go` | Application bootstrap, route registration, graceful shutdown |
| `internal/config/config.go` | Complete configuration schema and environment variable mapping |
| `commons/ctx.go` | Trace ID and user ID propagation — imported by every handler and core function |
| `commons/errors.go` | Typed domain errors — defines error classification used by all layers |
| `commons/response.go` | Standardized JSON response helpers — used by all API handlers |
| `internal/middleware/auth.go` | JWT authentication — Bearer header and cookie support |
| `internal/middleware/admin.go` | Role enforcement — DB lookup per request to verify admin role |
| `internal/utils/jwt.go` | JWT token generation and validation |
| `pkg/database/postgres.go` | Database initialization and `GetSQLDB()` for DAO injection |
| `pkg/database/migrations/` | 19 migration files defining complete schema (users, quizzes, attempts, jobs, notifications, live_sessions, educators) |
| `auth/core/auth.go` | Complete auth business logic (signup through password reset) |
| `livesession/core/session.go` | Live session lifecycle including speed-bonus point calculation |
| `livesession/websocket/hub.go` | Session-keyed WebSocket broadcast hub with event type constants |
| `ai-agent/worker/quiz_generator_worker.go` | Async AI job processing with concurrency semaphore |
| `notification/core/notification.go` | Notification service with WebSocket real-time delivery interface |

## Naming Conventions

**Files:**
- One domain concept per file, snake_case: `user.go`, `attempt_worker.go`, `file_parser.go`
- API handlers grouped by domain: `api/quiz.go`, `api/livesession.go`

**Packages:**
- Domain directories use kebab-case: `ai-agent/`, `livesession/`
- Go package names use no separator: `package websocket`, `package worker`, `package core`
- Import aliases used in `main.go` to avoid collisions: `livesessiondao`, `notificationws`, `aiworker`

**Types:**
- DAO types (database row structs) suffixed `DAO`: `UserDAO`, `LiveSessionDAO`
- API types (JSON response structs) suffixed `API`: `UserAPI`, `LiveSessionAPI`
- Request types suffixed `Request`: `SignupRequest`, `CreateLiveSessionRequest`
- Response types suffixed `Response`: `LoginResponse`, `SubmitAnswerResponse`

**Functions:**
- DAO functions: verb + noun, e.g. `CreateUser`, `FindByID`, `UpdateEmailVerified`
- Core functions: verb + noun, e.g. `Signup`, `CreateSession`, `SubmitAnswer`
- API handlers: PascalCase matching route intent, e.g. `Signup`, `GetQuiz`, `StartLiveSession`

## Where to Add New Code

**New domain feature (e.g., new quiz type):**
- Business logic: `quiz/core/quiz.go` or new file `quiz/core/{feature}.go`
- Database queries: `quiz/dao/{entity}.go`
- Data types: `quiz/model/{entity}.go`
- HTTP handler: `api/quiz.go` (add handler function)
- Route registration: `main.go` → `initExt()` under appropriate group

**New domain module (e.g., `payment/`):**
- Create `payment/core/`, `payment/dao/`, `payment/model/`
- Add `paymentdao.Init(sqlDB)` in `main.go`
- Add handler file `api/payment.go`
- Add migration to `pkg/database/migrations/`
- Register routes in `main.go initExt()`

**New background worker:**
- Create `{domain}/worker/{name}_worker.go` with `Start(ctx context.Context)` method
- Instantiate and launch with `go worker.Start(context.Background())` in `main.go`

**New WebSocket feature:**
- Follow pattern in `livesession/websocket/` or `notification/websocket/`
- Create `hub.go`, `client.go`, `handler.go`
- Wire hub into `main.go` startup sequence
- Inject hub into API layer via `api.Init{Feature}Service(hub)` pattern

**Shared utilities:**
- Cross-domain helpers: `commons/`
- Infrastructure (DB, cache, logger): `pkg/`
- Application config fields: `internal/config/config.go`
- Constants: `constants/`

**New migration:**
- Add `pkg/database/migrations/0000{N}_{description}.up.sql` and `*.down.sql`
- Increment N from current highest (19)
- Migrations run automatically on next startup

## Special Directories

**`uploads/`:**
- Purpose: Local disk storage for user-uploaded files (PDF, DOCX, images)
- Organized as `uploads/{user-uuid}/{filename}`
- Not committed to git (should be in `.gitignore`)
- In production, should be replaced with object storage (S3/GCS)

**`pkg/database/migrations/`:**
- Purpose: Versioned schema migration SQL files
- Generated: No — hand-written
- Committed: Yes

**`deployments/`:**
- Purpose: Infrastructure config (Docker, Kubernetes manifests)
- Committed: Yes

**`nginx/`:**
- Purpose: Nginx reverse proxy config and SSL certificates
- Committed: Config yes, SSL certs no

**`bin/`:**
- Purpose: Compiled binary output
- Not committed

**`internal/models/` and `internal/repositories/`:**
- Purpose: Legacy directories — not actively used; domain models live in `{domain}/model/` and queries in `{domain}/dao/`

---

*Structure analysis: 2026-03-29*
