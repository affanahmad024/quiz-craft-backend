package core

import (
	"context"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/admin/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	quizmodel "github.com/affanahmad024/quiz-craft/backend/quiz/model"
	userdao "github.com/affanahmad024/quiz-craft/backend/user/dao"
	"github.com/google/uuid"
)

// ListQuizzesAdmin retrieves all quizzes with admin view
func ListQuizzesAdmin(ctx context.Context, adminID string, page, limit int, isPublished *bool, isPublic *bool, creatorID *string, category *string, search *string, sortBy, sortOrder string, includeDeleted bool) ([]model.AdminQuizDetail, int, error) {
	commons.Log(ctx, "[core.ListQuizzesAdmin [admin_id=%s [page=%d", adminID, page)

	// Validate admin
	if !isAdmin(ctx, adminID) {
		return nil, 0, commons.Forbidden("admin-access-required")
	}

	// Set defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build search input
	searchInput := quizmodel.SearchQuizzesInput{
		Page:      page,
		Limit:     limit,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	if search != nil {
		searchInput.Query = *search
	}
	if category != nil {
		searchInput.Category = *category
	}
	if creatorID != nil {
		searchInput.CreatorID = *creatorID
	}

	// Get quizzes (use search function but with admin privileges)
	quizSummaries, total, err := quizdao.SearchQuizzes(ctx, searchInput, &adminID)
	if err != nil {
		commons.LogE(ctx, "[core.ListQuizzesAdmin::SearchErr [err=%v", err)
		return nil, 0, err
	}

	// Convert to AdminQuizDetail
	adminQuizzes := make([]model.AdminQuizDetail, len(quizSummaries))
	for i, q := range quizSummaries {
		adminQuizzes[i] = model.AdminQuizDetail{
			ID:            q.ID,
			Title:         q.Title,
			Description:   q.Description,
			Creator:       q.Creator,
			Category:      q.Category,
			Difficulty:    q.Difficulty,
			IsPublished:   false, // Default value, will need to get from full quiz if needed
			IsPublic:      false, // Default value, will need to get from full quiz if needed
			QuestionCount: q.QuestionCount,
			AttemptCount:  q.AttemptCount,
			ViewCount:     q.ViewCount,
			AverageScore:  q.AverageScore,
			CreatedAt:     q.CreatedAt,
			UpdatedAt:     time.Time{}, // QuizSummary doesn't have UpdatedAt, use default
		}
	}

	return adminQuizzes, total, nil
}

// GetQuizDetailAdmin retrieves detailed quiz info for admin
func GetQuizDetailAdmin(ctx context.Context, adminID string, quizID string) (*model.AdminQuizDetail, error) {
	commons.Log(ctx, "[core.GetQuizDetailAdmin [admin_id=%s [quiz_id=%s", adminID, quizID)

	// Validate admin
	if !isAdmin(ctx, adminID) {
		return nil, commons.Forbidden("admin-access-required")
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizDetailAdmin::FindErr [err=%v", err)
		return nil, err
	}

	// Get creator info
	creatorUUID, _ := uuid.Parse(quiz.CreatorID)
	creator, err := userdao.FindByID(ctx, creatorUUID)
	creatorSummary := quizmodel.UserSummary{
		ID:   quiz.CreatorID,
		Name: "Unknown",
	}
	if err == nil {
		creatorSummary.Name = creator.Name
	}

	// Get questions
	questions, err := quizdao.ListByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogW(ctx, "[core.GetQuizDetailAdmin::GetQuestionsErr [err=%v", err)
		questions = []*quizmodel.Question{}
	}

	// Get attempt stats
	stats, err := quizdao.GetQuizAttemptStats(ctx, quizUUID)
	var avgScore float64
	var attemptCount int
	if err == nil && stats != nil {
		avgScore = stats.AverageScore
		attemptCount = stats.TotalAttempts
	}

	// Helper to convert string to *string
	strPtr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}

	adminQuiz := &model.AdminQuizDetail{
		ID:            quiz.ID,
		Title:         quiz.Title,
		Description:   strPtr(quiz.Description),
		Creator:       creatorSummary,
		Category:      strPtr(quiz.Category),
		Difficulty:    strPtr(quiz.Difficulty),
		IsPublished:   quiz.IsPublished,
		IsPublic:      quiz.IsPublic,
		Questions:     questions,
		QuestionCount: len(questions),
		AttemptCount:  attemptCount,
		ViewCount:     quiz.ViewCount,
		AverageScore:  avgScore,
		CreatedAt:     quiz.CreatedAt,
		UpdatedAt:     quiz.UpdatedAt,
		DeletedAt:     quiz.DeletedAt,
	}

	return adminQuiz, nil
}

// DeleteQuizAdmin deletes a quiz (soft or hard delete)
func DeleteQuizAdmin(ctx context.Context, adminID string, quizID string, reason string, permanent bool) error {
	commons.Log(ctx, "[core.DeleteQuizAdmin [admin_id=%s [quiz_id=%s [permanent=%v", adminID, quizID, permanent)

	// Validate admin
	admin, err := userdao.FindByID(ctx, uuid.MustParse(adminID))
	if err != nil || admin.Role != "admin" {
		return commons.Forbidden("admin-access-required")
	}

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteQuizAdmin::FindErr [err=%v", err)
		return err
	}

	// Delete quiz
	if permanent {
		// Hard delete
		err = quizdao.HardDelete(ctx, quizUUID)
	} else {
		// Soft delete
		err = quizdao.SoftDelete(ctx, quizUUID)
	}

	if err != nil {
		commons.LogE(ctx, "[core.DeleteQuizAdmin::DeleteErr [err=%v", err)
		return err
	}

	// Log activity
	LogActivitySync(ctx, adminID, model.ActionQuizDeleted, "quiz", quizID, map[string]interface{}{
		"reason":     reason,
		"permanent":  permanent,
		"admin_name": admin.Name,
		"quiz_title": quiz.Title,
	})

	commons.Log(ctx, "[core.DeleteQuizAdmin::OK [quiz_id=%s [permanent=%v", quizID, permanent)
	return nil
}
