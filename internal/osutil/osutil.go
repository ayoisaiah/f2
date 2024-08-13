package osutil

import (
	"regexp"
)

var (
	// PartialWindowsForbiddenCharRegex is used to match the strings that contain forbidden
	// characters in Windows' file names. This does not include also forbidden
	// forward and back slash characters because their presence will cause a new
	// directory to be created.
	PartialWindowsForbiddenCharRegex = regexp.MustCompile(`<|>|:|"|\||\?|\*`)
	// CompleteWindowsForbiddenCharRegex is like windowsForbiddenRegex but includes
	// forward and backslashes.
	CompleteWindowsForbiddenCharRegex = regexp.MustCompile(
		`<|>|:|"|\||\?|\*|/|\\`,
	)
	// MacForbiddenCharRegex is used to match the strings that contain forbidden
	// characters in macOS' file names.
	MacForbiddenCharRegex = regexp.MustCompile(`:`)
)

const (
	Windows = "windows"
	Darwin  = "darwin"
)

type exitCode int

const (
	ExitOK    exitCode = 0
	ExitError exitCode = 1
)
