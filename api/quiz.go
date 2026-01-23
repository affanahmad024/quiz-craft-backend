package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/affanahmad024/quiz-craft/backend/quiz/core"
	"github.com/affanahmad024/quiz-craft/backend/quiz/model"
	"github.com/gin-gonic/gin"
)

// CreateQuiz creates a new quiz
func CreateQuiz(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	var req model.QuizInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.CreateQuiz::BindErr [user_id=%s [err=%v", userID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.CreateQuiz [user_id=%s [title=%s", userID, req.Title)

	quiz, err := core.CreateQuiz(ctx, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.CreateQuiz::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.CreateQuiz::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quiz.ID)
	commons.RespOKWithMsg(c, "Quiz created successfully", quiz)
}

// GetQuiz retrieves a quiz by ID
func GetQuiz(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	commons.Log(ctx, "[api.GetQuiz [user_id=%s [quiz_id=%s", userID, quizID)

	quiz, err := core.GetQuiz(ctx, quizID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetQuiz::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetQuiz::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz retrieved successfully", quiz)
}

// GetQuizWithQuestions retrieves a quiz with all questions and answers
func GetQuizWithQuestions(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	hideCorrectStr := c.Query("hide_correct") // Query param: ?hide_correct=true
	hideCorrect := hideCorrectStr == "true"

	commons.Log(ctx, "[api.GetQuizWithQuestions [user_id=%s [quiz_id=%s [hide_correct=%v",
		userID, quizID, hideCorrect)

	quizWithQuestions, err := core.GetQuizWithQuestions(ctx, quizID, userID.String(), hideCorrect)
	if err != nil {
		commons.LogC(ctx, "[api.GetQuizWithQuestions::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetQuizWithQuestions::OK [time_taken=%v [ip=%s [quiz_id=%s [question_count=%d",
		time.Since(st), ip, quizID, len(quizWithQuestions.Questions))
	commons.RespOKWithMsg(c, "Quiz with questions retrieved successfully", quizWithQuestions)
}

// ListMyQuizzes lists quizzes created by current user
func ListMyQuizzes(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	// Parse filters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	difficulty := c.Query("difficulty")

	var isPublished *bool
	if publishedStr := c.Query("is_published"); publishedStr != "" {
		val := publishedStr == "true"
		isPublished = &val
	}

	filter := model.QuizListFilter{
		Page:        page,
		Limit:       limit,
		Category:    category,
		Difficulty:  difficulty,
		IsPublished: isPublished,
	}

	commons.Log(ctx, "[api.ListMyQuizzes [user_id=%s [filter=%+v", userID, filter)

	quizzes, total, err := core.ListMyQuizzes(ctx, userID.String(), filter)
	if err != nil {
		commons.LogC(ctx, "[api.ListMyQuizzes::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	response := map[string]interface{}{
		"quizzes": quizzes,
		"total":   total,
		"page":    page,
		"limit":   limit,
	}

	commons.Log(ctx, "[api.ListMyQuizzes::OK [time_taken=%v [ip=%s [user_id=%s [total=%d",
		time.Since(st), ip, userID, total)
	commons.RespOKWithMsg(c, "Quizzes retrieved successfully", response)
}

// ListPublicQuizzes lists public published quizzes
func ListPublicQuizzes(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	// Extract user ID if authenticated (optional for public quizzes)
	userID, _ := commons.UsrID(ctx)

	// Parse filters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	category := c.Query("category")
	difficulty := c.Query("difficulty")
	search := c.Query("search")

	filter := model.QuizListFilter{
		Page:       page,
		Limit:      limit,
		Category:   category,
		Difficulty: difficulty,
		Search:     search,
	}

	commons.Log(ctx, "[api.ListPublicQuizzes [user_id=%s [filter=%+v", userID, filter)

	quizzes, total, err := core.ListPublicQuizzes(ctx, userID.String(), filter)
	if err != nil {
		commons.LogC(ctx, "[api.ListPublicQuizzes::KO [time_taken=%v [ip=%s [err=%v",
			time.Since(st), ip, err)

		commons.RespISE(c, err)
		return
	}

	response := map[string]interface{}{
		"quizzes": quizzes,
		"total":   total,
		"page":    page,
		"limit":   limit,
	}

	commons.Log(ctx, "[api.ListPublicQuizzes::OK [time_taken=%v [ip=%s [total=%d",
		time.Since(st), ip, total)
	commons.RespOKWithMsg(c, "Public quizzes retrieved successfully", response)
}

// UpdateQuiz updates a quiz
func UpdateQuiz(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")

	var req model.QuizUpdateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.UpdateQuiz::BindErr [quiz_id=%s [err=%v", quizID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.UpdateQuiz [user_id=%s [quiz_id=%s", userID, quizID)

	quiz, err := core.UpdateQuiz(ctx, quizID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.UpdateQuiz::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UpdateQuiz::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz updated successfully", quiz)
}

// PublishQuiz publishes a quiz
func PublishQuiz(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	commons.Log(ctx, "[api.PublishQuiz [user_id=%s [quiz_id=%s", userID, quizID)

	err := core.PublishQuiz(ctx, quizID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.PublishQuiz::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.PublishQuiz::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz published successfully", nil)
}

// UnpublishQuiz unpublishes a quiz
func UnpublishQuiz(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	commons.Log(ctx, "[api.UnpublishQuiz [user_id=%s [quiz_id=%s", userID, quizID)

	err := core.UnpublishQuiz(ctx, quizID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.UnpublishQuiz::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UnpublishQuiz::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz unpublished successfully", nil)
}

// DeleteQuiz soft deletes a quiz
func DeleteQuiz(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	commons.Log(ctx, "[api.DeleteQuiz [user_id=%s [quiz_id=%s", userID, quizID)

	err := core.DeleteQuiz(ctx, quizID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.DeleteQuiz::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DeleteQuiz::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz deleted successfully", nil)
}

// AddQuestion adds a question to a quiz
// Accepts both JSON and multipart/form-data
func AddQuestion(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")

	var req model.QuestionInput
	contentType := c.ContentType()

	// Handle multipart/form-data with optional image
	if strings.Contains(contentType, "multipart/form-data") {
		// Get optional image file
		imageFile, imageHeader, _ := c.Request.FormFile("image")
		if imageFile != nil {
			defer imageFile.Close()

			// Upload image
			fileUpload, err := fileUploadService.SaveFile(ctx, imageFile, imageHeader, userID.String())
			if err != nil {
				commons.LogE(ctx, "[api.AddQuestion::ImageUploadErr [err=%v", err)
				if commons.IsBadInput(err) {
					commons.RespBadReq(c, commons.GetCode(err))
					return
				}
				commons.RespISE(c, err)
				return
			}

			imageFileID := fileUpload.ID
			req.ImageFileID = &imageFileID
		}

		// Check if question_data field exists (JSON string containing all question data)
		if questionDataJSON := c.PostForm("question_data"); questionDataJSON != "" {
			// Parse the entire question from JSON
			if err := json.Unmarshal([]byte(questionDataJSON), &req); err != nil {
				commons.LogW(ctx, "[api.AddQuestion::QuestionDataParseErr [err=%v", err)
				commons.RespBadReq(c, "invalid question_data JSON")
				return
			}
		} else {
			// Parse individual form fields (backward compatibility)
			req.QuestionText = c.PostForm("question_text")
			req.QuestionType = c.PostForm("question_type")
			req.Explanation = c.PostForm("explanation")

			// Parse points
			if pointsStr := c.PostForm("points"); pointsStr != "" {
				points, err := strconv.Atoi(pointsStr)
				if err != nil {
					commons.RespBadReq(c, "invalid points value")
					return
				}
				req.Points = points
			} else {
				req.Points = 1
			}

			// Parse answers JSON if provided
			if answersJSON := c.PostForm("answers"); answersJSON != "" {
				var answers []model.AnswerInput
				if err := json.Unmarshal([]byte(answersJSON), &answers); err != nil {
					commons.LogW(ctx, "[api.AddQuestion::AnswersParseErr [err=%v", err)
					commons.RespBadReq(c, "invalid answers JSON")
					return
				}
				req.Answers = answers
			}

			// Parse meta JSON if provided
			if metaJSON := c.PostForm("meta"); metaJSON != "" {
				var meta model.JSONMap
				if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
					commons.LogW(ctx, "[api.AddQuestion::MetaParseErr [err=%v", err)
					commons.RespBadReq(c, "invalid meta JSON")
					return
				}
				req.Meta = meta
			}
		}
	} else {
		// Handle JSON (backward compatibility)
		if err := c.ShouldBindJSON(&req); err != nil {
			commons.LogW(ctx, "[api.AddQuestion::BindErr [quiz_id=%s [err=%v", quizID, err)
			commons.RespBadReq(c, "invalid-request-body")
			return
		}
	}

	commons.Log(ctx, "[api.AddQuestion [user_id=%s [quiz_id=%s", userID, quizID)

	question, err := core.AddQuestion(ctx, quizID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.AddQuestion::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.AddQuestion::OK [time_taken=%v [ip=%s [question_id=%s",
		time.Since(st), ip, question.ID)
	commons.RespOKWithMsg(c, "Question added successfully", question)
}

// AddQuestionsBulk adds multiple questions to a quiz
// Accepts JSON or multipart/form-data with images named "image_0", "image_1", etc.
func AddQuestionsBulk(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")

	var req model.BulkQuestionsInput
	contentType := c.ContentType()

	// Handle multipart/form-data with optional multiple images
	if strings.Contains(contentType, "multipart/form-data") {
		// Parse questions JSON - support both "questions" and "questions_data" field names
		questionsJSON := c.PostForm("questions_data")
		if questionsJSON == "" {
			questionsJSON = c.PostForm("questions")
		}
		if questionsJSON == "" {
			commons.RespBadReq(c, "questions or questions_data field is required")
			return
		}

		if err := json.Unmarshal([]byte(questionsJSON), &req.Questions); err != nil {
			commons.LogW(ctx, "[api.AddQuestionsBulk::QuestionsParseErr [err=%v", err)
			commons.RespBadReq(c, "invalid questions JSON")
			return
		}

		// Upload images and map them to questions by index
		form, err := c.MultipartForm()
		if err == nil && form.File != nil {
			for i := range req.Questions {
				imageKey := fmt.Sprintf("image_%d", i)
				if fileHeaders, exists := form.File[imageKey]; exists && len(fileHeaders) > 0 {
					fileHeader := fileHeaders[0]
					imageFile, err := fileHeader.Open()
					if err != nil {
						commons.LogW(ctx, "[api.AddQuestionsBulk::ImageOpenErr [index=%d [err=%v", i, err)
						continue
					}

					// Upload image
					fileUpload, err := fileUploadService.SaveFile(ctx, imageFile, fileHeader, userID.String())
					imageFile.Close()
					if err != nil {
						commons.LogE(ctx, "[api.AddQuestionsBulk::ImageUploadErr [index=%d [err=%v", i, err)
						if commons.IsBadInput(err) {
							commons.RespBadReq(c, fmt.Sprintf("image_%d: %s", i, commons.GetCode(err)))
							return
						}
						commons.RespISE(c, err)
						return
					}

					imageFileID := fileUpload.ID
					req.Questions[i].ImageFileID = &imageFileID
				}
			}
		}
	} else {
		// Handle JSON (backward compatibility)
		if err := c.ShouldBindJSON(&req); err != nil {
			commons.LogW(ctx, "[api.AddQuestionsBulk::BindErr [quiz_id=%s [err=%v", quizID, err)
			commons.RespBadReq(c, "invalid-request-body")
			return
		}
	}

	commons.Log(ctx, "[api.AddQuestionsBulk [user_id=%s [quiz_id=%s [count=%d", userID, quizID, len(req.Questions))

	questions, err := core.AddQuestionsBulk(ctx, quizID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.AddQuestionsBulk::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.AddQuestionsBulk::OK [time_taken=%v [ip=%s [count=%d",
		time.Since(st), ip, len(questions))
	commons.RespOKWithMsg(c, "Questions added successfully", gin.H{
		"questions": questions,
		"count":     len(questions),
	})
}

// GetQuestion retrieves a question with answers
func GetQuestion(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	questionID := c.Param("qid")
	commons.Log(ctx, "[api.GetQuestion [user_id=%s [question_id=%s", userID, questionID)

	questionWithAnswers, err := core.GetQuestionWithAnswers(ctx, questionID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetQuestion::KO [time_taken=%v [ip=%s [question_id=%s [err=%v",
			time.Since(st), ip, questionID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetQuestion::OK [time_taken=%v [ip=%s [question_id=%s",
		time.Since(st), ip, questionID)
	commons.RespOKWithMsg(c, "Question retrieved successfully", questionWithAnswers)
}

// UpdateQuestion updates a question
// Accepts both JSON and multipart/form-data with optional image
func UpdateQuestion(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	questionID := c.Param("qid")

	var req model.QuestionUpdateInput
	contentType := c.ContentType()

	// Handle multipart/form-data with optional image
	if strings.Contains(contentType, "multipart/form-data") {
		// Check for image removal flag
		if c.PostForm("remove_image") == "true" {
			emptyStr := ""
			req.ImageFileID = &emptyStr
		} else {
			// Get optional new image file
			imageFile, imageHeader, _ := c.Request.FormFile("image")
			if imageFile != nil {
				defer imageFile.Close()

				// Upload new image
				fileUpload, err := fileUploadService.SaveFile(ctx, imageFile, imageHeader, userID.String())
				if err != nil {
					commons.LogE(ctx, "[api.UpdateQuestion::ImageUploadErr [err=%v", err)
					if commons.IsBadInput(err) {
						commons.RespBadReq(c, commons.GetCode(err))
						return
					}
					commons.RespISE(c, err)
					return
				}

				imageFileID := fileUpload.ID
				req.ImageFileID = &imageFileID
			}
		}

		// Check if question_data field exists (JSON string containing all question data)
		if questionDataJSON := c.PostForm("question_data"); questionDataJSON != "" {
			// Parse the entire question update from JSON
			if err := json.Unmarshal([]byte(questionDataJSON), &req); err != nil {
				commons.LogW(ctx, "[api.UpdateQuestion::QuestionDataParseErr [err=%v", err)
				commons.RespBadReq(c, "invalid question_data JSON")
				return
			}
		} else {
			// Parse individual form fields (backward compatibility)
			if questionText := c.PostForm("question_text"); questionText != "" {
				req.QuestionText = questionText
			}
			if questionType := c.PostForm("question_type"); questionType != "" {
				req.QuestionType = questionType
			}
			if explanation := c.PostForm("explanation"); explanation != "" {
				req.Explanation = &explanation
			}

			// Parse points if provided
			if pointsStr := c.PostForm("points"); pointsStr != "" {
				points, err := strconv.Atoi(pointsStr)
				if err != nil {
					commons.RespBadReq(c, "invalid points value")
					return
				}
				req.Points = &points
			}

			// Parse meta JSON if provided
			if metaJSON := c.PostForm("meta"); metaJSON != "" {
				var meta model.JSONMap
				if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
					commons.LogW(ctx, "[api.UpdateQuestion::MetaParseErr [err=%v", err)
					commons.RespBadReq(c, "invalid meta JSON")
					return
				}
				req.Meta = meta
			}
		}
	} else {
		// Handle JSON (backward compatibility)
		if err := c.ShouldBindJSON(&req); err != nil {
			commons.LogW(ctx, "[api.UpdateQuestion::BindErr [question_id=%s [err=%v", questionID, err)
			commons.RespBadReq(c, "invalid-request-body")
			return
		}
	}

	commons.Log(ctx, "[api.UpdateQuestion [user_id=%s [question_id=%s", userID, questionID)

	question, err := core.UpdateQuestion(ctx, questionID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.UpdateQuestion::KO [time_taken=%v [ip=%s [question_id=%s [err=%v",
			time.Since(st), ip, questionID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UpdateQuestion::OK [time_taken=%v [ip=%s [question_id=%s",
		time.Since(st), ip, questionID)
	commons.RespOKWithMsg(c, "Question updated successfully", question)
}

// DeleteQuestion deletes a question
func DeleteQuestion(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	questionID := c.Param("qid")
	commons.Log(ctx, "[api.DeleteQuestion [user_id=%s [question_id=%s", userID, questionID)

	err := core.DeleteQuestion(ctx, questionID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.DeleteQuestion::KO [time_taken=%v [ip=%s [question_id=%s [err=%v",
			time.Since(st), ip, questionID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DeleteQuestion::OK [time_taken=%v [ip=%s [question_id=%s",
		time.Since(st), ip, questionID)
	commons.RespOKWithMsg(c, "Question deleted successfully", nil)
}

// AddAnswer adds an answer to a question
func AddAnswer(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	questionID := c.Param("qid")

	var req model.AnswerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.AddAnswer::BindErr [question_id=%s [err=%v", questionID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.AddAnswer [user_id=%s [question_id=%s", userID, questionID)

	answer, err := core.AddAnswer(ctx, questionID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.AddAnswer::KO [time_taken=%v [ip=%s [question_id=%s [err=%v",
			time.Since(st), ip, questionID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.AddAnswer::OK [time_taken=%v [ip=%s [answer_id=%s",
		time.Since(st), ip, answer.ID)
	commons.RespOKWithMsg(c, "Answer added successfully", answer)
}

// UpdateAnswer updates an answer
func UpdateAnswer(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	answerID := c.Param("aid")

	var req model.AnswerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogW(ctx, "[api.UpdateAnswer::BindErr [answer_id=%s [err=%v", answerID, err)
		commons.RespBadReq(c, "invalid-request-body")
		return
	}

	commons.Log(ctx, "[api.UpdateAnswer [user_id=%s [answer_id=%s", userID, answerID)

	answer, err := core.UpdateAnswer(ctx, answerID, userID.String(), req)
	if err != nil {
		commons.LogC(ctx, "[api.UpdateAnswer::KO [time_taken=%v [ip=%s [answer_id=%s [err=%v",
			time.Since(st), ip, answerID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.UpdateAnswer::OK [time_taken=%v [ip=%s [answer_id=%s",
		time.Since(st), ip, answerID)
	commons.RespOKWithMsg(c, "Answer updated successfully", answer)
}

// DeleteAnswer deletes an answer
func DeleteAnswer(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	answerID := c.Param("aid")
	commons.Log(ctx, "[api.DeleteAnswer [user_id=%s [answer_id=%s", userID, answerID)

	err := core.DeleteAnswer(ctx, answerID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.DeleteAnswer::KO [time_taken=%v [ip=%s [answer_id=%s [err=%v",
			time.Since(st), ip, answerID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.DeleteAnswer::OK [time_taken=%v [ip=%s [answer_id=%s",
		time.Since(st), ip, answerID)
	commons.RespOKWithMsg(c, "Answer deleted successfully", nil)
}

// Phase 6: Advanced Features - Statistics and Sharing

// GetQuizStatistics gets detailed statistics for a quiz (creator only)
func GetQuizStatistics(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	includeQuestions := c.Query("include_questions") == "true"

	commons.Log(ctx, "[api.GetQuizStatistics [user_id=%s [quiz_id=%s", userID, quizID)

	stats, err := core.GetQuizStatistics(ctx, quizID, userID.String(), includeQuestions)
	if err != nil {
		commons.LogC(ctx, "[api.GetQuizStatistics::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetQuizStatistics::OK [time_taken=%v [ip=%s [quiz_id=%s",
		time.Since(st), ip, quizID)
	commons.RespOKWithMsg(c, "Quiz statistics retrieved successfully", stats)
}

// GetUserStatistics gets aggregate statistics for current user
func GetUserStatistics(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.GetUserStatistics [user_id=%s", userID)

	stats, err := core.GetUserStatistics(ctx, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GetUserStatistics::KO [time_taken=%v [ip=%s [user_id=%s [err=%v",
			time.Since(st), ip, userID, err)

		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GetUserStatistics::OK [time_taken=%v [ip=%s [user_id=%s",
		time.Since(st), ip, userID)
	commons.RespOKWithMsg(c, "User statistics retrieved successfully", stats)
}

// GenerateShareLink generates a share link for a quiz
func GenerateShareLink(c *gin.Context) {
	st := time.Now()
	ctx := commons.Ctx(c)
	tid := commons.Tid(ctx)
	c.Set(commons.TidRespHdr, tid)
	ip := c.ClientIP()

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	quizID := c.Param("id")
	commons.Log(ctx, "[api.GenerateShareLink [user_id=%s [quiz_id=%s", userID, quizID)

	result, err := core.GenerateShareLink(ctx, quizID, userID.String())
	if err != nil {
		commons.LogC(ctx, "[api.GenerateShareLink::KO [time_taken=%v [ip=%s [quiz_id=%s [err=%v",
			time.Since(st), ip, quizID, err)

		if commons.IsNotFound(err) {
			commons.Resp404(c, commons.GetCode(err))
			return
		}
		if commons.IsForbidden(err) {
			commons.Resp403(c, commons.GetCode(err))
			return
		}
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	commons.Log(ctx, "[api.GenerateShareLink::OK [time_taken=%v [ip=%s [quiz_id=%s [code=%s",
		time.Since(st), ip, quizID, result.ShareCode)
	commons.RespOKWithMsg(c, "Share link generated successfully", result)
}
