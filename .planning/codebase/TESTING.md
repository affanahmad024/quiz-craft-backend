# Testing

**Analysis Date:** 2026-03-29

## Testing Framework

**Runner:** None configured. Zero test files exist in the codebase (`*_test.go` count: 0).

No test framework, test runner configuration, or CI test pipeline has been set up. The entire codebase — including all business logic in `auth/core/`, `quiz/core/`, `livesession/core/`, all DAO functions, all middleware, and all API handlers — is untested.

**Mocking library present (unused):**
- `go.uber.org/mock v0.6.0` is listed as an indirect dependency in `go.mod`, suggesting intent to use mock generation, but no mocks have been written.

## Test Organization

No test files exist. If tests were added, the Go standard convention for this project would be:

**Co-located test files** (Go standard):
```
auth/core/auth_test.go        # mirrors auth/core/auth.go
user/dao/user_test.go         # mirrors user/dao/user.go
quiz/core/quiz_test.go        # mirrors quiz/core/quiz.go
```

**Naming:** `{filename}_test.go` in the same package directory.

**Package declaration options:**
- `package core` — white-box tests (access unexported identifiers)
- `package core_test` — black-box tests (test only exported surface)

## Coverage & Strategy

**Current coverage: 0%**

No unit tests, integration tests, or end-to-end tests exist for any layer.

**What should be covered (by priority):**

**High — Core business logic:**
- `auth/core/auth.go` — `Signup`, `Login`, `VerifyEmail`, `ChangePassword`, `ResetPassword`
- `quiz/core/quiz.go` — `CreateQuiz`, `GetQuiz`, access control checks
- `livesession/core/session.go` — `SubmitAnswer`, `CalculatePoints`, `JoinSession`, `StartSession`
- `quiz/core/attempt.go` — answer submission, auto-submit logic

**High — Domain error mapping:**
- `commons/errors.go` — `Is*()` check functions, `GetCode()`, `DomainError.Unwrap()`

**Medium — DAO layer:**
- `user/dao/user.go` — requires a real or in-memory PostgreSQL instance; integration test territory
- `user/dao/password.go` — token expiry logic
- `quiz/dao/quiz.go`, `quiz/dao/attempt.go`

**Medium — Utilities:**
- `internal/utils/jwt.go` — token generation, validation, expiry
- `internal/utils/` — password hashing, strength validation, token generation

**Low — HTTP handlers (API layer):**
- `api/auth.go`, `api/quiz.go` — thin dispatch logic; mostly covered by core tests

## Test Utilities

No test helpers, fixtures, factories, or mocks exist.

**What should be built first if tests are added:**

**Interface mocks (using `go.uber.org/mock`):**
- `auth/core/auth.go` depends on `EmailService` interface — mockable immediately
- DAO functions are package-level functions (not interfaces), which makes them harder to mock without interface extraction

**Database setup for integration tests:**
- A test database or `dockertest` setup would be needed for DAO-layer tests
- No migration runner test helper exists yet; `database.RunMigrations()` in `pkg/database` could be reused in test setup

**Context helpers:**
- `commons.Ctx(c *gin.Context)` requires a gin context; a test helper creating a bare `context.WithValue(context.Background(), ...)` would be needed for core/dao tests

**Example test structure for `CalculatePoints` (pure function — no dependencies):**
```go
// livesession/core/session_test.go
package core

import "testing"

func TestCalculatePoints(t *testing.T) {
    tests := []struct {
        name        string
        timeTakenMs int
        maxTimeMs   int
        want        int
    }{
        {"fastest possible", 0, 10000, 150},     // 100 * (1 + 0.5)
        {"at time limit", 10000, 10000, 100},    // 100 * (1 + 0.0)
        {"over time limit", 15000, 10000, 100},  // clamped to maxTime
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculatePoints(tt.timeTakenMs, tt.maxTimeMs)
            if got != tt.want {
                t.Errorf("CalculatePoints(%d, %d) = %d, want %d",
                    tt.timeTakenMs, tt.maxTimeMs, got, tt.want)
            }
        })
    }
}
```

## Running Tests

No test commands are configured. Standard Go test commands that would apply:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run a specific package
go test ./auth/core/...

# Run a specific test function
go test -run TestCalculatePoints ./livesession/core/...
```

No `Makefile`, `justfile`, or CI pipeline (`.github/workflows/`, `.gitlab-ci.yml`) exists to standardize test execution.

## Gaps

**Everything is a gap.** There are zero tests in the entire codebase. Specific high-risk areas with no coverage:

**Authentication flows — critical, zero coverage:**
- Account lockout after 5 failed attempts (`constants.MaxFailedAttempts`)
- Token expiry validation in `VerifyEmail` and `ResetPassword`
- Password history enforcement (last 5 passwords blocked)
- JWT token type validation (`access` vs `refresh`)

**Authorization logic — critical, zero coverage:**
- `middleware.Auth` — token extraction from header vs cookie, invalid formats
- `middleware.RequireAdmin` — role check against DB
- `middleware.RequireEducator` — educator role enforcement
- Quiz access control: `quiz.IsAccessibleBy(userID)` in `core.GetQuiz`

**Live session logic — complex, zero coverage:**
- `CalculatePoints` — has edge cases (over-time clamping, speed bonus calculation)
- `JoinSession` — access code vs invitation validation, capacity check
- `AdvanceToNextQuestion` — manual vs automatic progression guard

**Worker behavior — zero coverage:**
- `QuizGeneratorWorker` polling loop
- `AttemptWorker` auto-submit behavior
- `SessionWorker` auto-start and auto-progression

**DAO SQL queries — zero coverage:**
- All raw SQL in `user/dao/`, `quiz/dao/`, `livesession/dao/`, `notification/dao/`
- Soft delete filtering (`AND deleted_at IS NULL`) correctness
- Dynamic query building in `ListUsers` (filter/pagination)

**Error code propagation — zero coverage:**
- `commons.GetCode(err)` when error is not a `*DomainError`
- All `Is*()` functions in `commons/errors.go`

**Priority for first tests to write:**
1. `livesession/core/session.go` — `CalculatePoints` (pure function, no dependencies)
2. `commons/errors.go` — all error type constructors and checkers (no external deps)
3. `internal/utils/jwt.go` — token generation and validation (self-contained)
4. `auth/core/auth.go` — `Login`, `Signup` with mocked `EmailService` and DAO interfaces
