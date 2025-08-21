// Package report provides details about the renaming operation in table or json
// format
package report

import (
	"os"
	"strings"

	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/localize"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
)

func ExitWithErr(err error) {
	pterm.EnableOutput()

	errPrefix := localize.T("report.error")
	errMessage := err.Error()

	s := strings.Split(errMessage, ":")
	if len(s) > 1 {
		errPrefix = strings.TrimSpace(s[0])
		errMessage = strings.TrimSpace(s[1])
	}

	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf("%s: %v", pterm.Red(errPrefix), errMessage),
	)
	os.Exit(int(osutil.ExitError))
}

func BackupFailed(err error) {
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			"%s: %v",
			pterm.Red(localize.T("report.backup_failed")),
			err,
		),
	)
}

func SearchEvalFailed(path, target string, err error) {
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			"%s: %v -> %s",
			pterm.Yellow(path),
			err,
			target,
		),
	)
}

func BackupFileRemovalFailed(err error) {
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			"%s: %v",
			pterm.Red(localize.T("report.backup_cleanup_failed")),
			err,
		),
	)
}

func ShortHelp(helpText string) {
	pterm.Fprintln(config.Stderr, helpText)
}

func DefaultOpt(opt, val string) {
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			localize.T(
				"report.default_opt",
			),
			pterm.Green(opt),
			pterm.Yellow(val),
		),
	)
}

func NonExistentFile(name string, row int) {
	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			"%s %d: %s",
			localize.T("report.non_existent_file"),
			row,
			name,
		),
	)
}

// NoMatches prints out a message indicating that the find string failed
// to match any files.
func NoMatches(conf *config.Config) {
	if conf.Quiet {
		os.Exit(int(osutil.ExitError))
	}

	msg := localize.T("report.no_matches")
	if conf.CSVFilename != "" {
		msg = localize.T("report.no_matches_csv")
	}

	if conf.Revert {
		msg = localize.T("report.no_matches_undo")
	}

	pterm.Fprintln(config.Stderr, pterm.Sprint(msg))
}

// Report prints a report of the renaming changes to be made.
func Report(
	conf *config.Config,
	fileChanges file.Changes,
	conflictDetected bool,
) {
	if conf.JSON {
		err := fileChanges.RenderJSON(config.Stdout)
		if err != nil {
			pterm.Fprintln(
				config.Stderr,
				pterm.Sprintf(
					"%s %v",
					pterm.Red(localize.T("report.error")),
					err,
				),
			)
		}

		return
	}

	fileChanges.RenderTable(config.Stdout, conf.NoColor)

	if conflictDetected || conf.JSON {
		return
	}

	pterm.Fprintln(
		config.Stderr,
		pterm.Sprintf(
			"%s %s",
			pterm.Green(localize.T("report.dry_run"), ":"),
			localize.T("report.commit_changes"),
		),
	)
}

// PrintResults prints the results of a renaming operation, including any errors
// encountered. It displays successful renames to stderr if verbose mode is
// enabled, and prints renamed paths to stdout if output is piped. Errors are
// always printed to stderr.
func PrintResults(conf *config.Config, fileChanges file.Changes, err error) {
	if err != nil {
		//nolint:errorlint // checking if err matches custom interface
		renameErr, ok := err.(*apperr.Error)
		if ok {
			errIndices, ok := renameErr.Context.([]int)
			if ok {
				for _, index := range errIndices {
					change := fileChanges[index]

					pterm.Fprintln(
						config.Stderr,
						pterm.Sprintf(
							"%s %v",
							pterm.Red(localize.T("report.error"), ":"),
							change.Error,
						),
					)
				}
			}
		}
	}

	if !conf.Verbose && !conf.PipeOutput {
		return
	}

	for i := range fileChanges {
		change := fileChanges[i]

		if conf.PipeOutput && change.Error == nil {
			pterm.Fprintln(config.Stdout, change.TargetPath)
		}

		if !conf.Verbose {
			continue
		}

		pterm.Fprintln(config.Stderr,
			pterm.Sprintf(
				"%s '%s' to '%s'",
				pterm.Green(localize.T("report.renamed"), ":"),
				change.SourcePath,
				change.TargetPath,
			),
		)
	}
}
