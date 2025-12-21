package utils

// ValidationError represents a validation error with a custom message
type ValidationError struct {
	message string
}

// NewValidationError creates a new validation error with the given message
func NewValidationError(message string) ValidationError {
	return ValidationError{message: message}
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return e.message
}
