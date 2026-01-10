package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/pkg/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestDB holds test database connection
var TestDB *sql.DB

// SetupTestDB initializes a test database connection
func SetupTestDB(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "quiz_craft_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	err := database.InitPostgres(cfg)
	require.NoError(t, err, "Failed to connect to test database")

	TestDB = database.GetSQLDB()
	require.NotNil(t, TestDB, "Test database connection is nil")
}

// TeardownTestDB closes test database connection
func TeardownTestDB(t *testing.T) {
	if TestDB != nil {
		err := database.Close()
		if err != nil {
			t.Logf("Warning: Failed to close test database: %v", err)
		}
	}
}

// CleanupTestData removes all test data from database
func CleanupTestData(t *testing.T) {
	tables := []string{
		"user_answers",
		"quiz_attempts",
		"activity_logs",
		"questions",
		"quizzes",
		"email_verifications",
		"users",
	}

	for _, table := range tables {
		_, err := TestDB.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to clean table %s: %v", table, err)
		}
	}
}

// CreateTestUser creates a test user and returns the ID
func CreateTestUser(t *testing.T, email, name, password, role string) uuid.UUID {
	userID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO users (id, email, name, password, role, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, $6, $7)
	`

	_, err := TestDB.Exec(query, userID, email, name, password, role, now, now)
	require.NoError(t, err, "Failed to create test user")

	return userID
}

// CreateTestQuiz creates a test quiz and returns the ID
func CreateTestQuiz(t *testing.T, creatorID uuid.UUID, title string, isPublished, isPublic bool, passingScore int) uuid.UUID {
	quizID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO quizzes (id, creator_id, title, description, category, difficulty, is_published, is_public, passing_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := TestDB.Exec(query,
		quizID,
		creatorID,
		title,
		"Test quiz description",
		"General",
		"easy",
		isPublished,
		isPublic,
		passingScore,
		now,
		now,
	)
	require.NoError(t, err, "Failed to create test quiz")

	return quizID
}

// CreateTestQuestion creates a test question and returns the ID
func CreateTestQuestion(t *testing.T, quizID uuid.UUID, questionType, text string, points int, correctAnswer interface{}) uuid.UUID {
	questionID := uuid.New()
	now := time.Now()

	var correctAnswerJSON interface{}
	switch questionType {
	case "multiple_choice":
		correctAnswerJSON = map[string]interface{}{"correct_index": correctAnswer}
	case "true_false":
		correctAnswerJSON = map[string]interface{}{"correct": correctAnswer}
	}

	query := `
		INSERT INTO questions (id, quiz_id, type, text, options, correct_answer, points, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	options := []string{"Option A", "Option B", "Option C", "Option D"}
	if questionType == "true_false" {
		options = []string{"True", "False"}
	}

	_, err := TestDB.Exec(query,
		questionID,
		quizID,
		questionType,
		text,
		options,
		correctAnswerJSON,
		points,
		now,
		now,
	)
	require.NoError(t, err, "Failed to create test question")

	return questionID
}

// CreateTestAttempt creates a test quiz attempt and returns the ID
func CreateTestAttempt(t *testing.T, userID, quizID uuid.UUID, status string) uuid.UUID {
	attemptID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO quiz_attempts (id, user_id, quiz_id, status, started_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := TestDB.Exec(query, attemptID, userID, quizID, status, now, now, now)
	require.NoError(t, err, "Failed to create test attempt")

	return attemptID
}

// CreateTestAnswer creates a test user answer
func CreateTestAnswer(t *testing.T, attemptID, questionID uuid.UUID, answer interface{}, isCorrect bool, pointsEarned int) {
	answerID := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO user_answers (id, attempt_id, question_id, answer, is_correct, points_earned, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := TestDB.Exec(query, answerID, attemptID, questionID, answer, isCorrect, pointsEarned, now, now)
	require.NoError(t, err, "Failed to create test answer")
}

// GetContext returns a test context with timeout
func GetContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}
