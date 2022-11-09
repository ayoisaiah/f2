package rename

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pterm/pterm"

	internaljson "github.com/ayoisaiah/f2/internal/json"
	internalos "github.com/ayoisaiah/f2/internal/os"
	internalpath "github.com/ayoisaiah/f2/internal/path"
	internalsort "github.com/ayoisaiah/f2/internal/sort"
	"github.com/ayoisaiah/f2/report"
)

var errUndoFailed = errors.New(
	"The undo operation failed due to the above errors",
)

var errBackupFileRemovalFailed = errors.New(
	"Unable to remove redundant backup file '%s' after successful undo operation. Please remove it manually",
)

// Undo reverses a renaming operation according to the relevant backup file.
// The undo file is deleted if the operation is successfully reverted.
func Undo(
	exec, includeDir, quiet, revert, verbose bool,
	jsonOpts *internaljson.OutputOpts,
) error {
	dir := strings.ReplaceAll(jsonOpts.WorkingDir, internalpath.Separator, "_")
	if runtime.GOOS == internalos.Windows {
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

	internalsort.FilesBeforeDirs(changes, revert)

	if !exec {
		report.Dry(changes, includeDir, quiet, revert, jsonOpts)

		return nil
	}

	errs := commit(changes, revert, verbose, jsonOpts)
	if len(errs) > 0 {
		report.Changes(changes, errs, quiet, jsonOpts)
		return errUndoFailed
	}

	if exec {
		if err = os.Remove(backupFilePath); err != nil {
			return fmt.Errorf(
				errBackupFileRemovalFailed.Error(),
				pterm.LightYellow(backupFilePath),
			)
		}
	}

	return nil
}
