package app

import (
	"bufio"
	"context"
	"io"
	"net/mail"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/report"
)

const (
	EnvDefaultOpts = "F2_DEFAULT_OPTS"
)

var VersionString = "unset"

// isInputFromPipe detects if input is being piped to F2.
func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

// isOutputToPipe detects if F2's output is being piped to another command.
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

// Get returns an F2 instance that reads from `reader` and writes to `writer`.
func Get(reader io.Reader, writer io.Writer) (*cli.Command, error) {
	err := handlePipeInput(reader)
	if err != nil {
		return nil, err
	}

	app := CreateCLIApp(reader, writer)

	origArgs := make([]string, len(os.Args))

	copy(origArgs, os.Args)

	if optsEnv, exists := os.LookupEnv(EnvDefaultOpts); exists {
		args := strings.Fields(optsEnv)

		for _, token := range args {
			if strings.HasPrefix(token, "-") {
				if !supportedDefaultOpts[token] {
					return nil, errDefaultOptsParsing.Fmt(token)
				}
			}
		}

		args = append(args, os.Args[1:]...)
		os.Args = append(os.Args[:1], args...)
	}

	app.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		// print short help and exit if no arguments or flags are present
		if cmd.NumFlags() == 0 && !cmd.Args().Present() || len(origArgs) <= 1 {
			report.ShortHelp(ShortHelp(cmd))
			os.Exit(int(osutil.ExitOK))
		}

		config.Stdout = cmd.Writer
		config.Stdin = cmd.Reader

		app.Metadata["ctx"] = cmd

		return ctx, nil
	}

	return app, nil
}

func CreateCLIApp(r io.Reader, w io.Writer) *cli.Command {
	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(cmd *cli.Command) {
		oldVersionPrinter(cmd)
		v := cmd.Version

		if strings.Contains(v, "nightly") {
			v = "nightly"
		}

		pterm.Fprint(
			w,
			"https://github.com/ayoisaiah/f2/releases/"+v,
		)
	}

	app := &cli.Command{
		Name: "f2",
		Authors: []any{
			&mail.Address{
				Name:    "Ayooluwa Isaiah",
				Address: "ayo@freshman.tech",
			},
		},
		Usage: `f2 bulk renames files and directories, matching files against a specified
pattern. It employs safety checks to prevent accidental overwrites and
offers several options for fine-grained control over the renaming process.`,
		Version:               VersionString,
		EnableShellCompletion: true,
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
			flagInclude,
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
		UseShortOptionHandling:    true,
		DisableSliceFlagSeparator: true,
		OnUsageError: func(_ context.Context, _ *cli.Command, err error, _ bool) error {
			return err
		},
		Writer: w,
		Reader: r,
	}

	// Override the default help template
	app.CustomRootCommandHelpTemplate = helpText(app)

	return app
}
