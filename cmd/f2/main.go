package main

import (
	"context"
	"os"

	"github.com/ayoisaiah/f2/v2"
	"github.com/ayoisaiah/f2/v2/report"
)

func main() {
	renamer, err := f2.New(os.Stdin, os.Stdout)
	if err != nil {
		report.ExitWithErr(err)
	}

	err = renamer.Run(context.Background(), os.Args)
	if err != nil {
		report.ExitWithErr(err)
	}
}
