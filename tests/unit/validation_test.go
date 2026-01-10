package unit

import (
	"testing"

	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/stretchr/testify/assert"
)

// ====== Quiz Validation Tests ======

// TestQuizValidate_ValidQuiz tests validation passes for valid quiz
func TestQuizValidate_ValidQuiz(t *testing.T) {
	quiz := &model.Quiz{
		Title:        "Valid Quiz Title",
		Description:  "A valid quiz description",
		Category:     "Science",
		Difficulty:   "medium",
		PassingScore: 70,
	}

	err := quiz.Validate()
	assert.NoError(t, err)
}

// TestQuizValidate_ShortTitle tests validation fails for too short title
func TestQuizValidate_ShortTitle(t *testing.T) {
	quiz := &model.Quiz{
		Title:        "Ab",
		PassingScore: 70,
	}

	err := quiz.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title must be at least 3 characters")
}

// TestQuizValidate_LongTitle tests validation fails for too long title
func TestQuizValidate_LongTitle(t *testing.T) {
	longTitle := make([]byte, 256)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	quiz := &model.Quiz{
		Title:        string(longTitle),
		PassingScore: 70,
	}

	err := quiz.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title must not exceed 255 characters")
}

// TestQuizValidate_LongDescription tests validation fails for too long description
func TestQuizValidate_LongDescription(t *testing.T) {
	longDesc := make([]byte, 2001)
	for i := range longDesc {
		longDesc[i] = 'a'
	}

	quiz := &model.Quiz{
		Title:        "Valid Title",
		Description:  string(longDesc),
		PassingScore: 70,
	}

	err := quiz.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description must not exceed 2000 characters")
}

// TestQuizValidate_InvalidDifficulty tests validation fails for invalid difficulty
func TestQuizValidate_InvalidDifficulty(t *testing.T) {
	quiz := &model.Quiz{
		Title:        "Valid Title",
		Difficulty:   "super_hard",
		PassingScore: 70,
	}

	err := quiz.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "difficulty must be 'easy', 'medium', or 'hard'")
}

// TestQuizValidate_ValidDifficulties tests all valid difficulty values
func TestQuizValidate_ValidDifficulties(t *testing.T) {
	difficulties := []string{"easy", "medium", "hard", ""}

	for _, diff := range difficulties {
		quiz := &model.Quiz{
			Title:        "Valid Title",
			Difficulty:   diff,
			PassingScore: 70,
		}

		err := quiz.Validate()
		assert.NoError(t, err, "Should accept difficulty: %s", diff)
	}
}

// TestQuizValidate_InvalidTimeLimit tests validation fails for invalid time limit
func TestQuizValidate_InvalidTimeLimit(t *testing.T) {
	testCases := []struct {
		name      string
		timeLimit int
	}{
		{"Zero time limit", 0},
		{"Negative time limit", -5},
		{"Too long time limit", 301},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			quiz := &model.Quiz{
				Title:            "Valid Title",
				TimeLimitMinutes: &tc.timeLimit,
				PassingScore:     70,
			}

			err := quiz.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "time limit must be between 1 and 300 minutes")
		})
	}
}

// TestQuizValidate_ValidTimeLimit tests validation passes for valid time limits
func TestQuizValidate_ValidTimeLimit(t *testing.T) {
	validTimeLimits := []int{1, 30, 60, 120, 300}

	for _, limit := range validTimeLimits {
		limitCopy := limit
		quiz := &model.Quiz{
			Title:            "Valid Title",
			TimeLimitMinutes: &limitCopy,
			PassingScore:     70,
		}

		err := quiz.Validate()
		assert.NoError(t, err, "Should accept time limit: %d", limit)
	}
}

// TestQuizValidate_InvalidPassingScore tests validation fails for invalid passing score
func TestQuizValidate_InvalidPassingScore(t *testing.T) {
	testCases := []struct {
		name         string
		passingScore int
	}{
		{"Negative passing score", -1},
		{"Over 100 passing score", 101},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			quiz := &model.Quiz{
				Title:        "Valid Title",
				PassingScore: tc.passingScore,
			}

			err := quiz.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "passing score must be between 0 and 100")
		})
	}
}

