package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ayoisaiah/f2/v2"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/report"
)

func init() {
	_, exists := os.LookupEnv(config.EnvDebug)
	if exists {
		slog.SetDefault(
			slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
		)
	}
}

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
