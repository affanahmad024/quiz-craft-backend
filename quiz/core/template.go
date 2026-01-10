package core

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// SaveQuizAsTemplate saves a quiz as a reusable template
func SaveQuizAsTemplate(ctx context.Context, quizID string, userID string, input model.SaveAsTemplateInput) (string, error) {
	commons.Log(ctx, "[core.SaveQuizAsTemplate [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return "", commons.BadInput("invalid-quiz-id")
	}

	// Get original quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.SaveQuizAsTemplate::FindErr [quiz_id=%s [err=%v", quizID, err)
		return "", err
	}

	// Check ownership
	if quiz.CreatorID != userID {
		commons.LogW(ctx, "[core.SaveQuizAsTemplate::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return "", commons.Forbidden("not-quiz-creator")
	}

	// Get questions and answers
	questions, err := quizdao.ListByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.SaveQuizAsTemplate::GetQuestionsErr [err=%v", err)
		return "", err
	}

	if len(questions) == 0 {
		return "", commons.BadInput("quiz-has-no-questions")
	}

	// Create template quiz
	templateID := uuid.New().String()
	template := &model.Quiz{
		ID:               templateID,
		Title:            input.TemplateName,
		Description:      input.TemplateDescription,
		CreatorID:        userID,
		Category:         quiz.Category,
		Difficulty:       quiz.Difficulty,
		TimeLimitMinutes: quiz.TimeLimitMinutes,
		PassingScore:     quiz.PassingScore,
		IsPublished:      input.MakePublic,
		IsPublic:         input.MakePublic,
		IsTemplate:       true,
		Tags:             quiz.Tags,
		Meta:             make(model.JSONMap),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := quizdao.CreateQuiz(ctx, template); err != nil {
		commons.LogE(ctx, "[core.SaveQuizAsTemplate::CreateErr [err=%v", err)
		return "", err
	}

	// Copy questions
	for _, q := range questions {
		newQuestionID := uuid.New().String()
		newQuestion := &model.Question{
			ID:           newQuestionID,
			QuizID:       templateID,
			QuestionText: q.QuestionText,
			QuestionType: q.QuestionType,
			Points:       q.Points,
			OrderIndex:   q.OrderIndex,
			Explanation:  q.Explanation,
			Meta:         make(model.JSONMap),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := quizdao.CreateQuestion(ctx, newQuestion); err != nil {
			commons.LogE(ctx, "[core.SaveQuizAsTemplate::CreateQuestionErr [err=%v", err)
			continue
		}

		// Copy answers
		questionUUID, _ := uuid.Parse(q.ID)
		answers, err := quizdao.ListByQuestionID(ctx, questionUUID)
		if err != nil {
			commons.LogW(ctx, "[core.SaveQuizAsTemplate::GetAnswersErr [question_id=%s [err=%v", q.ID, err)
			continue
		}

		for _, a := range answers {
			newAnswer := &model.Answer{
				ID:         uuid.New().String(),
				QuestionID: newQuestionID,
				AnswerText: a.AnswerText,
				IsCorrect:  a.IsCorrect,
				OrderIndex: a.OrderIndex,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			if err := quizdao.CreateAnswer(ctx, newAnswer); err != nil {
				commons.LogW(ctx, "[core.SaveQuizAsTemplate::CreateAnswerErr [err=%v", err)
			}
		}
	}

	commons.Log(ctx, "[core.SaveQuizAsTemplate::OK [template_id=%s", templateID)
	return templateID, nil
}

// CreateQuizFromTemplate creates a new quiz based on a template
func CreateQuizFromTemplate(ctx context.Context, templateID string, userID string, input model.UseTemplateInput) (string, error) {
	commons.Log(ctx, "[core.CreateQuizFromTemplate [template_id=%s [user_id=%s", templateID, userID)

	templateUUID, err := uuid.Parse(templateID)
	if err != nil {
		return "", commons.BadInput("invalid-template-id")
	}

	// Get template
	template, err := quizdao.FindByID(ctx, templateUUID)
	if err != nil {
		commons.LogE(ctx, "[core.CreateQuizFromTemplate::FindErr [template_id=%s [err=%v", templateID, err)
		return "", err
	}

	if !template.IsTemplate {
		return "", commons.BadInput("not-a-template")
	}

	// Check if template is accessible
	if !template.IsPublic && template.CreatorID != userID {
		return "", commons.Forbidden("template-not-accessible")
	}

	// Get template questions and answers
	questions, err := quizdao.ListByQuizID(ctx, templateUUID)
	if err != nil {
		commons.LogE(ctx, "[core.CreateQuizFromTemplate::GetQuestionsErr [err=%v", err)
		return "", err
	}

	// Create new quiz
	newQuizID := uuid.New().String()
	newQuiz := &model.Quiz{
		ID:                newQuizID,
		Title:             input.Title,
		Description:       template.Description,
		CreatorID:         userID,
		Category:          template.Category,
		Difficulty:        template.Difficulty,
		TimeLimitMinutes:  template.TimeLimitMinutes,
		PassingScore:      template.PassingScore,
		IsPublished:       false,
		IsPublic:          true,
		BasedOnTemplateID: &templateID,
		Tags:              template.Tags,
		Meta:              make(model.JSONMap),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := quizdao.CreateQuiz(ctx, newQuiz); err != nil {
		commons.LogE(ctx, "[core.CreateQuizFromTemplate::CreateErr [err=%v", err)
		return "", err
	}

	// Copy questions and answers (same as SaveQuizAsTemplate)
	for _, q := range questions {
		newQuestionID := uuid.New().String()
		newQuestion := &model.Question{
			ID:           newQuestionID,
			QuizID:       newQuizID,
			QuestionText: q.QuestionText,
			QuestionType: q.QuestionType,
			Points:       q.Points,
			OrderIndex:   q.OrderIndex,
			Explanation:  q.Explanation,
			Meta:         make(model.JSONMap),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := quizdao.CreateQuestion(ctx, newQuestion); err != nil {
			commons.LogE(ctx, "[core.CreateQuizFromTemplate::CreateQuestionErr [err=%v", err)
			continue
		}

		// Copy answers
		questionUUID, _ := uuid.Parse(q.ID)
		answers, err := quizdao.ListByQuestionID(ctx, questionUUID)
		if err != nil {
			commons.LogW(ctx, "[core.CreateQuizFromTemplate::GetAnswersErr [question_id=%s [err=%v", q.ID, err)
			continue
		}

		for _, a := range answers {
			newAnswer := &model.Answer{
				ID:         uuid.New().String(),
				QuestionID: newQuestionID,
				AnswerText: a.AnswerText,
				IsCorrect:  a.IsCorrect,
				OrderIndex: a.OrderIndex,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			if err := quizdao.CreateAnswer(ctx, newAnswer); err != nil {
				commons.LogW(ctx, "[core.CreateQuizFromTemplate::CreateAnswerErr [err=%v", err)
			}
		}
	}

	// Increment template usage count
	if err := quizdao.IncrementTemplateUsageCount(ctx, templateUUID); err != nil {
		commons.LogW(ctx, "[core.CreateQuizFromTemplate::IncrementErr [err=%v", err)
		// Non-critical, continue
	}

	commons.Log(ctx, "[core.CreateQuizFromTemplate::OK [quiz_id=%s", newQuizID)
	return newQuizID, nil
}

// ListTemplates lists available templates with filters
func ListTemplates(ctx context.Context, userID *string, page int, limit int, category string, myTemplates bool) ([]model.QuizSummary, int, error) {
	commons.Log(ctx, "[core.ListTemplates [page=%d [my_templates=%v", page, myTemplates)

	// Build search input for templates
	input := model.SearchQuizzesInput{
		Page:       page,
		Limit:      limit,
		IsTemplate: boolPtr(true),
		SortBy:     "newest",
		SortOrder:  "desc",
	}

	if category != "" {
		input.Category = category
	}

	if myTemplates && userID != nil {
		input.CreatorID = *userID
	}

	// Use search function
	templates, total, err := quizdao.SearchQuizzes(ctx, input, userID)
	if err != nil {
		commons.LogE(ctx, "[core.ListTemplates::SearchErr [err=%v", err)
		return nil, 0, err
	}

	commons.Log(ctx, "[core.ListTemplates::OK [count=%d [total=%d", len(templates), total)
	return templates, total, nil
}

func boolPtr(b bool) *bool {
	return &b
}
