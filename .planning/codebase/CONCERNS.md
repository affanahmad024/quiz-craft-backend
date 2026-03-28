# Technical Concerns

**Analysis Date:** 2026-03-29

---

## Security Concerns

### WebSocket Origin Validation Disabled
- **Files:** `livesession/websocket/handler.go:16`, `notification/websocket/handler.go:15`
- **Issue:** Both WebSocket upgraders use `CheckOrigin: func(r *http.Request) bool { return true }` — all origins accepted unconditionally. TODOs acknowledge this must be fixed for production but the fix has never landed.
- **Risk:** Cross-site WebSocket hijacking. Any page on the internet can open a WebSocket connection to the server while using the victim's cookies.
- **Fix:** Check `r.Header.Get("Origin")` against the allowed-origins list from `CORSConfig`.

### Hardcoded Credentials in Committed Config File
- **File:** `config/config.yaml`
- **Issue:** The committed YAML contains `password: 2024` (Postgres), `secret: your-secret-key-change-in-production` (JWT), and placeholder SMTP credentials. This file is tracked by git.
- **Risk:** Any developer who clones the repo or any leaked git history exposes database credentials and a weak JWT secret to anyone who reads the file.
- **Fix:** Remove all secret values from `config/config.yaml`; leave only structural defaults. Rely exclusively on environment variables for secrets. Add `config/config.yaml` entries to `.gitignore` or use a separate `config.yaml.example`.

### JWT Secret Defaults to Weak Value
- **File:** `config/config.yaml` line 18, `internal/config/config.go:254`
- **Issue:** `validateConfig` only checks the secret is non-empty. The literal string `"your-secret-key-change-in-production"` satisfies that check. No length or entropy validation is performed.
- **Risk:** A misconfigured deployment uses the well-known default secret; tokens can be forged.
- **Fix:** Enforce a minimum secret length (e.g., 32 bytes) in `validateConfig`.

### No JWT Token Revocation / Blacklist
- **Files:** `internal/utils/jwt.go`, `auth/core/auth.go`
- **Issue:** Refresh tokens are stateless JWTs. On `RefreshToken`, the old refresh token remains valid until its 168-hour expiry. There is no token revocation, logout-all-devices, or server-side session store.
- **Risk:** Stolen refresh tokens remain valid for 7 days and cannot be invalidated without rotating the global JWT secret.
- **Fix:** Store issued refresh token JTI values in a Redis set or database table; delete the record on logout; validate JTI on every refresh request.

### `rand.Read` Error Ignored in Share-Code Generation
- **File:** `quiz/core/search.go:158`
- **Issue:** `rand.Read(b)` return error is silently discarded. If the OS CSPRNG is unavailable (extremely rare but possible in hardened containers), the slice remains zeroed.
- **Risk:** Predictable share codes.
- **Fix:** `if _, err := rand.Read(b); err != nil { return "", err }`.

### Admin Role Re-Validated Per Request via DB Lookup
- **Files:** `admin/core/statistics.go`, `admin/core/user_management.go`
- **Issue:** Every admin operation calls `userdao.FindByID` and checks `role == "admin"` manually, in addition to the `Admin` middleware. This dual-check is inconsistent — some functions use the helper `isAdmin()` that hits the DB, others re-fetch the user directly.
- **Risk:** Inconsistency means a future code path may bypass one of the two checks, creating a privilege escalation vector.
- **Fix:** Trust the middleware-set context value; remove per-function DB re-validation of the admin role.

### File MIME Type Validated from Header Only
- **File:** `ai-agent/services/file_upload.go:205`
- **Issue:** `MimeType: fileHeader.Header.Get("Content-Type")` is stored in the database from the client-supplied multipart header, not from actual content inspection.
- **Risk:** Attackers can upload a PHP/Python/shell script with a forged MIME type of `application/pdf`. The extension check alone (`AllowedExtensions`) is the only real gate.
- **Fix:** Use `github.com/gabriel-vasile/mimetype` (already in `go.mod`) to detect actual MIME type from file bytes after writing to disk.

---

## Performance Concerns

### Session Worker Polls Every Second Against Full DB Query
- **File:** `livesession/worker/session_worker.go:17`
- **Issue:** `WorkerPollInterval = 1 * time.Second`. Every second the worker queries `GetScheduledSessions` and `GetLiveAutomaticSessions`, both of which scan the live sessions table. With many concurrent sessions this becomes O(sessions) DB hits per second.
- **Impact:** At moderate scale (hundreds of concurrent sessions) this generates thousands of queries per minute on performance-critical tables.
- **Fix:** Use a priority queue or Redis sorted set keyed by next-event-time; only query DB when events are due.

