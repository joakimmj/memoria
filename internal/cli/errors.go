package cli

import (
	"errors"
	"fmt"
)

// Error codes of the CLI itself.
const (
	codeUsage    = "usage"     // bad command line
	codeNotFound = "not_found" // e.g. unknown help topic
)

// coder is an error with a code for the envelope, see mia(1).
type coder interface {
	ErrorCode() string
}

// Error is a failure with a code a caller can act on.
type Error struct {
	Code string
	Err  error
}

// Error returns the message of the wrapped error.
func (e *Error) Error() string { return e.Err.Error() }

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error { return e.Err }

// ErrorCode returns the error's code.
func (e *Error) ErrorCode() string { return e.Code }

// codeOf returns the first code in err's chain, or "".
func codeOf(err error) string {
	var c coder
	if errors.As(err, &c) {
		return c.ErrorCode()
	}
	return ""
}

// usageErrorf returns a codeUsage error.
func usageErrorf(format string, args ...any) error {
	return &Error{Code: codeUsage, Err: fmt.Errorf(format, args...)}
}

// notFoundErrorf returns a codeNotFound error.
func notFoundErrorf(format string, args ...any) error {
	return &Error{Code: codeNotFound, Err: fmt.Errorf(format, args...)}
}
