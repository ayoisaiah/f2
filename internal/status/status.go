package status

type Status string

const (
	OK                     Status = "ok"
	Unchanged              Status = "unchanged"
	Overwriting            Status = "overwriting"
	EmptyFilename          Status = "empty filename"
	TrailingPeriod         Status = "trailing periods are prohibited"
	PathExists             Status = "path already exists"
	OverwritingNewPath     Status = "overwriting newly renamed path"
	InvalidCharacters      Status = "invalid characters present: (%s)"
	FilenameLengthExceeded Status = "max file name length exceeded: (%s)"
	TargetFileChanging     Status = "the target file is changing"
)
