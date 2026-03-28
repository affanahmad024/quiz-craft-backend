# Code Conventions

**Analysis Date:** 2026-03-29

## Naming Conventions

**Files:**
- Go files use `snake_case` for multi-word names: `quiz_generator_worker.go`, `error_handler.go`
- Single-concept files match their primary type or role: `auth.go`, `session.go`, `user.go`
- Files group by responsibility within their layer: `password.go` (passwords only), `user.go` (user entity only)

**Packages:**
- Domain modules use shorthand aliases when imported: `aiagentdao`, `authcore`, `livesessionws`, `notificationworker`
- Pattern: `{domain}{layer}` e.g., `livesessiondao`, `quizdao`, `userdao`
- Internal utilities use full package name: `middleware`, `utils`, `config`, `logger`

**Functions:**
- Exported functions use `PascalCase`: `CreateUser`, `FindByID`, `UpdateLastLogin`
- DAO functions describe the operation + subject: `FindByEmail`, `SoftDelete`, `HardDelete`, `SetVerificationToken`
- Core functions mirror DAO names at a higher level: `core.Login`, `core.Signup`, `core.ChangePassword`
- API handler functions use plain resource-action names: `Signup`, `Login`, `GetQuiz`, `CreateLiveSession`

**Variables:**
- Local variables use `camelCase`: `userDAO`, `accessToken`, `hashedPassword`, `participantDAO`
- DAO struct instances suffixed with `DAO`: `userDAO`, `pwdDAO`, `sessionDAO`, `participantDAO`
- API/response struct instances suffixed with `API` or `api`: `userAPI`, `api`
- Request structs suffixed with `Request`: `SignupRequest`, `LoginRequest`, `UpdateProfileRequest`
- Response structs suffixed with `Response`: `LoginResponse`, `SignupResponse`, `APIResponse`

**Types:**
- DAO types end with `DAO`: `UserDAO`, `LiveSessionDAO`, `ParticipantDAO`
- API types end with `API`: `UserAPI`, `LiveSessionAPI`, `ParticipantAPI`
- Filter types end with `Filter`: `UserListFilter`
- List response types end with `ListResponse` or `Response`: `UserListResponse`
- Errors use descriptive string codes in `kebab-case`: `"invalid-credentials"`, `"email-already-exists"`, `"user-not-found"`

**Constants:**
- Use `SCREAMING_SNAKE_CASE` for package-level const blocks: `MaxConcurrent`, `PollInterval`, `BasePoints`
- Role strings stored as package constants: `RoleUser = "user"`, `RoleAdmin = "admin"`
- Token types stored as constants: `TokenTypeAccess`, `TokenTypeRefresh`

## Code Organization

**Module layout (per domain):**
```
{domain}/
  core/      # Business logic; calls dao layer
  dao/       # Raw SQL queries; no business logic
  model/     # Structs: DAO types, API types, request/response types
  websocket/ # (if applicable) WebSocket hub, client, handler
  worker/    # (if applicable) Background goroutine-based processing
```

**File grouping within a layer:**
- `core/auth.go` — all auth business logic in one file per domain concern
- `dao/user.go` and `dao/password.go` — split by entity/table when a DAO grows large
- `model/user.go` and `model/password.go` — split by entity when models diverge

**`api/` package:**
- One file per feature area: `auth.go`, `quiz.go`, `livesession.go`, `admin.go`
- Handlers are thin: extract context, bind request, call core, handle error, respond

**`commons/` package:**
- `ctx.go` — context helpers (trace ID, user ID extraction)
- `errors.go` — domain error types and constructors
- `log.go` — context-aware logging wrappers
- `response.go` — standard HTTP response helpers

**`internal/` package:**
- `config/` — Viper-based config loading
- `middleware/` — Gin middleware (Auth, CORS, Logger, ErrorHandler, RequireAdmin, RequireEducator)
- `utils/` — JWT manager, password hashing, token generation
- `services/` — Email service (EmailService interface + Mailjet implementation)

**`pkg/` package:**
- `database/` — PostgreSQL init, migrations
- `logger/` — Logrus singleton
- `cache/` — (present but not analyzed further)

