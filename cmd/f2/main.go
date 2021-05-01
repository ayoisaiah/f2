package main

import (
	"os"

	f2 "github.com/ayoisaiah/f2/src"
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
