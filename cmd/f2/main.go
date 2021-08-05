package main

import (
	"os"

	"github.com/pterm/pterm"

	f2 "github.com/ayoisaiah/f2/src"
)

func run(args []string) error {
	return f2.GetApp().Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		pterm.EnableOutput()
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
