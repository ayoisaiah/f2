// Package rename handles the actual file renaming operations and manages
// backups for potential undo operations.
package rename

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/ayoisaiah/f2/internal/apperr"
	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/osutil"
	"github.com/ayoisaiah/f2/internal/status"
	"github.com/ayoisaiah/f2/report"
)

var errRenameFailed = &apperr.Error{
	Message: "some files could not be renamed",
}

// commit iterates over all the matches and renames them on the filesystem.
// Directories are auto-created if necessary, and errors are aggregated.
func commit(fileChanges file.Changes) []int {
	var errIndices []int

	for i := range fileChanges {
		change := fileChanges[i]

		if change.Status == status.Ignored {
			continue
		}

		targetPath := change.TargetPath

		// skip paths that are unchanged in every aspect
		if change.SourcePath == targetPath {
			continue
		}

		// Workaround for case insensitive filesystems where renaming a filename to
		// its upper or lowercase equivalent doesn't work. Fixing this involves the
		// following steps:
		// 1. Prefix and suffix <target> with __<time>__
		// 2. Rename <source> to <target>
		// 3. Rename __<time>__<target>__<time>__ to <target>
		var isCaseChangeOnly bool // only the target case is changing
		if strings.EqualFold(change.SourcePath, targetPath) {
			isCaseChangeOnly = true
			timeStr := fmt.Sprintf("%d", time.Now().UnixNano())
			targetPath = filepath.Join(
				change.BaseDir,
				"__"+timeStr+"__"+change.Target+"__"+timeStr+"__", // step 1
			)
		}

		// If target contains a slash, create all missing
		// directories before renaming the file
		if strings.Contains(change.Target, "/") ||
			strings.Contains(change.Target, `\`) &&
				runtime.GOOS == osutil.Windows {
			// No need to check if the `dir` exists or if there are several
			// consecutive slashes since `os.MkdirAll` handles that
			dir := filepath.Dir(change.Target)

			err := os.MkdirAll(
				filepath.Join(change.BaseDir, dir),
				osutil.DirPermission,
			)
			if err != nil {
				errIndices = append(errIndices, i)
				change.Error = err

				continue
			}
		}

		err := os.Rename(change.SourcePath, targetPath) // step 2
		// if the intermediate rename is successful,
		// proceed with the original renaming operation
		if err == nil && isCaseChangeOnly {
			err = os.Rename(targetPath, change.TargetPath) // step 3
		}

		if err != nil {
			errIndices = append(errIndices, i)
			change.Error = err
		}
	}

	return errIndices
}

// Rename renames files according to the provided changes and configuration
// handling conflicts and backups.
func Rename(
	fileChanges file.Changes,
) error {
	renameErrs := commit(fileChanges)
	if len(renameErrs) > 0 {
		return errRenameFailed.WithCtx(renameErrs)
	}

	return nil
}

// PostRename handles actions after a renaming operation, such as printing
// results and creating a backup file if applicable.
func PostRename(conf *config.Config, fileChanges file.Changes, err error) {
	report.PrintResults(conf, fileChanges, err)

	if len(fileChanges) != 0 && !conf.Revert {
		err := backupChanges(
			fileChanges,
			conf.BackupFilename,
			conf.BackupLocation,
		)
		if err != nil {
			report.BackupFailed(err)
		}
	}

	if conf.Revert {
		backupFilePath, err := xdg.SearchDataFile(
			filepath.Join("f2", "backups", conf.BackupFilename),
		)
		if err != nil {
			report.BackupFileRemovalFailed(err)
			return
		}

		if err = os.Remove(backupFilePath); err != nil {
			report.BackupFileRemovalFailed(err)
			return
		}
	}
}
