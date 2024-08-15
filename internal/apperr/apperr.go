package apperr

import "fmt"

type Error struct {
	Cause   error // The underlying error if any
	Message string
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

// Unwrap is used to make it work with errors.Is, errors.As.
func (e *Error) Unwrap() error {
	// Return the inner error.
	return e.Cause
}

// Wrap associates the underlying error.
func (e *Error) Wrap(err error) *Error {
	e.Cause = err
	return e
}

// Fmt calls fmt.Sprintf on the error message.
func (e *Error) Fmt(str ...any) *Error {
	e.Message = fmt.Sprintf(e.Message, str...)
	return e
}