## Error Handling Patterns

**Domain error system (all in `commons/errors.go`):**

Six sentinel errors with typed constructors:
```go
// Constructors — return *DomainError with a kebab-case code string
commons.NotFound("user-not-found")
commons.BadInput("invalid-credentials")
commons.Conflict("email-already-exists")
commons.Unauthorized("invalid-credentials")
commons.Forbidden("not-quiz-owner")
commons.Internal("unexpected failure message")
```

**Error propagation (DAO → core → api):**

1. DAO layer: wraps `sql.ErrNoRows` into `commons.NotFound(...)`, passes raw DB errors up
2. Core layer: calls DAO, checks return error, either wraps into domain errors or propagates
3. API layer: inspects error type with `commons.Is*()` checks and maps to HTTP status codes

```go
// Standard API handler error-dispatch pattern:
if err != nil {
    commons.LogC(ctx, "[api.Signup::KO [time_taken=%v ...")
    if commons.IsBadInput(err) {
        commons.RespBadReq(c, commons.GetCode(err))
        return
    }
    if commons.IsConflict(err) {
        commons.Resp409(c, commons.GetCode(err))
        return
    }
    commons.RespISE(c, err)
    return
}
```

**Error code extraction:**
- `commons.GetCode(err)` — extracts the kebab-case code string from a `*DomainError` and passes it directly to the response body

**Non-critical errors (best-effort operations):**
```go
// Pattern: log warning, do not return error
if err := dao.UpdateLastLogin(ctx, userDAO.ID); err != nil {
    commons.LogW(ctx, "[core.Login::UpdateLastLoginErr ...")
    // Don't fail login if this fails, just log the error
}
```

**Fire-and-forget (goroutine) for async operations:**
```go
// Email sends and notification dispatches are async, errors are silently dropped
if emailService != nil {
    go emailService.SendVerificationEmail(ctx, ...)
}
```

**Middleware error handling:**
- `gin.Recovery()` at server level catches panics
- `middleware.ErrorHandler()` recovers and logs panics, responds with 500
- `middleware.RespondWithError()` used in middleware for inline error responses

## Logging Patterns

**Library:** `github.com/sirupsen/logrus` via singleton in `pkg/logger`

**Two logging surfaces:**

1. `commons.Log/LogW/LogE/LogC(ctx, format, args...)` — context-aware, includes `trace_id` field automatically; used throughout `core/` and `dao/` layers
2. `logger.GetLogger().Info/Warn/Error/WithFields(...)` — direct logrus calls; used in `middleware/`, `main.go`, and worker files

**Log format by environment:**
- `production`: JSON formatter (`logrus.JSONFormatter`)
- `development`: Text formatter with full timestamp and colors (`logrus.TextFormatter`)

**Log message format (structured prefix pattern):**
```
"[{layer}.{Function}::{Event} [key=value [key=value"
```
Examples:
```
"[api.Signup::KO [time_taken=%v [ip=%s [email=%s [err=%v"
"[core.Login::UserNotFound [email=%s"
"[dao.CreateUser::OK [user_id=%s"
"[dao.FindByID::SQLErr [id=%s [err=%v"
```

**Log levels used:**
- `Info` — normal operation events, start/OK confirmations
- `Warn` (LogW) — expected business errors: invalid credentials, not found, already verified
- `Error` (LogE) — unexpected technical failures: SQL errors, hash failures, token generation failures
- `Critical` (LogC) — request-level KO with timing; implemented as `Error` with `level=CRITICAL` field

**Timing pattern in API handlers:**
```go
st := time.Now()
// ... handler logic ...
commons.Log(ctx, "[api.Signup::OK [time_taken=%v ...", time.Since(st), ...)
```

**Request tracing:**
- Each request gets a UUID trace ID via `X-Trace-ID` header or generated fresh
- Trace ID propagated via `context.Context` using `commons.TraceIDKey`
- Returned to client via `X-Trace-ID` response header

## Common Patterns

**DAO Init pattern:**
```go
var db *sql.DB

func Init(database *sql.DB) {
    db = database
}
```
Every DAO package has a package-level `*sql.DB` and an `Init()` function called from `main.go`.

