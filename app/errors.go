package app

import (
	"github.com/ayoisaiah/f2/v2/internal/apperr"
)

var (
	errDefaultOptsParsing = &apperr.Error{
		Message: "error parsing default options",
	}

	errSetDefaultOpt = &apperr.Error{
		Message: "failed to apply the default value '%s' for the option '%s'",
	}

	errPipeRead = &apperr.Error{
		Message: "error reading from pipe",
	}
)
