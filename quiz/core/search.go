package core

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// SearchQuizzes searches quizzes with filters and pagination
func SearchQuizzes(ctx context.Context, input model.SearchQuizzesInput, userID *string) (*model.SearchQuizzesResponse, error) {
	commons.Log(ctx, "[core.SearchQuizzes [query=%s [page=%d", input.Query, input.Page)

	// Validate and set defaults
	if input.Page < 1 {
		input.Page = 1
	}
	if input.Limit < 1 || input.Limit > 100 {
		input.Limit = 20
	}

	// Validate difficulty
	if input.Difficulty != "" {
		validDifficulties := []string{"easy", "medium", "hard"}
		if !contains(validDifficulties, input.Difficulty) {
			return nil, commons.BadInput("invalid-difficulty")
		}
	}

	// Validate sort options
	validSortOptions := []string{"newest", "popular", "most_attempted", ""}
	if !contains(validSortOptions, input.SortBy) {
		return nil, commons.BadInput("invalid-sort-by")
	}

	if input.SortOrder != "" && input.SortOrder != "asc" && input.SortOrder != "desc" {
		return nil, commons.BadInput("invalid-sort-order")
	}

	if input.SortBy == "" {
		input.SortBy = "newest"
	}
	if input.SortOrder == "" {
		input.SortOrder = "desc"
	}

	// Execute search
	quizzes, total, err := quizdao.SearchQuizzes(ctx, input, userID)
	if err != nil {
		commons.LogE(ctx, "[core.SearchQuizzes::SearchErr [err=%v", err)
		return nil, err
	}

	// Get facets
	facets, err := quizdao.GetSearchFacets(ctx, userID)
	if err != nil {
		commons.LogW(ctx, "[core.SearchQuizzes::FacetsErr [err=%v", err)
		// Non-critical, continue with empty facets
		facets = model.SearchFacets{}
	}

	// Calculate total pages
	totalPages := (total + input.Limit - 1) / input.Limit

	response := &model.SearchQuizzesResponse{
		Quizzes:    quizzes,
		Total:      total,
		Page:       input.Page,
		Limit:      input.Limit,
		TotalPages: totalPages,
		Facets:     facets,
	}

	commons.Log(ctx, "[core.SearchQuizzes::OK [count=%d [total=%d", len(quizzes), total)
	return response, nil
}

// GenerateShareLink generates a unique share link for a quiz
func GenerateShareLink(ctx context.Context, quizID string, userID string) (*model.ShareLinkResponse, error) {
	commons.Log(ctx, "[core.GenerateShareLink [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz and validate ownership
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GenerateShareLink::FindErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	if quiz.CreatorID != userID {
		commons.LogW(ctx, "[core.GenerateShareLink::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("not-quiz-creator")
	}

	// Check if share code already exists
	if quiz.ShareCode != nil {
		shareURL := fmt.Sprintf("https://quizcraft.com/share/%s", *quiz.ShareCode)
		return &model.ShareLinkResponse{
			ShareCode: *quiz.ShareCode,
			ShareURL:  shareURL,
		}, nil
	}

	// Generate new share code
	shareCode := generateShareCode()

	// Update quiz with share code
	err = quizdao.UpdateQuizShareCode(ctx, quizUUID, shareCode)
	if err != nil {
		commons.LogE(ctx, "[core.GenerateShareLink::UpdateErr [err=%v", err)
		return nil, err
	}

	shareURL := fmt.Sprintf("https://quizcraft.com/share/%s", shareCode)

	commons.Log(ctx, "[core.GenerateShareLink::OK [quiz_id=%s [code=%s", quizID, shareCode)
	return &model.ShareLinkResponse{
		ShareCode: shareCode,
		ShareURL:  shareURL,
	}, nil
}

// GetQuizByShareCode retrieves a quiz by share code and increments view count
func GetQuizByShareCode(ctx context.Context, shareCode string) (*model.Quiz, error) {
	commons.Log(ctx, "[core.GetQuizByShareCode [share_code=%s", shareCode)

	// Get quiz by share code
	quiz, err := quizdao.GetQuizByShareCode(ctx, shareCode)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizByShareCode::FindErr [err=%v", err)
		return nil, err
	}

	// Increment view count
	quizUUID, _ := uuid.Parse(quiz.ID)
	if err := quizdao.IncrementQuizViewCount(ctx, quizUUID); err != nil {
		commons.LogW(ctx, "[core.GetQuizByShareCode::IncrementErr [err=%v", err)
		// Non-critical, continue
	}

	commons.Log(ctx, "[core.GetQuizByShareCode::OK [quiz_id=%s", quiz.ID)
	return quiz, nil
}

// generateShareCode generates a random share code
func generateShareCode() string {
	b := make([]byte, 8)
	rand.Read(b)
	code := base32.StdEncoding.EncodeToString(b)
	code = strings.ToLower(code)
	code = strings.ReplaceAll(code, "=", "")
	if len(code) > 10 {
		code = code[:10]
	}
	return code
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
