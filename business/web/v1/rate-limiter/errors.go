package ratelimiter

import (
	"errors"
	"fmt"
)

// authError is used to pass an error during the request through the
// application with auth specific context.
type rateLimitError struct {
	msg string
}

// NewAuthError creates an AuthError for the provided message.
func NewRateLimitError(format string, args ...any) error {
	return &rateLimitError{
		msg: fmt.Sprintf(format, args...),
	}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (ae *rateLimitError) Error() string {
	return ae.msg
}

// IsAuthError checks if an error of type AuthError exists.
func IsRateLimitError(err error) bool {
	var ae *rateLimitError
	return errors.As(err, &ae)
}
