package main

import (
	"os"

	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2"
)

func main() {
	app := f2.GetApp(os.Stdin, os.Stdout)

	err := app.Run(os.Args)
	if err != nil {
		pterm.EnableOutput()
		pterm.Fprintln(os.Stderr, pterm.Error.Sprint(err))
		os.Exit(1)
	}
}
