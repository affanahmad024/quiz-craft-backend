package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// Email regex pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// Phone number regex (basic international format)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	if len(email) > 255 {
		return errors.New("email must be less than 255 characters")
	}

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidateName validates name
func ValidateName(name string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if len(name) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	if len(name) > 255 {
		return errors.New("name must be less than 255 characters")
	}

	// Remove extra spaces
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("name cannot be empty")
	}

	return nil
}

// ValidatePhoneNumber validates phone number (optional field)
func ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}

	if len(phone) > 20 {
		return errors.New("phone number must be less than 20 characters")
	}

	// Basic validation - can be enhanced based on requirements
	if !phoneRegex.MatchString(phone) {
		return errors.New("invalid phone number format")
	}

	return nil
}

// SanitizeString removes leading/trailing spaces and normalizes whitespace
func SanitizeString(s string) string {
	// Trim leading and trailing spaces
	s = strings.TrimSpace(s)

	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	return s
}

// NormalizeEmail normalizes email (lowercase, trimmed)
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
