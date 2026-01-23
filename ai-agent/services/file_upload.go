package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/dao"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/google/uuid"
)

const (
	MaxFileSize       = 52428800 // 50MB in bytes
	UploadStoragePath = "./uploads"
)

var (
	AllowedMimeTypes = map[string]bool{
		// Document types
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"text/plain":     true,
		"text/markdown":  true,
		"application/msword": true, // .doc files

		// Image types (for question images)
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	AllowedExtensions = map[string]bool{
		// Document extensions
		".pdf":  true,
		".docx": true,
		".doc":  true,
		".txt":  true,
		".md":   true,

		// Image extensions
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
)

// FileUploadService handles file upload operations
type FileUploadService struct{}

// NewFileUploadService creates a new file upload service
func NewFileUploadService() *FileUploadService {
	return &FileUploadService{}
}

// ValidateFile validates file type and size
func (s *FileUploadService) ValidateFile(fileHeader *multipart.FileHeader) error {
	// Check file size
	if fileHeader.Size > MaxFileSize {
		return commons.BadInput("file-too-large")
	}

	if fileHeader.Size == 0 {
		return commons.BadInput("file-empty")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !AllowedExtensions[ext] {
		return commons.BadInput("file-type-not-allowed")
	}

	// Validate filename
	if err := s.ValidateFilename(fileHeader.Filename); err != nil {
		return err
	}

	return nil
}

// ValidateFilename checks if filename is safe
func (s *FileUploadService) ValidateFilename(filename string) error {
	if filename == "" {
		return commons.BadInput("filename-empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return commons.BadInput("filename-invalid")
	}

	// Check length
	if len(filename) > 255 {
		return commons.BadInput("filename-too-long")
	}

	return nil
}

// SanitizeFilename removes special characters from filename
func (s *FileUploadService) SanitizeFilename(filename string) string {
	// Remove path components
	filename = filepath.Base(filename)

	// Get extension
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Remove special characters, keep only alphanumeric, dash, underscore
	reg := regexp.MustCompile("[^a-zA-Z0-9-_]")
	nameWithoutExt = reg.ReplaceAllString(nameWithoutExt, "_")

	// Trim underscores
	nameWithoutExt = strings.Trim(nameWithoutExt, "_")

	// If empty after sanitization, use default
	if nameWithoutExt == "" {
		nameWithoutExt = "file"
	}

	return nameWithoutExt + ext
}

// GenerateUniqueFilename creates a unique filename with timestamp
func (s *FileUploadService) GenerateUniqueFilename(originalFilename string) string {
	sanitized := s.SanitizeFilename(originalFilename)
	ext := filepath.Ext(sanitized)
	nameWithoutExt := strings.TrimSuffix(sanitized, ext)

	timestamp := time.Now().Format("20060102_150405")
	randomID := uuid.New().String()[:8]

	return fmt.Sprintf("%s_%s_%s%s", nameWithoutExt, timestamp, randomID, ext)
}

// GetUserUploadDir returns the upload directory for a specific user
func (s *FileUploadService) GetUserUploadDir(userID string) string {
	return filepath.Join(UploadStoragePath, userID)
}

// EnsureUploadDir creates the upload directory if it doesn't exist
func (s *FileUploadService) EnsureUploadDir(userID string) error {
	uploadDir := s.GetUserUploadDir(userID)
	return os.MkdirAll(uploadDir, 0755)
}

// SaveFile saves the uploaded file to disk
func (s *FileUploadService) SaveFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, userID string) (*model.FileUpload, error) {
	commons.Log(ctx, "[services.SaveFile [user_id=%s [filename=%s [size=%d", userID, fileHeader.Filename, fileHeader.Size)

	// Validate file
	if err := s.ValidateFile(fileHeader); err != nil {
		commons.LogE(ctx, "[services.SaveFile::ValidationErr [err=%v", err)
		return nil, err
	}

	// Ensure upload directory exists
	if err := s.EnsureUploadDir(userID); err != nil {
		commons.LogE(ctx, "[services.SaveFile::DirErr [err=%v", err)
		return nil, commons.Internal("failed to create upload directory")
	}

	// Generate unique filename
	uniqueFilename := s.GenerateUniqueFilename(fileHeader.Filename)
	filePath := filepath.Join(s.GetUserUploadDir(userID), uniqueFilename)

	// Create file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		commons.LogE(ctx, "[services.SaveFile::CreateErr [err=%v", err)
		return nil, commons.Internal("failed to create file")
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		commons.LogE(ctx, "[services.SaveFile::CopyErr [err=%v", err)
		// Clean up partially written file
		os.Remove(filePath)
		return nil, commons.Internal("failed to save file")
	}

	// Calculate expiration date (7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create file record
	fileUpload := &model.FileUpload{
		ID:               uuid.New().String(),
		UserID:           userID,
		Filename:         uniqueFilename,
		OriginalFilename: fileHeader.Filename,
		FilePath:         filePath,
		FileSize:         fileHeader.Size,
		MimeType:         fileHeader.Header.Get("Content-Type"),
		Status:           "uploaded",
		CreatedAt:        time.Now(),
		ExpiresAt:        &expiresAt,
	}

	// Save to database
	if err := dao.CreateFile(ctx, fileUpload); err != nil {
		commons.LogE(ctx, "[services.SaveFile::DBErr [err=%v", err)
		// Clean up file on disk
		os.Remove(filePath)
		return nil, err
	}

	commons.Log(ctx, "[services.SaveFile::OK [file_id=%s", fileUpload.ID)
	return fileUpload, nil
}

// DeleteFile deletes a file from disk and database
func (s *FileUploadService) DeleteFile(ctx context.Context, fileID string, userID string) error {
	commons.Log(ctx, "[services.DeleteFile [file_id=%s [user_id=%s", fileID, userID)

	// Get file record
	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		return commons.BadInput("invalid-file-id")
	}

	fileUpload, err := dao.GetFileByID(ctx, fileUUID)
	if err != nil {
		return err
	}

	// Verify ownership
	if fileUpload.UserID != userID {
		return commons.Forbidden("not-file-owner")
	}

	// Delete file from disk
	if err := os.Remove(fileUpload.FilePath); err != nil {
		commons.LogE(ctx, "[services.DeleteFile::RemoveErr [err=%v", err)
		// Continue even if file doesn't exist on disk
	}

	// Delete from database
	if err := dao.DeleteFile(ctx, fileUUID); err != nil {
		return err
	}

	commons.Log(ctx, "[services.DeleteFile::OK [file_id=%s", fileID)
	return nil
}

// CleanupExpiredFiles removes expired files
func (s *FileUploadService) CleanupExpiredFiles(ctx context.Context) (int, error) {
	commons.Log(ctx, "[services.CleanupExpiredFiles")

	files, err := dao.GetExpiredFiles(ctx)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, file := range files {
		// Delete file from disk
		if err := os.Remove(file.FilePath); err != nil {
			commons.LogE(ctx, "[services.CleanupExpiredFiles::RemoveErr [file_id=%s [err=%v", file.ID, err)
			continue
		}

		// Delete from database
		fileUUID, _ := uuid.Parse(file.ID)
		if err := dao.DeleteFile(ctx, fileUUID); err != nil {
			commons.LogE(ctx, "[services.CleanupExpiredFiles::DBErr [file_id=%s [err=%v", file.ID, err)
			continue
		}

		deleted++
	}

	commons.Log(ctx, "[services.CleanupExpiredFiles::OK [deleted=%d", deleted)
	return deleted, nil
}
