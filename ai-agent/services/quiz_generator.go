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
	db              *sql.DB
	imageProcessor  *ImageProcessorService
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

	// Create quiz
	quiz := &quizmodel.Quiz{
		ID:          uuid.New().String(),
		Title:       aiResponse.Title,
		Description: aiResponse.Description,
		CreatorID:   userID,
		Category:    aiResponse.Category,
		Difficulty:  aiResponse.Difficulty,
		IsPublished: false,
		IsPublic:    true,
		Tags:        []string{"ai-generated"},
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
