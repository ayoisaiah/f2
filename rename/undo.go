package rename

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/config"
	internaljson "github.com/ayoisaiah/f2/internal/json"
	internalpath "github.com/ayoisaiah/f2/internal/path"
	internalsort "github.com/ayoisaiah/f2/internal/sort"
	"github.com/ayoisaiah/f2/internal/utils"
	"github.com/ayoisaiah/f2/report"
)

var errUndoFailed = errors.New(
	"The undo operation failed due to the above errors",
)

// Undo reverses a renaming operation according to the relevant backup file.
// The undo file is deleted if the operation is successfully reverted.
func Undo() error {
	conf := config.Get()

	dir := strings.ReplaceAll(conf.WorkingDir(), internalpath.Separator, "_")
	if runtime.GOOS == utils.Windows {
		dir = strings.ReplaceAll(dir, ":", "_")
	}

	file := dir + ".json"

	backupFilePath, err := xdg.SearchDataFile(
		filepath.Join("f2", "backups", file),
	)
	if err != nil {
		return err
	}

	fileBytes, err := os.ReadFile(backupFilePath)
	if err != nil {
		return err
	}

	var o internaljson.Output

	err = json.Unmarshal(fileBytes, &o)
	if err != nil {
		return err
	}

	changes := o.Changes

	for i := range changes {
		ch := changes[i]

		target := ch.Target
		source := ch.Source

		ch.Source = target
		ch.Target = source

		changes[i] = ch
	}

	// Sort only in dry-run mode
	if !conf.ShouldExec() {
		internalsort.Alphabetically(changes)
	}

	errs := commit(changes)
	if len(errs) > 0 {
		report.Changes(changes, errs)
		return errUndoFailed
	}

	if conf.ShouldExec() {
		if err = os.Remove(backupFilePath); err != nil {
			pterm.Fprintln(conf.Stderr(),
				pterm.Warning.Sprintf(
					"Unable to remove redundant backup file '%s' after successful undo operation.",
					pterm.LightYellow(backupFilePath),
				),
			)
		}
	}

	return nil
}
