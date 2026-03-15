package app

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation error")
)

// AppError is a structured application-level error.
// Kind determines HTTP status mapping. UserMessage is the client-facing text.
// Cause carries full context for logs/debugging.
type AppError struct {
	Kind        error
	UserMessage string
	Cause       error
}

func (e *AppError) Error() string        { return e.Cause.Error() }
func (e *AppError) Is(target error) bool { return errors.Is(e.Kind, target) }
func (e *AppError) Unwrap() error        { return e.Cause }

func Validation(op string, err error) *AppError {
	return &AppError{Kind: ErrValidation, UserMessage: err.Error(), Cause: fmt.Errorf("%s: %w", op, err)}
}

func NotFound(op string, err error) *AppError {
	return &AppError{Kind: ErrNotFound, UserMessage: err.Error(), Cause: fmt.Errorf("%s: %w", op, err)}
}
