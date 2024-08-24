package status

type Status string

const (
	OK                     Status = "ok"
	Unchanged              Status = "unchanged"
	Overwriting            Status = "overwriting"
	EmptyFilename          Status = "empty filename"
	TrailingPeriod         Status = "trailing periods present"
	PathExists             Status = "target exists"
	OverwritingNewPath     Status = "overwriting new path"
	InvalidCharacters      Status = "invalid characters: (%s)"
	FilenameLengthExceeded Status = "filename too long: (%s)"
	TargetFileChanging     Status = "target file is changing"
)
