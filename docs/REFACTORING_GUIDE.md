# Quiz-Craft Backend Refactoring Guide

## 🎯 Objective

Refactor Quiz-Craft backend from GORM-based repository pattern to **Titan-style architecture** with:
- Raw SQL in DAO layer
- Separate DAO/API models
- Clear API → Core → DAO layering
- Trace ID propagation
- Domain error handling

---

## 📊 Architecture Comparison

### Before (Phase 2 - GORM)
```
Request
  ↓
Handler (internal/handlers)
  ↓
Service (internal/services) - Business logic
  ↓
Repository (internal/repositories) - GORM operations
  ↓
Model (internal/models) - Single model with GORM tags
  ↓
Database
```

### After (Titan Pattern - Raw SQL)
```
HTTP Request
  ↓
API Handler (api/user.go, api/auth.go)
  │ - Parse request & extract context/trace ID
  │ - Call CORE layer (NOT dao directly!)
  ↓
CORE Layer (user/core, auth/core)
  │ - Business validation & logic
  │ - Orchestrate multiple DAO calls
  │ - Transform DAO → API models
  │ - Call external services
  ↓
DAO Layer (user/dao, auth/dao)
  │ - Execute raw SQL
  │ - Map rows to DAO models
  │ - Return domain errors
  ↓
Database
```

---

## 🗂️ New Folder Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # UPDATED - Dual router pattern
│
├── api/                          # NEW - HTTP handlers
│   ├── user.go                  # User endpoints
│   └── auth.go                  # Auth endpoints
│
├── user/                         # NEW - User domain module
│   ├── core/
│   │   └── user.go              # User business logic
│   ├── dao/
│   │   ├── user.go              # User data access (raw SQL)
│   │   └── password.go          # Password data access
│   └── model/
│       ├── user.go              # User models (DAO + API)
│       └── password.go          # Password models
│
├── auth/                         # NEW - Auth domain module
│   ├── core/
│   │   └── auth.go              # Auth business logic
│   └── model/
│       └── auth.go              # Auth models
│
├── commons/                      # NEW - Shared utilities
│   ├── ctx.go                   # Context & trace ID management
│   ├── log.go                   # Structured logging with trace ID
│   ├── response.go              # Standard HTTP responses
│   └── errors.go                # Domain error types
│
├── constants/                    # NEW - Application constants
│   └── constants.go
│
├── internal/                     # KEEP - Internal packages
│   ├── config/                  # Keep as-is
│   ├── middleware/              # Update auth middleware
│   ├── utils/                   # Keep password, jwt, validator utils
│   └── workers/                 # Keep as-is
│
└── pkg/                          # KEEP - Database, Redis, etc.
    ├── database/
    └── redis/
```

---

## ✅ What's Been Created

### 1. Commons Package (`commons/`)
- ✅ **ctx.go**: Context management with trace ID
- ✅ **log.go**: Structured logging (`Log`, `LogW`, `LogE`, `LogC`)
- ✅ **response.go**: Standard HTTP responses (`RespOK`, `Resp404`, etc.)
- ✅ **errors.go**: Domain error types (`NotFound`, `BadInput`, `Conflict`, etc.)

### 2. Constants Package (`constants/`)
- ✅ **constants.go**: Application-wide constants

### 3. User Module (`user/`)

**Models (`user/model/`)**:
- ✅ **user.go**:
  - `UserDAO` - Database model with `sql.Null*` types
  - `UserAPI` - API response model with native types
  - `ToAPI()` method to convert DAO → API
  - Request/Response structs (Signup, Login, Update, List, Stats)

- ✅ **password.go**:
  - `UserPasswordDAO` - Database model
  - Password history structures
  - Request structs (VerifyEmail, ResetPassword, ChangePassword)

**DAO Layer (`user/dao/`)**:
- ✅ **user.go** - All user database operations:
  ```go
  - CreateUser()
  - FindByID()
  - FindByIDIncludingDeleted()
  - FindByEmail()
  - EmailExists()
  - EmailExistsIncludingDeleted()
  - UpdateUser()
  - UpdateEmailVerified()
  - UpdateEnabled()
  - SoftDelete()
  - Restore()
  - HardDelete()
  - ListUsers()
  - CountUsers()
  - CountDeletedUsers()
  ```

- ✅ **password.go** - All password database operations:
  ```go
  - CreatePassword()
  - FindPasswordByUserID()
  - UpdatePasswordHash()
  - SetVerificationToken()
  - ClearVerificationToken()
  - FindByVerificationToken()
  - SetResetToken()
  - ClearResetToken()
  - FindByResetToken()
  - IncrementFailedAttempts()
  - ResetFailedAttempts()
  - LockAccount()
  ```

---

## 🚧 What Needs to Be Done

### 1. User Core Layer (`user/core/user.go`) - PENDING

Create business logic layer that:
- Calls DAO functions
- Validates business rules
- Transforms DAO → API models
- Handles user operations

**Functions needed**:
```go
// User management
func GetUserByID(ctx context.Context, id uuid.UUID) (model.UserAPI, error)
func GetUserProfile(ctx context.Context, userID uuid.UUID) (model.UserAPI, error)
func UpdateUserProfile(ctx context.Context, userID uuid.UUID, req model.UpdateProfileRequest) error
func DeleteUser(ctx context.Context, userID uuid.UUID) error
func RestoreUser(ctx context.Context, userID uuid.UUID) error
func DisableUser(ctx context.Context, userID uuid.UUID) error
func EnableUser(ctx context.Context, userID uuid.UUID) error

