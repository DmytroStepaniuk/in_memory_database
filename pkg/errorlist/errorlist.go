package errorlist

import (
	"errors"
)

type ErrorList struct {
	Errors []error
}

// New creates a new ErrorList.
func New() *ErrorList {
	return &ErrorList{}
}

// Add adds an error to the list.
func (e *ErrorList) Add(err error) {
	e.Errors = append(e.Errors, err)
}

// Error joins multiple error lists into one.
func (e *ErrorList) Error() error {
	return errors.Join(e.Errors...)
}
