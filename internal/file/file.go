package file

import (
	"log/slog"

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
	Index         int      `json:"-"` // TODO: Rename to position?
	IsDir         bool     `json:"is_dir"`
	WillOverwrite bool     `json:"will_overwrite"`
}

func (c Change) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("error", c.Error),
		slog.String("original_source", c.OriginalSource),
		slog.Any("status", c.Status),
		slog.String("base_dir", c.BaseDir),
		slog.String("source", c.Source),
		slog.String("target", c.Target),
		slog.String("rel_source_path", c.RelSourcePath),
		slog.String("rel_target_path", c.RelTargetPath),
		slog.Any("csv_row", c.CSVRow),
		slog.Int("index", c.Index),
		slog.Bool("is_dir", c.IsDir),
		slog.Bool("will_overwrite", c.WillOverwrite),
	)
}
