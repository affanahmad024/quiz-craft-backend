package core

import (
	"context"
	"fmt"
	"time"

	aidao "github.com/affanahmad024/quiz-craft/backend/ai-agent/dao"
	aimodel "github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	quizdao "github.com/affanahmad024/quiz-craft/backend/quiz/dao"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/google/uuid"
)

// CreateQuiz creates a new quiz with questions and answers
func CreateQuiz(ctx context.Context, creatorID string, input model.QuizInput) (*model.Quiz, error) {
	commons.Log(ctx, "[core.CreateQuiz [creator_id=%s [title=%s", creatorID, input.Title)

	// Create quiz object
	quiz := &model.Quiz{
		ID:               uuid.New().String(),
		Title:            input.Title,
		Description:      input.Description,
		CreatorID:        creatorID,
		Category:         input.Category,
		Difficulty:       input.Difficulty,
		TimeLimitMinutes: input.TimeLimitMinutes,
		PassingScore:     input.PassingScore,
		IsPublished:      false, // New quizzes start unpublished
		PublishStatus:    "draft",
		IsPublic:         input.IsPublic,
		Tags:             input.Tags,
		Meta:             input.Meta,
	}

	// Validate quiz
	if err := quiz.Validate(); err != nil {
		commons.LogW(ctx, "[core.CreateQuiz::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	// Create quiz in database
	if err := quizdao.CreateQuiz(ctx, quiz); err != nil {
		commons.LogE(ctx, "[core.CreateQuiz::CreateErr [err=%v", err)
		return nil, err
	}

	quizUUID, _ := uuid.Parse(quiz.ID)
	if _, err := quizdao.EnsureDefaultSection(ctx, quizUUID); err != nil {
		commons.LogE(ctx, "[core.CreateQuiz::DefaultSectionErr [quiz_id=%s [err=%v", quiz.ID, err)
		return nil, err
	}

	commons.Log(ctx, "[core.CreateQuiz::OK [quiz_id=%s", quiz.ID)
	return quiz, nil
}

// GetQuiz retrieves a quiz by ID with access control
func GetQuiz(ctx context.Context, quizID string, userID string) (*model.Quiz, error) {
	commons.Log(ctx, "[core.GetQuiz [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuiz::FindErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	// Check access control
	if !quiz.IsAccessibleBy(userID) {
		commons.LogW(ctx, "[core.GetQuiz::AccessDenied [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	return quiz, nil
}

// GetQuizWithQuestions retrieves a quiz with all its questions and answers
func GetQuizWithQuestions(ctx context.Context, quizID string, userID string, hideCorrectAnswers bool) (*model.QuizWithQuestions, error) {
	commons.Log(ctx, "[core.GetQuizWithQuestions [quiz_id=%s [user_id=%s [hide_correct=%v", quizID, userID, hideCorrectAnswers)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizWithQuestions::FindErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	// Check access control
	if !quiz.IsAccessibleBy(userID) {
		commons.LogW(ctx, "[core.GetQuizWithQuestions::AccessDenied [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	sections, err := quizdao.ListSectionsByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizWithQuestions::ListSectionsErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}
	if len(sections) == 0 {
		defaultSection, err := quizdao.EnsureDefaultSection(ctx, quizUUID)
		if err != nil {
			return nil, err
		}
		sections = []*model.QuizSection{defaultSection}
	}

	// Get questions
	questions, err := quizdao.ListByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuizWithQuestions::ListQuestionsErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	// Get answers for each question
	questionsWithAnswers := []model.Question{}
	for _, question := range questions {
		questionUUID, _ := uuid.Parse(question.ID)

		if hideCorrectAnswers {
			// For quiz attempts - hide correct answers
			answerItems, err := quizdao.ListByQuestionIDHideCorrect(ctx, questionUUID)
			if err != nil {
				commons.LogE(ctx, "[core.GetQuizWithQuestions::ListAnswersErr [question_id=%s [err=%v", question.ID, err)
				continue
			}
			// Attach answers without is_correct field
			question.Answers = answerItems
			// Preserve existing meta fields (like image_url, image_file_id) and add answer_count
			if question.Meta == nil {
				question.Meta = model.JSONMap{}
			}
			question.Meta["answer_count"] = len(answerItems)
		} else {
			// For quiz creators/editors - show correct answers
			answers, err := quizdao.ListByQuestionID(ctx, questionUUID)
			if err != nil {
				commons.LogE(ctx, "[core.GetQuizWithQuestions::ListAnswersErr [question_id=%s [err=%v", question.ID, err)
				continue
			}
			// Convert []*Answer to []Answer for JSON serialization
			answerList := []model.Answer{}
			for _, ans := range answers {
				answerList = append(answerList, *ans)
			}
			question.Answers = answerList
			// Preserve existing meta fields (like image_url, image_file_id) and add answer_count
			if question.Meta == nil {
				question.Meta = model.JSONMap{}
			}
			question.Meta["answer_count"] = len(answers)
		}

		questionsWithAnswers = append(questionsWithAnswers, *question)
	}

	sectionModels := make([]model.QuizSection, 0, len(sections))
	sectionIndex := map[string]int{}
	for _, section := range sections {
		sectionModels = append(sectionModels, *section)
		sectionIndex[section.ID] = len(sectionModels) - 1
	}
	for _, question := range questionsWithAnswers {
		if idx, ok := sectionIndex[question.SectionID]; ok {
			sectionModels[idx].Questions = append(sectionModels[idx].Questions, question)
		}
	}

	result := &model.QuizWithQuestions{
		Quiz:      *quiz,
		Questions: questionsWithAnswers,
		Sections:  sectionModels,
	}

	commons.Log(ctx, "[core.GetQuizWithQuestions::OK [quiz_id=%s [question_count=%d", quizID, len(questions))
	return result, nil
}

// ListMyQuizzes lists quizzes created by the user
func ListMyQuizzes(ctx context.Context, creatorID string, filter model.QuizListFilter) ([]*model.Quiz, int, error) {
	commons.Log(ctx, "[core.ListMyQuizzes [creator_id=%s [page=%d", creatorID, filter.Page)

	creatorUUID, err := uuid.Parse(creatorID)
	if err != nil {
		return nil, 0, commons.BadInput("invalid-user-id")
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	quizzes, total, err := quizdao.ListByCreator(ctx, creatorUUID, filter)
	if err != nil {
		commons.LogE(ctx, "[core.ListMyQuizzes::ListErr [creator_id=%s [err=%v", creatorID, err)
		return nil, 0, err
	}

	// Fetch latest attempt information for each quiz
	if len(quizzes) > 0 {
		quizIDs := make([]string, len(quizzes))
		for i, quiz := range quizzes {
			quizIDs[i] = quiz.ID
		}

		latestAttempts, err := quizdao.GetLatestAttemptsByQuizIDs(ctx, creatorUUID, quizIDs)
		if err != nil {
			commons.LogW(ctx, "[core.ListMyQuizzes::GetAttemptsErr [err=%v", err)
			// Non-critical error, continue without attempt info
		} else {
			// Attach latest attempt info to each quiz
			for _, quiz := range quizzes {
				if attemptInfo, ok := latestAttempts[quiz.ID]; ok {
					quiz.LatestAttempt = attemptInfo
				}
			}
		}
	}

	commons.Log(ctx, "[core.ListMyQuizzes::OK [creator_id=%s [total=%d [returned=%d", creatorID, total, len(quizzes))
	return quizzes, total, nil
}

// ListPublicQuizzes lists public published quizzes
func ListPublicQuizzes(ctx context.Context, userID string, filter model.QuizListFilter) ([]*model.Quiz, int, error) {
	commons.Log(ctx, "[core.ListPublicQuizzes [user_id=%s [page=%d", userID, filter.Page)

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	quizzes, total, err := quizdao.ListPublic(ctx, filter)
	if err != nil {
		commons.LogE(ctx, "[core.ListPublicQuizzes::ListErr [err=%v", err)
		return nil, 0, err
	}

	// Fetch latest attempt information for authenticated user
	if userID != "" && len(quizzes) > 0 {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			quizIDs := make([]string, len(quizzes))
			for i, quiz := range quizzes {
				quizIDs[i] = quiz.ID
			}

			latestAttempts, err := quizdao.GetLatestAttemptsByQuizIDs(ctx, userUUID, quizIDs)
			if err != nil {
				commons.LogW(ctx, "[core.ListPublicQuizzes::GetAttemptsErr [err=%v", err)
				// Non-critical error, continue without attempt info
			} else {
				// Attach latest attempt info to each quiz
				for _, quiz := range quizzes {
					if attemptInfo, ok := latestAttempts[quiz.ID]; ok {
						quiz.LatestAttempt = attemptInfo
					}
				}
			}
		}
	}

	commons.Log(ctx, "[core.ListPublicQuizzes::OK [total=%d [returned=%d", total, len(quizzes))
	return quizzes, total, nil
}

// UpdateQuiz updates a quiz
func UpdateQuiz(ctx context.Context, quizID string, userID string, input model.QuizUpdateInput) (*model.Quiz, error) {
	commons.Log(ctx, "[core.UpdateQuiz [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get existing quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateQuiz::FindErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	// Check ownership
	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.UpdateQuiz::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("not-quiz-owner")
	}

	// Update fields
	if input.Title != "" {
		quiz.Title = input.Title
	}
	if input.Description != nil {
		quiz.Description = *input.Description
	}
	if input.Category != nil {
		quiz.Category = *input.Category
	}
	if input.Difficulty != nil {
		quiz.Difficulty = *input.Difficulty
	}
	if input.TimeLimitMinutes != nil {
		quiz.TimeLimitMinutes = input.TimeLimitMinutes
	}
	if input.PassingScore != nil {
		quiz.PassingScore = *input.PassingScore
	}
	if input.IsPublic != nil {
		quiz.IsPublic = *input.IsPublic
	}
	if input.Tags != nil {
		quiz.Tags = input.Tags
	}
	if input.Meta != nil {
		quiz.Meta = input.Meta
	}

	// Validate updated quiz
	if err := quiz.Validate(); err != nil {
		commons.LogW(ctx, "[core.UpdateQuiz::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	// Update in database
	if err := quizdao.UpdateQuiz(ctx, quiz); err != nil {
		commons.LogE(ctx, "[core.UpdateQuiz::UpdateErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	commons.Log(ctx, "[core.UpdateQuiz::OK [quiz_id=%s", quizID)
	return quiz, nil
}

// PublishQuiz publishes a quiz (makes it available for taking)
func PublishQuiz(ctx context.Context, quizID string, userID string) error {
	commons.Log(ctx, "[core.PublishQuiz [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.PublishQuiz::FindErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	// Check ownership
	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.PublishQuiz::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return commons.Forbidden("not-quiz-owner")
	}

	// Check if can publish
	if err := quiz.CanPublish(); err != nil {
		commons.LogW(ctx, "[core.PublishQuiz::CannotPublish [quiz_id=%s [err=%v", quizID, err)
		return commons.BadInput(err.Error())
	}

	// Check if quiz has at least one question
	count, err := quizdao.CountByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.PublishQuiz::CountErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}
	if count == 0 {
		commons.LogW(ctx, "[core.PublishQuiz::NoQuestions [quiz_id=%s", quizID)
		return commons.BadInput("quiz-must-have-at-least-one-question")
	}

	jobID := uuid.New()
	job := &aimodel.GenerationJob{
		ID:      jobID.String(),
		UserID:  userID,
		QuizID:  &quizID,
		JobType: "enrichment",
		Options: aimodel.QuizGenOptions{
			NumQuestions: count,
			Difficulty:   "medium",
			QuestionType: "mixed",
			Category:     "general",
			Language:     "en",
		},
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
	}

	if err := aidao.CreateJob(ctx, job); err != nil {
		commons.LogE(ctx, "[core.PublishQuiz::CreateEnrichmentJobErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	if err := quizdao.UpdatePublishWorkflowStatus(ctx, quizUUID, false, "enrichment_pending", &jobID); err != nil {
		commons.LogE(ctx, "[core.PublishQuiz::WorkflowStatusErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	commons.Log(ctx, "[core.PublishQuiz::Queued [quiz_id=%s [job_id=%s", quizID, job.ID)
	return nil
}

// UnpublishQuiz unpublishes a quiz
func UnpublishQuiz(ctx context.Context, quizID string, userID string) error {
	commons.Log(ctx, "[core.UnpublishQuiz [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UnpublishQuiz::FindErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	// Check ownership
	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.UnpublishQuiz::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return commons.Forbidden("not-quiz-owner")
	}

	// Update publish status
	if err := quizdao.UpdatePublishStatus(ctx, quizUUID, false); err != nil {
		commons.LogE(ctx, "[core.UnpublishQuiz::UpdateErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	commons.Log(ctx, "[core.UnpublishQuiz::OK [quiz_id=%s", quizID)
	return nil
}

// DeleteQuiz soft deletes a quiz
func DeleteQuiz(ctx context.Context, quizID string, userID string) error {
	commons.Log(ctx, "[core.DeleteQuiz [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return commons.BadInput("invalid-quiz-id")
	}

	// Get quiz
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteQuiz::FindErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	// Check ownership
	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.DeleteQuiz::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return commons.Forbidden("not-quiz-owner")
	}

	// Soft delete
	if err := quizdao.SoftDelete(ctx, quizUUID); err != nil {
		commons.LogE(ctx, "[core.DeleteQuiz::DeleteErr [quiz_id=%s [err=%v", quizID, err)
		return err
	}

	commons.Log(ctx, "[core.DeleteQuiz::OK [quiz_id=%s", quizID)
	return nil
}

// AddQuestion adds a question to a quiz
func AddQuestion(ctx context.Context, quizID string, userID string, input model.QuestionInput) (*model.Question, error) {
	commons.Log(ctx, "[core.AddQuestion [quiz_id=%s [user_id=%s", quizID, userID)

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz and check ownership
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddQuestion::FindQuizErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.AddQuestion::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("not-quiz-owner")
	}

	// Get current question count for order_index
	count, err := quizdao.CountByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddQuestion::CountErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	section, err := resolveQuestionSection(ctx, quizUUID, input.SectionID)
	if err != nil {
		return nil, err
	}
	sectionUUID, _ := uuid.Parse(section.ID)
	sectionOrderIndex, err := quizdao.NextSectionQuestionOrderIndex(ctx, sectionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddQuestion::SectionOrderErr [quiz_id=%s [section_id=%s [err=%v", quizID, section.ID, err)
		return nil, err
	}

	// Initialize meta if nil
	if input.Meta == nil {
		input.Meta = model.JSONMap{}
	}

	// Process image file if provided
	if input.ImageFileID != nil && *input.ImageFileID != "" {
		commons.Log(ctx, "[core.AddQuestion::ProcessingImage [image_file_id=%s", *input.ImageFileID)
		fileUUID, err := uuid.Parse(*input.ImageFileID)
		if err != nil {
			commons.LogW(ctx, "[core.AddQuestion::InvalidImageFileID [err=%v", err)
			return nil, commons.BadInput("invalid-image-file-id")
		}

		// Verify file exists and belongs to user
		file, err := aidao.GetFileByID(ctx, fileUUID)
		if err != nil {
			commons.LogE(ctx, "[core.AddQuestion::GetFileErr [file_id=%s [err=%v", *input.ImageFileID, err)
			if commons.IsNotFound(err) {
				return nil, commons.BadInput("image-file-not-found")
			}
			return nil, err
		}

		if file.UserID != userID {
			commons.LogW(ctx, "[core.AddQuestion::ImageFileNotOwned [file_id=%s [user_id=%s", *input.ImageFileID, userID)
			return nil, commons.Forbidden("not-image-file-owner")
		}

		// Populate meta with image information
		input.Meta["image_url"] = fmt.Sprintf("/api/v1/files/serve/%s", file.ID)
		input.Meta["image_file_id"] = file.ID
		commons.Log(ctx, "[core.AddQuestion [attached_image=%s", file.ID)
	}

	// Create question
	question := &model.Question{
		ID:                uuid.New().String(),
		QuizID:            quizID,
		SectionID:         section.ID,
		QuestionText:      input.QuestionText,
		QuestionType:      input.QuestionType,
		Difficulty:        input.Difficulty,
		Tags:              input.Tags,
		Points:            input.Points,
		OrderIndex:        count, // Append to end
		SectionOrderIndex: sectionOrderIndex,
		Explanation:       input.Explanation,
		Meta:              input.Meta,
	}

	// Validate question
	if err := question.Validate(); err != nil {
		commons.LogW(ctx, "[core.AddQuestion::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	// Create question in database
	if err := quizdao.CreateQuestion(ctx, question); err != nil {
		commons.LogE(ctx, "[core.AddQuestion::CreateErr [err=%v", err)
		return nil, err
	}

	// If answers provided, create them
	if len(input.Answers) > 0 {
		answers := []*model.Answer{}
		for i, answerInput := range input.Answers {
			answer := &model.Answer{
				ID:         uuid.New().String(),
				QuestionID: question.ID,
				AnswerText: answerInput.AnswerText,
				IsCorrect:  answerInput.IsCorrect,
				OrderIndex: i,
			}
			answers = append(answers, answer)
		}

		// Validate answers with question
		answerModels := []model.Answer{}
		for _, a := range answers {
			answerModels = append(answerModels, *a)
		}
		if err := question.ValidateWithAnswers(answerModels); err != nil {
			// Delete the question since answers are invalid
			questionUUID, _ := uuid.Parse(question.ID)
			quizdao.DeleteQuestion(ctx, questionUUID)
			commons.LogW(ctx, "[core.AddQuestion::AnswerValidationErr [err=%v", err)
			return nil, commons.BadInput(err.Error())
		}

		// Create answers in bulk
		if err := quizdao.CreateAnswersBulk(ctx, answers); err != nil {
			// Delete the question since answers failed to create
			questionUUID, _ := uuid.Parse(question.ID)
			quizdao.DeleteQuestion(ctx, questionUUID)
			commons.LogE(ctx, "[core.AddQuestion::CreateAnswersErr [err=%v", err)
			return nil, err
		}
	}

	commons.Log(ctx, "[core.AddQuestion::OK [question_id=%s [answer_count=%d", question.ID, len(input.Answers))
	return question, nil
}

// AddQuestionsBulk adds multiple questions to a quiz in a single transaction
func AddQuestionsBulk(ctx context.Context, quizID string, userID string, input model.BulkQuestionsInput) ([]*model.Question, error) {
	commons.Log(ctx, "[core.AddQuestionsBulk [quiz_id=%s [user_id=%s [count=%d", quizID, userID, len(input.Questions))

	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, commons.BadInput("invalid-quiz-id")
	}

	// Get quiz and check ownership
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddQuestionsBulk::FindQuizErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.AddQuestionsBulk::NotOwner [quiz_id=%s [user_id=%s", quizID, userID)
		return nil, commons.Forbidden("not-quiz-owner")
	}

	// Get current question count for order_index starting point
	startOrderIndex, err := quizdao.CountByQuizID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddQuestionsBulk::CountErr [quiz_id=%s [err=%v", quizID, err)
		return nil, err
	}

	defaultSection, err := quizdao.EnsureDefaultSection(ctx, quizUUID)
	if err != nil {
		return nil, err
	}
	sectionOrderByID := map[string]int{}

	// Prepare questions and answers
	questions := make([]*model.Question, 0, len(input.Questions))
	allAnswers := make([]*model.Answer, 0)

	for i, questionInput := range input.Questions {
		sectionID := questionInput.SectionID
		if sectionID == "" {
			sectionID = defaultSection.ID
		}
		section, err := resolveQuestionSection(ctx, quizUUID, sectionID)
		if err != nil {
			return nil, err
		}
		if _, ok := sectionOrderByID[section.ID]; !ok {
			sectionUUID, _ := uuid.Parse(section.ID)
			nextIndex, err := quizdao.NextSectionQuestionOrderIndex(ctx, sectionUUID)
			if err != nil {
				return nil, err
			}
			sectionOrderByID[section.ID] = nextIndex
		}
		sectionOrderIndex := sectionOrderByID[section.ID]
		sectionOrderByID[section.ID] = sectionOrderIndex + 1
		// Initialize meta if nil
		if questionInput.Meta == nil {
			questionInput.Meta = model.JSONMap{}
		}

		// Process image file if provided
		if questionInput.ImageFileID != nil && *questionInput.ImageFileID != "" {
			fileUUID, err := uuid.Parse(*questionInput.ImageFileID)
			if err != nil {
				commons.LogW(ctx, "[core.AddQuestionsBulk::InvalidImageFileID [question_index=%d [err=%v", i, err)
				return nil, commons.BadInput(fmt.Sprintf("question %d: invalid-image-file-id", i))
			}

			// Verify file exists and belongs to user
			file, err := aidao.GetFileByID(ctx, fileUUID)
			if err != nil {
				commons.LogE(ctx, "[core.AddQuestionsBulk::GetFileErr [question_index=%d [file_id=%s [err=%v", i, *questionInput.ImageFileID, err)
				if commons.IsNotFound(err) {
					return nil, commons.BadInput(fmt.Sprintf("question %d: image-file-not-found", i))
				}
				return nil, err
			}

			if file.UserID != userID {
				commons.LogW(ctx, "[core.AddQuestionsBulk::ImageFileNotOwned [question_index=%d [file_id=%s [user_id=%s", i, *questionInput.ImageFileID, userID)
				return nil, commons.Forbidden(fmt.Sprintf("question %d: not-image-file-owner", i))
			}

			// Populate meta with image information
			questionInput.Meta["image_url"] = fmt.Sprintf("/api/v1/files/serve/%s", file.ID)
			questionInput.Meta["image_file_id"] = file.ID
			commons.Log(ctx, "[core.AddQuestionsBulk [question_index=%d [attached_image=%s", i, file.ID)
		}

		// Create question
		question := &model.Question{
			ID:                uuid.New().String(),
			QuizID:            quizID,
			SectionID:         section.ID,
			QuestionText:      questionInput.QuestionText,
			QuestionType:      questionInput.QuestionType,
			Difficulty:        questionInput.Difficulty,
			Tags:              questionInput.Tags,
			Points:            questionInput.Points,
			OrderIndex:        startOrderIndex + i,
			SectionOrderIndex: sectionOrderIndex,
			Explanation:       questionInput.Explanation,
			Meta:              questionInput.Meta,
		}

		// Validate question
		if err := question.Validate(); err != nil {
			commons.LogW(ctx, "[core.AddQuestionsBulk::ValidationErr [question_index=%d [err=%v", i, err)
			return nil, commons.BadInput("question " + string(rune(i)) + ": " + err.Error())
		}

		// Prepare answers if provided
		if len(questionInput.Answers) > 0 {
			answers := make([]*model.Answer, 0, len(questionInput.Answers))
			for j, answerInput := range questionInput.Answers {
				answer := &model.Answer{
					ID:         uuid.New().String(),
					QuestionID: question.ID,
					AnswerText: answerInput.AnswerText,
					IsCorrect:  answerInput.IsCorrect,
					OrderIndex: j,
				}
				answers = append(answers, answer)
			}

			// Validate answers with question
			answerModels := []model.Answer{}
			for _, a := range answers {
				answerModels = append(answerModels, *a)
			}
			if err := question.ValidateWithAnswers(answerModels); err != nil {
				commons.LogW(ctx, "[core.AddQuestionsBulk::AnswerValidationErr [question_index=%d [err=%v", i, err)
				return nil, commons.BadInput("question " + string(rune(i)) + " answers: " + err.Error())
			}

			allAnswers = append(allAnswers, answers...)
		}

		questions = append(questions, question)
	}

	// Create all questions in a transaction
	if err := quizdao.CreateQuestionsBulk(ctx, questions); err != nil {
		commons.LogE(ctx, "[core.AddQuestionsBulk::CreateQuestionsErr [err=%v", err)
		return nil, err
	}

	// Create all answers in a transaction
	if len(allAnswers) > 0 {
		if err := quizdao.CreateAnswersBulk(ctx, allAnswers); err != nil {
			commons.LogE(ctx, "[core.AddQuestionsBulk::CreateAnswersErr [err=%v", err)
			// Note: Questions are already committed, but we return error
			// Consider implementing a cleanup mechanism or single transaction for both
			return nil, err
		}
	}

	commons.Log(ctx, "[core.AddQuestionsBulk::OK [question_count=%d [answer_count=%d", len(questions), len(allAnswers))
	return questions, nil
}

// UpdateQuestion updates a question
func UpdateQuestion(ctx context.Context, questionID string, userID string, input model.QuestionUpdateInput) (*model.Question, error) {
	commons.Log(ctx, "[core.UpdateQuestion [question_id=%s [user_id=%s", questionID, userID)

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, commons.BadInput("invalid-question-id")
	}

	// Get question
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateQuestion::FindErr [question_id=%s [err=%v", questionID, err)
		return nil, err
	}

	// Get quiz and check ownership
	quizUUID, _ := uuid.Parse(question.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateQuestion::FindQuizErr [quiz_id=%s [err=%v", question.QuizID, err)
		return nil, err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.UpdateQuestion::NotOwner [quiz_id=%s [user_id=%s", question.QuizID, userID)
		return nil, commons.Forbidden("not-quiz-owner")
	}

	if input.SectionID != nil && *input.SectionID != "" && *input.SectionID != question.SectionID {
		section, err := resolveQuestionSection(ctx, quizUUID, *input.SectionID)
		if err != nil {
			return nil, err
		}
		sectionUUID, _ := uuid.Parse(section.ID)
		sectionOrderIndex, err := quizdao.NextSectionQuestionOrderIndex(ctx, sectionUUID)
		if err != nil {
			return nil, err
		}
		question.SectionID = section.ID
		question.SectionOrderIndex = sectionOrderIndex
	}

	// Update fields
	if input.QuestionText != "" {
		question.QuestionText = input.QuestionText
	}
	if input.QuestionType != "" {
		question.QuestionType = input.QuestionType
	}
	if input.Difficulty != nil {
		question.Difficulty = *input.Difficulty
	}
	if input.Tags != nil {
		question.Tags = input.Tags
	}
	if input.Points != nil {
		question.Points = *input.Points
	}
	if input.Explanation != nil {
		question.Explanation = *input.Explanation
	}
	if input.Meta != nil {
		question.Meta = input.Meta
	}

	// Handle image file updates
	if input.ImageFileID != nil {
		// Initialize meta if nil
		if question.Meta == nil {
			question.Meta = model.JSONMap{}
		}

		if *input.ImageFileID == "" {
			// Empty string means remove image
			delete(question.Meta, "image_url")
			delete(question.Meta, "image_file_id")
			commons.Log(ctx, "[core.UpdateQuestion [removed_image]")
		} else {
			// New image file ID - validate and update
			fileUUID, err := uuid.Parse(*input.ImageFileID)
			if err != nil {
				commons.LogW(ctx, "[core.UpdateQuestion::InvalidImageFileID [err=%v", err)
				return nil, commons.BadInput("invalid-image-file-id")
			}

			// Verify file exists and belongs to user
			file, err := aidao.GetFileByID(ctx, fileUUID)
			if err != nil {
				commons.LogE(ctx, "[core.UpdateQuestion::GetFileErr [file_id=%s [err=%v", *input.ImageFileID, err)
				if commons.IsNotFound(err) {
					return nil, commons.BadInput("image-file-not-found")
				}
				return nil, err
			}

			if file.UserID != userID {
				commons.LogW(ctx, "[core.UpdateQuestion::ImageFileNotOwned [file_id=%s [user_id=%s", *input.ImageFileID, userID)
				return nil, commons.Forbidden("not-image-file-owner")
			}

			// Update meta with new image information
			question.Meta["image_url"] = fmt.Sprintf("/api/v1/files/serve/%s", file.ID)
			question.Meta["image_file_id"] = file.ID
			commons.Log(ctx, "[core.UpdateQuestion [updated_image=%s", file.ID)
		}
	}

	// Validate
	if err := question.Validate(); err != nil {
		commons.LogW(ctx, "[core.UpdateQuestion::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	// Update in database
	if err := quizdao.UpdateQuestion(ctx, question); err != nil {
		commons.LogE(ctx, "[core.UpdateQuestion::UpdateErr [question_id=%s [err=%v", questionID, err)
		return nil, err
	}

	commons.Log(ctx, "[core.UpdateQuestion::OK [question_id=%s", questionID)
	return question, nil
}

// DeleteQuestion deletes a question
func DeleteQuestion(ctx context.Context, questionID string, userID string) error {
	commons.Log(ctx, "[core.DeleteQuestion [question_id=%s [user_id=%s", questionID, userID)

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return commons.BadInput("invalid-question-id")
	}

	// Get question
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteQuestion::FindErr [question_id=%s [err=%v", questionID, err)
		return err
	}

	// Get quiz and check ownership
	quizUUID, _ := uuid.Parse(question.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteQuestion::FindQuizErr [quiz_id=%s [err=%v", question.QuizID, err)
		return err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.DeleteQuestion::NotOwner [quiz_id=%s [user_id=%s", question.QuizID, userID)
		return commons.Forbidden("not-quiz-owner")
	}

	// Delete question (CASCADE will delete answers)
	if err := quizdao.DeleteQuestion(ctx, questionUUID); err != nil {
		commons.LogE(ctx, "[core.DeleteQuestion::DeleteErr [question_id=%s [err=%v", questionID, err)
		return err
	}

	commons.Log(ctx, "[core.DeleteQuestion::OK [question_id=%s", questionID)
	return nil
}

// GetQuestionWithAnswers retrieves a question with its answers
func GetQuestionWithAnswers(ctx context.Context, questionID string, userID string) (*model.QuestionWithAnswers, error) {
	commons.Log(ctx, "[core.GetQuestionWithAnswers [question_id=%s [user_id=%s", questionID, userID)

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, commons.BadInput("invalid-question-id")
	}

	// Get question
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuestionWithAnswers::FindErr [question_id=%s [err=%v", questionID, err)
		return nil, err
	}

	// Get quiz and check access
	quizUUID, _ := uuid.Parse(question.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuestionWithAnswers::FindQuizErr [quiz_id=%s [err=%v", question.QuizID, err)
		return nil, err
	}

	if !quiz.IsAccessibleBy(userID) {
		commons.LogW(ctx, "[core.GetQuestionWithAnswers::AccessDenied [quiz_id=%s [user_id=%s", question.QuizID, userID)
		return nil, commons.Forbidden("access-denied")
	}

	// Get answers
	answers, err := quizdao.ListByQuestionID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.GetQuestionWithAnswers::ListAnswersErr [question_id=%s [err=%v", questionID, err)
		return nil, err
	}

	// Convert to model
	answerModels := []model.Answer{}
	for _, a := range answers {
		answerModels = append(answerModels, *a)
	}

	result := &model.QuestionWithAnswers{
		Question: *question,
		Answers:  answerModels,
	}

	commons.Log(ctx, "[core.GetQuestionWithAnswers::OK [question_id=%s [answer_count=%d", questionID, len(answers))
	return result, nil
}

// AddAnswer adds an answer to a question
func AddAnswer(ctx context.Context, questionID string, userID string, input model.AnswerInput) (*model.Answer, error) {
	commons.Log(ctx, "[core.AddAnswer [question_id=%s [user_id=%s", questionID, userID)

	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, commons.BadInput("invalid-question-id")
	}

	// Get question
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddAnswer::FindQuestionErr [question_id=%s [err=%v", questionID, err)
		return nil, err
	}

	// Get quiz and check ownership
	quizUUID, _ := uuid.Parse(question.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddAnswer::FindQuizErr [quiz_id=%s [err=%v", question.QuizID, err)
		return nil, err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.AddAnswer::NotOwner [quiz_id=%s [user_id=%s", question.QuizID, userID)
		return nil, commons.Forbidden("not-quiz-owner")
	}

	// Get current answer count for order_index
	count, err := quizdao.CountByQuestionID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.AddAnswer::CountErr [question_id=%s [err=%v", questionID, err)
		return nil, err
	}

	// Create answer
	answer := &model.Answer{
		ID:         uuid.New().String(),
		QuestionID: questionID,
		AnswerText: input.AnswerText,
		IsCorrect:  input.IsCorrect,
		OrderIndex: count, // Append to end
	}

	// Validate answer
	if err := answer.Validate(); err != nil {
		commons.LogW(ctx, "[core.AddAnswer::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	// Create answer in database
	if err := quizdao.CreateAnswer(ctx, answer); err != nil {
		commons.LogE(ctx, "[core.AddAnswer::CreateErr [err=%v", err)
		return nil, err
	}

	commons.Log(ctx, "[core.AddAnswer::OK [answer_id=%s", answer.ID)
	return answer, nil
}

// UpdateAnswer updates an answer
func UpdateAnswer(ctx context.Context, answerID string, userID string, input model.AnswerInput) (*model.Answer, error) {
	commons.Log(ctx, "[core.UpdateAnswer [answer_id=%s [user_id=%s", answerID, userID)

	answerUUID, err := uuid.Parse(answerID)
	if err != nil {
		return nil, commons.BadInput("invalid-answer-id")
	}

	// Get answer
	answer, err := quizdao.FindAnswerByID(ctx, answerUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateAnswer::FindErr [answer_id=%s [err=%v", answerID, err)
		return nil, err
	}

	// Get question
	questionUUID, _ := uuid.Parse(answer.QuestionID)
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateAnswer::FindQuestionErr [question_id=%s [err=%v", answer.QuestionID, err)
		return nil, err
	}

	// Get quiz and check ownership
	quizUUID, _ := uuid.Parse(question.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.UpdateAnswer::FindQuizErr [quiz_id=%s [err=%v", question.QuizID, err)
		return nil, err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.UpdateAnswer::NotOwner [quiz_id=%s [user_id=%s", question.QuizID, userID)
		return nil, commons.Forbidden("not-quiz-owner")
	}

	// Update fields
	answer.AnswerText = input.AnswerText
	answer.IsCorrect = input.IsCorrect
	if input.OrderIndex >= 0 {
		answer.OrderIndex = input.OrderIndex
	}

	// Validate
	if err := answer.Validate(); err != nil {
		commons.LogW(ctx, "[core.UpdateAnswer::ValidationErr [err=%v", err)
		return nil, commons.BadInput(err.Error())
	}

	// Update in database
	if err := quizdao.UpdateAnswer(ctx, answer); err != nil {
		commons.LogE(ctx, "[core.UpdateAnswer::UpdateErr [answer_id=%s [err=%v", answerID, err)
		return nil, err
	}

	commons.Log(ctx, "[core.UpdateAnswer::OK [answer_id=%s", answerID)
	return answer, nil
}

// DeleteAnswer deletes an answer
func DeleteAnswer(ctx context.Context, answerID string, userID string) error {
	commons.Log(ctx, "[core.DeleteAnswer [answer_id=%s [user_id=%s", answerID, userID)

	answerUUID, err := uuid.Parse(answerID)
	if err != nil {
		return commons.BadInput("invalid-answer-id")
	}

	// Get answer
	answer, err := quizdao.FindAnswerByID(ctx, answerUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteAnswer::FindErr [answer_id=%s [err=%v", answerID, err)
		return err
	}

	// Get question
	questionUUID, _ := uuid.Parse(answer.QuestionID)
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteAnswer::FindQuestionErr [question_id=%s [err=%v", answer.QuestionID, err)
		return err
	}

	// Get quiz and check ownership
	quizUUID, _ := uuid.Parse(question.QuizID)
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		commons.LogE(ctx, "[core.DeleteAnswer::FindQuizErr [quiz_id=%s [err=%v", question.QuizID, err)
		return err
	}

	if !quiz.IsOwnedBy(userID) {
		commons.LogW(ctx, "[core.DeleteAnswer::NotOwner [quiz_id=%s [user_id=%s", question.QuizID, userID)
		return commons.Forbidden("not-quiz-owner")
	}

	// Check if this is the last answer for a question that requires answers
	if question.RequiresAnswers() {
		count, err := quizdao.CountByQuestionID(ctx, questionUUID)
		if err != nil {
			commons.LogE(ctx, "[core.DeleteAnswer::CountErr [question_id=%s [err=%v", answer.QuestionID, err)
			return err
		}
		if count <= question.MinimumAnswers() {
			return commons.BadInput(fmt.Sprintf("cannot-delete-answer-minimum-%d-required", question.MinimumAnswers()))
		}
	}

	// Delete answer
	if err := quizdao.DeleteAnswer(ctx, answerUUID); err != nil {
		commons.LogE(ctx, "[core.DeleteAnswer::DeleteErr [answer_id=%s [err=%v", answerID, err)
		return err
	}

	commons.Log(ctx, "[core.DeleteAnswer::OK [answer_id=%s", answerID)
	return nil
}

func resolveQuestionSection(ctx context.Context, quizID uuid.UUID, sectionID string) (*model.QuizSection, error) {
	if sectionID == "" {
		return quizdao.EnsureDefaultSection(ctx, quizID)
	}
	sectionUUID, err := uuid.Parse(sectionID)
	if err != nil {
		return nil, commons.BadInput("invalid-section-id")
	}
	section, err := quizdao.FindSectionByID(ctx, sectionUUID)
	if err != nil {
		return nil, err
	}
	if section.QuizID != quizID.String() {
		return nil, commons.BadInput("section-does-not-belong-to-quiz")
	}
	return section, nil
}

func requireQuizOwner(ctx context.Context, quizID string, userID string) (*model.Quiz, uuid.UUID, error) {
	quizUUID, err := uuid.Parse(quizID)
	if err != nil {
		return nil, uuid.Nil, commons.BadInput("invalid-quiz-id")
	}
	quiz, err := quizdao.FindByID(ctx, quizUUID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if !quiz.IsOwnedBy(userID) {
		return nil, uuid.Nil, commons.Forbidden("not-quiz-owner")
	}
	return quiz, quizUUID, nil
}

// ListSections lists sections for a quiz.
func ListSections(ctx context.Context, quizID string, userID string) ([]model.QuizSection, error) {
	_, quizUUID, err := requireQuizOwner(ctx, quizID, userID)
	if err != nil {
		return nil, err
	}
	sections, err := quizdao.ListSectionsByQuizID(ctx, quizUUID)
	if err != nil {
		return nil, err
	}
	if len(sections) == 0 {
		section, err := quizdao.EnsureDefaultSection(ctx, quizUUID)
		if err != nil {
			return nil, err
		}
		sections = []*model.QuizSection{section}
	}
	result := make([]model.QuizSection, 0, len(sections))
	for _, section := range sections {
		result = append(result, *section)
	}
	return result, nil
}

// CreateSection creates a section for a quiz.
func CreateSection(ctx context.Context, quizID string, userID string, input model.QuizSectionInput) (*model.QuizSection, error) {
	_, quizUUID, err := requireQuizOwner(ctx, quizID, userID)
	if err != nil {
		return nil, err
	}
	if input.Title == "" {
		return nil, commons.BadInput("section-title-required")
	}
	orderIndex, err := quizdao.NextSectionOrderIndex(ctx, quizUUID)
	if err != nil {
		return nil, err
	}
	section := &model.QuizSection{
		ID:               uuid.New().String(),
		QuizID:           quizID,
		Title:            input.Title,
		Description:      input.Description,
		Instructions:     input.Instructions,
		OrderIndex:       orderIndex,
		TimeLimitMinutes: input.TimeLimitMinutes,
		IsDefault:        false,
		Meta:             input.Meta,
	}
	if section.Meta == nil {
		section.Meta = model.JSONMap{}
	}
	if err := quizdao.CreateSection(ctx, section); err != nil {
		return nil, err
	}
	return section, nil
}

// UpdateSection updates section metadata.
func UpdateSection(ctx context.Context, quizID string, sectionID string, userID string, input model.QuizSectionUpdateInput) (*model.QuizSection, error) {
	_, quizUUID, err := requireQuizOwner(ctx, quizID, userID)
	if err != nil {
		return nil, err
	}
	section, err := resolveQuestionSection(ctx, quizUUID, sectionID)
	if err != nil {
		return nil, err
	}
	if section.IsDefault && input.Title != nil && *input.Title != "General" {
		return nil, commons.BadInput("default-section-title-cannot-change")
	}
	if input.Title != nil {
		if *input.Title == "" {
			return nil, commons.BadInput("section-title-required")
		}
		section.Title = *input.Title
	}
	if input.Description != nil {
		section.Description = *input.Description
	}
	if input.Instructions != nil {
		section.Instructions = *input.Instructions
	}
	if input.TimeLimitMinutes != nil {
		section.TimeLimitMinutes = input.TimeLimitMinutes
	}
	if input.Meta != nil {
		section.Meta = input.Meta
	}
	if err := quizdao.UpdateSection(ctx, section); err != nil {
		return nil, err
	}
	return section, nil
}

// DeleteSection deletes a section and moves questions to General.
func DeleteSection(ctx context.Context, quizID string, sectionID string, userID string) error {
	_, quizUUID, err := requireQuizOwner(ctx, quizID, userID)
	if err != nil {
		return err
	}
	sectionUUID, err := uuid.Parse(sectionID)
	if err != nil {
		return commons.BadInput("invalid-section-id")
	}
	return quizdao.DeleteSectionMoveQuestionsToDefault(ctx, quizUUID, sectionUUID)
}

// ReorderSections reorders quiz sections.
func ReorderSections(ctx context.Context, quizID string, userID string, input model.SectionReorderInput) error {
	_, quizUUID, err := requireQuizOwner(ctx, quizID, userID)
	if err != nil {
		return err
	}
	return quizdao.ReorderSections(ctx, quizUUID, input.Sections)
}

// ReorderQuestionsInSections reorders and moves questions between sections.
func ReorderQuestionsInSections(ctx context.Context, quizID string, userID string, input model.QuestionSectionReorderInput) error {
	_, quizUUID, err := requireQuizOwner(ctx, quizID, userID)
	if err != nil {
		return err
	}
	return quizdao.ReorderQuestionsInSections(ctx, quizUUID, input.Questions)
}

// MoveQuestionToSection moves a single question to the end of a section.
func MoveQuestionToSection(ctx context.Context, questionID string, userID string, input model.QuestionSectionUpdateInput) (*model.Question, error) {
	questionUUID, err := uuid.Parse(questionID)
	if err != nil {
		return nil, commons.BadInput("invalid-question-id")
	}
	question, err := quizdao.FindQuestionByID(ctx, questionUUID)
	if err != nil {
		return nil, err
	}
	_, quizUUID, err := requireQuizOwner(ctx, question.QuizID, userID)
	if err != nil {
		return nil, err
	}
	section, err := resolveQuestionSection(ctx, quizUUID, input.SectionID)
	if err != nil {
		return nil, err
	}
	sectionUUID, _ := uuid.Parse(section.ID)
	sectionOrderIndex, err := quizdao.NextSectionQuestionOrderIndex(ctx, sectionUUID)
	if err != nil {
		return nil, err
	}
	question.SectionID = section.ID
	question.SectionOrderIndex = sectionOrderIndex
	if err := quizdao.UpdateQuestion(ctx, question); err != nil {
		return nil, err
	}
	return question, nil
}
