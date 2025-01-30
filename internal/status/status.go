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
	ForbiddenCharacters    Status = "forbidden characters present"
	FilenameLengthExceeded Status = "filename too long"
	SourceAlreadyRenamed   Status = "source already renamed"
	SourceNotFound         Status = "source not found"
	Ignored                Status = "ignored"
)
