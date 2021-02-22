package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func checkForUpdates(app *cli.App) {
	fmt.Println("Checking for updates...")

	c := http.Client{Timeout: 20 * time.Second}
	resp, err := c.Get("https://github.com/ayoisaiah/f2/releases/latest")
	if err != nil {
		fmt.Println("HTTP Error: Failed to check for update")
		return
	}
	var version string
	_, err = fmt.Sscanf(resp.Request.URL.String(), "https://github.com/ayoisaiah/f2/releases/tag/%s", &version)
	if err != nil {
		fmt.Println("Failed to get latest version")
		return
	}
	version = "v" + version

	if version == app.Version {
		fmt.Printf("Congratulations, you are using the latest version of %s\n", app.Name)
	} else {
		fmt.Printf("%s: %s at %s\n", green("Update available"), version, resp.Request.URL.String())
	}
}

func getApp() *cli.App {
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
		Version:              "v1.0.0",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Search `string` or regular expression.",
			},
			&cli.StringFlag{
				Name:    "replace",
				Aliases: []string{"r"},
				Usage:   "Replacement `string`. If omitted, defaults to an empty string.",
			},
			&cli.IntFlag{
				Name:        "start-num",
				Aliases:     []string{"n"},
				Usage:       "Starting number when using numbering scheme in replacement string such as %03d",
				Value:       1,
				DefaultText: "1",
			},
			&cli.StringFlag{
				Name:    "output-file",
				Aliases: []string{"o"},
				Usage:   "Output a map file for the current operation",
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
			&cli.StringFlag{
				Name:    "undo",
				Aliases: []string{"u"},
				Usage:   "Undo a successful operation using a previously created map file",
			},
			&cli.BoolFlag{
				Name:    "ignore-case",
				Aliases: []string{"i"},
				Usage:   "Ignore case",
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
				Name:    "force",
				Aliases: []string{"F"},
				Usage:   "Force the renaming operation even when there are conflicts (may cause data loss).",
			},
		},
		Action: func(c *cli.Context) error {
			op, err := NewOperation(c)
			if err != nil {
				return err
			}

			return op.Run()
		},
	}
}

func run(args []string) error {
	app := getApp()

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

WEBSITE:
	https://github.com/ayoisaiah/f2
`

	return app.Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if contains(os.Args, "-v") || contains(os.Args, "--version") {
		checkForUpdates(getApp())
	}
}
