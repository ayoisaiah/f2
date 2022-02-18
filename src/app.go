package f2

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const (
	envUpdateNotifier = "F2_UPDATE_NOTIFIER"
	envNoColor        = "NO_COLOR"
	envF2NoColor      = "F2_NO_COLOR"
	envDefaultOpts    = "F2_DEFAULT_OPTS"
)

// supportedDefaultFlags contains those flags that can be
// overridden through the `F2_DEFAULT_OPTS` environmental variable.
var supportedDefaultFlags = []string{
	"hidden", "allow-overwrites", "exclude", "exec", "fix-conflicts", "include-dir", "ignore-case", "ignore-ext", "max-depth", "no-color", "only-dir", "quiet", "recursive", "replace-limit", "sort", "sortr", "string-mode", "verbose",
}

// getDefaultOptsCtx creates a new `cli.Context` that represents the
// program's options if it were run solely with the flags and arguments
// represented in the `F2_DEFAULT_OPTS` environmental variable.
// If this variable does not exist in the env, the returned Context
// is `nil`.
func getDefaultOptsCtx() *cli.Context {
	var defaultCtx *cli.Context

	if optsEnv, exists := os.LookupEnv(envDefaultOpts); exists {
		var defaultOpts = make([]string, len(os.Args))

		copy(defaultOpts, os.Args)

		defaultOpts = append(defaultOpts[:1], strings.Split(optsEnv, " ")...)

		app := newApp()

		app.Before = func(c *cli.Context) error {
			if c.IsSet("find") || c.IsSet("replace") || c.IsSet("csv") ||
				c.IsSet("undo") {
				pterm.Warning.Printfln(
					"%s are not supported as default options",
					"'find', 'replace', 'csv' and 'undo'",
				)
			}

			defaultCtx = c

			return nil
		}

		app.Action = func(c *cli.Context) error {
			return nil
		}

		_ = app.Run(defaultOpts)
	}

	return defaultCtx
}

func GetApp(reader io.Reader, writer io.Writer) *cli.App {
	app := newApp()

	defaultCtx := getDefaultOptsCtx()

	app.Before = func(c *cli.Context) error {
		app.Metadata["reader"] = reader
		app.Metadata["writer"] = writer

		if c.NumFlags() == 0 {
			app.Metadata["simple-mode"] = true
		} else if defaultCtx != nil {
			// defaultCtx will be nil if `F2_DEFAULT_OPTS` is not set
			// in the environment
			for _, v := range supportedDefaultFlags {
				value := fmt.Sprintf("%v", defaultCtx.Value(v))

				if !c.IsSet(v) && defaultCtx.IsSet(v) {
					if x, ok := defaultCtx.Value(v).(cli.StringSlice); ok {
						value = strings.Join(x.Value(), "|")
					}

					err := c.Set(v, value)
					if err != nil {
						pterm.Warning.Printfln(
							"Unable to set default option for: %s",
							v,
						)
					}
				}
			}
		}

		return nil
	}

	return app
}

func init() {
	// Disable colour output if NO_COLOR is set
	if _, exists := os.LookupEnv(envNoColor); exists {
		disableStyling()
	}

	// Disable colour output if F2_NO_COLOR is set
	if _, exists := os.LookupEnv(envF2NoColor); exists {
		disableStyling()
	}

	// Override the default help template
	cli.AppHelpTemplate = helpText()

	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(c *cli.Context) {
		oldVersionPrinter(c)
		fmt.Printf(
			"https://github.com/ayoisaiah/f2/releases/%s\n",
			c.App.Version,
		)

		if _, found := os.LookupEnv(envUpdateNotifier); found {
			checkForUpdates(c.App)
		}
	}

	pterm.Error.MessageStyle = pterm.NewStyle(pterm.FgRed)
	pterm.Error.Prefix = pterm.Prefix{
		Text:  "ERROR",
		Style: pterm.NewStyle(pterm.BgRed, pterm.FgBlack),
	}
}

// disableStyling disables all styling provided by pterm.
func disableStyling() {
	pterm.DisableColor()
	pterm.DisableStyling()
	pterm.Debug.Prefix.Text = ""
	pterm.Info.Prefix.Text = ""
	pterm.Success.Prefix.Text = ""
	pterm.Warning.Prefix.Text = ""
	pterm.Error.Prefix.Text = ""
	pterm.Fatal.Prefix.Text = ""
}

// checkForUpdates alerts the user if an updated version of F2 is available.
func checkForUpdates(app *cli.App) {
	spinner, _ := pterm.DefaultSpinner.Start("Checking for updates...")
	c := http.Client{Timeout: 10 * time.Second}

	resp, err := c.Get("https://github.com/ayoisaiah/f2/releases/latest")
	if err != nil {
		pterm.Error.Println("Failed to check for update")
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
		pterm.Error.Println("Failed to get latest version")
		return
	}

	if version == app.Version {
		text := pterm.Sprintf(
			"Congratulations, you are using the latest version of %s",
			app.Name,
		)
		spinner.Success(text)
	} else {
		pterm.Warning.Prefix = pterm.Prefix{
			Text:  "UPDATE AVAILABLE",
			Style: pterm.NewStyle(pterm.BgYellow, pterm.FgBlack),
		}
		pterm.Warning.Printfln("A new release of F2 is available: %s at %s", version, resp.Request.URL.String())
	}
}

