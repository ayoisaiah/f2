package main

import (
	"fmt"
	"os"

	"github.com/ayoisaiah/f2"
)

func run(args []string) error {
	return f2.GetApp().Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
