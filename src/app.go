package f2

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func init() {
	// Override the default help template
	cli.AppHelpTemplate = helpText()

	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(c *cli.Context) {
		oldVersionPrinter(c)
		checkForUpdates(GetApp())
	}

	// Disable colour output if NO_COLOR is set
	if _, exists := os.LookupEnv("NO_COLOR"); exists {
		disableStyling()
	}

	// Disable colour output if F2_NO_COLOR is set
	if _, exists := os.LookupEnv("F2_NO_COLOR"); exists {
		disableStyling()
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

// checkForUpdates alerts the user if there is
// an updated version of F2 from the one currently installed.
func checkForUpdates(app *cli.App) {
	spinner, _ := pterm.DefaultSpinner.Start("Checking for updates...")
	c := http.Client{Timeout: 10 * time.Second}

	resp, err := c.Get("https://github.com/ayoisaiah/f2/releases/latest")
	if err != nil {
		pterm.Error.Println("HTTP Error: Failed to check for update")
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

// GetApp retrieves the f2 app instance.
func GetApp() *cli.App {
	return &cli.App{
		Name: "F2",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage:                "F2 is a command-line tool for batch renaming multiple files and directories quickly and safely.",
		UsageText:            "FLAGS [OPTIONS] [PATHS TO FILES OR DIRECTORIES...]",
		Version:              "v1.7.1",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
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
			&cli.StringFlag{
				Name:        "csv",
				Usage:       "Load a CSV file, and rename according to its contents.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Renaming-from-a-CSV-file.",
				DefaultText: "<csv file>",
			},
			&cli.IntFlag{
				Name:        "replace-limit",
				Aliases:     []string{"l"},
				Usage:       "Limit the number of replacements to be made on each matched file (replaces all matches if set to 0).\n\t\t\t\tCan be set to a negative integer to start replacing from the end of the file name.",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.BoolFlag{
				Name:    "string-mode",
				Aliases: []string{"s"},
				Usage:   "Treats the search pattern as a non-regex string.",
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
				Name:    "recursive",
				Aliases: []string{"R"},
				Usage:   "Recursively traverse directories when searching for matches.",
			},
			&cli.UintFlag{
				Name:        "max-depth",
				Aliases:     []string{"m"},
				Usage:       "Indicates the maximum depth for a recursive search (set to 0 by default for no limit).",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.StringFlag{
				Name: "sort",
				Usage: `Sort the matches according to the provided '<sort>'.
					Allowed sort values:
						'default': alphabetical order
						'size': file size
						'mtime': file last modified time
						'btime': file creation time (Windows and macOS only)
						'atime': file last access time
						'ctime': file metadata last change time`,
				DefaultText: "<sort>",
			},
			&cli.StringFlag{
				Name:        "sortr",
				Usage:       "Same as --sort but presents the matches in the reverse order.",
				DefaultText: "<sort>",
			},
			&cli.BoolFlag{
				Name:    "ignore-case",
				Aliases: []string{"i"},
				Usage:   "Search for matches case insensitively.",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Don't print out any information (except errors).",
			},
			&cli.BoolFlag{
				Name:    "ignore-ext",
				Aliases: []string{"e"},
				Usage:   "Ignore the file extension when searching for matches.",
			},
			&cli.BoolFlag{
				Name:    "include-dir",
				Aliases: []string{"d"},
				Usage:   "Include directories (they are exempted by default).",
			},
			&cli.BoolFlag{
				Name:    "only-dir",
				Aliases: []string{"D"},
				Usage:   "Rename only directories, not files (implies --include-dir).",
			},
			&cli.BoolFlag{
				Name:    "hidden",
				Aliases: []string{"H"},
				Usage:   "Include hidden files (they are skipped by default).",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Enable verbose output.",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Disable coloured output.",
			},
			&cli.BoolFlag{
				Name:    "fix-conflicts",
				Aliases: []string{"F"},
				Usage:   "Automatically fix conflicts based on predefined rules.\n\t\t\t\tLearn more: https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection.",
			},
			&cli.BoolFlag{
				Name:  "allow-overwrites",
				Usage: "Allow the overwriting of existing files.",
			},
		},
		UseShortOptionHandling: true,
		Action: func(c *cli.Context) error {
			if c.NumFlags() == 0 {
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

			return op.run()
		},
	}
}
