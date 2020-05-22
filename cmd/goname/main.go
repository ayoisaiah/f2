package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"gopkg.in/gookit/color.v1"
)

func main() {
	app := &cli.App{
		Name:  "goname",
		Usage: "Goname is a command-line utility for renaming files in bulk",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Search pattern",
			},
			&cli.StringFlag{
				Name:    "replace",
				Aliases: []string{"r"},
				Usage:   "Replacement string",
			},
			&cli.BoolFlag{
				Name:    "exec",
				Aliases: []string{"x"},
				Usage:   "Execute bulk rename operation",
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"F"},
				Usage:   "Force renaming operation even if there are conflicts",
			},
			&cli.StringFlag{
				Name:    "template",
				Aliases: []string{"t"},
				Usage:   "Rename using a template",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NumFlags() == 0 && c.NArg() == 0 {
				return fmt.Errorf("goname: not enough arguments\nTry 'goname --help' for more information.")
			}

			op, err := NewOperation(c)
			if err != nil {
				return err
			}

			op.FindMatches()
			if c.String("template") != "" {
				op.UseTemplate()
			} else {
				op.Replace()
			}

			return op.Apply()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		red := color.New(color.FgRed, color.OpBold).Render
		fmt.Fprintln(os.Stderr, red(err))
	}
}
