# Technology Stack

**Analysis Date:** 2026-03-29

## Primary Language & Runtime

**Language:** Go 1.24.1
- Module: `github.com/affanahmad024/quiz-craft/backend`
- Single binary output, compiled with `CGO_ENABLED=0` for static Linux/amd64 builds
- Entry point: `main.go` at repo root

**Runtime:**
- Bare Alpine Linux (alpine:3.19) in production Docker image
- Development: native Go toolchain on host

## Frameworks & Libraries

**HTTP Framework:**
- `github.com/gin-gonic/gin v1.11.0` — Core HTTP router and middleware framework
- `github.com/gin-contrib/cors v1.7.6` — CORS middleware for Gin

**WebSocket:**
- `github.com/gorilla/websocket v1.5.3` — WebSocket upgrades for real-time notifications and live sessions

**Database / ORM:**
- `gorm.io/gorm v1.31.1` — ORM used for database initialization and connection management
- `gorm.io/driver/postgres v1.6.0` — GORM PostgreSQL driver
- `github.com/lib/pq v1.10.9` — Raw `database/sql` PostgreSQL driver (DAO layer uses raw SQL, not GORM)
- `github.com/golang-migrate/migrate/v4 v4.19.1` — SQL migration runner (`pkg/database/migrations/`)

**Authentication:**
- `github.com/golang-jwt/jwt/v5 v5.3.0` — JWT access and refresh token generation/validation (HS256)
- `golang.org/x/crypto v0.46.0` — bcrypt password hashing (cost=12)

**AI / ML:**
- `github.com/google/generative-ai-go v0.20.1` — Google Gemini AI SDK; used for quiz generation and vision extraction
- Default model: `gemini-1.5-pro` (configurable via `AIService.Model`)

**Email:**
- `github.com/mailjet/mailjet-apiv3-go/v4 v4.0.8` — Mailjet transactional email API (replaces direct SMTP; avoids cloud provider port blocks)

**File Parsing:**
- `github.com/ledongthuc/pdf v0.0.0-20250511090121-5959a4027728` — PDF text extraction
- `github.com/nguyenthenguyen/docx v0.0.0-20230621112118-9c8e795a11db` — DOCX text extraction

**Configuration:**
- `github.com/spf13/viper v1.21.0` — Config loading from YAML + env vars with `AutomaticEnv()`
- `github.com/joho/godotenv v1.5.1` — `.env` file loader (optional, loaded at startup)

**Logging:**
- `github.com/sirupsen/logrus v1.9.3` — Structured logger; text format in dev, JSON format in production

**Utilities:**
- `github.com/google/uuid v1.6.0` — UUID generation for user IDs, token IDs, etc.

## Build & Toolchain

**Package Manager:** Go modules (`go.mod` + `go.sum`)
- No external package manager; all deps are managed via `go mod`

**Build Command (production):**
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o quizcraft-api .
```

**Docker:** Multi-stage build
- Stage 1: `golang:1.21-alpine` — builds the binary
- Stage 2: `alpine:3.19` — runtime-only image with the static binary
- Dockerfile: `Dockerfile`

**Container Orchestration:**
- `deployments/docker/docker-compose.yml` — development stack
- `docker-compose.prod.yml` — production stack (Postgres 15, API, Nginx 1.25, Redis 7)

**Binaries in repo:**
- `backend.exe`, `quiz-craft.exe`, `quiz-craft-server`, `bin/quizcraft` — pre-built artifacts (Windows + Linux)

## Configuration & Environment

**Config loading priority** (highest to lowest):
1. Environment variables (bound explicitly via `viper.BindEnv`)
2. `config/config.yaml` — YAML config file (committed to repo with dev defaults)
3. `.env` file — loaded via `godotenv` if present (gitignored)
4. Viper defaults set in `setDefaults()` in `internal/config/config.go`

**Config struct:** `internal/config/config.go` — typed `Config` struct with sub-structs:
- `ServerConfig` — host, port, gin mode
- `DatabaseConfig` — PostgreSQL connection pool settings; DSN built via `GetDSN()`
- `JWTConfig` — secret, access/refresh token durations
- `EmailConfig` — Mailjet API key + secret, sender identity, frontend URL
- `AppConfig` — name, env (development/staging/production), log level
- `CORSConfig` — allowed origins/methods/headers
- `RateLimitConfig` — requests/window (config only, not enforced by middleware yet)
- `FileUploadConfig` — max size (10 MB default), upload dir, allowed MIME types
- `GoogleAIConfig` — API key

**Key env vars:**
- `SERVER_PORT` / `PORT` — HTTP port (default `8080`)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` — PostgreSQL
- `JWT_SECRET` — required; no default
- `GOOGLE_AI_API_KEY` — required for AI features
- `MAILJET_API_KEY`, `MAILJET_SECRET_KEY` — required for email
- `SMTP_FROM_EMAIL`, `SMTP_FROM_NAME`, `FRONTEND_URL` — email sender identity and link base URL
- `GIN_MODE` — `debug` (dev) / `release` (prod)
- `APP_ENV` — switches logger format to JSON in production
- `LOG_LEVEL` — `debug`, `info`, `warn`, `error`
- `CORS_ALLOWED_ORIGINS` — comma-separated list

**Example files:**
- `.env.example` — development defaults
- `.env.production.example` — production checklist with all required vars

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `gin-gonic/gin` | v1.11.0 | HTTP router |
| `gorm.io/gorm` | v1.31.1 | ORM / DB init |
| `golang-jwt/jwt/v5` | v5.3.0 | JWT auth |
| `golang.org/x/crypto` | v0.46.0 | bcrypt passwords |
| `google/generative-ai-go` | v0.20.1 | Gemini AI integration |
| `mailjet/mailjet-apiv3-go/v4` | v4.0.8 | Transactional email |
| `gorilla/websocket` | v1.5.3 | Real-time WebSocket |
| `golang-migrate/migrate/v4` | v4.19.1 | DB migrations |
| `spf13/viper` | v1.21.0 | Configuration |
| `sirupsen/logrus` | v1.9.3 | Structured logging |
| `ledongthuc/pdf` | latest | PDF parsing |
| `nguyenthenguyen/docx` | latest | DOCX parsing |
| `google/uuid` | v1.6.0 | UUID generation |

---

*Stack analysis: 2026-03-29*
