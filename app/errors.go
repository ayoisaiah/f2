package app

import (
	"fmt"
)

var (
	errDefaultOptsParsing = &AppErr{
		Message: "error parsing default options",
	}

	errSetDefaultOpt = &AppErr{
		Message: "failed to apply the default value '%s' for the option '%s'",
	}

	errPipeRead = &AppErr{
		Message: "error reading from pipe",
	}
)

type AppErr struct {
	Cause   error // The underlying error if any
	Message string
}

func (e *AppErr) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

// Unwrap is used to make it work with errors.Is, errors.As.
func (e *AppErr) Unwrap() error {
	// Return the inner error.
	return e.Cause
}

// Wrap associates the underlying error.
func (e *AppErr) Wrap(err error) *AppErr {
	e.Cause = err
	return e
}

// Fmt calls fmt.Sprintf on the error message.
func (e *AppErr) Fmt(str ...any) *AppErr {
	e.Message = fmt.Sprintf(e.Message, str...)
	return e
}
