package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

const (
	MinContentLength = 100
	MaxContentLength = 50000
)

// FileParserService handles file parsing operations
type FileParserService struct{}

// NewFileParserService creates a new file parser service
func NewFileParserService() *FileParserService {
	return &FileParserService{}
}

// ParseFile parses a file and extracts text content
func (s *FileParserService) ParseFile(ctx context.Context, filePath string) (string, error) {
	commons.Log(ctx, "[services.ParseFile [file_path=%s", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", commons.NotFound("file-not-found")
	}

	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(filePath))

	var content string
	var err error

	switch ext {
	case ".pdf":
		content, err = s.ParsePDF(ctx, filePath)
	case ".docx":
		content, err = s.ParseDOCX(ctx, filePath)
	case ".txt", ".md":
		content, err = s.ParseText(ctx, filePath)
	default:
		return "", commons.BadInput("unsupported-file-type")
	}

	if err != nil {
		return "", err
	}

	// Validate content length
	if len(content) < MinContentLength {
		return "", commons.BadInput("content-too-short")
	}

	// Truncate if too long
	if len(content) > MaxContentLength {
		commons.Log(ctx, "[services.ParseFile::Warning [content_truncated [original_length=%d", len(content))
		content = content[:MaxContentLength] + "\n\n[Content truncated to 50,000 characters]"
	}

	commons.Log(ctx, "[services.ParseFile::OK [content_length=%d", len(content))
	return content, nil
}

// ParsePDF extracts text from PDF files
func (s *FileParserService) ParsePDF(ctx context.Context, filePath string) (string, error) {
	commons.Log(ctx, "[services.ParsePDF [file_path=%s", filePath)

	f, r, err := pdf.Open(filePath)
	if err != nil {
		commons.LogE(ctx, "[services.ParsePDF::OpenErr [err=%v", err)
		return "", commons.BadInput("unable-to-parse-pdf")
	}
	defer f.Close()

	var content strings.Builder
	totalPages := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPages; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			commons.LogE(ctx, "[services.ParsePDF::PageErr [page=%d [err=%v", pageIndex, err)
			continue
		}

		content.WriteString(text)
		content.WriteString("\n\n")
	}

	result := strings.TrimSpace(content.String())

	if result == "" {
		return "", commons.BadInput("pdf-contains-no-text")
	}

	commons.Log(ctx, "[services.ParsePDF::OK [pages=%d [content_length=%d", totalPages, len(result))
	return result, nil
}

// ParseDOCX extracts text from DOCX files
func (s *FileParserService) ParseDOCX(ctx context.Context, filePath string) (string, error) {
	commons.Log(ctx, "[services.ParseDOCX [file_path=%s", filePath)

	r, err := docx.ReadDocxFile(filePath)
	if err != nil {
		commons.LogE(ctx, "[services.ParseDOCX::OpenErr [err=%v", err)
		return "", commons.BadInput("unable-to-parse-docx")
	}
	defer r.Close()

	docx1 := r.Editable()
	content := docx1.GetContent()

	// Clean up content
	content = strings.TrimSpace(content)

	if content == "" {
		return "", commons.BadInput("docx-contains-no-text")
	}

	commons.Log(ctx, "[services.ParseDOCX::OK [content_length=%d", len(content))
	return content, nil
}

// ParseText reads text from TXT or MD files
func (s *FileParserService) ParseText(ctx context.Context, filePath string) (string, error) {
	commons.Log(ctx, "[services.ParseText [file_path=%s", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		commons.LogE(ctx, "[services.ParseText::OpenErr [err=%v", err)
		return "", commons.Internal("failed-to-read-file")
	}
	defer file.Close()

	// Read file contents
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		commons.LogE(ctx, "[services.ParseText::ReadErr [err=%v", err)
		return "", commons.Internal("failed-to-read-file")
	}

	content := string(contentBytes)
	content = strings.TrimSpace(content)

	if content == "" {
		return "", commons.BadInput("file-is-empty")
	}

	commons.Log(ctx, "[services.ParseText::OK [content_length=%d", len(content))
	return content, nil
}

// ValidateContent checks if content meets requirements
func (s *FileParserService) ValidateContent(content string) error {
	length := len(content)

	if length < MinContentLength {
		return commons.BadInput(fmt.Sprintf("content-too-short-min-%d-chars", MinContentLength))
	}

	if length > MaxContentLength {
		return commons.BadInput(fmt.Sprintf("content-too-long-max-%d-chars", MaxContentLength))
	}

	return nil
}
