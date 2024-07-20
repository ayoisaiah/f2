package rename

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/jsonutil"
	"github.com/ayoisaiah/f2/internal/sortfiles"
	"github.com/ayoisaiah/f2/report"
)

var errUndoFailed = errors.New(
	"reverting the renaming operation failed due to the above errors",
)

var errNothingToUndo = errors.New(
	"nothing to undo",
)

var errBackupFileRemovalFailed = errors.New(
	"unable to remove redundant backup file '%s' after reverting the changes. Please remove it manually",
)

// Undo reverses a renaming operation according to the relevant backup file.
// The undo file is deleted if the operation is successfully reverted.
func Undo(conf *config.Config) error {
	filename := workingDirHash(conf.WorkingDir)

	backupFilePath, err := xdg.SearchDataFile(
		filepath.Join("f2", "backups", filename),
	)
	if err != nil {
		return errNothingToUndo
	}

	fileBytes, err := os.ReadFile(backupFilePath)
	if err != nil {
		return err
	}

	var o jsonutil.Output

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

	// Always sort files before directories when undoing an operation
	sortfiles.ForRenamingAndUndo(changes, conf.Revert)

	err = Rename(conf, changes)
	if err != nil {
		report.NonInteractive(changes)
		return errUndoFailed
	}

	if conf.Exec {
		if err = os.Remove(backupFilePath); err != nil {
			return fmt.Errorf(
				errBackupFileRemovalFailed.Error(),
				pterm.LightYellow(backupFilePath),
			)
		}
	}

	return nil
}
