package jsonutil

import (
	"encoding/json"
	"time"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
)

// Output represents the structure of the output produced by the
// `--json` flag. It is also used for backup files.
type Output struct {
	WorkingDir string         `json:"working_dir"`
	Date       string         `json:"date"`
	Changes    []*file.Change `json:"changes"`
	DryRun     bool           `json:"dry_run"`
}

func GetOutput(
	changes []*file.Change,
) ([]byte, error) {
	conf := config.Get()

	out := Output{
		WorkingDir: conf.WorkingDir,
		Date:       conf.Date.Format(time.RFC3339),
		DryRun:     !conf.Exec,
		Changes:    changes,
	}

	// prevent empty matches from being encoded as `null`
	if out.Changes == nil {
		out.Changes = make([]*file.Change, 0)
	}

	b, err := json.MarshalIndent(out, "", "    ")
	if err != nil {
		return b, err
	}

	return b, nil
}
