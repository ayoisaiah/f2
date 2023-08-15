package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2"
)

func main() {
	f, err1 := os.Create("f2.prof")
	if err1 != nil {
		log.Fatal(err1)
	}

	err1 = pprof.StartCPUProfile(f)
	if err1 != nil {
		log.Fatal(err1)
	}

	app := f2.GetApp(os.Stdin, os.Stdout)

	err := app.Run(os.Args)
	if err != nil {
		pterm.EnableOutput()
		pterm.Fprintln(os.Stderr, pterm.Error.Sprint(err))
		os.Exit(1)
	}

	pprof.StopCPUProfile()
}
