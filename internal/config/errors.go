package config

import (
	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/localize"
)

var (
	errInvalidArgument = &apperr.Error{
		Message: localize.T("error.invalid_argument"),
	}

	errParsingFixConflictsPattern = &apperr.Error{
		Message: localize.T("error.parsing_fix_conflicts_pattern"),
	}

	errInvalidSort = &apperr.Error{
		Message: localize.T("error.invalid_sort"),
	}

	errInvalidSortVariable = &apperr.Error{
		Message: localize.T("error.invalid_sort_variable"),
	}

	errInvalidTargetDir = &apperr.Error{
		Message: localize.T("error.invalid_target_dir"),
	}
)