// TestQuizValidate_ValidPassingScore tests validation passes for valid passing scores
func TestQuizValidate_ValidPassingScore(t *testing.T) {
	validScores := []int{0, 50, 70, 100}

	for _, score := range validScores {
		quiz := &model.Quiz{
			Title:        "Valid Title",
			PassingScore: score,
		}

		err := quiz.Validate()
		assert.NoError(t, err, "Should accept passing score: %d", score)
	}
}

// TestQuizValidate_LongCategory tests validation fails for too long category
func TestQuizValidate_LongCategory(t *testing.T) {
	longCategory := make([]byte, 101)
	for i := range longCategory {
		longCategory[i] = 'a'
	}

	quiz := &model.Quiz{
		Title:        "Valid Title",
		Category:     string(longCategory),
		PassingScore: 70,
	}

	err := quiz.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "category must not exceed 100 characters")
}

// ====== Question Validation Tests ======

// TestQuestionValidate_ValidQuestion tests validation passes for valid question
func TestQuestionValidate_ValidQuestion(t *testing.T) {
	question := &model.Question{
		QuestionText: "What is 2+2?",
		QuestionType: "multiple_choice",
		Points:       10,
		OrderIndex:   0,
		Explanation:  "Addition of two and two",
	}

	err := question.Validate()
	assert.NoError(t, err)
}

// TestQuestionValidate_EmptyText tests validation fails for empty question text
func TestQuestionValidate_EmptyText(t *testing.T) {
	question := &model.Question{
		QuestionText: "",
		QuestionType: "multiple_choice",
		Points:       10,
	}

	err := question.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question text is required")
}

// TestQuestionValidate_LongText tests validation fails for too long question text
func TestQuestionValidate_LongText(t *testing.T) {
	longText := make([]byte, 1001)
	for i := range longText {
		longText[i] = 'a'
	}

	question := &model.Question{
		QuestionText: string(longText),
		QuestionType: "multiple_choice",
		Points:       10,
	}

	err := question.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question text must not exceed 1000 characters")
}

// TestQuestionValidate_InvalidType tests validation fails for invalid question type
func TestQuestionValidate_InvalidType(t *testing.T) {
	question := &model.Question{
		QuestionText: "Valid question?",
		QuestionType: "essay",
		Points:       10,
	}

	err := question.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "question type must be")
}

// TestQuestionValidate_ValidTypes tests all valid question types
func TestQuestionValidate_ValidTypes(t *testing.T) {
	types := []string{"multiple_choice", "true_false", "short_answer"}

	for _, qType := range types {
		question := &model.Question{
			QuestionText: "Valid question?",
			QuestionType: qType,
			Points:       10,
		}

		err := question.Validate()
		assert.NoError(t, err, "Should accept question type: %s", qType)
	}
}

// TestQuestionValidate_InvalidPoints tests validation fails for invalid points
func TestQuestionValidate_InvalidPoints(t *testing.T) {
	testCases := []struct {
		name   string
		points int
	}{
		{"Zero points", 0},
		{"Negative points", -5},
		{"Too many points", 101},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			question := &model.Question{
				QuestionText: "Valid question?",
				QuestionType: "multiple_choice",
				Points:       tc.points,
			}

			err := question.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "points must be between 1 and 100")
		})
	}
}

// TestQuestionValidate_ValidPoints tests validation passes for valid points
func TestQuestionValidate_ValidPoints(t *testing.T) {
	validPoints := []int{1, 5, 10, 50, 100}

	for _, points := range validPoints {
		question := &model.Question{
			QuestionText: "Valid question?",
			QuestionType: "multiple_choice",
			Points:       points,
		}

		err := question.Validate()
		assert.NoError(t, err, "Should accept points: %d", points)
	}
}

