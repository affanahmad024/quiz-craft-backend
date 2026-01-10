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
	userID := c.GetString("user_id")

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
	fileUpload, err := fileUploadService.SaveFile(ctx, file, fileHeader, userID)
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
	userID := c.GetString("user_id")

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
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		commons.RespBadReq(c, "invalid user ID")
		return
	}

	files, total, err := dao.ListFilesByUserID(ctx, userUUID, page, limit)
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
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	commons.Log(ctx, "[api.DeleteFile [user_id=%s [file_id=%s", userID, fileID)

	if err := fileUploadService.DeleteFile(ctx, fileID, userID); err != nil {
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
	userID := c.GetString("user_id")

	commons.Log(ctx, "[api.GenerateQuizFromText [user_id=%s", userID)

	var req model.GenerateQuizFromTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commons.LogE(ctx, "[api.GenerateQuizFromText::BindErr [err=%v", err)
		commons.RespBadReq(c, "invalid request body")
		return
	}

	// Create generation job
	job := &model.GenerationJob{
		ID:          uuid.New().String(),
		UserID:      userID,
		TextContent: &req.Text,
		Options:     req.Options,
		Status:      "pending",
		Progress:    0,
		CreatedAt:   time.Now(),
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
			"generation_id":         job.ID,
			"status":                "pending",
			"estimated_time_seconds": 30,
		},
	})
}

// GenerateQuizFromFile generates quiz from uploaded file
// POST /api/v1/ai/generate-quiz-from-file
func GenerateQuizFromFile(c *gin.Context) {
	ctx := commons.Ctx(c)
	userID := c.GetString("user_id")

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

	if file.UserID != userID {
		commons.Resp403(c, "not file owner")
		return
	}

	// Create generation job
	job := &model.GenerationJob{
		ID:        uuid.New().String(),
		UserID:    userID,
		FileID:    &req.FileID,
		Options:   req.Options,
		Status:    "pending",
		Progress:  0,
		CreatedAt: time.Now(),
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
		},
	})
}

// GetGenerationStatus checks the status of a generation job
// GET /api/v1/ai/generation-status/:id
func GetGenerationStatus(c *gin.Context) {
	ctx := commons.Ctx(c)
	userID := c.GetString("user_id")
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
	if job.UserID != userID {
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
	userID := c.GetString("user_id")

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
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		commons.RespBadReq(c, "invalid user ID")
		return
	}

	jobs, total, err := dao.ListJobsByUserID(ctx, userUUID, page, limit)
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
