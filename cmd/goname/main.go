package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
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
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "By default, goname will do a 'dry run' so that you can inspect the results and confirm that it looks correct. Add this flag to proceed with renaming the files.",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"F"},
				Usage:   "If there are conflicts after a replacement operation (such as when overwriting existing files), goname will report them to you. Use this flag to force the renaming operation even if there are conflicts.",
			},
		},
		Action: func(c *cli.Context) error {
			op, err := NewOperation(c)
			if err != nil {
				return err
			}

			op.FindMatches()
			if err := op.Replace(); err != nil {
				return err
			}

			return op.Apply()
		},
	}

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
{{if .Commands}}
FLAGS:{{range .VisibleFlags}}
	 {{.}}
	 {{end}}{{end}}
WEBSITE:
	https://github.com/ayoisaiah/goname
`

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
