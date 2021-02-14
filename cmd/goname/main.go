package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func getApp() *cli.App {
	return &cli.App{
		Name: "goname",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage:     "Batch rename multiple files and directories. Hidden files and directories are skipped automatically.",
		UsageText: "[options] [files...]",
		Version:   "v0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Search `string` or regular expression. If omitted, the whole filename will be matched and replaced.",
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
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "By default, goname will do a 'dry run' so that you can inspect the results and confirm that it looks correct. Add this flag to proceed with renaming the files.",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"R"},
				Usage:   "Rename files recursively",
			},
			&cli.BoolFlag{
				Name:    "undo",
				Aliases: []string{"U"},
				Usage:   "Undo the LAST successful operation",
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
				Aliases: []string{"D"},
				Usage:   "Rename directories",
			},
			&cli.BoolFlag{
				Name:    "hidden",
				Aliases: []string{"H"},
				Usage:   "Include hidden files and directories",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"F"},
				Usage:   "If there are conflicts after a replacement operation (such as when overwriting existing files), it will be reported to you. Use this flag to force the renaming operation even if there are conflicts.",
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("undo") {
				op := &Operation{}
				op.ignoreConflicts = c.Bool("force")
				op.exec = c.Bool("exec")
				return op.Undo()
			}

			op, err := NewOperation(c)
			if err != nil {
				return err
			}

			err = op.FindMatches()
			if err != nil {
				return err
			}

			if op.includeDir {
				op.SortMatches()
			}
			if err := op.Replace(); err != nil {
				return err
			}

			return op.Apply()
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
	 {{.}}
	 {{end}}{{end}}
WEBSITE:
	https://github.com/ayoisaiah/goname
`

	return app.Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
