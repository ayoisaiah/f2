package conflict

// Name refers to a specific conflict.
type Name string

type (
	// Collection represents all conflicts detected during a renaming operation.
	Collection map[Name][]Conflict

	// Conflict represents a single renaming operation conflict.
	Conflict struct {
		Target  string   `json:"target"`
		Cause   string   `json:"cause"`
		Sources []string `json:"sources"`
	}
)

const (
	EmptyFilename             Name = "emptyFilename"
	FileExists                Name = "fileExists"
	OverwritingNewPath        Name = "overwritingNewPath"
	MaxFilenameLengthExceeded Name = "maxFilenameLengthExceeded"
	InvalidCharacters         Name = "invalidCharacters"
	TrailingPeriod            Name = "trailingPeriod"
)
