// Package rename commits the renaming operation to the filesystem and reports
// errors if any. It also creates a backup file for the operation and provides a
// way to undo any renaming operation
package rename

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	internaljson "github.com/ayoisaiah/f2/internal/json"
	internalos "github.com/ayoisaiah/f2/internal/os"
	"github.com/ayoisaiah/f2/internal/sortfiles"
	"github.com/ayoisaiah/f2/report"
)

var errRenameFailed = errors.New(
	"some files could not be renamed. Revert the changes through the --undo flag",
)

var errs []int

// rename iterates over all the matches and renames them on the filesystem.
// Directories are auto-created if necessary, and errors are aggregated.
func rename(
	changes []*file.Change,
) []int {
	for i := range changes {
		change := changes[i]

		targetPath := change.RelTargetPath

		// skip paths that are unchanged in every aspect
		if change.RelSourcePath == targetPath {
			continue
		}

		// Account for case insensitive filesystems where renaming a filename to its
		// upper or lowercase equivalent doesn't work. Fixing this involves the
		// following steps:
		// 1. Prefix <target> with __<time>__ if case insensitive FS
		// 2. Rename <source> to <target>
		// 3. Rename __<time>__<target> to <target> if case insensitive FS
		var caseInsensitiveFS bool
		if strings.EqualFold(change.RelSourcePath, targetPath) {
			caseInsensitiveFS = true
			timeStr := fmt.Sprintf("%d", time.Now().UnixNano())
			targetPath = filepath.Join(
				change.BaseDir,
				"__"+timeStr+"__"+change.Target, // step 1
			)
		}

		// If target contains a slash, create all missing
		// directories before renaming the file
		if strings.Contains(change.Target, "/") ||
			strings.Contains(change.Target, `\`) &&
				runtime.GOOS == internalos.Windows {
			// No need to check if the `dir` exists or if there are several
			// consecutive slashes since `os.MkdirAll` handles that
			dir := filepath.Dir(change.Target)

			//nolint:gomnd // number can be understood from context
			err := os.MkdirAll(filepath.Join(change.BaseDir, dir), 0o750)
			if err != nil {
				errs = append(errs, i)
				change.Error = err

				continue
			}
		}

		err := os.Rename(change.RelSourcePath, targetPath) // step 2
		// if the intermediate rename is successful,
		// proceed with the original renaming operation
		if err == nil && caseInsensitiveFS {
			err = os.Rename(targetPath, change.RelTargetPath) // step 3
		}

		if err != nil {
			errs = append(errs, i)
			change.Error = err

			continue
		}
	}

	return errs
}

func workingDirHash(workingDir string) string {
	h := md5.New()
	h.Write([]byte(workingDir))

	return fmt.Sprintf("%x", h.Sum(nil)) + ".json"
}

// backupChanges records the details of a renaming operation to the filesystem
// so that it may be reverted if necessary.
func backupChanges(changes []*file.Change, workingDir string) error {
	filename := workingDirHash(workingDir)

	backupFilePath, err := xdg.DataFile(
		filepath.Join("f2", "backups", filename),
	)
	if err != nil {
		return err
	}

	// Create or truncate backupFile
	backupFile, err := os.Create(backupFilePath)
	if err != nil {
		return err
	}

	defer func() {
		ferr := backupFile.Close()
		if ferr != nil {
			err = ferr
		}
	}()

	successfulChanges := make([]*file.Change, len(changes))

	copy(successfulChanges, changes)

	// remove files that errored out
	for i := len(successfulChanges) - 1; i >= 0; i-- {
		if successfulChanges[i].Error != nil {
			successfulChanges = append(
				successfulChanges[:i],
				successfulChanges[i+1:]...)
		}
	}

	b, err := internaljson.GetOutput(successfulChanges)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(backupFile)

	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	return writer.Flush()
}

// commit applies the renaming operation to the filesystem.
// A backup file is auto created as long as at least one file
// was renamed and it wasn't an undo operation.
func commit(
	fileChanges []*file.Change,
	conf *config.Config,
) []int {
	errs = rename(fileChanges)

	if conf.Verbose {
		for _, change := range fileChanges {
			if change.Error != nil {
				pterm.Fprintln(report.Stderr,
					pterm.Error.Sprintf(
						"Failed to rename %s to %s",
						change.RelSourcePath,
						change.RelTargetPath,
					),
				)

				continue
			}

			pterm.Fprintln(report.Stderr,
				pterm.Success.Printfln(
					"Renamed '%s' to '%s'",
					pterm.Yellow(change.RelSourcePath),
					pterm.Yellow(change.RelTargetPath),
				),
			)
		}
	}

	if !conf.Revert {
		err := backupChanges(fileChanges, conf.WorkingDir)
		if err != nil {
			report.BackupFailed(err)
		}
	}

	if len(errs) > 0 {
		sort.SliceStable(fileChanges, func(i, _ int) bool {
			compareElement1 := fileChanges[i]

			return compareElement1.Error == nil
		})
	}

	return errs
}

// Rename prints the changes to be made in dry-run mode
// or commits the operation to the filesystem if in execute mode.
func Rename(
	conf *config.Config,
	fileChanges []*file.Change,
) error {
	if conf.IncludeDir {
		fileChanges = sortfiles.FilesBeforeDirs(fileChanges, conf.Revert)
	}

	if !conf.Interactive && !conf.Exec && !conf.JSON {
		report.NonInteractive(fileChanges)
		return nil
	}

	if conf.JSON {
		report.JSON(fileChanges)
	} else if conf.Interactive {
		report.Interactive(fileChanges)
	}

	if !conf.Exec {
		return nil
	}

	renameErrs := commit(fileChanges, conf)
	if renameErrs != nil {
		// TODO: Print the errors
		return errRenameFailed
	}

	return nil
}
