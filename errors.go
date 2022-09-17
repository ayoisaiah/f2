package f2

import "errors"

var (
	errInvalidArgument = errors.New(
		"Invalid argument: one of `-f`, `-r`, `-csv` or `-u` must be present and set to a non empty string value. Use 'f2 --help' for more information",
	)

	errInvalidSimpleModeArgs = errors.New(
		"At least one argument must be specified in simple mode",
	)

	errConflictDetected = errors.New(
		"Resolve conflicts before proceeding or use the -F flag to auto fix all conflicts",
	)

	errCSVReadFailed = errors.New("Unable to read CSV file")

	errBackupNotFound = errors.New(
		"Unable to find the backup file for the current directory",
	)

	errInvalidSubmatches = errors.New("Invalid number of submatches")
)
