package f2

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/urfave/cli/v2"

	"github.com/ayoisaiah/f2/app"
	"github.com/ayoisaiah/f2/find"
	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/rename"
	"github.com/ayoisaiah/f2/replace"
	"github.com/ayoisaiah/f2/report"
	"github.com/ayoisaiah/f2/validate"
)

var errConflictDetected = errors.New(
	"resolve conflicts before proceeding or use -F/--fix-conflicts to auto-fix",
)

// execute initiates a new renaming operation based on the provided CLI context
func execute(ctx *cli.Context) error {
	appConfig := config.Get()

	report.Stdout = appConfig.Stdout
	report.Stderr = appConfig.Stderr

	if appConfig.Revert {
		return rename.Undo(appConfig)
	}

	findMatches, err := find.Find(appConfig)
	if err != nil {
		return err
	}

	if len(findMatches) == 0 {
		slog.Info("find matches completed: no matches found")
		report.NoMatches(appConfig.JSON)

		return nil
	}

	slog.Info(
		fmt.Sprintf(
			"find matches completed: found %d matches",
			len(findMatches),
		),
		slog.Any("find_matches", findMatches),
		slog.Int("num_matches", len(findMatches)),
	)

	changes, err := replace.Replace(appConfig, findMatches)
	if err != nil {
		return err
	}

	slog.Info("bulk renaming completed", slog.Any("changes", changes))

	hasConflicts := validate.Validate(
		changes,
		appConfig.AutoFixConflicts,
		appConfig.AllowOverwrites,
	)

	if hasConflicts {
		report.NonInteractive(changes, hasConflicts)

		return errConflictDetected
	}

	if !appConfig.Exec {
		report.Report(appConfig, changes)
		return nil
	}

	err = rename.Rename(appConfig, changes)
	if err != nil {
		return err
	}

	if appConfig.Print {
		for i := range changes {
			fmt.Println(changes[i].RelTargetPath)
		}
	}

	return nil
}

// New creates a new CLI application for f2
func New(reader io.Reader, writer io.Writer) (*cli.App, error) {
	renamer, err := app.Get(reader, writer)
	if err != nil {
		return nil, err
	}

	renamer.Action = execute

	return renamer, nil
}
