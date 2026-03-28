# quiz-craft backend

## What This Is

A Go/Gin REST API backend for a quiz platform where educators create quizzes, students take them, and live sessions run in real time. The platform includes AI-powered quiz generation via Google Gemini, WebSocket-based live session management with leaderboards, an async job queue for AI tasks, and a full auth system with email verification and account lockout. Admin, educator, and student roles are enforced via JWT middleware.

## Core Value

Educators can create and run live quiz sessions that students experience in real time — everything else supports that flow.

## Requirements

### Validated

- ✓ User authentication — signup, login, email verification, password reset, account lockout, JWT access + refresh tokens — existing
- ✓ User profile management — view/update profile, avatar — existing
- ✓ Quiz CRUD — create, edit, delete quizzes with questions and answers — existing
- ✓ Quiz attempts — student can attempt quiz, answers scored, results returned — existing
- ✓ Quiz discovery — search by keyword, filter by category/tags, public share links — existing
- ✓ AI quiz generation — upload PDF/DOCX/image → async Gemini job → quiz created — existing
- ✓ Live sessions — educator starts session, students join via WebSocket, real-time scoring and leaderboard — existing
- ✓ Notification system — per-user WebSocket notifications, session invitations — existing
- ✓ Educator profiles — follower/subscription system, public educator discovery — existing
- ✓ Admin panel — user management, quiz moderation, platform stats, activity log — existing
- ✓ File upload — PDF/DOCX/image upload with size and MIME type validation — existing
- ✓ Versioned DB migrations — golang-migrate, runs automatically on startup — existing

### Active

- [ ] Remove admin role DB re-validation from per-function core code — trust middleware-set context; delete `isAdmin()` helper and direct `FindByID` role checks inside admin core functions
- [ ] Audit all role checks across admin, educator, and student routes for consistency and privilege escalation vectors
- [ ] Fix WebSocket CORS — replace `CheckOrigin: func() { return true }` with origin validation against configured CORS allowlist
- [ ] Remove committed secrets — `config/config.yaml` commits real DB password and a placeholder JWT secret; move to `.env.example`-only pattern
- [ ] Replace `uuid.MustParse` panic calls in admin core with proper error handling

### Out of Scope

- Horizontal scaling / Redis pub-sub for WebSocket — architectural change, defer to future milestone
- JWT token revocation / server-side token store — larger auth overhaul, defer
- Test coverage — important but separate effort from security hardening
- Educator quiz listing stub — functional gap, separate feature milestone

## Context

- **Codebase map:** `.planning/codebase/` — full architecture, stack, conventions, and concerns analysis
- **Auth middleware:** `internal/middleware/auth.go` validates JWT and sets `user_id` in context. `RequireAdmin` and `RequireEducator` do a DB lookup per request to validate role — this is the intended role enforcement point.
- **Problem pattern:** `admin/core/statistics.go` and `admin/core/user_management.go` additionally call `userdao.FindByID` and check `role == "admin"` inline, independent of the middleware. Some paths use a helper `isAdmin()`, others re-fetch directly — inconsistent.
- **Risk:** Inconsistency means future code paths may bypass one check, creating a privilege escalation vector. Fix: remove the redundant checks, trust the middleware.
- **Zero tests:** The entire codebase has no test files despite `go.uber.org/mock` being a dependency.

## Constraints

- **Stack:** Go 1.24 + Gin + PostgreSQL — no stack changes
- **API compatibility:** Existing API contracts (routes, request/response shapes) must not change
- **No behavior changes:** Security fixes must not alter observable behavior for legitimate callers

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Raw SQL in DAO layer | Full SQL control; GORM used only for init/migrations | ✓ Good |
| In-process WebSocket hubs | Simpler, single-instance; not horizontally scalable | — Pending revisit |
| Polling-based AI job queue | No Redis/queue dep; 5s poll with semaphore concurrency | ✓ Good for current scale |
| Trust middleware for auth context | Middleware is the correct enforcement boundary; per-function DB re-checks are redundant | — Pending (active work) |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd:transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-03-29 after initialization*
