# Backend Documentation

Welcome to the QuizCraft Backend documentation!

---

## 📚 Documentation Structure

### For Frontend Developers

1. **[COMPLETE_API_REFERENCE.md](COMPLETE_API_REFERENCE.md)** ⭐ **START HERE**
   - **All 105 API endpoints** with complete request/response formats
   - Authentication, Users, Quizzes, Questions, Attempts
   - AI Generation, Notifications, Live Sessions, Educators
   - Search, Templates, Admin endpoints
   - Error codes and examples
   - **This is your single source of truth for the API**

2. **[FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md)**
   - Practical integration guide with **code examples**
   - React hooks and components
   - WebSocket connections (Notifications, Live Sessions)
   - Authentication flow with token refresh
   - File uploads and error handling
   - Ready-to-use service modules
   - **Copy-paste friendly examples**

### For Backend Developers

3. **[QUICKSTART.md](QUICKSTART.md)**
   - How to run the backend locally
   - Database setup and migrations
   - Environment configuration
   - Troubleshooting guide
   - **Start here if you're new to the project**

4. **[EDUCATOR_ROLE_IMPLEMENTATION.md](EDUCATOR_ROLE_IMPLEMENTATION.md)**
   - Architecture and design decisions
   - Database schema for educators, live sessions, notifications
   - WebSocket infrastructure
   - Implementation phases and roadmap
   - **Reference for understanding system architecture**

---

## 🚀 Quick Links

### Getting Started
- New to the project? → [QUICKSTART.md](QUICKSTART.md)
- Building a frontend? → [FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md)
- Need API reference? → [COMPLETE_API_REFERENCE.md](COMPLETE_API_REFERENCE.md)

### Common Tasks

**Authentication:**
- Signup/Login → [COMPLETE_API_REFERENCE.md#authentication](COMPLETE_API_REFERENCE.md#authentication)
- Token Refresh → [FRONTEND_INTEGRATION.md#authentication-flow](FRONTEND_INTEGRATION.md#authentication-flow)

**Quiz Management:**
- Create/Update Quiz → [COMPLETE_API_REFERENCE.md#quizzes](COMPLETE_API_REFERENCE.md#quizzes)
- Add Questions → [COMPLETE_API_REFERENCE.md#questions--answers](COMPLETE_API_REFERENCE.md#questions--answers)
- Quiz Attempts → [COMPLETE_API_REFERENCE.md#quiz-attempts](COMPLETE_API_REFERENCE.md#quiz-attempts)

**AI Features:**
- Generate Quiz from Text/File → [COMPLETE_API_REFERENCE.md#ai-quiz-generation](COMPLETE_API_REFERENCE.md#ai-quiz-generation)
- Extract from Image → [COMPLETE_API_REFERENCE.md#ai-quiz-generation](COMPLETE_API_REFERENCE.md#ai-quiz-generation)

**Real-time Features:**
- Notifications → [COMPLETE_API_REFERENCE.md#notifications](COMPLETE_API_REFERENCE.md#notifications)
- Live Quiz Sessions → [COMPLETE_API_REFERENCE.md#live-quiz-sessions](COMPLETE_API_REFERENCE.md#live-quiz-sessions)
- WebSocket Integration → [FRONTEND_INTEGRATION.md#notifications-real-time](FRONTEND_INTEGRATION.md#notifications-real-time)

**Educator Features:**
- Become Educator → [COMPLETE_API_REFERENCE.md#educators](COMPLETE_API_REFERENCE.md#educators)
- Follow System → [COMPLETE_API_REFERENCE.md#educators](COMPLETE_API_REFERENCE.md#educators)

---

## 📋 API Overview

### Endpoint Summary

| Module | Endpoints | Description |
|--------|-----------|-------------|
| **Authentication** | 9 | Signup, Login, Password management |
| **Users** | 7 | Profile management, Statistics |
| **Quizzes** | 13 | Create, Update, Publish quizzes |
| **Questions** | 8 | Add, Update questions and answers |
| **Attempts** | 6 | Take quizzes, Submit answers |
| **AI Generation** | 8 | Generate/Extract quizzes using AI |
| **Files** | 4 | Upload, Serve, Delete files |
| **Notifications** | 6 | Real-time notifications + WebSocket |
| **Live Sessions** | 14 | Real-time competitive quizzes |
| **Educators** | 13 | Educator profiles, Follow system |
| **Search** | 4 | Search quizzes, Categories, Tags |
| **Templates** | 2 | Quiz templates |
| **Admin** | 14 | User/Quiz management, Platform stats |
| **Total** | **105** | |

---

## 🔐 Authentication

All protected endpoints require a Bearer token:

```bash
Authorization: Bearer {access_token}
```

Get your token by calling:
```
POST /api/v1/auth/login
```

---

## 🌐 Base URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.quizcraft.com/api/v1
```

---

## 📖 Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": "error-code",
  "message": "Error description"
}
```

---

## 🛠️ Tools & Technologies

- **Language:** Go 1.21+
- **Framework:** Gin
- **Database:** PostgreSQL 15+
- **Cache:** Redis (optional)
- **AI:** Google Gemini API
- **WebSocket:** gorilla/websocket
- **Docs:** Swagger/OpenAPI

---

## 📦 Project Structure

```
backend/
├── api/               # HTTP handlers (11 files)
├── auth/              # Authentication logic
├── user/              # User management
├── quiz/              # Quiz management
├── ai-agent/          # AI quiz generation
├── notification/      # Notification system
├── livesession/       # Live quiz sessions
├── educator/          # Educator features
├── internal/
│   ├── middleware/    # JWT, CORS, logging
│   ├── services/      # Email, AI services
│   └── utils/         # JWT, password utils
├── pkg/
│   ├── database/      # DB connection, migrations
│   └── logger/        # Structured logging
└── docs/              # 📚 THIS FOLDER
```

---

## 🔄 Recent Updates (Feb 2026)

### Documentation Overhaul ✅
- ✅ Created **COMPLETE_API_REFERENCE.md** with all 105 endpoints
- ✅ Created **FRONTEND_INTEGRATION.md** with practical examples
- ✅ Removed 7 redundant/outdated documentation files
- ✅ Consolidated API documentation into single source of truth

### Implemented Features
- ✅ Real-time notifications with WebSocket
- ✅ Live quiz sessions with leaderboards
- ✅ Educator role and follow system
- ✅ AI quiz generation from text/files
- ✅ AI quiz extraction from images
- ✅ File upload and serving
- ✅ Admin dashboard

---

## 🆘 Support

- **Issues:** Check [QUICKSTART.md](QUICKSTART.md) troubleshooting section
- **API Questions:** See [COMPLETE_API_REFERENCE.md](COMPLETE_API_REFERENCE.md)
- **Integration Help:** See [FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md)
- **GitHub:** [Report Issues](https://github.com/your-repo/issues)

---

## 📝 Contributing

When adding new endpoints:
1. Update [COMPLETE_API_REFERENCE.md](COMPLETE_API_REFERENCE.md)
2. Add integration examples to [FRONTEND_INTEGRATION.md](FRONTEND_INTEGRATION.md)
3. Update this README's endpoint count

---

**Last Updated:** February 7, 2026
**API Version:** 2.0
**Total Endpoints:** 105