### N+1 Pattern in Session Worker End
- **File:** `livesession/worker/session_worker.go:279-283`, `livesession/core/session.go:253-258`
- **Issue:** `EndSession` calls `dao.ListParticipants` then loops calling `dao.UpdateParticipantStatus` individually per participant.
- **Fix:** Replace with a single bulk `UPDATE ... WHERE session_id = $1 AND status = 'active'` query.

### In-Memory Rate Limiter Lost on Restart
- **File:** `ai-agent/services/rate_limiter.go`
- **Issue:** `RateLimiter` stores upload and AI generation counts in a Go `map`. Process restart resets all counters. Multiple server instances have independent counters.
- **Impact:** Users can bypass AI generation quotas (20/day) by hitting different replicas or restarting the service.
- **Fix:** Back rate limit windows with Redis using `INCR` + `EXPIRE`.

### Repeated Scan/Nullable Field Boilerplate in DAO Layer
- **Files:** `quiz/dao/attempt.go`, `quiz/dao/quiz.go`, `livesession/dao/session.go`
- **Issue:** Every list query duplicates 30–50 lines of `sql.NullString` / `sql.NullTime` / `json.Unmarshal` boilerplate per row. `FindByID` and `FindByIDIncludingDeleted` in `quiz/dao/quiz.go` are nearly identical.
- **Impact:** Maintenance overhead; bugs fixed in one copy may not be fixed in the other.
- **Fix:** Extract a `scanQuiz(row)` helper. Consider `sqlx` or a lightweight scan helper.

### Admin Stats Query Not Cached
- **File:** `admin/core/statistics.go`, `admin/dao/admin_stats.go`
- **Issue:** `GetPlatformStats` runs aggregate COUNT queries across large tables on every admin request. No caching layer exists (Redis is unused in the platform stats path).
- **Fix:** Cache platform stats for 60 seconds in Redis.

---

## Technical Debt

### Educator Quiz Listing Is a Stub
- **File:** `api/educator.go:438-464`
- **Issue:** `GetEducatorQuizzes` parses pagination parameters and immediately discards them (`_ = 20`, `_ = 0`, `_ = parsed`). The handler returns `{"quizzes": [], "total": 0, "message": "Quiz listing integration pending"}`.
- **Impact:** The educator quiz listing feature is non-functional. Clients calling this endpoint get empty data silently.
- **Fix:** Implement the educator quiz listing by reusing `quiz/core` and `quiz/dao` with the educator's user ID.

### `TODO: Get actual inviter name` in Live Session Invitations
- **File:** `livesession/core/session.go:416`
- **Issue:** Notification sent to invited users reads `"A user invited you to a live quiz"` instead of the inviter's name. The inviter's user record is already available in the calling scope (`inviterID`).
- **Fix:** Fetch inviter name with `userdao.FindByID(ctx, inviterID)` before sending notifications.

### Ignored JSON Unmarshal Errors in DAO and Model Layers
- **Files:** `quiz/dao/quiz.go:119`, `quiz/dao/attempt.go:122`, `quiz/dao/attempt.go:418`, `user/model/password.go:47`, `user/model/password.go:56`, `educator/model/profile.go:78`, `educator/model/profile.go:84`, `notification/model/notification.go:56`, `livesession/model/session.go:129`, and 10+ more
- **Issue:** Calls like `json.Unmarshal(metaJSON, &quiz.Meta)` silently discard errors. Corrupted JSON in the database causes silent field omission without log output.
- **Fix:** Log the error with the record ID: `if err := json.Unmarshal(...); err != nil { commons.LogW(ctx, "...") }`.

### `buildPrompt` Function Is Unreachable Dead Code
- **File:** `ai-agent/core/ai_service.go:285`
- **Issue:** `buildPrompt` exists alongside `buildPromptWithImages` but is never called from any code path. All callers use `buildPromptWithImages`.
- **Fix:** Remove `buildPrompt` or consolidate into `buildPromptWithImages`.

### Duplicate `min` Helper That Go 1.21 Now Provides Natively
- **File:** `ai-agent/core/ai_service.go:417`
- **Issue:** A local `min(a, b int)` function is defined. The module uses Go 1.24.1 which includes the built-in `min` function.
- **Fix:** Delete the local definition.

### Committed Binary Artifacts
- **Files:** `backend.exe`, `quiz-craft.exe`, `bin/quizcraft`, `quiz-craft-server`
- **Issue:** Compiled binaries are committed to the repository.
- **Impact:** Inflates repository size; creates confusion about which binary is canonical.
- **Fix:** Add `*.exe`, `bin/`, and the Linux binary names to `.gitignore` and remove them from git tracking.

