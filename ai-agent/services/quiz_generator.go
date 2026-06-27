package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	quizmodel "github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// QuizGeneratorService handles quiz generation from AI responses
type QuizGeneratorService struct {
	db             *sql.DB
	imageProcessor *ImageProcessorService
}

// NewQuizGeneratorService creates a new quiz generator service
func NewQuizGeneratorService(db *sql.DB) *QuizGeneratorService {
	return &QuizGeneratorService{
		db:             db,
		imageProcessor: NewImageProcessorService(),
	}
}

// CreateQuizFromAIResponse creates a quiz in database from AI response
func (s *QuizGeneratorService) CreateQuizFromAIResponse(ctx context.Context, aiResponse *model.AIQuizResponse, userID string, options model.QuizGenOptions) (string, error) {
	commons.Log(ctx, "[services.CreateQuizFromAIResponse [user_id=%s [title=%s [questions=%d",
		userID, aiResponse.Title, len(aiResponse.Questions))

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		commons.LogE(ctx, "[services.CreateQuizFromAIResponse::TxErr [err=%v", err)
		return "", commons.Internal("failed to begin transaction")
	}
	defer tx.Rollback()

	metadata := options.Metadata
	title := metadata.Title
	if title == "" {
		title = aiResponse.Title
	}

	description := metadata.Description
	if description == "" {
		description = aiResponse.Description
	}

	category := options.Category
	if category == "" {
		category = aiResponse.Category
	}

	difficulty := options.Difficulty
	if difficulty == "" {
		difficulty = aiResponse.Difficulty
	}

	passingScore := 0
	if metadata.PassingScorePercentage != nil {
		passingScore = *metadata.PassingScorePercentage
	}

	// Create quiz
	quiz := &quizmodel.Quiz{
		ID:               uuid.New().String(),
		Title:            title,
		Description:      description,
		CreatorID:        userID,
		Category:         category,
		Difficulty:       difficulty,
		TimeLimitMinutes: metadata.TimeLimitMinutes,
		PassingScore:     passingScore,
		IsPublished:      false,
		IsPublic:         true,
		Tags:             metadata.Tags,
		Meta: map[string]interface{}{
			"ai_generated":    true,
			"generation_date": time.Now().Format(time.RFC3339),
			"model":           "gemini-1.5-pro",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create quiz in database
	if err := quizdao.CreateQuiz(ctx, quiz); err != nil {
		commons.LogE(ctx, "[services.CreateQuizFromAIResponse::QuizErr [err=%v", err)
		return "", err
	}

	// Create questions and answers
	for i, aiQuestion := range aiResponse.Questions {
		// Create question
		question := &quizmodel.Question{
			ID:           uuid.New().String(),
			QuizID:       quiz.ID,
			QuestionText: aiQuestion.QuestionText,
			QuestionType: aiQuestion.QuestionType,
			Difficulty:   normalizeAIDifficulty(aiQuestion.Difficulty, options.Difficulty),
			Tags:         aiQuestion.Tags,
			Points:       aiQuestion.Points,
			OrderIndex:   i,
			Explanation:  aiQuestion.Explanation,
			Meta:         map[string]interface{}{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Process and save question image if present
		if aiQuestion.ImageData != "" {
			commons.Log(ctx, "[services.CreateQuizFromAIResponse [question=%d has_image=true", i+1)

			imageFile, err := s.imageProcessor.ProcessAndSaveImage(ctx, aiQuestion.ImageData, userID, question.ID)
			if err != nil {
				commons.LogE(ctx, "[services.CreateQuizFromAIResponse::ImageErr [question_num=%d [err=%v", i+1, err)
				// Continue without image rather than failing entire quiz
				commons.Log(ctx, "[services.CreateQuizFromAIResponse [question=%d continuing_without_image", i+1)
			} else {
				// Add image URL to question meta
				question.Meta["image_url"] = fmt.Sprintf("/api/v1/files/serve/%s", imageFile.ID)
				question.Meta["image_file_id"] = imageFile.ID
				commons.Log(ctx, "[services.CreateQuizFromAIResponse [question=%d image_saved file_id=%s", i+1, imageFile.ID)
			}
		}

		if err := quizdao.CreateQuestion(ctx, question); err != nil {
			commons.LogE(ctx, "[services.CreateQuizFromAIResponse::QuestionErr [question_num=%d [err=%v", i+1, err)
			return "", err
		}

		// Create answers
		for j, aiAnswer := range aiQuestion.Answers {
			answer := &quizmodel.Answer{
				ID:         uuid.New().String(),
				QuestionID: question.ID,
				AnswerText: aiAnswer.AnswerText,
				IsCorrect:  aiAnswer.IsCorrect,
				OrderIndex: j,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			if err := quizdao.CreateAnswer(ctx, answer); err != nil {
				commons.LogE(ctx, "[services.CreateQuizFromAIResponse::AnswerErr [question_num=%d [answer_num=%d [err=%v",
					i+1, j+1, err)
				return "", err
			}
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		commons.LogE(ctx, "[services.CreateQuizFromAIResponse::CommitErr [err=%v", err)
		return "", commons.Internal("failed to commit transaction")
	}

	commons.Log(ctx, "[services.CreateQuizFromAIResponse::OK [quiz_id=%s", quiz.ID)
	return quiz.ID, nil
}

// AddQuestionsToExistingQuiz adds AI-generated questions to an existing quiz
func (s *QuizGeneratorService) AddQuestionsToExistingQuiz(ctx context.Context, quizID string, aiResponse *model.AIQuizResponse, userID string) error {
	commons.Log(ctx, "[services.AddQuestionsToExistingQuiz [quiz_id=%s [user_id=%s [questions=%d",
		quizID, userID, len(aiResponse.Questions))

	// Verify quiz exists and user owns it
	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return commons.BadInput("invalid-quiz-id")
	}

	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::QuizNotFound [quiz_id=%s [err=%v", quizID, err)
		return commons.NotFound("quiz-not-found")
	}

	if quiz.CreatorID != userID {
		commons.LogW(ctx, "[services.AddQuestionsToExistingQuiz::Forbidden [quiz_id=%s [user_id=%s", quizID, userID)
		return commons.Forbidden("not-quiz-owner")
	}

	// Get current max order index
	existingQuestions, err := quizdao.ListByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::ListQuestionsErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	maxOrderIndex := -1
	for _, q := range existingQuestions {
		if q.OrderIndex > maxOrderIndex {
			maxOrderIndex = q.OrderIndex
		}
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::TxErr [err=%v", err)
		return commons.Internal("failed to begin transaction")
	}
	defer tx.Rollback()

	// Add questions to existing quiz
	for i, aiQuestion := range aiResponse.Questions {
		// Create question
		question := &quizmodel.Question{
			ID:           uuid.New().String(),
			QuizID:       quizID,
			QuestionText: aiQuestion.QuestionText,
			QuestionType: aiQuestion.QuestionType,
			Difficulty:   normalizeAIDifficulty(aiQuestion.Difficulty, "medium"),
			Tags:         aiQuestion.Tags,
			Points:       aiQuestion.Points,
			OrderIndex:   maxOrderIndex + i + 1,
			Explanation:  aiQuestion.Explanation,
			Meta:         map[string]interface{}{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Process and save question image if present
		if aiQuestion.ImageData != "" {
			commons.Log(ctx, "[services.AddQuestionsToExistingQuiz [question=%d has_image=true", i+1)

			imageFile, err := s.imageProcessor.ProcessAndSaveImage(ctx, aiQuestion.ImageData, userID, question.ID)
			if err != nil {
				commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::ImageErr [question_num=%d [err=%v", i+1, err)
				commons.Log(ctx, "[services.AddQuestionsToExistingQuiz [question=%d continuing_without_image", i+1)
			} else {
				question.Meta["image_url"] = fmt.Sprintf("/api/v1/files/serve/%s", imageFile.ID)
				question.Meta["image_file_id"] = imageFile.ID
				commons.Log(ctx, "[services.AddQuestionsToExistingQuiz [question=%d image_saved file_id=%s", i+1, imageFile.ID)
			}
		}

		if err := quizdao.CreateQuestion(ctx, question); err != nil {
			commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::QuestionErr [question_num=%d [err=%v", i+1, err)
			return err
		}

		// Create answers
		for j, aiAnswer := range aiQuestion.Answers {
			answer := &quizmodel.Answer{
				ID:         uuid.New().String(),
				QuestionID: question.ID,
				AnswerText: aiAnswer.AnswerText,
				IsCorrect:  aiAnswer.IsCorrect,
				OrderIndex: j,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			if err := quizdao.CreateAnswer(ctx, answer); err != nil {
				commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::AnswerErr [question_num=%d [answer_num=%d [err=%v",
					i+1, j+1, err)
				return err
			}
		}
	}

	// Update quiz updated_at timestamp
	quiz.UpdatedAt = time.Now()
	if err := quizdao.UpdateQuiz(ctx, quiz); err != nil {
		commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::UpdateQuizErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		commons.LogE(ctx, "[services.AddQuestionsToExistingQuiz::CommitErr [err=%v", err)
		return commons.Internal("failed to commit transaction")
	}

	commons.Log(ctx, "[services.AddQuestionsToExistingQuiz::OK [quiz_id=%s [added=%d", quizID, len(aiResponse.Questions))
	return nil
}

// ValidateAIQuizResponse validates AI quiz response structure
func (s *QuizGeneratorService) ValidateAIQuizResponse(aiResponse *model.AIQuizResponse) error {
	if aiResponse.Title == "" {
		return commons.BadInput("quiz-title-missing")
	}

	if len(aiResponse.Title) > 255 {
		return commons.BadInput("quiz-title-too-long")
	}

	if len(aiResponse.Description) > 1000 {
		return commons.BadInput("quiz-description-too-long")
	}

	if len(aiResponse.Questions) == 0 {
		return commons.BadInput("no-questions-in-response")
	}

	if len(aiResponse.Questions) > 50 {
		return commons.BadInput("too-many-questions")
	}

	// Validate each question
	for i, question := range aiResponse.Questions {
		if err := s.validateQuestion(&question, i+1); err != nil {
			return err
		}
	}

	return nil
}

// ValidateAIQuestionsOnly validates only questions (for adding to existing quiz)
func (s *QuizGeneratorService) ValidateAIQuestionsOnly(aiResponse *model.AIQuizResponse) error {
	if len(aiResponse.Questions) == 0 {
		return commons.BadInput("no-questions-in-response")
	}

	if len(aiResponse.Questions) > 50 {
		return commons.BadInput("too-many-questions")
	}

	// Validate each question
	for i, question := range aiResponse.Questions {
		if err := s.validateQuestion(&question, i+1); err != nil {
			return err
		}
	}

	return nil
}

// validateQuestion validates a single question
func (s *QuizGeneratorService) validateQuestion(question *model.AIQuestion, questionNum int) error {
	if question.QuestionText == "" {
		return commons.BadInput("question-text-missing")
	}

	if len(question.QuestionText) < 10 {
		return commons.BadInput("question-text-too-short")
	}

	if len(question.QuestionText) > 1000 {
		return commons.BadInput("question-text-too-long")
	}

	if question.Points <= 0 || question.Points > 100 {
		question.Points = 1 // Default to 1 point
	}

	if len(question.Answers) < 2 {
		return commons.BadInput("insufficient-answers")
	}

	if len(question.Answers) > 6 {
		return commons.BadInput("too-many-answers")
	}

	// Check for exactly one correct answer
	correctCount := 0
	for _, answer := range question.Answers {
		if answer.IsCorrect {
			correctCount++
		}

		if len(answer.AnswerText) == 0 {
			return commons.BadInput("answer-text-empty")
		}

		if len(answer.AnswerText) > 500 {
			return commons.BadInput("answer-text-too-long")
		}
	}

	if correctCount == 0 {
		return commons.BadInput("no-correct-answer")
	}

	if correctCount > 1 && question.QuestionType == "multiple_choice" {
		return commons.BadInput("multiple-correct-answers-not-allowed")
	}

	return nil
}

// ParseAIResponse parses AI JSON response
func (s *QuizGeneratorService) ParseAIResponse(responseText string) (*model.AIQuizResponse, error) {
	var aiResponse model.AIQuizResponse

	if err := json.Unmarshal([]byte(responseText), &aiResponse); err != nil {
		return nil, commons.BadInput("invalid-json-response")
	}

	return &aiResponse, nil
}

func normalizeAIDifficulty(value string, fallback string) string {
	switch value {
	case "easy", "medium", "hard":
		return value
	}
	switch fallback {
	case "easy", "medium", "hard":
		return fallback
	}
	return "medium"
}
