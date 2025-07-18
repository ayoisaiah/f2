package f2

import (
	"context"
	"io"
	"log/slog"
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
func execute(ctx context.Context, cmd *cli.Command) error {
	appConfig, err := config.Init(cmd, isOutputToPipe())
	if err != nil {
		return err
	}

	slog.DebugContext(
		ctx,
		"working configuration",
		slog.Any("config", appConfig),
	)

	matches, err := find.Find(appConfig)
	if err != nil {
		return err
	}

	slog.Debug(
		"find results",
		slog.Int("count", len(matches)),
		slog.Any("matches", matches),
	)

	if len(matches) == 0 {
		report.NoMatches(appConfig)

		return nil
	}

	if !appConfig.Revert {
		matches, err = replace.Replace(appConfig, matches)
		if err != nil {
			return err
		}
	}

	hasConflicts := validate.Validate(
		matches,
		appConfig.AutoFixConflicts,
		appConfig.AllowOverwrites,
	)

	if hasConflicts {
		report.Report(appConfig, matches, hasConflicts)

		return errConflictDetected
	}

	if !appConfig.Exec {
		report.Report(appConfig, matches, hasConflicts)
		return nil
	}

	err = rename.Rename(appConfig, matches)

	rename.PostRename(appConfig, matches, err)

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