// TestQuestionValidate_NegativeOrderIndex tests validation fails for negative order index
func TestQuestionValidate_NegativeOrderIndex(t *testing.T) {
	question := &model.Question{
		QuestionText: "Valid question?",
		QuestionType: "multiple_choice",
		Points:       10,
		OrderIndex:   -1,
	}

	err := question.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order index must be non-negative")
}

// TestQuestionValidate_LongExplanation tests validation fails for too long explanation
func TestQuestionValidate_LongExplanation(t *testing.T) {
	longExplanation := make([]byte, 501)
	for i := range longExplanation {
		longExplanation[i] = 'a'
	}

	question := &model.Question{
		QuestionText: "Valid question?",
		QuestionType: "multiple_choice",
		Points:       10,
		Explanation:  string(longExplanation),
	}

	err := question.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "explanation must not exceed 500 characters")
}

// TestQuestionMinimumAnswers tests minimum answers requirement
func TestQuestionMinimumAnswers(t *testing.T) {
	testCases := []struct {
		questionType   string
		expectedMin    int
	}{
		{"multiple_choice", 2},
		{"true_false", 2},
		{"short_answer", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.questionType, func(t *testing.T) {
			question := &model.Question{
				QuestionType: tc.questionType,
			}

			min := question.MinimumAnswers()
			assert.Equal(t, tc.expectedMin, min)
		})
	}
}

// TestQuestionRequiresAnswers tests answers requirement logic
func TestQuestionRequiresAnswers(t *testing.T) {
	testCases := []struct {
		questionType     string
		requiresAnswers  bool
	}{
		{"multiple_choice", true},
		{"true_false", true},
		{"short_answer", false},
	}

	for _, tc := range testCases {
		t.Run(tc.questionType, func(t *testing.T) {
			question := &model.Question{
				QuestionType: tc.questionType,
			}

			requires := question.RequiresAnswers()
			assert.Equal(t, tc.requiresAnswers, requires)
		})
	}
}

// ====== Answer Validation Tests ======

// TestAnswerValidate_ValidAnswer tests validation passes for valid answer
func TestAnswerValidate_ValidAnswer(t *testing.T) {
	answer := &model.Answer{
		AnswerText: "Option A",
		IsCorrect:  true,
		OrderIndex: 0,
	}

	err := answer.Validate()
	assert.NoError(t, err)
}

// TestAnswerValidate_EmptyText tests validation fails for empty answer text
func TestAnswerValidate_EmptyText(t *testing.T) {
	answer := &model.Answer{
		AnswerText: "",
		IsCorrect:  true,
	}

	err := answer.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer text is required")
}

// TestAnswerValidate_LongText tests validation fails for too long answer text
func TestAnswerValidate_LongText(t *testing.T) {
	longText := make([]byte, 501)
	for i := range longText {
		longText[i] = 'a'
	}

	answer := &model.Answer{
		AnswerText: string(longText),
		IsCorrect:  true,
	}

	err := answer.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer text must not exceed 500 characters")
}

// TestAnswerValidate_NegativeOrderIndex tests validation fails for negative order index
func TestAnswerValidate_NegativeOrderIndex(t *testing.T) {
	answer := &model.Answer{
		AnswerText: "Valid answer",
		IsCorrect:  true,
		OrderIndex: -1,
	}

	err := answer.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order index must be non-negative")
}

// TestAnswerInputValidate_ValidInput tests answer input validation passes
func TestAnswerInputValidate_ValidInput(t *testing.T) {
	input := &model.AnswerInput{
		AnswerText: "Valid answer",
		IsCorrect:  true,
		OrderIndex: 0,
	}

	err := input.Validate()
	assert.NoError(t, err)
}

// TestAnswerInputValidate_EmptyText tests answer input validation fails for empty text
func TestAnswerInputValidate_EmptyText(t *testing.T) {
	input := &model.AnswerInput{
		AnswerText: "",
		IsCorrect:  true,
	}

	err := input.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "answer text is required")
}
