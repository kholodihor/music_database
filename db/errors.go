package db

import "fmt"

// Custom error types
type DbError struct {
	Code    int
	Message string
	Err     error
}

func (e *DbError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

const (
	ErrCodeNotFound     = 404
	ErrCodeInvalidInput = 400
	ErrCodeInternal     = 500
)
