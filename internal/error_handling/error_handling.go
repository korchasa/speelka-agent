// Package error_handling provides utilities for error handling in the application.
// Responsibility: Unified handling and categorization of errors throughout the application
// Features: Supports error categorization, retry with exponential backoff, and sensitive information sanitization
package error_handling

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
)

// ErrorCategory represents an error category.
// Responsibility: Classification of errors by type to determine their handling strategy
// Features: Used as an enumeration to indicate the error type
type ErrorCategory int

const (
	// ErrorCategoryUnknown represents an unknown error category.
	ErrorCategoryUnknown ErrorCategory = iota
	// ErrorCategoryValidation represents a validation error.
	ErrorCategoryValidation
	// ErrorCategoryTransient represents a temporary error that can be successfully handled with a retry.
	ErrorCategoryTransient
	// ErrorCategoryExternal represents an external service error.
	ErrorCategoryExternal
	// ErrorCategoryInternal represents an internal error.
	ErrorCategoryInternal
)

// AppError represents an application error with a category and optional cause.
// Responsibility: Encapsulation of error information, including its category and root cause
// Features: Implements the standard error interface and provides additional methods for working with category and cause
type AppError struct {
	message  string
	category ErrorCategory
	cause    error
}

// Error returns the error message.
// Responsibility: Implementation of the standard error interface
// Features: Includes the cause error message if it exists
func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s", e.message, e.cause.Error())
	}
	return e.message
}

// Category returns the error category.
// Responsibility: Providing access to the error category
// Features: Simple getter for the category field
func (e *AppError) Category() ErrorCategory {
	return e.category
}

// Cause returns the cause of the error, if any.
// Responsibility: Providing access to the root cause of the error
// Features: Simple getter for the cause field
func (e *AppError) Cause() error {
	return e.cause
}

// Unwrap returns the cause of the error for compatibility with errors.Is and errors.As.
// Responsibility: Supporting the standard Go functionality for working with nested errors
// Features: Complies with the unwrap specification in the Go standard library
func (e *AppError) Unwrap() error {
	return e.cause
}

// NewError creates a new AppError with the given message and category.
// Responsibility: Factory method for creating a new error without a cause
// Features: Initializes all fields of AppError, leaving cause as nil
func NewError(message string, category ErrorCategory) *AppError {
	return &AppError{
		message:  message,
		category: category,
		cause:    nil,
	}
}

// WrapError wraps an existing error with additional context and category.
// Responsibility: Factory method for creating a new error containing another error as the cause
// Features: Preserves the original error in the cause field
func WrapError(err error, message string, category ErrorCategory) *AppError {
	return &AppError{
		message:  message,
		category: category,
		cause:    err,
	}
}

// IsTransient checks if an error is transient (can be retried).
// Responsibility: Determining the possibility of a retry for a given error
// Features: Uses type assertion to check if the error is of type AppError, and if so, checks its category
func IsTransient(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.category == ErrorCategoryTransient
	}
	return false
}

// RetryConfig defines the configuration for retry with exponential backoff.
// Responsibility: Storing parameters for the retry strategy
// Features: Contains the maximum number of attempts, initial delay, multiplier, and maximum delay
type RetryConfig struct {
	// MaxRetries - maximum number of retry attempts.
	MaxRetries int
	// InitialBackoff - initial delay.
	InitialBackoff time.Duration
	// BackoffMultiplier - multiplier for delay after each retry attempt.
	BackoffMultiplier float64
	// MaxBackoff - maximum delay.
	MaxBackoff time.Duration
}

// RetryWithBackoff retries a function with exponential backoff.
// Responsibility: Implementation of a retry strategy for handling transient errors
// Features: Increases the wait time between attempts exponentially, not exceeding the maximum delay
func RetryWithBackoff(ctx context.Context, fn func() error, config RetryConfig) error {
	var err error
	backoff := config.InitialBackoff

	// Initial attempt
	if err = fn(); err == nil {
		return nil
	}

	// Don't retry if the error is not transient
	if !IsTransient(err) {
		return err
	}

	// Retry with delay
	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			if err = fn(); err == nil {
				return nil
			}

			// Don't retry if the error is not transient
			if !IsTransient(err) {
				return err
			}

			// Increase delay for the next attempt, but don't exceed the maximum delay
			backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}
	}

	// If all retry attempts are exhausted, return the last error
	return WrapError(err, fmt.Sprintf("failed after %d retries", config.MaxRetries), ErrorCategoryTransient)
}

// SanitizeError sanitizes the error message of sensitive information.
// Responsibility: Preventing leakage of sensitive data in error messages
// Features: Uses regular expressions to find and replace potentially sensitive information
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	// Define patterns for sensitive information
	patterns := []*regexp.Regexp{
		// API keys (e.g., sk-1234567890abcdef)
		regexp.MustCompile(`(sk-[a-zA-Z0-9]{16,})`),
		// OAuth tokens
		regexp.MustCompile(`(Bearer\s+[a-zA-Z0-9_\-\.]+)`),
		// Basic auth
		regexp.MustCompile(`(Basic\s+[a-zA-Z0-9_\-\.]+)`),
		// Passwords
		regexp.MustCompile(`(password|passwd|pwd)[:=]\s*([^\s,;]+)`),
		// Credit card numbers
		regexp.MustCompile(`(\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4})`),
	}

	// Sanitize the error message
	message := err.Error()
	for _, pattern := range patterns {
		message = pattern.ReplaceAllString(message, "[REDACTED]")
	}

	// If the message hasn't changed, return the original error
	if message == err.Error() {
		return err
	}

	// Create a new error with the sanitized message
	var appErr *AppError
	if errors.As(err, &appErr) {
		return NewError(message, appErr.category)
	}
	return errors.New(message)
}
