package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/affanahmad024/quiz-craft/backend/ai-agent/model"
	"github.com/affanahmad024/quiz-craft/backend/commons"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	DefaultModel   = "gemini-1.5-pro"
	DefaultTimeout = 60 * time.Second
	MaxRetries     = 1
)

// AIService handles Google Gemini AI operations
type AIService struct {
	APIKey string
	Model  string
}

// NewAIService creates a new AI service
func NewAIService(apiKey string, model string) *AIService {
	if model == "" {
		model = DefaultModel
	}
	return &AIService{
		APIKey: apiKey,
		Model:  model,
	}
}

// GenerateQuizFromText generates a quiz from text content using Gemini AI
func (s *AIService) GenerateQuizFromText(ctx context.Context, content string, options model.QuizGenOptions, referenceImagePaths []string) (*model.AIQuizResponse, error) {
	commons.Log(ctx, "[core.GenerateQuizFromText [content_length=%d [num_questions=%d [difficulty=%s [ref_images=%d",
		len(content), options.NumQuestions, options.Difficulty, len(referenceImagePaths))

	// Build prompt
	prompt := s.buildPromptWithImages(content, options, len(referenceImagePaths) > 0)

	// Call Gemini AI with retry (with or without images)
	var response string
	var err error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			commons.Log(ctx, "[core.GenerateQuizFromText::Retry [attempt=%d", attempt)
			time.Sleep(2 * time.Second)
		}

		if len(referenceImagePaths) > 0 {
			// Use Vision API with images
			response, err = s.callGeminiVisionAPIWithMultipleImages(ctx, referenceImagePaths, prompt)
		} else {
			// Use text-only API
			response, err = s.callGeminiAPI(ctx, prompt)
		}

		if err == nil {
			break
		}

		if attempt == MaxRetries {
			commons.LogE(ctx, "[core.GenerateQuizFromText::MaxRetries [err=%v", err)
			return nil, err
		}
	}

	// Parse JSON response
	quizResponse, err := s.parseAIResponse(ctx, response)
	if err != nil {
		return nil, err
	}

	commons.Log(ctx, "[core.GenerateQuizFromText::OK [questions_generated=%d", len(quizResponse.Questions))
	return quizResponse, nil
}

// callGeminiAPI makes the actual API call to Gemini
func (s *AIService) callGeminiAPI(ctx context.Context, prompt string) (string, error) {
	commons.Log(ctx, "[core.callGeminiAPI [model=%s", s.Model)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.APIKey))
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiAPI::ClientErr [err=%v", err)
		return "", commons.Internal("failed to initialize AI client")
	}
	defer client.Close()

	// Get model
	geminiModel := client.GenerativeModel(s.Model)

	// Set generation config
	geminiModel.GenerationConfig.Temperature = genai.Ptr(float32(0.7))
	geminiModel.GenerationConfig.TopP = genai.Ptr(float32(0.95))
	geminiModel.GenerationConfig.TopK = genai.Ptr(int32(40))
	geminiModel.GenerationConfig.MaxOutputTokens = genai.Ptr(int32(8000))

	// Generate content
	resp, err := geminiModel.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiAPI::GenerateErr [err=%v", err)
		return "", commons.Internal("AI generation failed")
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		commons.LogE(ctx, "[core.callGeminiAPI::EmptyResponse")
		return "", commons.Internal("AI returned empty response")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		result.WriteString(fmt.Sprintf("%v", part))
	}

	responseText := result.String()
	commons.Log(ctx, "[core.callGeminiAPI::OK [response_length=%d", len(responseText))
	return responseText, nil
}

