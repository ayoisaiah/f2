package apperr

import (
	"errors"
	"fmt"
)

type Error struct {
	Cause   error
	Context any
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

// Wrap returns a copy of the error with the cause set.
func (e *Error) Wrap(err error) *Error {
	newE := *e
	newE.Cause = err

	return &newE
}

// Fmt returns a copy of the error with the message formatted.
func (e *Error) Fmt(str ...any) *Error {
	newE := *e
	newE.Message = fmt.Sprintf(e.Message, str...)

	return &newE
}

// WithCtx returns a copy of the error with context attached.
func (e *Error) WithCtx(ctx any) *Error {
	newE := *e
	newE.Context = ctx

	return &newE
}

func Unwrap(err error) error {
	ae := &Error{}

	ok := errors.As(err, &ae)
	if !ok {
		return err
	}

	return ae.Cause
}
