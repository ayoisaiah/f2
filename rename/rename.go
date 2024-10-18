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

// traversedDirs records the directories that were traversed during a renaming
// operation.
var traversedDirs = make(map[string]string)

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
				change.TargetDir,
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
				filepath.Join(change.TargetDir, dir),
				osutil.DirPermission,
			)
			if err != nil {
				errIndices = append(errIndices, i)
				change.Error = err

				continue
			}
		}

		traversedDirs[change.BaseDir] = change.BaseDir

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
	conf *config.Config,
	fileChanges file.Changes,
) error {
	if conf.TargetDir != "" {
		err := os.MkdirAll(conf.TargetDir, osutil.DirPermission)
		if err != nil {
			return err
		}
	}

	renameErrs := commit(fileChanges)
	if len(renameErrs) > 0 {
		return errRenameFailed.WithCtx(renameErrs)
	}

	return nil
}

// PostRename handles actions after a renaming operation, such as printing
// results, cleaning empty directories, and creating a backup file if applicable.
func PostRename(
	conf *config.Config,
	fileChanges file.Changes,
	renameErr error,
) {
	report.PrintResults(conf, fileChanges, renameErr)

	var cleanedDirs []string

	if conf.Clean {
		for _, dir := range traversedDirs {
			if dir == "." { // don't try to clean the working directory
				continue
			}

			// This will fail if the directory is not empty so no need
			// to check before hand
			err := os.Remove(dir)
			if err == nil {
				cleanedDirs = append(cleanedDirs, dir)
			}
		}
	}

	if len(fileChanges) != 0 && !conf.Revert {
		err := backupChanges(
			fileChanges,
			cleanedDirs,
			conf.BackupFilename,
			conf.BackupLocation,
		)
		if err != nil {
			report.BackupFailed(err)
		}
	}

	if conf.Revert && renameErr == nil {
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
