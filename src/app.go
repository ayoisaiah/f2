package f2

import (
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
)

func init() {
	// Override the default help template
	cli.AppHelpTemplate = `DESCRIPTION:
	{{.Usage}}

USAGE:
   {{.HelpName}} {{if .UsageText}}{{ .UsageText }}{{end}}
{{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}{{end}}
{{if .Version}}
VERSION:
	 {{.Version}}{{end}}
{{if .VisibleFlags}}
FLAGS:{{range .VisibleFlags}}{{ if (eq .Name "find" "undo" "replace") }}
		 {{if .Aliases}}-{{range $element := .Aliases}}{{$element}},{{end}}{{end}} --{{.Name}} {{.DefaultText}}
				 {{.Usage}}
		 {{end}}{{end}}
OPTIONS:{{range .VisibleFlags}}{{ if not (eq .Name "find" "replace" "undo") }}
		 {{if .Aliases}}-{{range $element := .Aliases}}{{$element}},{{end}}{{end}} --{{.Name}} {{ .DefaultText }}
				 {{.Usage}}
		 {{end}}{{end}}{{end}}
DOCUMENTATION:
	https://github.com/ayoisaiah/f2/wiki

WEBSITE:
	https://github.com/ayoisaiah/f2
`

	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(c *cli.Context) {
		oldVersionPrinter(c)
		checkForUpdates(GetApp())
	}
}

func checkForUpdates(app *cli.App) {
	fmt.Println("Checking for updates...")

	c := http.Client{Timeout: 20 * time.Second}
	resp, err := c.Get("https://github.com/ayoisaiah/f2/releases/latest")
	if err != nil {
		fmt.Println("HTTP Error: Failed to check for update")
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
		fmt.Println("Failed to get latest version")
		return
	}

	if version == app.Version {
		fmt.Printf(
			"Congratulations, you are using the latest version of %s\n",
			app.Name,
		)
	} else {
		fmt.Printf("%s: %s at %s\n", printColor("green", "Update available"), version, resp.Request.URL.String())
	}
}

// GetApp retrieves the f2 app instance
func GetApp() *cli.App {
	return &cli.App{
		Name: "F2",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage:                "F2 is a command-line tool for batch renaming multiple files and directories quickly and safely",
		UsageText:            "FLAGS [OPTIONS] [PATHS...]",
		Version:              "v1.6.2",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "find",
				Aliases:     []string{"f"},
				Usage:       "Search pattern. Treated as a regular expression by default unless --string-mode is also used. If omitted, it defaults to the entire file name (including the extension).",
				DefaultText: "<pattern>",
			},
			&cli.StringFlag{
				Name:        "replace",
				Aliases:     []string{"r"},
				Usage:       "Replacement string. If omitted, defaults to an empty string. Supports built-in and regex capture variables. Learn more about variable support here: https://github.com/ayoisaiah/f2/wiki/Built-in-variables",
				DefaultText: "<string>",
			},
			&cli.IntFlag{
				Name:        "replace-limit",
				Aliases:     []string{"l"},
				Usage:       "Limit the number of replacements to be made on the file name (replaces all matches if set to 0). Can be set to a negative integer to start replacing from the end of the file name.",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.BoolFlag{
				Name:    "string-mode",
				Aliases: []string{"s"},
				Usage:   "Opt into string literal mode. The presence of this flag causes the search pattern to be treated as a non-regex string.",
			},
			&cli.StringSliceFlag{
				Name:        "exclude",
				Aliases:     []string{"E"},
				Usage:       "Exclude files/directories that match the given search pattern. Treated as a regular expression. Multiple exclude patterns can be specified.",
				DefaultText: "<pattern>",
			},
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "Execute the batch renaming operation. This will commit the changes to your filesystem.",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"R"},
				Usage:   "Recursively traverse all directories when searching for matches. Use the --max-depth flag to control the maximum allowed depth (no limit by default).",
			},
			&cli.UintFlag{
				Name:        "max-depth",
				Aliases:     []string{"m"},
				Usage:       "Positive integer indicating the maximum depth for a recursive search (set to 0 for no limit).",
				Value:       0,
				DefaultText: "<integer>",
			},
			&cli.BoolFlag{
				Name:    "undo",
				Aliases: []string{"u"},
				Usage:   "Undo the last operation performed in the current working directory if possible. Learn more: https://github.com/ayoisaiah/f2/wiki/Undoing-a-renaming-operation",
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
				Usage:   "When this flag is provided, the given pattern will be searched case insensitively.",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Activate silent mode which doesn't print out any information including errors",
			},
			&cli.BoolFlag{
				Name:    "ignore-ext",
				Aliases: []string{"e"},
				Usage:   "Ignore the file extension when searching for matches.",
			},
			&cli.BoolFlag{
				Name:    "include-dir",
				Aliases: []string{"d"},
				Usage:   "Include directories when searching for matches as they are exempted by default.",
			},
			&cli.BoolFlag{
				Name:    "only-dir",
				Aliases: []string{"D"},
				Usage:   "Rename only directories, not files (implies --include-dir)",
			},
			&cli.BoolFlag{
				Name:    "hidden",
				Aliases: []string{"H"},
				Usage:   "Include hidden directories and files in the matches (they are skipped by default). A hidden file or directory is one whose name starts with a period (all operating systems) or one whose hidden attribute is set to true (Windows only)",
			},
			&cli.BoolFlag{
				Name:    "fix-conflicts",
				Aliases: []string{"F"},
				Usage:   "Automatically fix conflicts based on predefined rules. Learn more: https://github.com/ayoisaiah/f2/wiki/Validation-and-conflict-detection",
			},
		},
		UseShortOptionHandling: true,
		Action: func(c *cli.Context) error {
			op, err := newOperation(c)
			if err != nil {
				printError(false, err)
				return err
			}

			err = op.run()
			if err != nil {
				printError(op.quiet, err)
			}

			return err
		},
	}
}