---

## Maintainability Issues

### Context Key Inconsistency (`user_id` as string vs `commons.SetUserID`)
- **Files:** `internal/middleware/auth.go:63-64`, `livesession/websocket/handler.go:27`
- **Issue:** The HTTP auth middleware uses `commons.SetUserID(c, claims.UserID)` which stores a `uuid.UUID`. The livesession WebSocket handler reads back with `c.Get("user_id")` returning a raw `interface{}` and casts to `string`. The notification WebSocket handler uses `commons.UsrID(ctx)`. Three different retrieval patterns exist for the same value.
- **Fix:** Standardize all context reads through `commons.UsrID` / `commons.SetUserID`.

### Admin Authorization Done Two Different Ways
- **Files:** `admin/core/statistics.go`, `admin/core/user_management.go`
- **Issue:** `GetPlatformStats` and `GetUserActivityTimeline` check `admin.Role != "admin"` after a DB fetch. `ListUsers` and `GetUserDetail` call the `isAdmin(ctx, adminID)` helper which also does a DB fetch. `DisableUser` / `EnableUser` / `ChangeUserRole` / `DeleteUser` do `uuid.MustParse(adminID)` (panics on bad UUID) then another `FindByID`.
- **Fix:** Single helper; use the middleware-injected role claim from the JWT.

### `uuid.MustParse` Panic Risk in Admin Core
- **Files:** `admin/core/user_management.go:73`, `admin/core/user_management.go:118`, `admin/core/user_management.go:157`, `admin/core/user_management.go:210`
- **Issue:** `uuid.MustParse(adminID)` panics if `adminID` is not a valid UUID. The admin middleware sets this value from JWT claims which should always be valid, but a future refactor or middleware bypass could trigger a panic in production.
- **Fix:** Use `uuid.Parse` and return an error.

### No Tests Exist Anywhere
- **Finding:** Zero `*_test.go` files exist in the entire codebase.
- **Impact:** No automated confidence for regressions. Live session scoring, AI quiz generation, auth flows, and admin operations are all completely untested. Any refactor or dependency update is high-risk.
- **Fix:** Start with unit tests for `internal/utils/jwt.go`, `internal/utils/password.go`, `livesession/core/session.go` (score calculation), and integration tests for the auth flow.

### Logger Level Magic Number
- **File:** `pkg/database/postgres.go:23`
- **Issue:** GORM log level is selected with `case 5: // Debug`. The magic number `5` maps to `logrus.DebugLevel` but this is fragile: if `logrus` changes its constant values the mapping silently breaks.
- **Fix:** `case logrus.DebugLevel:`.

---

## Missing Features / Gaps

### Short-Answer Question Type Not Supported
- **File:** `quiz/core/attempt.go:183`
- **Issue:** `SubmitAnswer` explicitly rejects `question_type != "multiple_choice" && question_type != "true_false"` with error `"question-type-not-supported"`. The AI generator and database schema both support `short_answer` but no scoring or submission path exists.
- **Impact:** AI-generated short-answer quizzes cannot be taken.

### Educator Quiz Listing Endpoint Returns Placeholder
- **File:** `api/educator.go:459-464`
- **Issue:** As noted in Technical Debt — the endpoint exists in the router but returns empty data. Frontend consumers will see zero quizzes for any educator.

### No Global HTTP Rate Limiting Middleware
- **File:** `internal/config/config.go:238` (`rate_limit` config exists), but no middleware applies it
- **Issue:** `RateLimitConfig` is parsed from config/env but no HTTP middleware uses it. Only the AI-specific in-memory limiter exists.
- **Impact:** Auth endpoints (`/auth/login`, `/auth/signup`, `/auth/forgot-password`) are not rate-limited at the HTTP layer, making brute-force and enumeration easier.
- **Fix:** Add a per-IP sliding window middleware using Redis or `golang.org/x/time/rate` applied globally in `main.go`.

### No Refresh Token Rotation
- **File:** `auth/core/auth.go:513`
- **Issue:** `RefreshToken` returns the same refresh token (`RefreshToken: refreshToken`). Best practice is to issue a new refresh token and invalidate the old one on each use.

### Uploaded Files Stored Locally Without Cleanup Guarantee
- **File:** `ai-agent/services/file_cleanup.go`, `ai-agent/services/file_upload.go:195`
- **Issue:** Files expire after 7 days and a cleanup goroutine runs, but the `uploads/` directory is local disk. In any multi-replica or ephemeral deployment (Docker, Kubernetes), files written by one replica are invisible to others, and restart wipes them.
- **Impact:** AI generation jobs will fail with "file not found" in scaled deployments.
- **Fix:** Use object storage (S3, GCS) or a shared volume.

