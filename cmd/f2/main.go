package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"github.com/ayoisaiah/f2/v2"
	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/report"
)

func init() {
	_, exists := os.LookupEnv(config.EnvDebug)
	if exists {
		slog.SetDefault(
			slog.New(tint.NewHandler(os.Stderr, &tint.Options{
				Level: slog.LevelDebug,
				ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
						return slog.Attr{}
					}

					return a
				},
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
