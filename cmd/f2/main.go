package main

import (
	"log/slog"
	"os"

	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2"

	slogctx "github.com/veqryn/slog-context"
)

func initLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	h := slogctx.NewHandler(slog.NewJSONHandler(os.Stderr, opts), nil)

	l := slog.New(h)

	slog.SetDefault(l)
}

func main() {
	initLogger()

	app := f2.New(os.Stdin, os.Stdout)

	err := app.Run(os.Args)
	if err != nil {
		pterm.EnableOutput()
		pterm.Fprintln(os.Stderr, pterm.Error.Sprint(err))
		os.Exit(1)
	}
}
