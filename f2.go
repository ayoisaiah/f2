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

// run starts a new renaming operation.
func run(ctx *cli.Context) error {
	// TODO: Log the final context

	conf, err := config.Init(ctx)
	if err != nil {
		return err
	}

	slog.Info("configuration loaded", slog.Any("config", conf))

	report.Stdout = conf.Stdout
	report.Stderr = conf.Stderr

	if conf.Revert {
		return rename.Undo(conf)
	}

	matches, err := find.Find(conf)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		slog.Info("find matches completed: no matches found")
		report.NoMatches(conf.JSON)

		return nil
	}

	slog.Info(
		fmt.Sprintf("find matches completed: found %d matches", len(matches)),
		slog.Any("find_matches", matches),
		slog.Int("num_matches", len(matches)),
	)

	changes, err := replace.Replace(conf, matches)
	if err != nil {
		return err
	}

	slog.Info("bulk renaming completed", slog.Any("changes", changes))

	conflicts := validate.Validate(
		changes,
		conf.AutoFixConflicts,
		conf.AllowOverwrites,
	)

	if len(conflicts) > 0 {
		report.Conflicts(conflicts, conf.JSON)

		return errConflictDetected
	}

	return rename.Rename(conf, changes)
}

func New(reader io.Reader, writer io.Writer) *cli.App {
	f2App := app.Get(reader, writer)
	f2App.Action = run

	return f2App
}
