package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/report"
)

const (
	EnvDefaultOpts = "F2_DEFAULT_OPTS"
)

// supportedDefaultOpts contains flags whose values can be
// overridden through the `F2_DEFAULT_OPTS` environmental variable.
var supportedDefaultOpts = []string{
	flagClean.Name,
	flagExclude.Name,
	flagExcludeDir.Name,
	flagExec.Name,
	flagExiftoolOpts.Name,
	flagFixConflicts.Name,
	flagFixConflictsPattern.Name,
	flagHidden.Name,
	flagIgnoreCase.Name,
	flagIgnoreExt.Name,
	flagIncludeDir.Name,
	flagJSON.Name,
	flagNoColor.Name,
	flagQuiet.Name,
	flagRecursive.Name,
	flagSort.Name,
	flagSortr.Name,
	flagResetIndexPerDir.Name,
	flagStringMode.Name,
	flagVerbose.Name,
}

// isInputFromPipe detects if input is being piped to F2.
func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

// isOutputToPipe detects if F2's output is being piped to another command.
func isOutputToPipe() bool {
	fileInfo, _ := os.Stdout.Stat()

	return ((fileInfo.Mode() & os.ModeCharDevice) != os.ModeCharDevice)
}

// handlePipeInput processes input from a pipe and appends it to os.Args.
func handlePipeInput(reader io.Reader) error {
	if !isInputFromPipe() {
		return nil
	}

	scanner := bufio.NewScanner(bufio.NewReader(reader))

	for scanner.Scan() {
		os.Args = append(os.Args, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return errPipeRead.Wrap(err)
	}

	return nil
}

// loadDefaultOpts creates a CLI context with default options (F2_DEFAULT_OPTS)
// from the environment. Returns `nil` if default options do not exist.
func loadDefaultOpts() (*cli.Context, error) {
	var defaultCtx *cli.Context

	if optsEnv, exists := os.LookupEnv(EnvDefaultOpts); exists {
		defaultOpts := make([]string, len(os.Args))

		copy(defaultOpts, os.Args)

		defaultOpts = append(defaultOpts[:1], strings.Split(optsEnv, " ")...)

		app := CreateCLIApp(bytes.NewReader(nil), io.Discard)

		// override the default action to do nothing since only the
		// cli context contstructed from default opts is needed
		app.Action = func(ctx *cli.Context) error {
			defaultCtx = ctx
			return nil
		}

		// Run needs to be called here so that `defaultCtx` is populated.
		// The only expected error is if the provided flags or arguments
		// are incorrect
		err := app.Run(defaultOpts)
		if err != nil {
			return nil, errDefaultOptsParsing.Wrap(err)
		}
	}

	return defaultCtx, nil
}

// Get returns an F2 instance that reads from `reader` and writes to `writer`.
func Get(reader io.Reader, writer io.Writer) (*cli.App, error) {
	err := handlePipeInput(reader)
	if err != nil {
		return nil, err
	}

	app := CreateCLIApp(reader, writer)

	defaultCtx, err := loadDefaultOpts()
	if err != nil {
		return nil, err
	}

	app.Before = func(ctx *cli.Context) (err error) {
		// print short help and exit if no arguments or flags are present
		if ctx.NumFlags() == 0 && !ctx.Args().Present() {
			report.ShortHelp(ShortHelp(ctx.App))
			os.Exit(int(osutil.ExitOK))
		}

		config.Stdout = ctx.App.Writer
		config.Stdin = ctx.App.Reader

		defer (func() {
			_, initErr := config.Init(ctx, isOutputToPipe())
			if initErr != nil && err == nil {
				err = initErr
				return
			}
		})()

		app.Metadata["ctx"] = ctx

		// defaultCtx will be nil if `F2_DEFAULT_OPTS` is not set
		// in the environment
		if defaultCtx == nil {
			return nil
		}

		verbose := ctx.Bool("verbose")
		if !verbose {
			verbose = defaultCtx.Bool("verbose")
		}

		for _, defaultOpt := range supportedDefaultOpts {
			defaultValue := fmt.Sprintf("%v", defaultCtx.Value(defaultOpt))

			if ctx.IsSet(defaultOpt) && defaultCtx.IsSet(defaultOpt) {
				continue
			}

			if !ctx.IsSet(defaultOpt) && defaultCtx.IsSet(defaultOpt) {
				if x, ok := defaultCtx.Value(defaultOpt).(cli.StringSlice); ok {
					defaultValue = strings.Join(x.Value(), "|")
				}

				err := ctx.Set(defaultOpt, defaultValue)
				if err != nil {
					return errSetDefaultOpt.Wrap(err).
						Fmt(defaultValue, defaultOpt)
				}

				if verbose {
					report.DefaultOpt(defaultOpt, defaultValue)
				}
			}
		}

		return nil
	}

	return app, nil
}

func CreateCLIApp(r io.Reader, w io.Writer) *cli.App {
	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(ctx *cli.Context) {
		oldVersionPrinter(ctx)
		pterm.Fprint(
			w,
			"https://github.com/ayoisaiah/f2/releases/"+ctx.App.Version,
		)
	}

	app := &cli.App{
		Name: "f2",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage: `f2 bulk renames files and directories, matching files against a specified
pattern. It employs safety checks to prevent accidental overwrites and
offers several options for fine-grained control over the renaming process.`,
		Version:              "v2.0.1",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			flagCSV,
			flagExiftoolOpts,
			flagFind,
			flagReplace,
			flagUndo,
			flagAllowOverwrites,
			flagClean,
			flagExclude,
			flagExcludeDir,
			flagExec,
			flagFixConflicts,
			flagFixConflictsPattern,
			flagHidden,
			flagIncludeDir,
			flagIgnoreCase,
			flagIgnoreExt,
			flagJSON,
			flagMaxDepth,
			flagNoColor,
			flagOnlyDir,
			flagPair,
			flagPairOrder,
			flagQuiet,
			flagRecursive,
			flagReplaceLimit,
			flagResetIndexPerDir,
			flagSort,
			flagSortr,
			flagSortPerDir,
			flagSortVar,
			flagStringMode,
			flagTargetDir,
			flagVerbose,
		},
		UseShortOptionHandling: true,
		OnUsageError: func(_ *cli.Context, err error, _ bool) error {
			return err
		},
		Writer: w,
		Reader: r,
	}

	// Override the default help template
	cli.AppHelpTemplate = helpText(app)

	return app
}
