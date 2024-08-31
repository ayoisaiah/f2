// Package report provides details about the renaming operation in table or json
// format
package report

import (
	"os"

	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/osutil"
)

func ExitWithErr(err error) {
	pterm.EnableOutput()
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf("%s %v", pterm.Red("error:"), err),
	)
	os.Exit(int(osutil.ExitError))
}

func BackupFailed(err error) {
	pterm.Fprintln(config.Stderr, pterm.Sprintf("backup failed: %v", err))
}

func BackupFileRemovalFailed(err error) {
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf("backup file cleanup failed: %v", err),
	)
}

// NoMatches prints out a message indicating that the find string failed
// to match any files.
func NoMatches(conf *config.Config) {
	if conf.Quiet {
		os.Exit(int(osutil.ExitError))
	}

	msg := "the search criteria didn't match any files"
	if conf.Revert {
		msg = "nothing to undo"
	}

	pterm.Fprintln(config.Stderr, pterm.Sprint(msg))
}

// Report prints a report of the renaming changes to be made
func Report(
	conf *config.Config,
	fileChanges file.Changes,
	conflictDetected bool,
) {
	if conf.JSON {
		fileChanges.RenderJSON(config.Stdout)
		return
	}

	fileChanges.RenderTable(config.Stdout)

	if conflictDetected || conf.JSON {
		return
	}

	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			"%s commit the above changes with the -x/--exec flag",
			pterm.Green("dry run:"),
		),
	)
}

func PrintResults(conf *config.Config, fileChanges file.Changes) {
	if !conf.Verbose && !conf.IsOutputToPipe {
		return
	}

	for i := range fileChanges {
		change := fileChanges[i]

		if conf.IsOutputToPipe && change.Error != nil {
			pterm.Println(change.TargetPath)
		}

		if !conf.Verbose {
			continue
		}

		if change.Error != nil {
			pterm.Fprintln(config.Stderr,
				pterm.Error.Sprintf(
					"renaming '%s' to '%s' failed: %v",
					change.SourcePath,
					change.TargetPath,
					change.Error,
				),
			)

			continue
		}

		pterm.Fprintln(config.Stderr,
			pterm.Sprintf(
				"%s '%s' to '%s'",
				pterm.Green("renamed:"),
				change.SourcePath,
				change.TargetPath,
			),
		)
	}
}
