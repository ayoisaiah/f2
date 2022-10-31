// Package rename commits the renaming operation to the filesystem and reports
// errors if any. It also creates a backup file for the operation and provides a
// way to undo any renaming operation
package rename

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/config"
	"github.com/ayoisaiah/f2/internal/file"
	internaljson "github.com/ayoisaiah/f2/internal/json"
	internalos "github.com/ayoisaiah/f2/internal/os"
	internalpath "github.com/ayoisaiah/f2/internal/path"
	internalsort "github.com/ayoisaiah/f2/internal/sort"
	"github.com/ayoisaiah/f2/report"
)

var errs []int

// rename iterates over all the matches and renames them on the filesystem.
// Directories are auto-created if necessary, and errors are aggregated.
func rename(changes []*file.Change) []int {
	conf := config.Get()

	for i := range changes {
		change := changes[i]

		source, target := change.Source, change.Target
		source = filepath.Join(change.BaseDir, source)
		target = filepath.Join(change.BaseDir, target)

		// skip unchanged file names
		if source == target {
			continue
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
				change.Error = err.Error()

				continue
			}
		}

		if err := os.Rename(source, target); err != nil {
			errs = append(errs, i)
			change.Error = err.Error()

			if conf.IsVerbose() {
				pterm.Fprintln(conf.Stderr(),
					pterm.Error.Sprintf(
						"Failed to rename %s to %s",
						source,
						target,
					),
				)
			}
		} else if conf.IsVerbose() && !conf.JSON() {
			pterm.Success.Printfln("Renamed '%s' to '%s'", pterm.Yellow(source), pterm.Yellow(target))
		}
	}

	return errs
}

// backupChanges records the details of a renaming operation to the filesystem
// so that it may be reverted if necessary.
func backupChanges(changes []*file.Change, errs []int) error {
	conf := config.Get()

	workingDir := strings.ReplaceAll(
		conf.WorkingDir(),
		internalpath.Separator,
		"_",
	)
	if runtime.GOOS == internalos.Windows {
		workingDir = strings.ReplaceAll(workingDir, ":", "_")
	}

	filename := workingDir + ".json"

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

	b, err := internaljson.GetOutput(changes, errs)
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
func commit(changes []*file.Change) []int {
	conf := config.Get()

	changes = internalsort.FilesBeforeDirs(changes)

	errs = rename(changes)

	if !conf.ShouldRevert() {
		err := backupChanges(changes, errs)
		if err != nil {
			pterm.Fprintln(conf.Stderr(),
				pterm.Warning.Sprintf(
					"Failed to backup renaming operation due to error: %s",
					err.Error(),
				),
			)
		}
	}

	if len(errs) > 0 {
		sort.SliceStable(changes, func(i, _ int) bool {
			compareElement1 := changes[i]

			return compareElement1.Error == ""
		})
	}

	return errs
}

// Execute prints the changes to be made in dry-run mode
// or commits the operation to the filesystem if in execute mode.
func Execute(changes []*file.Change) []int {
	conf := config.Get()

	if conf.SimpleMode() {
		report.Changes(changes, nil)

		if conf.JSON() {
			return nil
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("\033[s")
		fmt.Print("Press ENTER to commit the above changes")

		// Block until user input before beginning next session
		_, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			pterm.Fprintln(conf.Stderr(), pterm.Error.Print(err))
			return nil
		}
	}

	return commit(changes)
}

func GetErrs() []int {
	return errs
}