// User listing
func ListUsers(ctx context.Context, filter model.UserListFilter) (model.UserListResponse, error)
func GetUserStats(ctx context.Context) (model.UserStatsResponse, error)
```

### 2. Auth Core Layer (`auth/core/auth.go`) - PENDING

Create authentication business logic:
- User signup
- Email verification
- Login with password check
- Password reset flow
- Token generation

**Functions needed**:
```go
func Signup(ctx context.Context, req model.SignupRequest) (model.SignupResponse, error)
func VerifyEmail(ctx context.Context, req model.VerifyEmailRequest) error
func ResendVerification(ctx context.Context, req model.ResendVerificationRequest) error
func Login(ctx context.Context, req model.LoginRequest) (model.LoginResponse, error)
func ForgotPassword(ctx context.Context, req model.ForgotPasswordRequest) error
func ResetPassword(ctx context.Context, req model.ResetPasswordRequest) error
func ChangePassword(ctx context.Context, userID uuid.UUID, req model.ChangePasswordRequest) error
func RefreshToken(ctx context.Context, refreshToken string) (model.LoginResponse, error)
```

### 3. API Handlers (`api/`) - PENDING

**`api/user.go`** - User HTTP handlers:
```go
func GetUserProfile(c *gin.Context)      // GET /users/me
func UpdateProfile(c *gin.Context)       // PATCH /users/me
func DeleteAccount(c *gin.Context)       // DELETE /users/me
func ListUsers(c *gin.Context)           // GET /users
func GetUserStats(c *gin.Context)        // GET /users/stats
func GetUserByID(c *gin.Context)         // GET /users/:id
func RestoreUser(c *gin.Context)         // POST /users/:id/restore
func DisableUser(c *gin.Context)         // POST /users/:id/disable
func EnableUser(c *gin.Context)          // POST /users/:id/enable
```

**`api/auth.go`** - Auth HTTP handlers:
```go
func Signup(c *gin.Context)               // POST /auth/signup
func VerifyEmail(c *gin.Context)          // POST /auth/verify-email
func ResendVerification(c *gin.Context)   // POST /auth/resend-verification
func Login(c *gin.Context)                // POST /auth/login
func RefreshToken(c *gin.Context)         // POST /auth/refresh
func Logout(c *gin.Context)               // POST /auth/logout
func ForgotPassword(c *gin.Context)       // POST /auth/forgot-password
func ResetPassword(c *gin.Context)        // POST /auth/reset-password
func ChangePassword(c *gin.Context)       // POST /auth/change-password
```

**Handler Pattern** (following Titan):
```go
func GetUserProfile(c *gin.Context) {
    st := time.Now()
    ctx := commons.Ctx(c)                    // Get context with trace ID
    tid := commons.Tid(ctx)                  // Extract trace ID
    c.Set(commons.TidRespHdr, tid)           // Set trace header
    ip := c.ClientIP()

    // Extract user ID from context (set by auth middleware)
    userID, ok := commons.UsrID(ctx)
    if !ok {
        commons.Resp401(c, "unauthorized")
        return
    }

    commons.Log(ctx, "[api.GetUserProfile [user_id=%s", userID)

    // Call CORE layer (NOT dao!)
    user, err := core.GetUserProfile(ctx, userID)
    if err != nil {
        commons.LogC(ctx, "[api.GetUserProfile::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
            time.Since(st), ip, userID, err)

        if commons.IsNotFound(err) {
            commons.Resp404(c, commons.GetCode(err))
            return
        }
        commons.RespISE(c, err)
        return
    }

    commons.Log(ctx, "[api.GetUserProfile::OK [time_taken=%v [ip=%s [user_id=%s",
        time.Since(st), ip, userID)
    commons.RespOKWithMsg(c, "Profile retrieved successfully", user)
}
```

### 4. Update main.go - PENDING

Implement dual router pattern:
```go
func main() {
    // Initialize database
    db := initDatabase()

    // Initialize DAOs
    userdao.Init(db)

    // Start routers
    initInt()  // Internal APIs (no auth)
    initExt()  // External APIs (with auth)
}

func initInt() {
    r := gin.New()
    r.Use(TraceIDMiddleware())

    v1 := r.Group("/api/v1")
    {
        // Internal endpoints (service-to-service)
        v1.GET("/health", healthCheck)
    }

    go r.Run(":8081") // Internal port
}