// callGeminiVisionAPIWithMultipleImages makes API call with multiple reference images
func (s *AIService) callGeminiVisionAPIWithMultipleImages(ctx context.Context, imagePaths []string, prompt string) (string, error) {
	commons.Log(ctx, "[core.callGeminiVisionAPIWithMultipleImages [model=%s [images=%d", s.Model, len(imagePaths))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.APIKey))
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiVisionAPIWithMultipleImages::ClientErr [err=%v", err)
		return "", commons.Internal("failed to initialize AI client")
	}
	defer client.Close()

	// Get model
	geminiModel := client.GenerativeModel(s.Model)

	// Set generation config
	geminiModel.GenerationConfig.Temperature = genai.Ptr(float32(0.7))
	geminiModel.GenerationConfig.TopP = genai.Ptr(float32(0.95))
	geminiModel.GenerationConfig.TopK = genai.Ptr(int32(40))
	geminiModel.GenerationConfig.MaxOutputTokens = genai.Ptr(int32(8000))

	// Build parts array with text prompt and images
	var parts []genai.Part
	parts = append(parts, genai.Text(prompt))

	// Add all reference images
	for i, imagePath := range imagePaths {
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			commons.LogE(ctx, "[core.callGeminiVisionAPIWithMultipleImages::ReadFileErr [image=%d [err=%v", i, err)
			continue // Skip this image
		}

		// Determine image format
		imageFormat := "image/jpeg"
		if strings.HasSuffix(strings.ToLower(imagePath), ".png") {
			imageFormat = "image/png"
		}

		parts = append(parts, genai.ImageData(imageFormat, imageData))
		commons.Log(ctx, "[core.callGeminiVisionAPIWithMultipleImages [image=%d loaded [size=%d", i, len(imageData))
	}

	// Generate content
	resp, err := geminiModel.GenerateContent(ctx, parts...)
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiVisionAPIWithMultipleImages::GenerateErr [err=%v", err)
		return "", commons.Internal("AI generation with images failed")
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		commons.LogE(ctx, "[core.callGeminiVisionAPIWithMultipleImages::EmptyResponse")
		return "", commons.Internal("AI returned empty response")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		result.WriteString(fmt.Sprintf("%v", part))
	}

	responseText := result.String()
	commons.Log(ctx, "[core.callGeminiVisionAPIWithMultipleImages::OK [response_length=%d", len(responseText))
	return responseText, nil
}

// buildPromptWithImages creates the prompt for Gemini API with optional image support
func (s *AIService) buildPromptWithImages(content string, options model.QuizGenOptions, hasImages bool) string {
	imageInstructions := ""
	if hasImages {
		imageInstructions = `
- Reference images are provided to help generate contextual questions
- If a question relates to content in the reference images, include the base64 image data in "image_data" field
- You may use the images to create visual questions (e.g., "What is shown in this image?", "Identify the component in the diagram")
- Not all questions need images - only include images where relevant`
	}

	prompt := fmt.Sprintf(`You are an expert quiz generator. Generate %d %s quiz questions based on this content.

CONTENT:
%s%s

REQUIREMENTS:
- Generate exactly %d questions
- Difficulty: %s
- Question type: %s
- Category: %s
- For multiple choice: provide 4 options
- Mark the correct answer with "is_correct": true
- Include brief explanations for correct answers
- Vary question difficulty across the quiz
- Make questions clear and unambiguous

RESPONSE FORMAT (strict JSON, no markdown code blocks):
{
  "title": "Generated quiz title (50 chars max)",
  "description": "Brief description (200 chars max)",
  "category": "%s",
  "difficulty": "%s",
  "questions": [
    {
      "question_text": "Question here?",
      "question_type": "multiple_choice",
      "points": 1,
      "explanation": "Explanation here",
      "image_data": "base64_encoded_image_if_relevant",
      "answers": [
        {"answer_text": "Option A", "is_correct": false},
        {"answer_text": "Option B", "is_correct": true},
        {"answer_text": "Option C", "is_correct": false},
        {"answer_text": "Option D", "is_correct": false}
      ]
    }
  ]
}

IMPORTANT:
- Return ONLY valid JSON, no markdown code blocks or additional text
- Ensure exactly %d questions are generated
- Each multiple choice question must have exactly 4 answers
- Exactly ONE answer must be marked as correct per question`,
		options.NumQuestions,
		options.Difficulty,
		content,
		imageInstructions,
		options.NumQuestions,
		options.Difficulty,
		options.QuestionType,
		options.Category,
		options.Category,
		options.Difficulty,
		options.NumQuestions,
	)

	return prompt
}