---

## Scalability Concerns

### Single-Process WebSocket Hub
- **Files:** `livesession/websocket/hub.go`, `notification/websocket/hub.go`
- **Issue:** Both `Hub` types are in-process goroutines storing client connections in memory maps. WebSocket sessions are pinned to a single server process.
- **Impact:** Cannot horizontally scale WebSocket connections. A second server process has no clients registered. Load balancers must use sticky sessions.
- **Fix:** Introduce a pub/sub backend (Redis pub/sub) so any replica can broadcast to any session.

### Session Worker Is a Single Goroutine Polling Entire Sessions Table
- **File:** `livesession/worker/session_worker.go:63-78`
- **Issue:** `ProcessScheduledStarts` fetches all sessions due to start; `ProcessAutomaticProgression` fetches all live automatic sessions. With many concurrent live sessions, these are full table scans every second.
- **Fix:** Move scheduling to Redis sorted sets or a persistent job queue (e.g., Asynq, River).

### Local File Storage Blocks Horizontal Scaling
- **File:** `ai-agent/services/file_upload.go:23` (`UploadStoragePath = "./uploads"`)
- **Issue:** Uploads are stored on the local filesystem. As noted under Missing Features, this is incompatible with multi-replica deployments.

---

## Dependency Risks

### Google Generative AI SDK Uses Deprecated `gemini-1.5-pro` Model
- **File:** `ai-agent/core/ai_service.go:19`
- **Issue:** `DefaultModel = "gemini-1.5-pro"`. Google has announced Gemini 1.5 Pro deprecation timelines in favor of Gemini 2.x. When the model is retired, all AI generation features will fail.
- **Fix:** Upgrade to `gemini-2.0-flash` or `gemini-2.5-pro` and make the model name a required config value with no default.

### `github.com/nguyenthenguyen/docx` Is Effectively Unmaintained
- **File:** `go.mod:17`
- **Issue:** The docx parser (`v0.0.0-20230621112118`) was last updated in 2023 and has no active maintenance. It is used in `ai-agent/services/file_parser.go` for `.docx` parsing.
- **Risk:** Security vulnerabilities in OOXML parsing will not receive patches.
- **Fix:** Evaluate `github.com/unidoc/unioffice` or convert `.docx` server-side using a system tool such as LibreOffice.

### Both GORM and `database/sql` Used Simultaneously
- **Files:** `pkg/database/postgres.go`, all `*/dao/*.go` files
- **Issue:** GORM manages the connection pool and migrations, but all DAO operations use the underlying `*sql.DB` directly via `database/sql`. This dual-layer provides no benefit of GORM's ORM features while still carrying its dependency weight.
- **Risk:** Confusion for contributors; GORM logging config and pool settings must be applied to the GORM instance but queries run through the raw `*sql.DB`.
- **Fix:** Either commit fully to raw `database/sql` (drop GORM) or use GORM consistently.

---

## Recommended Quick Wins

1. **Fix WebSocket origin check** — `livesession/websocket/handler.go` and `notification/websocket/handler.go`: replace `return true` with an origin allowlist check. ~10 lines of code, high security impact.

2. **Remove secrets from `config/config.yaml`** — Replace `password: 2024` and `secret: your-secret-key-change-in-production` with empty strings or `${ENV_VAR}` placeholders, add the file to `.gitignore`. Immediate security improvement.

3. **Add `.gitignore` entries for binaries** — `*.exe`, `bin/`, `quiz-craft-server`. Reduces repo bloat with no code changes.

4. **Implement educator quiz listing** — `api/educator.go:459-464`: route the existing `quiz/core.ListByCreator` logic using the educator's ID. ~20 lines, fixes a broken API endpoint.

5. **Fix inviter name in live session notification** — `livesession/core/session.go:416`: one DAO call to get the inviter's name. ~5 lines.

6. **Log ignored `json.Unmarshal` errors** — Add `if err != nil { commons.LogW(ctx, ...) }` after each silent unmarshal. ~30 one-line additions; makes data corruption visible in logs.

7. **Replace `uuid.MustParse` with `uuid.Parse` in admin core** — `admin/core/user_management.go` lines 73, 118, 157, 210. Eliminates panic risk. ~8 lines changed.

8. **Add rate limiting to auth endpoints** — Apply `golang.org/x/time/rate` (already transitively available) per-IP limiter to `/auth/login`, `/auth/signup`, `/auth/forgot-password`. Closes brute-force exposure with ~30 lines of middleware.

---

*Concerns audit: 2026-03-29*
