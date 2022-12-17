package file

import "github.com/ayoisaiah/f2/internal/status"

// Change represents a single renaming change.
type Change struct {
	OriginalSource string        `json:"-"`
	Status         status.Status `json:"status"`
	BaseDir        string        `json:"base_dir"`
	Source         string        `json:"source"`
	Target         string        `json:"target"`
	Error          error         `json:"error,omitempty"`
	CSVRow         []string      `json:"-"`
	Index          int           `json:"-"`
	IsDir          bool          `json:"is_dir"`
	WillOverwrite  bool          `json:"will_overwrite"`
}