func initExt() {
    r := gin.New()
    r.Use(TraceIDMiddleware())

    v1 := r.Group("/api/v1")

    // Public auth endpoints
    auth := v1.Group("/auth")
    {
        auth.POST("/signup", api.Signup)
        auth.POST("/verify-email", api.VerifyEmail)
        auth.POST("/login", api.Login)
        auth.POST("/refresh", api.RefreshToken)
        auth.POST("/forgot-password", api.ForgotPassword)
        auth.POST("/reset-password", api.ResetPassword)
    }

    // Protected user endpoints
    users := v1.Group("/users")
    users.Use(middleware.Auth()) // JWT middleware
    {
        users.GET("/me", api.GetUserProfile)
        users.PATCH("/me", api.UpdateProfile)
        users.DELETE("/me", api.DeleteAccount)
        users.GET("", api.ListUsers)
        users.GET("/stats", api.GetUserStats)
        users.GET("/:id", api.GetUserByID)
        users.POST("/:id/restore", api.RestoreUser)
        users.POST("/:id/disable", api.DisableUser)
        users.POST("/:id/enable", api.EnableUser)
    }

    // Protected auth endpoints
    protectedAuth := v1.Group("/auth")
    protectedAuth.Use(middleware.Auth())
    {
        protectedAuth.POST("/change-password", api.ChangePassword)
        protectedAuth.POST("/logout", api.Logout)
    }

    r.Run(":8080") // External port
}
```

### 5. Update Auth Middleware - PENDING

Update `internal/middleware/auth.go` to use commons:
```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := commons.Ctx(c)

        // Extract token
        token := extractToken(c)
        if token == "" {
            commons.Resp401(c, "authorization-token-not-found")
            c.Abort()
            return
        }

        // Validate JWT
        claims, err := jwtManager.ValidateToken(token)
        if err != nil {
            commons.Resp401(c, "invalid-token")
            c.Abort()
            return
        }

        // Set user ID in context
        commons.SetUserID(c, claims.UserID)

        c.Next()
    }
}
```

### 6. Trace ID Middleware - NEW

Create `api/middleware.go`:
```go
func TraceIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := commons.Ctx(c)
        tid := commons.Tid(ctx)
        c.Set(commons.TidRespHdr, tid)
        c.Header(commons.TidRespHdr, tid)
        c.Next()
    }
}
```

---

## 🔑 Key Principles (from Titan)

### 1. **Layering**
- ✅ API handlers NEVER call DAO directly
- ✅ Always go through CORE layer
- ✅ CORE contains business logic, DAO contains SQL only

### 2. **Models**
- ✅ Separate DAO models (with `sql.Null*`) from API models (native types)
- ✅ DAO models map directly to database columns
- ✅ API models are for JSON responses

### 3. **Error Handling**
- ✅ DAO returns domain errors (`commons.NotFound`, `commons.BadInput`)
- ✅ CORE propagates or transforms errors
- ✅ API maps domain errors to HTTP status codes

### 4. **Logging**
- ✅ Every function logs with trace ID
- ✅ Log entry, errors, and success with timing
- ✅ Use structured format: `[layer.Function [param=value`

### 5. **Context**
- ✅ Always pass `context.Context` as first parameter
- ✅ Extract trace ID and user ID from context
- ✅ Propagate context through all layers

---

## 📝 Migration Checklist

- [x] Create commons package (ctx, log, response, errors)
- [x] Create constants package
- [x] Create user/model with DAO and API models
- [x] Create user/dao with raw SQL operations
- [ ] Create user/core with business logic
- [ ] Create auth/core with authentication logic
- [ ] Create api/user.go with HTTP handlers
- [ ] Create api/auth.go with HTTP handlers
- [ ] Update main.go with dual router pattern
- [ ] Update middleware/auth.go to use commons
- [ ] Create trace ID middleware
- [ ] Test all endpoints
- [ ] Update API documentation

---

## 🧪 Testing Strategy

After refactoring:

1. **Database Setup**:
   ```bash
   docker-compose up -d
   ```

2. **Build**:
   ```bash
   go build -o bin/server cmd/server/main.go
   ```

3. **Test Endpoints** (use existing `API_TESTING_GUIDE.md`):
   - All endpoints should work identically
   - Check logs for trace IDs
   - Verify soft delete still works

4. **Verify Logs**:
   - Every request should have trace ID
   - Timing logged for each operation
   - Errors include trace ID for debugging

---

## 💡 Benefits of New Architecture

1. **Clear Separation**: Each layer has single responsibility
2. **Testability**: DAO, Core, API can be tested independently
3. **Performance**: Raw SQL is faster than GORM for complex queries
4. **Traceability**: Trace IDs make debugging distributed systems easy
5. **Consistency**: Every module follows same pattern
6. **Scalability**: Easy to add new domains (quiz/, question/, etc.)
7. **Maintainability**: Clear where logic belongs

---

## 🚀 Next Steps

1. Complete User Core layer
2. Complete Auth Core layer
3. Create API handlers
4. Update main.go
5. Test thoroughly
6. Apply same pattern to Quiz module (Phase 3)

---

## 📚 Reference

- Titan project structure: `C:\Users\affan\OneDrive\Documents\quizMaster\titan`
- Current implementation: Phase 2 (GORM-based)
- Target: Titan-style (Raw SQL, 3-layer architecture)
