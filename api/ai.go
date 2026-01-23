package api

import (
	"net/http"
	"time"
	"strconv"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/dao"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/services"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var fileUploadService = services.NewFileUploadService()

// UploadFile handles file upload
// POST /api/v1/files/upload
func UploadFile(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.UploadFile [user_id=%s", userID)

	// Get file from request
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		commons.LogE(ctx, "[api.UploadFile::FormFileErr [err=%v", err)
		commons.RespBadReq(c, "file is required")
		return
	}
	defer file.Close()

	// Save file
	fileUpload, err := fileUploadService.SaveFile(ctx, file, fileHeader, userID.String())
	if err != nil {
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	// Prepare response
	response := model.FileUploadResponse{
		FileID:   fileUpload.ID,
		Filename: fileUpload.Filename,
		Size:     fileUpload.FileSize,
		MimeType: fileUpload.MimeType,
	}

	commons.RespOKWithMsg(c, "File uploaded successfully", response)
}

// ListFiles lists user's uploaded files with pagination
// GET /api/v1/files?page=1&limit=20
func ListFiles(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.ListFiles [user_id=%s", userID)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get files
	files, total, err := dao.ListFilesByUserID(ctx, userID, page, limit)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	response := model.ListFilesResponse{
		Files: files,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	commons.RespOK(c, response)
}

// DeleteFile deletes a user's uploaded file
// DELETE /api/v1/files/:id
func DeleteFile(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	fileID := c.Param("id")

	commons.Log(ctx, "[api.DeleteFile [user_id=%s [file_id=%s", userID, fileID)

	if err := fileUploadService.DeleteFile(ctx, fileID, userID.String()); err != nil {
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

	commons.RespOKWithMsg(c, "File deleted successfully", nil)
}

// GenerateQuizFromText generates quiz from text content
// POST /api/v1/ai/generate-quiz-from-text
func GenerateQuizFromText(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.GenerateQuizFromText [user_id=%s", userID)

	var req model.GenerateQuizFromTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogE(ctx, "[api.GenerateQuizFromText::BindErr [err=%v", err)
		commons.RespBadReq(c, "invalid request body")
		return
	}

	// Create generation job
	job := &model.GenerationJob{
		ID:              uuid.New().String(),
		UserID:          userID.String(),
		TextContent:     &req.Text,
		JobType:         "generation",
		ReferenceImages: req.ReferenceImages,
		Options:         req.Options,
		Status:          "pending",
		Progress:        0,
		CreatedAt:       time.Now(),
	}

	if err := dao.CreateJob(ctx, job); err != nil {
		commons.RespISE(c, err)
		return
	}

	// Return accepted with job ID
	c.JSON(http.StatusAccepted, commons.APIResponse{
		Success: true,
		Message: "Quiz generation started",
		Data: gin.H{
			"generation_id":          job.ID,
			"status":                 "pending",
			"estimated_time_seconds": 30,
			"reference_images_count": len(req.ReferenceImages),
		},
	})
}

// GenerateQuizFromFile generates quiz from uploaded file
// POST /api/v1/ai/generate-quiz-from-file
func GenerateQuizFromFile(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.GenerateQuizFromFile [user_id=%s", userID)

	var req model.GenerateQuizFromFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogE(ctx, "[api.GenerateQuizFromFile::BindErr [err=%v", err)
		commons.RespBadReq(c, "invalid request body")
		return
	}

	// Verify file exists and belongs to user
	fileUUID, err := uuid.Parse(req.FileID)
	if err != nil {
		commons.RespBadReq(c, "invalid file ID")
		return
	}

	file, err := dao.GetFileByID(ctx, fileUUID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, "file not found")
			return
		}
		commons.RespISE(c, err)
		return
	}

	if file.UserID != userID.String() {
		commons.Resp403(c, "not file owner")
		return
	}

	// Create generation job
	job := &model.GenerationJob{
		ID:              uuid.New().String(),
		UserID:          userID.String(),
		FileID:          &req.FileID,
		JobType:         "generation",
		ReferenceImages: req.ReferenceImages,
		Options:         req.Options,
		Status:          "pending",
		Progress:        0,
		CreatedAt:       time.Now(),
	}

	if err := dao.CreateJob(ctx, job); err != nil {
		commons.RespISE(c, err)
		return
	}

	// Return accepted with job ID
	c.JSON(http.StatusAccepted, commons.APIResponse{
		Success: true,
		Message: "Quiz generation started",
		Data: gin.H{
			"generation_id":          job.ID,
			"status":                 "pending",
			"estimated_time_seconds": 60,
			"reference_images_count": len(req.ReferenceImages),
		},
	})
}

// GetGenerationStatus checks the status of a generation job
// GET /api/v1/ai/generation-status/:id
func GetGenerationStatus(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	jobID := c.Param("id")

	commons.Log(ctx, "[api.GetGenerationStatus [user_id=%s [job_id=%s", userID, jobID)

	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		commons.RespBadReq(c, "invalid job ID")
		return
	}

	job, err := dao.GetJobByID(ctx, jobUUID)
	if err != nil {
		if commons.IsNotFound(err) {
			commons.Resp404(c, "job not found")
			return
		}
		commons.RespISE(c, err)
		return
	}

	// Verify ownership
	if job.UserID != userID.String() {
		commons.Resp403(c, "not job owner")
		return
	}

	response := model.GenerationStatusResponse{
		Status:       job.Status,
		Progress:     job.Progress,
		QuizID:       job.QuizID,
		ErrorMessage: job.ErrorMessage,
	}

	// Add estimated time if still processing
	if job.Status == "pending" || job.Status == "processing" {
		estimatedMs := 30000 // 30 seconds default
		response.EstimatedTimeMs = &estimatedMs
	}

	commons.RespOK(c, response)
}

// ListJobs lists user's generation jobs
// GET /api/v1/ai/jobs?page=1&limit=20
func ListJobs(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.ListJobs [user_id=%s", userID)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get jobs
	jobs, total, err := dao.ListJobsByUserID(ctx, userID, page, limit)
	if err != nil {
		commons.RespISE(c, err)
		return
	}

	response := model.ListJobsResponse{
		Jobs:  jobs,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	commons.RespOK(c, response)
}

// ExtractQuizFromFile extracts existing quiz questions from uploaded file (image/PDF)
// POST /api/v1/ai/extract-quiz-from-file
func ExtractQuizFromFile(c *gin.Context) {
	ctx := commons.Ctx(c)

	userID, ok := commons.UsrID(ctx)
	if !ok {
		commons.Resp401(c, "unauthorized")
		return
	}

	commons.Log(ctx, "[api.ExtractQuizFromFile [user_id=%s", userID)

	// Get file from request
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		commons.LogE(ctx, "[api.ExtractQuizFromFile::FormFileErr [err=%v", err)
		commons.RespBadReq(c, "file is required")
		return
	}
	defer file.Close()

	// Get extract_images parameter (default: true)
	extractImages := c.DefaultPostForm("extract_images", "true") == "true"

	// Validate file type (image or PDF)
	contentType := fileHeader.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"application/pdf": true,
	}

	if !validTypes[contentType] {
		commons.RespBadReq(c, "invalid file type. Only JPG, PNG, and PDF are supported")
		return
	}

	// Validate file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024)
	if fileHeader.Size > maxSize {
		commons.RespBadReq(c, "file size exceeds 10MB limit")
		return
	}

	// Save uploaded file temporarily
	tempFile, err := fileUploadService.SaveFile(ctx, file, fileHeader, userID.String())
	if err != nil {
		if commons.IsBadInput(err) {
			commons.RespBadReq(c, commons.GetCode(err))
			return
		}
		commons.RespISE(c, err)
		return
	}

	// Create extraction job
	job := &model.GenerationJob{
		ID:       uuid.New().String(),
		UserID:   userID.String(),
		FileID:   &tempFile.ID,
		JobType:  "extraction",
		Status:   "pending",
		Progress: 0,
		Options: model.QuizGenOptions{
			NumQuestions: 0, // Will be determined by extraction
			Difficulty:   "medium",
			QuestionType: "mixed",
			Category:     "extracted",
			Language:     "en",
		},
		CreatedAt: time.Now(),
	}

	if err := dao.CreateJob(ctx, job); err != nil {
		commons.RespISE(c, err)
		return
	}

	// Return accepted with job ID
	c.JSON(http.StatusAccepted, commons.APIResponse{
		Success: true,
		Message: "Quiz extraction started",
		Data: gin.H{
			"generation_id":          job.ID,
			"status":                 "pending",
			"extract_images":         extractImages,
			"estimated_time_seconds": 45,
		},
	})
}
