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
FLAGS:{{range .VisibleFlags}}
	 {{.}}{{end}}{{end}}

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
		fmt.Printf("%s: %s at %s\n", green.Sprint("Update available"), version, resp.Request.URL.String())
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
		Version:              "v1.5.3",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Search `<pattern>`. Treated as a regular expression by default. Use -s or --string-mode to opt out",
			},
			&cli.StringFlag{
				Name:    "replace",
				Aliases: []string{"r"},
				Usage:   "Replacement `<string>`. If omitted, defaults to an empty string. Supports built-in and regex capture variables",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"E"},
				Usage:   "Exclude files/directories that match the given find pattern. Treated as a regular expression. Multiple exclude `<pattern>`s can be specified.",
			},
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "Execute the batch renaming operation",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"R"},
				Usage:   "Rename files recursively",
			},
			&cli.IntFlag{
				Name:        "max-depth",
				Aliases:     []string{"m"},
				Usage:       "positive `<integer>` indicating the maximum depth for a recursive search (set to 0 for no limit)",
				Value:       0,
				DefaultText: "0",
			},
			&cli.BoolFlag{
				Name:    "undo",
				Aliases: []string{"u"},
				Usage:   "Undo the last operation performed in the current working directory.",
			},
			&cli.StringFlag{
				Name:  "sort",
				Usage: "Sort the matches according to the provided `<sort>` (possible values: default, size, mtime, btime, atime, ctime)",
			},
			&cli.StringFlag{
				Name:  "sortr",
				Usage: "Same as `<sort>` but presents the matches in the reverse order (possible values: default, size, mtime, btime, atime, ctime)",
			},
			&cli.BoolFlag{
				Name:    "ignore-case",
				Aliases: []string{"i"},
				Usage:   "Ignore case",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Don't print out any information including errors",
			},
			&cli.BoolFlag{
				Name:    "ignore-ext",
				Aliases: []string{"e"},
				Usage:   "Ignore extension",
			},
			&cli.BoolFlag{
				Name:    "include-dir",
				Aliases: []string{"d"},
				Usage:   "Include directories",
			},
			&cli.BoolFlag{
				Name:    "only-dir",
				Aliases: []string{"D"},
				Usage:   "Rename only directories (implies include-dir)",
			},
			&cli.BoolFlag{
				Name:    "hidden",
				Aliases: []string{"H"},
				Usage:   "Include hidden files and directories",
			},
			&cli.BoolFlag{
				Name:    "fix-conflicts",
				Aliases: []string{"F"},
				Usage:   "Fix any detected conflicts with auto indexing",
			},
			&cli.BoolFlag{
				Name:    "string-mode",
				Aliases: []string{"s"},
				Usage:   "Opt into string literal mode by treating find expressions as non-regex strings",
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
