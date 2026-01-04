package commons

import "errors"

// Domain error types following Titan pattern
var (
	ErrNotFound     = errors.New("resource not found")
	ErrBadInput     = errors.New("invalid input")
	ErrConflict     = errors.New("resource conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInternal     = errors.New("internal error")
)

// DomainError wraps errors with additional context
type DomainError struct {
	Err  error
	Code string
	Msg  string
}

func (e *DomainError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Err.Error()
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Error constructors
func NotFound(code string) *DomainError {
	return &DomainError{Err: ErrNotFound, Code: code}
}

func BadInput(code string) *DomainError {
	return &DomainError{Err: ErrBadInput, Code: code}
}

func Conflict(code string) *DomainError {
	return &DomainError{Err: ErrConflict, Code: code}
}

func Unauthorized(code string) *DomainError {
	return &DomainError{Err: ErrUnauthorized, Code: code}
}

func Forbidden(code string) *DomainError {
	return &DomainError{Err: ErrForbidden, Code: code}
}

func Internal(msg string) *DomainError {
	return &DomainError{Err: ErrInternal, Msg: msg}
}

// Error checkers
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsBadInput(err error) bool {
	return errors.Is(err, ErrBadInput)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// GetCode extracts error code from DomainError
func GetCode(err error) string {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr.Code
	}
	return ""
}
