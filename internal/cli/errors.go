package cli

import (
	"errors"
	"fmt"

	"github.com/joakimmj/memoria/internal/brain"
	"github.com/joakimmj/memoria/internal/engram"
)

const (
	codeUsage     = "usage"     // bad command line
	codeNotFound  = "not_found" // e.g. unknown help topic
	codeCollision = "collision"
	codeBrain     = "brain" // a command needing one brain got none
	codeEmptySlug = "empty_slug"
	codePrefix    = "prefix"
	codeTemplate  = "template"
)

var sentinelCodes = []struct {
	err  error
	code string
}{
	{brain.ErrCollision, codeCollision},
	{brain.ErrNotFound, codeNotFound},
	{brain.ErrInvalidName, codeUsage},
	{engram.ErrCollision, codeCollision},
	{engram.ErrEmptySlug, codeEmptySlug},
	{engram.ErrPrefix, codePrefix},
	{engram.ErrTemplate, codeTemplate},
	{engram.ErrNoTitle, codeUsage},
}

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
	for _, sc := range sentinelCodes {
		if errors.Is(err, sc.err) {
			return sc.code
		}
	}
	return ""
}

func usageErrorf(format string, args ...any) error {
	return &Error{Code: codeUsage, Err: fmt.Errorf(format, args...)}
}

func notFoundErrorf(format string, args ...any) error {
	return &Error{Code: codeNotFound, Err: fmt.Errorf(format, args...)}
}
