package app

import (
	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/localize"
)

var (
	errDefaultOptsParsing = &apperr.Error{
		Message: localize.T("error.default_opts_parsing"),
	}

	errPipeRead = &apperr.Error{
		Message: localize.T("error.pipe_read"),
	}
)