// newApp creates a new app instance.
func newApp() *cli.App {
	usageText := `FLAGS [OPTIONS] [PATHS TO FILES OR DIRECTORIES...]
or: f2 FIND [REPLACE] [PATHS TO FILES OR DIRECTORIES...]`

	return &cli.App{
		Name: "f2",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage:                "f2 is a command-line tool for batch renaming multiple files and directories quickly and safely.",
		UsageText:            usageText,
		Version:              "v1.7.2",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "csv",
				Usage:       "Load a CSV file, and rename according to its contents.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Renaming-from-a-CSV-file.",
				DefaultText: "<csv file>",
			},
			&cli.StringSliceFlag{
				Name:        "find",
				Aliases:     []string{"f"},
				Usage:       "Search pattern. Treated as a regular expression by default unless --string-mode is also used.\n\t\t\t\tDefaults to the entire file name if omitted.",
				DefaultText: "<pattern>",
			},
			&cli.StringSliceFlag{
				Name:        "replace",
				Aliases:     []string{"r"},
				Usage:       "Replacement string. If omitted, defaults to an empty string. Supports several kinds of variables.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Built-in-variables.",
				DefaultText: "<string>",
			},
			&cli.BoolFlag{
				Name:    "undo",
				Aliases: []string{"u"},
				Usage:   "Undo the last operation performed in the current working directory if possible.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Undoing-a-renaming-operation.",
			},
			&cli.BoolFlag{
				Name:  "allow-overwrites",
				Usage: "Allow the overwriting of existing files.",
			},
			&cli.StringSliceFlag{
				Name:        "exclude",
				Aliases:     []string{"E"},
				Usage:       "Exclude files/directories that match the given search pattern. Treated as a regular expression.\n\t\t\t\tMultiple exclude patterns can be specified by repeating this option.",
				DefaultText: "<pattern>",
			},
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "Commit the renaming operation to the filesystem.",
			},
			&cli.BoolFlag{
				Name:    "fix-conflicts",
				Aliases: []string{"F"},
				Usage:   "Automatically fix conflicts based on predefined rules.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection.",
			},
			&cli.BoolFlag{
				Name:    "hidden",
				Aliases: []string{"H"},
				Usage:   "Include hidden files (they are skipped by default).",
			},
			&cli.BoolFlag{
				Name:    "include-dir",
				Aliases: []string{"d"},
				Usage:   "Include directories (they are exempted by default).",
			},
			&cli.BoolFlag{
				Name:    "ignore-case",
				Aliases: []string{"i"},
				Usage:   "Search for matches case insensitively.",
			},
			&cli.BoolFlag{
				Name:    "ignore-ext",
				Aliases: []string{"e"},
				Usage:   "Ignore the file extension when searching for matches.",
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
				Name:    "only-dir",
				Aliases: []string{"D"},
				Usage:   "Rename only directories, not files (implies --include-dir).",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Don't print out any information (except errors).",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"R"},
				Usage:   "Recursively traverse directories when searching for matches.",
			},
			&cli.IntFlag{
				Name:        "replace-limit",
				Aliases:     []string{"l"},
				Usage:       "Limit the number of replacements to be made on each matched file (replaces all matches if set to 0).\n\t\t\t\tCan be set to a negative integer to start replacing from the end of the file name.",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.StringFlag{
				Name: "sort",
				Usage: `Sort the matches in ascending order according to the provided '<sort>'.
					Allowed sort values:
						'default' : alphabetical order
						'size'    : sort by file size
						'mtime'   : sort by file last modified time
						'btime'   : sort by file creation time
						'atime'   : sort by file last access time
						'ctime'   : sort by file metadata last change time`,
				DefaultText: "<sort>",
			},
			&cli.StringFlag{
				Name:        "sortr",
				Usage:       "Same options as --sort but presents the matches in the descending order.",
				DefaultText: "<sort>",
			},
			&cli.BoolFlag{
				Name:    "string-mode",
				Aliases: []string{"s"},
				Usage:   "Treats the search pattern as a non-regex string.",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Enable verbose output.",
			},
		},
		UseShortOptionHandling: true,
		Action: func(c *cli.Context) error {
			// print short help if no arguments or flags are present
			if c.NumFlags() == 0 && !c.Args().Present() {
				pterm.Println(shortHelp(c.App))
				os.Exit(1)
			}

			if c.Bool("no-color") {
				disableStyling()
			}

			if c.Bool("quiet") {
				pterm.DisableOutput()
			}

			op, err := newOperation(c)
			if err != nil {
				return err
			}

			c.App.Metadata["op"] = op

			return op.run()
		},
		OnUsageError: func(context *cli.Context, err error, isSubcommand bool) error {
			return err
		},
	}
}
