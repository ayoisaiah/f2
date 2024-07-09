package main

import (
	"os"

	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2"
	"github.com/ayoisaiah/f2/app"
)

func main() {
	app.InitLogger()

	f2App := f2.New(os.Stdin, os.Stdout)

	err := f2App.Run(os.Args)
	if err != nil {
		pterm.EnableOutput()
		pterm.Fprintln(os.Stderr, pterm.Error.Sprint(err))
		os.Exit(1)
	}
}
