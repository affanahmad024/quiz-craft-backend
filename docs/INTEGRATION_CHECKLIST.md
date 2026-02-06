# Integration Checklist — Frontend ↔ Backend

This checklist helps make the `quiz-craft-frontend` and backend integrate reliably and quickly for users.

1. Environment & base URLs

- Frontend env: ensure `VITE_API_BASE_URL` points to the backend base (example `https://api.example.com/api/v1`).
- Development default: `http://localhost:8080/api/v1` (see `.env.development` in frontend).

2. CORS and cookies

- Add your frontend origin to backend `config/config.yaml` CORS allowed origins.
- For browser clients, ensure frontend HTTP clients set `withCredentials: true` / `credentials: 'include'` so auth cookies are sent.
- In production, set cookie `Secure=true` (backend currently sets Secure=false for local dev; enable secure cookies behind HTTPS).

3. Auth flows

- Use `POST /auth/login` to sign-in. The server sets `access_token` and `refresh_token` cookies and returns tokens in JSON.
- For single-page app flows, prefer cookie-based auth (cookies are HttpOnly so store access token only in cookies + optionally mirror in localStorage if needed).
- Implement client-side token refresh: either call `POST /auth/refresh` on 401 or rely on cookie session with short-lived access tokens.

4. API client configuration (frontend)

- Use the central `useNetReq` hook (exists at `src/hooks/NetworkRequest.js`) — it already supports `withCredentials: true` and `VITE_API_BASE_URL`.
- Ensure all existing calls pass relative `url` values that the hook expects (example `auth/login` or `quizzes` — hook concatenates base + `/${url}`).

5. WebSocket integration

- Notification WS: `/api/v1/notifications/ws` — frontend currently attaches `?token=<access_token>`; that works but using cookie authentication is preferable for browsers.
- Live-session WS: `/api/v1/live-sessions/:id/ws` — same auth considerations; ensure the client sends a token or cookies at connection time.

6. Files & images

- Static image serving is via `/api/v1/files/serve/:id`. Frontend components already rewrite `/api/v1` relative URLs to the configured base.

7. Frontend navigation & discoverability

- Ensure top-level navigation includes: Create Quiz, My Quizzes, Browse Public Quizzes, AI Tools, Live Sessions, Notifications, Educators, Profile, Admin (conditionally).
- Expose templates in a Templates browse page and 'Use Template' button that calls `POST /templates/:id/use`.

8. Tests & verification (quick smoke)

- Dev steps:
  - Start backend: `go run main.go` (or build + run binary). Ensure DB migrations ran.
  - Start frontend: `npm install` then `npm run dev` in `quiz-craft-frontend`.
  - Register a new user → verify email flow (use mailjet config or mock)
  - Login, create a quiz, add questions, publish quiz, take attempt, check attempt results.
  - Start a live session, join another user, verify websocket messages for question reveal and leaderboard.

9. Deploy notes

- Use the provided `Dockerfile` and `docker-compose.prod.yml` in `backend/` if deploying with Docker.
- Ensure secrets (JWT secret, mailjet keys, DB credentials) are provided from secure env or secret manager.

10. Next improvements I can implement for you (pick any):

- Generate an OpenAPI (swagger) spec from the API and add interactive docs.
- Create Postman collection / curl examples for the most common flows.
- Implement central axios instance and replace direct fetch calls in the frontend to ensure consistent headers and error handling.
- Harden cookie security and CSRF protection for production.
