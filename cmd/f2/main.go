package main

import (
	"os"

	"github.com/ayoisaiah/f2"
)

func run(args []string) error {
	return f2.GetApp().Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		os.Exit(1)
	}
}
