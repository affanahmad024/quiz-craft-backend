# QuizCraft Backend — API Reference (generated)

This is a concise reference for all public backend endpoints in the current repository.

Base path: `/api/v1`

Common response envelope:

```json
{
  "success": true,
  "message": "Human readable message",
  "data": {
    /* endpoint specific */
  }
}
```

Authentication

- The app uses JWT access/refresh tokens. `POST /auth/login` returns `access_token` and `refresh_token` in the JSON body and also sets `access_token` and `refresh_token` cookies. For browser clients, send requests with credentials (cookies) enabled.
- Protected endpoints require the Auth middleware; some endpoints require educator or admin roles.

Auth

- POST /auth/signup — Register new user
  - Body: {"email":"...","password":"...","name":"..."}
  - Response data: {"user_id":"..."}

- POST /auth/verify-email — Verify signup
  - Body: {"user_id":"...","token":"..."}

- POST /auth/resend-verification — Resend verification email
  - Body: {"email":"..."}

- POST /auth/login — Login
  - Body: {"email":"...","password":"..."}
  - Response data: {"access_token":"...","refresh_token":"...","expires_in":3600,"user":{...}}

- POST /auth/refresh — Refresh access token
  - Body: {"refresh_token":"..."}

- POST /auth/forgot-password — Initiate password reset
  - Body: {"email":"..."}

- POST /auth/reset-password — Complete password reset
  - Body: {"user_id":"...","token":"...","new_password":"..."}

- POST /auth/change-password (protected)
  - Body: {"old_password":"...","new_password":"..."}

- POST /auth/logout (protected)
  - Clears auth cookies

Users

- GET /users/me (protected) — Get current user profile
- PATCH /users/me (protected) — Update profile
  - Body: fields to update (see frontend model `UpdateProfileRequest`)
- DELETE /users/me (protected) — Soft delete account
- GET /users (protected/admin) — Admin list users (query: page, limit, role, enabled, search)
- GET /users/:id (protected) — Get user by id (admin)
- GET /users/stats (protected) — Get user stats (admin-level listing exists)

Quizzes

- POST /quizzes (protected) — Create quiz
  - Body: quiz object (title, description, questions optional)
- GET /quizzes (protected) — List my quizzes (query: page, limit, category, difficulty, is_published)
- GET /quizzes/public (protected) — List public quizzes (query filters; optional auth)
- GET /quizzes/:id (protected) — Get quiz by id
- GET /quizzes/:id/full (protected) — Get quiz with questions
  - Query: `hide_correct=true` to hide correct answers
- PATCH /quizzes/:id (protected) — Update quiz
- DELETE /quizzes/:id (protected) — Delete (soft) quiz
- POST /quizzes/:id/publish (protected) — Publish
- POST /quizzes/:id/unpublish (protected) — Unpublish
- POST /quizzes/:id/save-as-template (protected) — Save a quiz as template

Question & Answer management

- POST /quizzes/:id/questions (protected) — Add question
  - Accepts JSON or multipart/form-data (for optional image). JSON body uses `QuestionInput` model; multipart supports `question_data` JSON plus `image` file.
- POST /quizzes/:id/questions/bulk (protected) — Add multiple questions
  - Accepts JSON or multipart (image_0, image_1, ...)
- GET /quizzes/questions/:qid (protected) — Get question
- PATCH /quizzes/questions/:qid (protected) — Update question (JSON or multipart)
- DELETE /quizzes/questions/:qid (protected) — Delete question

- POST /quizzes/questions/:qid/answers (protected) — Add answer
- PATCH /quizzes/answers/:aid (protected) — Update answer
- DELETE /quizzes/answers/:aid (protected) — Delete answer

Attempts

- POST /quizzes/:id/attempts (protected) — Start attempt
- GET /attempts (protected) — List my attempts
- GET /attempts/:id (protected) — Get an attempt
- POST /attempts/:id/answers (protected) — Submit answer
- POST /attempts/:id/finish (protected) — Finish attempt
- GET /attempts/:id/results (protected) — Get attempt result

AI / Generation

- POST /ai/generate-quiz-from-text (protected) — Generate quiz from text
  - Body: {"text":"...","options":{...},"reference_images":[...]} — returns accepted (202) with generation_id
- POST /ai/generate-quiz-from-file (protected, multipart) — Upload file + options; returns generation job id (202)
- POST /ai/extract-quiz-from-file (protected, multipart) — Extract existing quiz questions from image/pdf (202)
- GET /ai/generation-status/:id (protected) — Get status of generation job
- GET /ai/jobs (protected) — List generation jobs

Quiz AI addition to existing quiz

- POST /quizzes/:id/add-questions-from-text (protected) — Add AI generated questions from text to quiz (202)
- POST /quizzes/:id/add-questions-from-file (protected, multipart) — Add AI-generated questions from an uploaded file (202)
- POST /quizzes/:id/extract-questions-from-file (protected, multipart) — Extract questions from file and add to quiz (202)

Files