// buildPrompt creates the prompt for Gemini API
func (s *AIService) buildPrompt(content string, options model.QuizGenOptions) string {
	prompt := fmt.Sprintf(`You are an expert quiz generator. Generate %d %s quiz questions based on this content.

CONTENT:
%s

REQUIREMENTS:
- Generate exactly %d questions
- Difficulty: %s
- Question type: %s
- Category: %s
- For multiple choice: provide 4 options
- Mark the correct answer with "is_correct": true
- Include brief explanations for correct answers
- Vary question difficulty across the quiz
- Make questions clear and unambiguous

RESPONSE FORMAT (strict JSON, no markdown code blocks):
{
  "title": "Generated quiz title (50 chars max)",
  "description": "Brief description (200 chars max)",
  "category": "%s",
  "difficulty": "%s",
  "questions": [
    {
      "question_text": "Question here?",
      "question_type": "multiple_choice",
      "points": 1,
      "explanation": "Explanation here",
      "answers": [
        {"answer_text": "Option A", "is_correct": false},
        {"answer_text": "Option B", "is_correct": true},
        {"answer_text": "Option C", "is_correct": false},
        {"answer_text": "Option D", "is_correct": false}
      ]
    }
  ]
}

IMPORTANT:
- Return ONLY valid JSON, no markdown code blocks or additional text
- Ensure exactly %d questions are generated
- Each multiple choice question must have exactly 4 answers
- Exactly ONE answer must be marked as correct per question`,
		options.NumQuestions,
		options.Difficulty,
		content,
		options.NumQuestions,
		options.Difficulty,
		options.QuestionType,
		options.Category,
		options.Category,
		options.Difficulty,
		options.NumQuestions,
	)

	return prompt
}

// parseAIResponse parses the JSON response from AI
func (s *AIService) parseAIResponse(ctx context.Context, response string) (*model.AIQuizResponse, error) {
	commons.Log(ctx, "[core.parseAIResponse [response_length=%d", len(response))

	// Clean up response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Parse JSON
	var quizResponse model.AIQuizResponse
	if err := json.Unmarshal([]byte(response), &quizResponse); err != nil {
		commons.LogE(ctx, "[core.parseAIResponse::JSONErr [err=%v [response_preview=%s",
			err, response[:min(len(response), 500)])
		return nil, commons.BadInput("ai-response-malformed")
	}

	// Validate response structure
	if err := s.validateQuizResponse(ctx, &quizResponse); err != nil {
		return nil, err
	}

	commons.Log(ctx, "[core.parseAIResponse::OK")
	return &quizResponse, nil
}

