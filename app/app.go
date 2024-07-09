package app

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
	slogctx "github.com/veqryn/slog-context"
)

const (
	EnvUpdateNotifier = "F2_UPDATE_NOTIFIER"
	EnvNoColor        = "NO_COLOR"
	EnvF2NoColor      = "F2_NO_COLOR"
	EnvDefaultOpts    = "F2_DEFAULT_OPTS"
	EnvDebug          = "F2_DEBUG"
)

// supportedDefaultOptions contains those flags that can be
// overridden through the `F2_DEFAULT_OPTS` environmental variable.
var supportedDefaultOptions = []string{
	"hidden", "allow-overwrites", "exclude", "exclude-dir", "exec", "fix-conflicts", "fix-conflicts-pattern", "include-dir", "ignore-case", "ignore-ext", "interactive", "json", "max-depth", "no-color", "only-dir", "quiet", "recursive", "replace-limit", "sort", "sortr", "string-mode", "verbose", "exiftool-opts",
}

func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}

	if debugVal, ok := os.LookupEnv(EnvDebug); ok {
		opts.Level = slog.LevelInfo
	} else if debugVal == "0" {
		opts.Level = slog.LevelDebug
	}

	h := slogctx.NewHandler(slog.NewJSONHandler(os.Stderr, opts), nil)

	l := slog.New(h)

	slog.SetDefault(l)
}

func init() {
	// Disable colour output if NO_COLOR is set
	if _, exists := os.LookupEnv(EnvNoColor); exists {
		pterm.DisableStyling()
	}

	// Disable colour output if F2_NO_COLOR is set
	if _, exists := os.LookupEnv(EnvF2NoColor); exists {
		pterm.DisableStyling()
	}

	// Override the default help template
	cli.AppHelpTemplate = helpText()

	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(c *cli.Context) {
		oldVersionPrinter(c)
		pterm.Printfln(
			"https://github.com/ayoisaiah/f2/releases/%s",
			c.App.Version,
		)

		if _, found := os.LookupEnv(EnvUpdateNotifier); found {
			checkForUpdates(c.App)
		}
	}

	pterm.Error.MessageStyle = pterm.NewStyle(pterm.FgRed)
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgRed, pterm.FgBlack),
	}
}

// getDefaultOptsCtx creates a new `cli.Context` that represents the
// program's options if it were run solely with the flags and arguments
// represented in the `F2_DEFAULT_OPTS` environmental variable.
// If this variable does not exist in the env, the returned Context
// is `nil`.
func getDefaultOptsCtx() *cli.Context {
	var defaultCtx *cli.Context

	if optsEnv, exists := os.LookupEnv(EnvDefaultOpts); exists {
		slog.Debug(
			"found default options in environment",
			slog.String("default_opts", optsEnv),
		)

		defaultOpts := make([]string, len(os.Args))

		copy(defaultOpts, os.Args)

		defaultOpts = append(defaultOpts[:1], strings.Split(optsEnv, " ")...)

		app := New()

		// override the default action to do nothing since only the
		// cli context contstructed from default opts is needed
		app.Action = func(ctx *cli.Context) error {
			defaultCtx = ctx
			return nil
		}

		// Run needs to be called here so that `defaultCtx` is populated
		// The only expected error is if the provided flags or arguments
		// are incorrect
		err := app.Run(defaultOpts)
		if err != nil {
			slog.Debug("default options parse error",
				slog.String("error", fmt.Sprintf("%v", err)),
			)

			pterm.Fprintln(
				os.Stderr,
				pterm.Error.Sprintf(
					"error parsing %s: %v",
					EnvDefaultOpts,
					err,
				),
			)

			os.Exit(1)
		}
	}

	return defaultCtx
}

// checkForUpdates alerts the user if an updated version of F2 is available.
func checkForUpdates(app *cli.App) {
	spinner, _ := pterm.DefaultSpinner.Start("Checking for updates...")
	c := http.Client{Timeout: 10 * time.Second}

	resp, err := c.Get("https://github.com/ayoisaiah/f2/releases/latest")
	if err != nil {
		pterm.Fprintln(
			os.Stderr,
			pterm.Error.Sprint("Failed to check for update"),
		)

		return
	}

	defer resp.Body.Close()

	var version string

	_, err = fmt.Sscanf(
		resp.Request.URL.String(),
		"https://github.com/ayoisaiah/f2/releases/tag/%s",
		&version,
	)
	if err != nil {
		pterm.Fprintln(
			os.Stderr,
			pterm.Error.Sprint("Failed to get latest version"),
		)

		return
	}

	if version == app.Version {
		text := pterm.Sprintf(
			"Congratulations, you are using the latest version of %s",
			app.Name,
		)
		spinner.Success(text)
	} else {
		pterm.Info.Prefix = pterm.Prefix{
			Text:  "UPDATE AVAILABLE",
			Style: pterm.NewStyle(pterm.BgYellow, pterm.FgBlack),
		}
		pterm.Info.Printfln("A new release of F2 is available: %s at %s", version, resp.Request.URL.String())
	}
}