- POST /files/upload (protected, multipart) — Upload file
  - Field: `file` — returns FileUploadResponse {file_id, filename, size, mime_type}
- GET /files (protected) — List files (page, limit)
- GET /files/serve/:id (protected) — Serve/download file (sets `Content-Type` and streams content)
- DELETE /files/:id (protected) — Delete file

Search & Discovery

- GET /quizzes/search (public) — Search quizzes (query params bind to `SearchQuizzesInput`)
- GET /categories (public) — Get categories
- GET /tags (public) — Get popular tags
- GET /share/:code (public) — Retrieve quiz by share code

Templates

- GET /templates (protected) — List templates (query: page, limit, category, my_templates)
- POST /templates/:id/use (protected) — Create quiz from template (body: `UseTemplateInput`)

Notifications

- GET /notifications (protected) — List notifications (query: unread_only, type, limit, offset)
- GET /notifications/unread-count (protected) — Unread count
- PATCH /notifications/:id/read (protected) — Mark single notification read
- PATCH /notifications/read-all (protected) — Mark all read
- DELETE /notifications/:id (protected) — Delete notification
- GET /notifications/ws (protected) — WebSocket endpoint for live notifications

Live Sessions (Realtime)

- POST /live-sessions (protected) — Create live session
- GET /live-sessions (protected) — List my sessions
- GET /live-sessions/public (protected) — List public sessions
- GET /live-sessions/invitations (protected) — List invitations
- GET /live-sessions/:id (protected) — Get session details
- POST /live-sessions/:id/join (protected) — Join session (body may include access_code)
- POST /live-sessions/:id/start (protected) — Start session (creator)
- POST /live-sessions/:id/advance (protected) — Advance to next question (creator)
- POST /live-sessions/:id/end (protected) — End session (creator)
- POST /live-sessions/:id/answer (protected) — Submit live answer
- GET /live-sessions/:id/leaderboard (protected) — Get leaderboard
- GET /live-sessions/:id/results (protected) — Get results after session
- POST /live-sessions/:id/invite (protected) — Invite users
- GET /live-sessions/:id/ws (protected) — WebSocket endpoint for live session

Educator features

- POST /educator/promote (protected) — Promote current user to educator
- GET /educator/profile (protected) — Get my educator profile
- PATCH /educator/profile (protected) — Update my educator profile
- GET /educators (public) — List educators
- GET /educators/search (public) — Search educators
- GET /educators/featured (public) — Featured educators
- GET /educators/:id (public) — Get educator profile
- POST /educators/:id/follow (protected) — Follow educator
- DELETE /educators/:id/unfollow (protected) — Unfollow educator
- GET /following (protected) — Get my following
- GET /educator/followers (protected, educator-only) — My followers
- GET /educator/followers/stats (protected, educator-only) — Follower stats

Admin (requires `RequireAdmin` middleware)

- GET /admin/users — List all users (admin)
- GET /admin/users/:id — Get user details
- PATCH /admin/users/:id/role — Change user role
- POST /admin/users/:id/disable — Disable user (body: reason)
- POST /admin/users/:id/enable — Enable user
- POST /admin/users/:id/restore — Restore soft-deleted user
- DELETE /admin/users/:id — Permanently delete user (`?permanent=true` required)
- GET /admin/quizzes — List all quizzes (admin)
- GET /admin/quizzes/:id — Get quiz detail (admin)
- DELETE /admin/quizzes/:id — Delete quiz (admin)
- GET /admin/stats — Platform statistics
- GET /admin/logs — Activity logs

Integration notes & frontend checklist

- Base URL: set `REACT_APP_API_BASE` (or equivalent) to `https://<api-host>/api/v1` in frontend env.
- CORS: server uses `middleware.CORS(&cfg.CORS)`. Ensure frontend origin is included in backend config `config/config.yaml`.
- Auth (browser): The backend sets `access_token` and `refresh_token` cookies. For browser requests use `fetch`/axios with `credentials: 'include'` or `withCredentials: true` so cookies are sent.
- Auth (mobile/API clients): Use the `Authorization: Bearer <access_token>` header with token obtained from login/refresh endpoints.
- Refresh tokens: Call `POST /auth/refresh` if access token expires; server also issues cookies on login flow.
- Websockets: Use the WebSocket endpoints for real-time notifications and live sessions. You must authenticate the socket connection (either by sending `access_token` in a query param or sending a session cookie; adjust the frontend to match server-side WS auth implementation).

Quick dev / deploy checklist

1. Ensure `config/config.yaml` contains correct CORS origins and mailjet/email settings.
2. Build and run backend: `go build ./...` then `./quizcraft` (or use Dockerfile in `backend/`).
3. Configure frontend `REACT_APP_API_BASE` and run `npm run dev` / build.
4. Test login -> create quiz -> publish -> attempt flows end-to-end.

If you want, I can:

- Generate a machine-readable OpenAPI spec from these routes (partial/manual conversion).
- Update the frontend `src/api` clients to use these endpoints and enable credentials.
- Add example Postman collection and example curl commands for common flows.

File created: `backend/docs/API_REFERENCE.md`
