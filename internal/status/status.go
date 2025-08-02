package status

import (
	"encoding/json"

	"github.com/ayoisaiah/f2/v2/internal/localize"
)

type Status struct {
	ID      string
	Message string
}

func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.ID + `"`), nil
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var id string

	if err := json.Unmarshal(data, &id); err != nil {
		return err
	}

	s.ID = id
	s.Message = localize.T("status." + id)

	return nil
}

func (s Status) String() string {
	return s.Message
}

func newStatus(id string) Status {
	return Status{
		ID:      id,
		Message: localize.T("status." + id),
	}
}

var (
	OK                     = newStatus("ok")
	Unchanged              = newStatus("unchanged")
	Overwriting            = newStatus("overwriting")
	EmptyFilename          = newStatus("empty_filename")
	TrailingPeriod         = newStatus("trailing_periods_present")
	PathExists             = newStatus("target_exists")
	OverwritingNewPath     = newStatus("overwriting_new_path")
	ForbiddenCharacters    = newStatus("forbidden_characters_present")
	FilenameLengthExceeded = newStatus("filename_too_long")
	SourceAlreadyRenamed   = newStatus("source_already_renamed")
	SourceNotFound         = newStatus("source_not_found")
	Ignored                = newStatus("ignored")
)