// Get returns an F2 instance that reads from `reader` and writes to `writer`.
func Get(reader io.Reader, writer io.Writer) *cli.App {
	app := New()

	defaultCtx := getDefaultOptsCtx()

	app.Before = func(ctx *cli.Context) error {
		if ctx.Bool("no-color") {
			slog.Debug("disabling styling")
			pterm.DisableStyling()
		}

		if ctx.Bool("quiet") {
			slog.Debug("disabling output")
			pterm.DisableOutput()
		}

		// print short help and exit if no arguments or flags are present
		if ctx.NumFlags() == 0 && !ctx.Args().Present() {
			slog.Debug("print short help and exit")
			pterm.Println(ShortHelp(ctx.App))
			os.Exit(1)
		}

		app.Metadata["reader"] = reader
		app.Metadata["writer"] = writer

		if ctx.NumFlags() == 0 {
			slog.Debug(
				"simple mode detected",
				slog.Int("num_flags", ctx.NumFlags()),
			)

			app.Metadata["simple-mode"] = true
		}

		// defaultCtx will be nil if `F2_DEFAULT_OPTS` is not set
		// in the environment
		if defaultCtx == nil {
			return nil
		}

		for _, defaultOpt := range supportedDefaultOptions {
			defaultValue := fmt.Sprintf("%v", defaultCtx.Value(defaultOpt))

			if ctx.IsSet(defaultOpt) && defaultCtx.IsSet(defaultOpt) {
				cliValue := fmt.Sprintf("%v", ctx.Value(defaultOpt))
				slog.Debug(
					fmt.Sprintf(
						"command line flag overrides default option for: %s",
						defaultOpt,
					),
					slog.String("flag", defaultOpt),
					slog.String("command_line_value", cliValue),
					slog.String("default_value", defaultValue),
				)

				continue
			}

			if !ctx.IsSet(defaultOpt) && defaultCtx.IsSet(defaultOpt) {
				if x, ok := defaultCtx.Value(defaultOpt).(cli.StringSlice); ok {
					defaultValue = strings.Join(x.Value(), "|")
				}

				slog.Debug(
					fmt.Sprintf("set default option for flag: %s", defaultOpt),
					slog.String("flag", defaultOpt),
					slog.String("default_value", defaultValue),
				)

				err := ctx.Set(defaultOpt, defaultValue)
				if err != nil {
					slog.Debug("failed to set default option for: %s",
						slog.String("flag", defaultOpt),
						slog.String("default_value", defaultValue),
					)
					pterm.Fprintln(os.Stderr,
						pterm.Warning.Sprintf(
							"Unable to set default option for: %s",
							defaultOpt,
						),
					)
				}
			}
		}

		return nil
	}

	return app
}

