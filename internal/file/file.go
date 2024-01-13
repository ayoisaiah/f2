package file

import (
	"path/filepath"

	"github.com/ayoisaiah/f2/internal/status"
)

// Change represents a single renaming change.
type Change struct {
	Error          error         `json:"error,omitempty"`
	OriginalSource string        `json:"-"`
	Status         status.Status `json:"status"`
	BaseDir        string        `json:"base_dir"`
	Source         string        `json:"source"`
	Target         string        `json:"target"`
	// RelSourcePath is BaseDir + Source
	RelSourcePath string `json:"-"`
	// RelTargetPath is BaseDir + Target
	RelTargetPath string   `json:"-"`
	CSVRow        []string `json:"-"`
	Index         int      `json:"-"`
	IsDir         bool     `json:"is_dir"`
	WillOverwrite bool     `json:"will_overwrite"`
}

func (c *Change) SourcePath() string {
	return filepath.Join(c.BaseDir, c.Source)
}

func (c *Change) TargetPath() string {
	return filepath.Join(c.BaseDir, c.Target)
}
