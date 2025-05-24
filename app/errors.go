package app

import (
	"github.com/ayoisaiah/f2/v2/internal/apperr"
)

var (
	errDefaultOptsParsing = &apperr.Error{
		Message: "F2_DEFAULT_OPTS error: unsupported flag '%s'",
	}

	errPipeRead = &apperr.Error{
		Message: "error reading from pipe",
	}
)
