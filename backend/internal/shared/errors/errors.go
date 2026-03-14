package errors

import "fmt"

// DomainError is a typed application error that carries a machine-readable code,
// a human-readable message, and an optional cause.
type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *DomainError) Unwrap() error { return e.Cause }

// New creates a new DomainError.
func New(code, message string, cause error) *DomainError {
	return &DomainError{Code: code, Message: message, Cause: cause}
}

// Sentinel domain errors
var (
	ErrUserNotFound       = New("USER_NOT_FOUND", "user not found", nil)
	ErrEmailTaken         = New("EMAIL_TAKEN", "email already registered", nil)
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "invalid email or password", nil)
	ErrTokenExpired       = New("TOKEN_EXPIRED", "token has expired", nil)
	ErrTokenInvalid       = New("TOKEN_INVALID", "token is invalid", nil)
	ErrUnauthorized       = New("UNAUTHORIZED", "authentication required", nil)
	ErrForbidden          = New("FORBIDDEN", "access denied", nil)
	ErrSettingsNotFound   = New("SETTINGS_NOT_FOUND", "settings not found", nil)
	ErrInternal           = New("INTERNAL_ERROR", "an internal error occurred", nil)
)

// Is implements errors.Is compatibility for sentinel errors (by code).
func Is(err error, target *DomainError) bool {
	if de, ok := err.(*DomainError); ok {
		return de.Code == target.Code
	}
	return false
}
