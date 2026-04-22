package core

import (
	"errors"
	"fmt"
)

// NotFoundError represents a resource that could not be found.
type NotFoundError struct {
	Resource string
	ID       string
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(resource, id string) error {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// IsNotFound checks if an error is a NotFoundError.
func IsNotFound(err error) bool {
	return IsNotFoundError(err)
}

// IsNotFoundError checks if an error is or wraps a NotFoundError.
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	var notFoundErr *NotFoundError
	return errors.As(err, &notFoundErr)
}
