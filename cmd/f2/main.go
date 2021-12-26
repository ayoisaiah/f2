package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	f2 "github.com/ayoisaiah/f2/src"
)

// supportedDefaultFlags are those that can be overridden through the
// F2_DEFAULT_OPTS environmental variable.
var supportedDefaultFlags = []string{
	"hidden", "allow-overwrites", "exclude", "exec", "fix-conflicts", "include-dir", "ignore-case", "ignore-ext", "max-depth", "no-color", "only-dir", "quiet", "recursive", "replace-limit", "sort", "sortr", "string-mode", "verbose",
}

func setDefaultOpts() *cli.Context {
	var defaultCtx *cli.Context

	if optsEnv, exists := os.LookupEnv("F2_DEFAULT_OPTS"); exists {
		var defaultOpts = make([]string, len(os.Args))

		copy(defaultOpts, os.Args)

		defaultOpts = append(defaultOpts[:1], strings.Split(optsEnv, " ")...)

		app := f2.GetApp()

		app.Before = func(c *cli.Context) error {
			if c.IsSet("find") || c.IsSet("replace") || c.IsSet("csv") ||
				c.IsSet("undo") {
				pterm.Warning.Printfln(
					"%s are not supported as default options",
					"'find', 'replace', 'csv' and 'undo'",
				)
			}

			defaultCtx = c

			return nil
		}

		app.Action = func(c *cli.Context) error {
			return nil
		}

		_ = app.Run(defaultOpts)
	}

	return defaultCtx
}

func run(args []string) error {
	app := f2.GetApp()

	defaultCtx := setDefaultOpts()

	if defaultCtx == nil {
		return app.Run(args)
	}

	app.Before = func(c *cli.Context) error {
		if c.NumFlags() == 0 {
			app.Metadata["simple-mode"] = true
		}

		for _, v := range supportedDefaultFlags {
			value := fmt.Sprintf("%v", defaultCtx.Value(v))

			if !c.IsSet(v) && defaultCtx.IsSet(v) {
				if x, ok := defaultCtx.Value(v).(cli.StringSlice); ok {
					value = strings.Join(x.Value(), "|")
				}

				err := c.Set(v, value)
				if err != nil {
					pterm.Warning.Printfln(
						"Unable to set default option for: %s",
						v,
					)
				}
			}
		}

		return nil
	}

	return app.Run(args)
}

func main() {
	err := run(os.Args)
	if err != nil {
		pterm.EnableOutput()
		pterm.Error.Println(err)
		os.Exit(1)
	}
}
