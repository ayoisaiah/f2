package f2

import (
	"errors"
	"io"

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
	conf, err := config.Init(ctx)
	if err != nil {
		return err
	}

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
		report.NoMatches(conf.JSON)
		return nil
	}

	changes, err := replace.Replace(conf, matches)
	if err != nil {
		return err
	}

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

func GetApp(reader io.Reader, writer io.Writer) *cli.App {
	f2App := app.Get(reader, writer)
	f2App.Action = run

	return f2App
}

// NewApp creates a new app instance.
func NewApp() *cli.App {
	f2App := app.New()
	f2App.Action = run

	return f2App
}
