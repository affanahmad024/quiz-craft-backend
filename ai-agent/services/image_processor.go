package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/dao"
	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/google/uuid"
)

const (
	QuestionImagesPath = "./uploads/questions"
	ImageQuality       = 75 // JPEG quality (1-100, lower = more compression)
	MaxImageWidth      = 1200
	MaxImageHeight     = 1200
)

// ImageProcessorService handles image processing operations
type ImageProcessorService struct{}

// NewImageProcessorService creates a new image processor service
func NewImageProcessorService() *ImageProcessorService {
	return &ImageProcessorService{}
}

// ProcessAndSaveImage processes base64 image data, compresses it, and saves to disk
func (s *ImageProcessorService) ProcessAndSaveImage(ctx context.Context, base64Data string, userID string, questionID string) (*model.FileUpload, error) {
	commons.Log(ctx, "[services.ProcessAndSaveImage [user_id=%s [question_id=%s", userID, questionID)

	// Decode base64 data
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		commons.LogE(ctx, "[services.ProcessAndSaveImage::DecodeErr [err=%v", err)
		return nil, commons.BadInput("invalid-base64-image")
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		commons.LogE(ctx, "[services.ProcessAndSaveImage::ImageDecodeErr [err=%v", err)
		return nil, commons.BadInput("invalid-image-format")
	}

	commons.Log(ctx, "[services.ProcessAndSaveImage [format=%s [original_size=%d", format, len(imageData))

	// Resize if needed
	img = s.ResizeImage(img, MaxImageWidth, MaxImageHeight)

	// Compress image to JPEG
	var compressedBuf bytes.Buffer
	if err := jpeg.Encode(&compressedBuf, img, &jpeg.Options{Quality: ImageQuality}); err != nil {
		commons.LogE(ctx, "[services.ProcessAndSaveImage::CompressErr [err=%v", err)
		return nil, commons.Internal("failed-to-compress-image")
	}

	compressedData := compressedBuf.Bytes()
	compressionRatio := float64(len(compressedData)) / float64(len(imageData)) * 100
	commons.Log(ctx, "[services.ProcessAndSaveImage [compressed_size=%d [ratio=%.2f%%", len(compressedData), compressionRatio)

	// Ensure upload directory exists
	uploadDir := filepath.Join(QuestionImagesPath, userID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		commons.LogE(ctx, "[services.ProcessAndSaveImage::DirErr [err=%v", err)
		return nil, commons.Internal("failed-to-create-upload-directory")
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102_150405")
	randomID := uuid.New().String()[:8]
	filename := fmt.Sprintf("q_%s_%s_%s.jpg", questionID[:8], timestamp, randomID)
	filePath := filepath.Join(uploadDir, filename)

	// Save to disk
	if err := os.WriteFile(filePath, compressedData, 0644); err != nil {
		commons.LogE(ctx, "[services.ProcessAndSaveImage::WriteErr [err=%v", err)
		return nil, commons.Internal("failed-to-save-image")
	}

	// Create file record
	fileUpload := &model.FileUpload{
		ID:               uuid.New().String(),
		UserID:           userID,
		Filename:         filename,
		OriginalFilename: fmt.Sprintf("question_%s.jpg", questionID[:8]),
		FilePath:         filePath,
		FileSize:         int64(len(compressedData)),
		MimeType:         "image/jpeg",
		Status:           "uploaded",
		CreatedAt:        time.Now(),
		ExpiresAt:        nil, // No expiration for question images
	}

	// Save to database
	if err := dao.CreateFile(ctx, fileUpload); err != nil {
		commons.LogE(ctx, "[services.ProcessAndSaveImage::DBErr [err=%v", err)
		// Clean up file on disk
		os.Remove(filePath)
		return nil, err
	}

	commons.Log(ctx, "[services.ProcessAndSaveImage::OK [file_id=%s [size=%d", fileUpload.ID, fileUpload.FileSize)
	return fileUpload, nil
}

// ResizeImage resizes an image to fit within maxWidth and maxHeight while maintaining aspect ratio
func (s *ImageProcessorService) ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if resizing is needed
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// Calculate aspect ratio
	aspectRatio := float64(width) / float64(height)
	var newWidth, newHeight int

	if float64(maxWidth)/float64(maxHeight) > aspectRatio {
		// Height is the limiting factor
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * aspectRatio)
	} else {
		// Width is the limiting factor
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / aspectRatio)
	}

	// Create new image with calculated dimensions
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor resampling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := x * width / newWidth
			srcY := y * height / newHeight
			resized.Set(x, y, img.At(srcX, srcY))
		}
	}

	return resized
}

// ValidateImageData validates base64 image data
func (s *ImageProcessorService) ValidateImageData(base64Data string) error {
	if base64Data == "" {
		return commons.BadInput("image-data-empty")
	}

	// Check size (max 5MB base64)
	if len(base64Data) > 5*1024*1024 {
		return commons.BadInput("image-data-too-large")
	}

	// Try to decode
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return commons.BadInput("invalid-base64-encoding")
	}

	// Try to decode as image
	_, _, err = image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return commons.BadInput("invalid-image-format")
	}

	return nil
}
