package errors

import (
	stderrors "errors"
	"fmt"
)

// StatusError carries an HTTP status code and a message while allowing error wrapping.
type StatusError struct {
	Status  int
	Message string
	Err     error
}

func (e *StatusError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *StatusError) Unwrap() error {
	return e.Err
}

// NewStatusError constructs a StatusError without an underlying cause.
func NewStatusError(status int, message string) *StatusError {
	return &StatusError{Status: status, Message: message}
}

// WrapStatus attaches an HTTP status and message to an underlying error.
func WrapStatus(err error, status int, message string) *StatusError {
	return &StatusError{Status: status, Message: message, Err: err}
}

// AsStatusError tries to extract a StatusError from the error chain.
func AsStatusError(err error) (*StatusError, bool) {
	var se *StatusError
	if stderrors.As(err, &se) {
		return se, true
	}
	return nil, false
}