func New() *cli.App {
	usageText := `FLAGS [OPTIONS] [PATHS TO FILES AND DIRECTORIES...]
or: FIND [REPLACE] [PATHS TO FILES AND DIRECTORIES...]`

	return &cli.App{
		Name: "F2",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage:                "F2 is a command-line tool for batch renaming multiple files and directories quickly and safely.",
		UsageText:            usageText,
		Version:              "v1.9.1",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "csv",
				Usage:       "Load a CSV file, and rename according to its contents.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Renaming-from-a-CSV-file.",
				DefaultText: "<path/to/csv/file>",
				TakesFile:   true,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug mode",
			},
			&cli.StringFlag{
				Name:  "exiftool-opts",
				Usage: "Provide custom options when using ExifTool variables",
			},
			&cli.StringSliceFlag{
				Name:        "find",
				Aliases:     []string{"f"},
				Usage:       "Search pattern. Treated as a regular expression unless combined with s/--string-mode.\n\t\t\t\tDefaults to the entire file name if omitted.",
				DefaultText: "<pattern>",
			},
			&cli.StringSliceFlag{
				Name:        "replace",
				Aliases:     []string{"r"},
				Usage:       "Replacement string or pattern. Supports several kinds of variables.\n\t\t\t\tDefaults to an empty string if omitted.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Built-in-variables.",
				DefaultText: "<string>",
			},
			&cli.BoolFlag{
				Name:    "undo",
				Aliases: []string{"u"},
				Usage:   "Undo the last operation performed in the current working directory if possible.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Undoing-a-renaming-operation.",
			},
			&cli.BoolFlag{
				Name:  "allow-overwrites",
				Usage: "Allow the renaming operation to overwite existing files.\n\t\t\t\tNote that using this option can lead to unrecoverable data loss in the renamed files.",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"E"},
				// TODO: Add note mentioning that this doesn't affect recursing into
				// directories that are matched
				Usage:       "Exclude files and directories that match the provided regular expression pattern. \n\t\t\t\tMultiple exclude patterns can be specified by repeating this option in a command.\n\n\t\t\t\tE.g: `-E 'json' -E 'yml'` filters out JSON and YAML files from the matched files.\n\t\t\t\tIt is equivalent to `-E 'json|yaml'`.",
				DefaultText: "<pattern>",
			},
			&cli.StringSliceFlag{
				Name:        "exclude-dir",
				Usage:       "Prevent F2 from recursing into directories that match the provided regular expression pattern.",
				DefaultText: "<pattern>",
			},
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "Execute the renaming operation and commit the changes to the filesystem.",
			},
			&cli.BoolFlag{
				Name:    "fix-conflicts",
				Aliases: []string{"F"},
				Usage:   "Automatically fix renaming conflicts based on predefined rules.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection.",
			},
			&cli.StringFlag{
				Name:  "fix-conflicts-pattern",
				Usage: "Custom pattern for renaming files when conflicts occur. This pattern should be a valid Go format string containing a single '%d' placeholder for the conflict index (e.g., '_%02d' to generate _01, _02, etc.). If not specified, the default pattern '(%d)' is used.",
			},
			&cli.BoolFlag{
				Name:    "hidden",
				Aliases: []string{"H"},
				Usage:   "Match hidden files (skipped by default) and search hidden directories for matches\n\t\t\t\t(if -R/--recursive is used).\n\t\t\t\tHidden files are those that start with a dot character '. (all OSes).\n\t\t\t\tOn Windows, files with the `hidden` attribute are also considered hidden.\n\t\t\t\tIf you want to match hidden directories as well, combine this the -d/--include-dir",
			},
			&cli.BoolFlag{
				Name:    "include-dir",
				Aliases: []string{"d"},
				Usage:   "Match directories in the renaming operation (they are exempted by default).",
			},
			&cli.BoolFlag{
				Name:    "ignore-case",
				Aliases: []string{"i"},
				Usage:   "Ignore string casing when searching for matches.",
			},
			&cli.BoolFlag{
				Name:    "ignore-ext",
				Aliases: []string{"e"},
				Usage:   "Ignore the file extension when searching for matches.",
			},
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"n"},
				Usage:   "Prompt to execute renaming operation after a dry-run.",
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Always produce JSON output except for error messages which go to the standard error",
			},
			&cli.UintFlag{
				Name:        "max-depth",
				Aliases:     []string{"m"},
				Usage:       "Indicates the maximum depth for a recursive search (set to 0 by default for no limit).",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Disable coloured output.",
			},
			&cli.BoolFlag{
				Name:  "non-interactive",
				Usage: "Print file paths after successful renaming",
			},
			&cli.BoolFlag{
				Name:    "only-dir",
				Aliases: []string{"D"},
				Usage:   "Rename only directories, not files (implies -d/--include-dir).",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Don't print out any information to the standard output.\n\t\t\t\tErrors will continue being sent to the standard error",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"R"},
				Usage:   "Recursively traverse directories when searching for matches.",
			},
			&cli.IntFlag{
				Name:        "replace-limit",
				Aliases:     []string{"l"},
				Usage:       "Limit the number of replacements to be made on each matched file.\n\t\t\t\tIt's set to 0 by default indicating that all matches should be replaced.\n\t\t\t\tCan be set to a negative integer to start replacing from the end of the file name.",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.StringFlag{
				Name: "sort",
				Usage: `Sort the matches in ascending order according to the provided '<sort>'.
					Allowed sort values:
						'default' : lexographical order.
						'size'    : sort by file size.
						'natural' : sort according to natural order.
						'mtime'   : sort by file last modified time.
						'btime'   : sort by file creation time.
						'atime'   : sort by file last access time.
						'ctime'   : sort by file metadata last change time.

        To sort results in reverse or descending order, use the --sortr flag. Also,
        this flag overrides --sortr. 
        `,
				DefaultText: "<sort>",
			},
			&cli.StringFlag{
				Name:        "sortr",
				Usage:       "Same options as --sort but presents the matches in the reverse order.",
				DefaultText: "<sort>",
			},
			&cli.BoolFlag{
				Name:    "string-mode",
				Aliases: []string{"s"},
				Usage:   "Treats the search pattern (specified by -f/--find) as a non-regex string.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Enable verbose output during the renaming operation.",
			},
		},
		UseShortOptionHandling: true,
		OnUsageError: func(_ *cli.Context, err error, _ bool) error {
			return err
		},
	}
}
