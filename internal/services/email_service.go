package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"

	"github.com/affanahmad024/quiz-craft/backend/internal/config"
	"github.com/affanahmad024/quiz-craft/backend/pkg/logger"
	"github.com/google/uuid"
	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
)

// EmailService handles email operations
type EmailService struct {
	config    *config.EmailConfig
	mjClient  *mailjet.Client
}

// NewEmailService creates a new email service using Mailjet API
func NewEmailService(cfg *config.EmailConfig) *EmailService {
	// Initialize Mailjet client with API credentials
	mjClient := mailjet.NewMailjetClient(cfg.MailjetAPIKey, cfg.MailjetSecretKey)

	logger.GetLogger().Infof("Email service initialized with Mailjet API")

	return &EmailService{
		config:    cfg,
		mjClient:  mjClient,
	}
}

// SendVerificationEmail sends email verification link
func (s *EmailService) SendVerificationEmail(ctx context.Context, toEmail, toName, token string, userID uuid.UUID) error {
	subject := "Verify Your Email - QuizCraft"

	// URL-encode the token to handle special characters safely
	encodedToken := url.QueryEscape(token)

	// Create verification link using configured frontend URL
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s&user_id=%s", s.config.FrontendURL, encodedToken, userID.String())

	// Email body
	body := s.renderVerificationEmail(toName, verificationLink)

	return s.sendEmail(toEmail, toName, subject, body)
}

// SendPasswordResetEmail sends password reset link
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, toEmail, toName, token string, userID uuid.UUID) error {
	subject := "Reset Your Password - QuizCraft"

	// URL-encode the token to handle special characters safely
	encodedToken := url.QueryEscape(token)

	// Create reset link using configured frontend URL
	resetLink := fmt.Sprintf("%s/reset-password?token=%s&user_id=%s", s.config.FrontendURL, encodedToken, userID.String())

	// Email body
	body := s.renderPasswordResetEmail(toName, resetLink)

	return s.sendEmail(toEmail, toName, subject, body)
}

// SendWelcomeEmail sends welcome email after successful verification
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail, toName string) error {
	subject := "Welcome to QuizCraft!"

	body := s.renderWelcomeEmail(toName)

	return s.sendEmail(toEmail, toName, subject, body)
}

// SendPasswordChangedEmail sends notification when password is changed
func (s *EmailService) SendPasswordChangedEmail(ctx context.Context, toEmail, toName string) error {
	subject := "Your Password Has Been Changed - QuizCraft"

	body := s.renderPasswordChangedEmail(toName)

	return s.sendEmail(toEmail, toName, subject, body)
}

// sendEmail sends an email using Mailjet API (HTTPS - not blocked by cloud providers)
func (s *EmailService) sendEmail(toEmail, toName, subject, body string) error {
	logger.GetLogger().Infof("Attempting to send email via Mailjet API: from=%s to=%s subject=%s", s.config.FromEmail, toEmail, subject)

	// Build Mailjet message
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.config.FromEmail,
				Name:  s.config.FromName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: toEmail,
					Name:  toName,
				},
			},
			Subject:  subject,
			HTMLPart: body,
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}

	// Send the email
	res, err := s.mjClient.SendMailV31(&messages)
	if err != nil {
		logger.GetLogger().Errorf("Failed to send email via Mailjet API to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Check response status
	if res.ResultsV31 != nil && len(res.ResultsV31) > 0 {
		result := res.ResultsV31[0]
		if result.Status == "success" {
			logger.GetLogger().Infof("✅ Email sent successfully via Mailjet API (from=%s to=%s, MessageID=%d)", s.config.FromEmail, toEmail, result.To[0].MessageID)
			return nil
		} else {
			logger.GetLogger().Errorf("Mailjet API returned non-success status: %s", result.Status)
			return fmt.Errorf("mailjet returned status: %s", result.Status)
		}
	}

	logger.GetLogger().Warnf("⚠️  Email sent but no result data returned from Mailjet")
	return nil
}

// renderVerificationEmail renders the verification email HTML
func (s *EmailService) renderVerificationEmail(name, link string) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9fafb; padding: 30px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>QuizCraft</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}},</h2>
            <p>Thank you for signing up with QuizCraft! To complete your registration, please verify your email address by clicking the button below:</p>
            <div style="text-align: center;">
                <a href="{{.Link}}" class="button">Verify Email</a>
            </div>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4F46E5;">{{.Link}}</p>
            <p>This link will expire in 24 hours.</p>
            <p>If you didn't create an account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 QuizCraft. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	return s.renderTemplate(tmpl, map[string]string{"Name": name, "Link": link})
}

// renderPasswordResetEmail renders the password reset email HTML
func (s *EmailService) renderPasswordResetEmail(name, link string) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9fafb; padding: 30px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>QuizCraft</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}},</h2>
            <p>We received a request to reset your password. Click the button below to create a new password:</p>
            <div style="text-align: center;">
                <a href="{{.Link}}" class="button">Reset Password</a>
            </div>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4F46E5;">{{.Link}}</p>
            <p>This link will expire in 1 hour.</p>
            <p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 QuizCraft. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	return s.renderTemplate(tmpl, map[string]string{"Name": name, "Link": link})
}

// renderWelcomeEmail renders the welcome email HTML
func (s *EmailService) renderWelcomeEmail(name string) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9fafb; padding: 30px; }
        .button { display: inline-block; padding: 14px 32px; background-color: #10B981; color: white; text-decoration: none; border-radius: 8px; margin: 20px 0; font-weight: bold; font-size: 16px; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to QuizCraft!</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}},</h2>
            <p>Your email has been verified successfully! Welcome to QuizCraft.</p>
            <p>You can now:</p>
            <ul>
                <li>Create custom quizzes</li>
                <li>Upload documents and extract questions with AI</li>
                <li>Take quizzes and track your progress</li>
                <li>Share quizzes with others</li>
            </ul>
            <div style="text-align: center;">
                <a href="{{.DashboardURL}}" style="display: inline-block; padding: 14px 32px; background-color: #10B981; color: #ffffff; text-decoration: none; border-radius: 8px; margin: 20px 0; font-weight: bold; font-size: 16px;">Go to Dashboard</a>
            </div>
            <p>If you have any questions, feel free to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 QuizCraft. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	dashboardURL := fmt.Sprintf("%s/dashboard", s.config.FrontendURL)
	return s.renderTemplate(tmpl, map[string]string{"Name": name, "DashboardURL": dashboardURL})
}

// renderPasswordChangedEmail renders the password changed notification email
func (s *EmailService) renderPasswordChangedEmail(name string) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9fafb; padding: 30px; }
        .footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>QuizCraft</h1>
        </div>
        <div class="content">
            <h2>Hi {{.Name}},</h2>
            <p>This is a confirmation that your password was successfully changed.</p>
            <p>If you didn't make this change, please contact our support team immediately.</p>
        </div>
        <div class="footer">
            <p>&copy; 2025 QuizCraft. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	return s.renderTemplate(tmpl, map[string]string{"Name": name})
}

// renderTemplate renders an email template
func (s *EmailService) renderTemplate(tmplStr string, data map[string]string) string {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		logger.GetLogger().Errorf("Failed to parse email template: %v", err)
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.GetLogger().Errorf("Failed to execute email template: %v", err)
		return ""
	}

	return buf.String()
}
