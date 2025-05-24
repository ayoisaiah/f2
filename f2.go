package f2

import (
	"context"
	"io"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/ayoisaiah/f2/v2/app"
	"github.com/ayoisaiah/f2/v2/find"
	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/rename"
	"github.com/ayoisaiah/f2/v2/replace"
	"github.com/ayoisaiah/f2/v2/report"
	"github.com/ayoisaiah/f2/v2/validate"
)

var errConflictDetected = &apperr.Error{
	Message: "conflict: resolve manually or use -F/--fix-conflicts",
}

func isOutputToPipe() bool {
	fileInfo, _ := os.Stdout.Stat()

	return ((fileInfo.Mode() & os.ModeCharDevice) != os.ModeCharDevice)
}

// execute initiates a new renaming operation based on the provided CLI context.
func execute(_ context.Context, cmd *cli.Command) error {
	appConfig, err := config.Init(cmd, isOutputToPipe())
	if err != nil {
		return err
	}

	changes, err := find.Find(appConfig)
	if err != nil {
		return err
	}

	if len(changes) == 0 {
		report.NoMatches(appConfig)

		return nil
	}

	if !appConfig.Revert {
		changes, err = replace.Replace(appConfig, changes)
		if err != nil {
			return err
		}
	}

	hasConflicts := validate.Validate(
		changes,
		appConfig.AutoFixConflicts,
		appConfig.AllowOverwrites,
	)

	if hasConflicts {
		report.Report(appConfig, changes, hasConflicts)

		return errConflictDetected
	}

	if !appConfig.Exec {
		report.Report(appConfig, changes, hasConflicts)
		return nil
	}

	err = rename.Rename(appConfig, changes)

	rename.PostRename(appConfig, changes, err)

	return err
}

// New creates a new CLI application for f2.
func New(reader io.Reader, writer io.Writer) (*cli.Command, error) {
	renamer, err := app.Get(reader, writer)
	if err != nil {
		return nil, err
	}

	renamer.Action = execute

	return renamer, nil
}
