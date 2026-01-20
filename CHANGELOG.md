# Changelog

All notable changes to the QuizCraft Backend will be documented in this file.

## [1.1.0] - 2026-01-20

### Added
- **Bulk Question Creation**: New endpoint `POST /api/v1/quizzes/:id/questions/bulk` to add multiple questions with answers in a single atomic transaction
  - Automatic order indexing
  - Full validation before creation
  - All-or-nothing transaction support
  - No limit on number of questions per request

- **Latest Attempt Information in Quiz Listings**: Quiz list endpoints now include user's latest attempt data
  - `GET /api/v1/quizzes` (My Quizzes) includes latest attempt info
  - `GET /api/v1/quizzes/public` (Public Quizzes) includes latest attempt info for authenticated users
  - Shows: attempt_id, status, score, percentage, passed, timestamps
  - Helps users track progress and decide which quizzes to retake

- **Admin User Restore Endpoint**: New endpoint `POST /admin/users/:id/restore` to restore soft-deleted users
  - Admin-only access
  - Activity logging for audit trail

### Fixed
- **Security: User Management Endpoints**: Removed security vulnerability where regular users could access admin-only user management endpoints
  - Removed `POST /users/:id/disable` from general users group
  - Removed `POST /users/:id/enable` from general users group
  - Removed `POST /users/:id/restore` from general users group
  - All user management operations now properly restricted to admin endpoints only
  - Added admin authorization checks in core functions

### Changed
- Updated `ListMyQuizzes` core function to fetch and attach latest attempt information
- Updated `ListPublicQuizzes` core function to fetch and attach latest attempt information for authenticated users
- Modified Quiz model to include optional `LatestAttempt` field
- Updated API documentation with new endpoints and security notes

### Technical
- Added `LatestAttemptInfo` model for attempt summary data
- Added `GetLatestAttemptsByQuizIDs` DAO function for efficient batch fetching
- Added `CreateQuestionsBulk` DAO function with transaction support
- Added `AddQuestionsBulk` core function with validation
- Added `BulkQuestionsInput` model for bulk question creation
- Updated admin routes with proper separation from user routes

### Documentation
- Updated API_DOCUMENTATION.md with new endpoints and detailed examples
- Updated README.md with recent enhancements section
- Added security notes for admin endpoints
- Removed outdated phase implementation documents
- Removed outdated IMPLEMENTATION_PLAN.md
- Removed outdated CURRENT_STATE.md
- Removed outdated approach-document.md
- Created CHANGELOG.md for tracking changes

## [1.0.0] - 2026-01-10

### Initial Release
- Complete quiz management platform with 8 phases fully implemented
- 64 REST API endpoints
- JWT authentication with email verification
- AI-powered quiz generation (Google Gemini)
- File upload support (PDF, DOCX, TXT)
- Quiz attempts with auto-submit on expiry
- Real-time scoring system
- Search and discovery features
- Admin dashboard with user and quiz management
- Platform statistics and activity logging
- 69 unit tests
- Swagger API documentation
- Docker deployment support
- Production-ready with Nginx reverse proxy