**DAO → API conversion:**
Every DAO struct has a `ToAPI()` method converting database types (`sql.NullString`, `[]byte`) to API-safe types:
```go
func (u *UserDAO) ToAPI() UserAPI { ... }
func (s *LiveSessionDAO) ToAPI() LiveSessionAPI { ... }
```

**Core Init with interface injection:**
```go
var (
    jwtManager   *utils.JWTManager
    emailService EmailService
)

func Init(jwt *utils.JWTManager, email EmailService) {
    jwtManager = jwt
    emailService = email
}
```
Dependencies injected at startup via `Init()`, not constructor injection per call.

**Context propagation:**
- Gin handler extracts context: `ctx := commons.Ctx(c)` (adds trace ID + user ID)
- Context passed through: `core.Signup(ctx, req)` → `dao.CreateUser(ctx, ...)` → `db.QueryRowContext(ctx, ...)`

**Request binding:**
```go
var req model.SignupRequest
if err := c.ShouldBindJSON(&req); err != nil {
    commons.LogW(ctx, "[api.Signup::BindErr ...")
    commons.RespBadReq(c, "invalid-request-body")
    return
}
```
All handlers use `ShouldBindJSON` followed by immediate early return on error.

**Background workers:**
```go
type QuizGeneratorWorker struct {
    stopChan       chan struct{}
    concurrencySem chan struct{}
}

func (w *Worker) Start(ctx context.Context) {
    ticker := time.NewTicker(PollInterval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done(): return
        case <-w.stopChan: return
        case <-ticker.C: // do work
        }
    }
}
```
All workers follow ticker+select pattern with context cancellation and stop channel.

**WebSocket Hub pattern:**
- Hub runs in its own goroutine: `go wsHub.Run()`
- Clients register/unregister via channels: `hub.register <- client`
- WritePump and ReadPump run in separate goroutines per client

**Soft delete pattern:**
- All user deletes set `deleted_at = NOW()`
- All queries filter `AND deleted_at IS NULL` by default
- Dedicated `FindByIDIncludingDeleted` variants exist where admin access requires it

**Pagination:**
```go
offset := (filter.Page - 1) * filter.Limit
// COUNT(*) query runs first, then paginated SELECT with LIMIT/OFFSET
```

## Anti-patterns Observed

**Inconsistent middleware usage in `auth.go`:**
Both `commons.UsrID(ctx)` (preferred) and direct `c.Get("user_id")` (raw gin key) appear in different files. The `OptionalAuth` middleware sets `c.Set("user_id", ...)` while `Auth` middleware uses `commons.SetUserID()`. Use `commons.UsrID(ctx)` exclusively.

**Ignored errors in non-critical paths:**
Pattern `dao.UpdateParticipantStatus(ctx, p.ID, ...)` (no error check) appears in `EndSession` and `EndSessionDueToTimeout`. Silent drops should at minimum be logged.

**Hardcoded string in live session invitation:**
`inviterName := "A user"` with `// TODO: Get actual inviter name` at `livesession/core/session.go:416`. Invited users receive generic sender name.

**WebSocket origin check disabled:**
Both `notification/websocket/handler.go` and `livesession/websocket/handler.go` have `CheckOrigin: func(...) bool { return true }` with `// TODO: In production, check against allowed origins`. Origin validation is skipped entirely.

**Mixed ID types between modules:**
`quiz/model` uses `string` for IDs (`ID string`, `CreatorID string`), while `user/model` and `livesession/model` use `uuid.UUID`. Causes string↔UUID conversions like `sessionDAO.CreatorID.String() != userID.String()`.

**Direct `os.Getenv` in worker:**
`quiz_generator_worker.go` reads `GOOGLE_AI_API_KEY` via `os.Getenv` directly and panics if missing. Config should flow through `internal/config` instead.

**Duplicate response struct:**
`commons.APIResponse` and `middleware.ErrorResponse` are parallel structs with the same shape (`success bool`, `message string`, `error string`). Both are used in different parts of the codebase.
