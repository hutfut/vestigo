package domain

import "fmt"

// ErrNotFound indicates a requested entity does not exist.
type ErrNotFound struct {
	Entity string
	ID     string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s %s not found", e.Entity, e.ID)
}

// ErrConflict indicates a request conflicts with current state.
type ErrConflict struct {
	Message string
}

func (e *ErrConflict) Error() string {
	return e.Message
}

// ErrValidation indicates invalid input.
type ErrValidation struct {
	Field   string
	Message string
}

func (e *ErrValidation) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}
