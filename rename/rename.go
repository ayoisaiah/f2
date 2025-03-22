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

	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/internal/status"
	"github.com/ayoisaiah/f2/v2/report"
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
		ch := fileChanges[i]

		if ch.Status == status.Ignored {
			continue
		}

		targetPath := ch.TargetPath

		// skip paths that are unchanged in every aspect
		if ch.SourcePath == targetPath {
			continue
		}

		// Workaround for case insensitive filesystems where renaming a filename to
		// its upper or lowercase equivalent doesn't work. Fixing this involves the
		// following steps:
		// 1. Prefix and suffix <target> with __<time>__
		// 2. Rename <source> to <target>
		// 3. Rename __<time>__<target>__<time>__ to <target>
		var isCaseChangeOnly bool // only the target case is changing
		if strings.EqualFold(ch.SourcePath, targetPath) {
			isCaseChangeOnly = true
			timeStr := fmt.Sprintf("%d", time.Now().UnixNano())
			targetPath = filepath.Join(
				ch.TargetDir,
				"__"+timeStr+"__"+ch.Target+"__"+timeStr+"__", // step 1
			)
		}

		// If target contains a slash, create all missing
		// directories before renaming the file
		if strings.Contains(ch.Target, "/") ||
			strings.Contains(ch.Target, `\`) &&
				runtime.GOOS == osutil.Windows {
			// No need to check if the `dir` exists or if there are several
			// consecutive slashes since `os.MkdirAll` handles that
			dir := filepath.Dir(ch.Target)

			err := os.MkdirAll(
				filepath.Join(ch.TargetDir, dir),
				osutil.DirPermission,
			)
			if err != nil {
				errIndices = append(errIndices, i)
				ch.Error = err

				continue
			}
		}

		traversedDirs[ch.BaseDir] = ch.BaseDir

		err := os.Rename(ch.SourcePath, targetPath) // step 2
		// if the intermediate rename is successful,
		// proceed with the original renaming operation
		if err == nil && isCaseChangeOnly {
			err = os.Rename(targetPath, ch.TargetPath) // step 3
		}

		if err != nil {
			errIndices = append(errIndices, i)
			ch.Error = err
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
	var cleanedDirs []string

	if conf.Clean && !conf.Revert {
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
		backupFilePath := filepath.Join(
			os.TempDir(),
			"f2",
			"backups",
			conf.BackupFilename,
		)

		if err := os.Remove(backupFilePath); err != nil {
			report.BackupFileRemovalFailed(err)
		}
	}

	report.PrintResults(conf, fileChanges, renameErr)
}