// validateQuizResponse validates the quiz structure
func (s *AIService) validateQuizResponse(ctx context.Context, quiz *model.AIQuizResponse) error {
	if quiz.Title == "" {
		return commons.BadInput("quiz-title-missing")
	}

	if len(quiz.Questions) == 0 {
		return commons.BadInput("no-questions-generated")
	}

	if len(quiz.Questions) > 50 {
		return commons.BadInput("too-many-questions")
	}

	// Validate each question
	for i, question := range quiz.Questions {
		if question.QuestionText == "" {
			return commons.BadInput(fmt.Sprintf("question-%d-text-missing", i+1))
		}

		if len(question.QuestionText) > 1000 {
			return commons.BadInput(fmt.Sprintf("question-%d-text-too-long", i+1))
		}

		if len(question.Answers) < 2 {
			return commons.BadInput(fmt.Sprintf("question-%d-insufficient-answers", i+1))
		}

		// Check for at least one correct answer
		hasCorrect := false
		for _, answer := range question.Answers {
			if answer.IsCorrect {
				hasCorrect = true
				break
			}
		}

		if !hasCorrect {
			return commons.BadInput(fmt.Sprintf("question-%d-no-correct-answer", i+1))
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExtractQuizFromImage extracts quiz structure from an image or PDF using Gemini Vision
func (s *AIService) ExtractQuizFromImage(ctx context.Context, imagePath string, extractImages bool) (*model.AIQuizResponse, error) {
	commons.Log(ctx, "[core.ExtractQuizFromImage [image_path=%s [extract_images=%v", imagePath, extractImages)

	// Build extraction prompt
	prompt := s.buildExtractionPrompt(extractImages)

	// Call Gemini Vision API with retry
	var response string
	var err error

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		if attempt > 0 {
			commons.Log(ctx, "[core.ExtractQuizFromImage::Retry [attempt=%d", attempt)
			time.Sleep(2 * time.Second)
		}

		response, err = s.callGeminiVisionAPI(ctx, imagePath, prompt)
		if err == nil {
			break
		}

		if attempt == MaxRetries {
			commons.LogE(ctx, "[core.ExtractQuizFromImage::MaxRetries [err=%v", err)
			return nil, err
		}
	}

	// Parse JSON response
	quizResponse, err := s.parseAIResponse(ctx, response)
	if err != nil {
		return nil, err
	}

	commons.Log(ctx, "[core.ExtractQuizFromImage::OK [questions_extracted=%d", len(quizResponse.Questions))
	return quizResponse, nil
}

// callGeminiVisionAPI makes API call to Gemini with image input
func (s *AIService) callGeminiVisionAPI(ctx context.Context, imagePath string, prompt string) (string, error) {
	commons.Log(ctx, "[core.callGeminiVisionAPI [model=%s [image_path=%s", s.Model, imagePath)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	// Initialize Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.APIKey))
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiVisionAPI::ClientErr [err=%v", err)
		return "", commons.Internal("failed to initialize AI client")
	}
	defer client.Close()

	// Get model
	geminiModel := client.GenerativeModel(s.Model)

	// Set generation config
	geminiModel.GenerationConfig.Temperature = genai.Ptr(float32(0.5))
	geminiModel.GenerationConfig.TopP = genai.Ptr(float32(0.95))
	geminiModel.GenerationConfig.TopK = genai.Ptr(int32(40))
	geminiModel.GenerationConfig.MaxOutputTokens = genai.Ptr(int32(8000))

	// Read image file
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiVisionAPI::ReadFileErr [err=%v", err)
		return "", commons.Internal("failed to read image file")
	}

	// Determine image format
	imageFormat := "image/jpeg"
	if strings.HasSuffix(strings.ToLower(imagePath), ".png") {
		imageFormat = "image/png"
	} else if strings.HasSuffix(strings.ToLower(imagePath), ".pdf") {
		imageFormat = "application/pdf"
	}

	// Generate content with image and text
	resp, err := geminiModel.GenerateContent(ctx,
		genai.ImageData(imageFormat, imageData),
		genai.Text(prompt),
	)
	if err != nil {
		commons.LogE(ctx, "[core.callGeminiVisionAPI::GenerateErr [err=%v", err)
		return "", commons.Internal("AI extraction failed")
	}

	// Extract text from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		commons.LogE(ctx, "[core.callGeminiVisionAPI::EmptyResponse")
		return "", commons.Internal("AI returned empty response")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		result.WriteString(fmt.Sprintf("%v", part))
	}

	responseText := result.String()
	commons.Log(ctx, "[core.callGeminiVisionAPI::OK [response_length=%d", len(responseText))
	return responseText, nil
}

// buildExtractionPrompt creates the prompt for quiz extraction
func (s *AIService) buildExtractionPrompt(extractImages bool) string {
	imageInstructions := ""
	if extractImages {
		imageInstructions = `
- If a question contains an image, include the base64 encoded image data in "image_data" field
- Extract all visual content (diagrams, charts, photos) associated with questions`
	}

	prompt := fmt.Sprintf(`You are an expert quiz extraction system. Extract the quiz structure from this image/PDF.

TASK:
- Identify the quiz title and description (if present)
- Extract all questions with their text
- Extract all answer options for each question
- Identify the correct answer(s) for each question
- Detect question types (multiple_choice, true_false, short_answer)
- Extract any explanations provided%s

RESPONSE FORMAT (strict JSON, no markdown code blocks):
{
  "title": "Extracted quiz title (or generate appropriate title)",
  "description": "Brief description (or empty string)",
  "category": "general",
  "difficulty": "medium",
  "questions": [
    {
      "question_text": "Question text here?",
      "question_type": "multiple_choice",
      "points": 1,
      "explanation": "Explanation if provided (or empty string)",
      "image_data": "base64_encoded_image_data_if_present",
      "answers": [
        {"answer_text": "Option A", "is_correct": false},
        {"answer_text": "Option B", "is_correct": true},
        {"answer_text": "Option C", "is_correct": false},
        {"answer_text": "Option D", "is_correct": false}
      ]
    }
  ]
}

IMPORTANT:
- Return ONLY valid JSON, no markdown code blocks or additional text
- If no correct answer is marked, make your best inference
- For true/false questions, ensure exactly 2 answers (True and False)
- For multiple choice, extract all visible options
- Maintain original question numbering and order
- If question has associated image, include it in image_data as base64`, imageInstructions)

	return prompt
}
