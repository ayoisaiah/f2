package app

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func helpText() string {
	description := fmt.Sprintf(
		"%s\n\t\t{{.Usage}}\n\n",
		pterm.Yellow("DESCRIPTION"),
	)
	usage := fmt.Sprintf(
		"%s\n\t\t{{.HelpName}} {{if .UsageText}}{{ .UsageText }}{{end}}\n\n",
		pterm.Yellow("USAGE"),
	)
	author := fmt.Sprintf(
		"{{if len .Authors}}%s\n\t\t{{range .Authors}}{{ . }}{{end}}{{end}}\n\n",
		pterm.Yellow("AUTHOR"),
	)

	version := fmt.Sprintf(
		"{{if .Version}}%s\n\t\t{{.Version}}{{end}}\n\n",
		pterm.Yellow("VERSION"),
	)
	flags := fmt.Sprintf(
		"{{if .VisibleFlags}}%s\n{{range .VisibleFlags}}{{ if (eq .Name `find` `undo` `replace` `csv`) }}\t\t{{if .Aliases}}-{{range $element := .Aliases}}%s,{{end}}{{end}} %s\n\t\t\t\t{{.Usage}}\n\n{{end}}{{end}}",
		pterm.Yellow("FLAGS"),
		pterm.Green("{{$element}}"),
		pterm.Green("--{{.Name}} {{.DefaultText}}"),
	)
	options := fmt.Sprintf(
		"%s\n{{range .VisibleFlags}}{{ if not (eq .Name `find` `undo` `replace` `csv`) }}\t\t{{if .Aliases}}-{{range $element := .Aliases}}%s,{{end}}{{end}} %s\n\t\t\t\t{{.Usage}}\n\n{{end}}{{end}}{{end}}",
		pterm.Yellow("OPTIONS"),
		pterm.Green("{{$element}}"),
		pterm.Green("--{{.Name}} {{.DefaultText}}"),
	)

	env := fmt.Sprintf(
		"%s\n\t\t%s\n\n",
		pterm.Yellow("ENVIRONMENTAL VARIABLES"),
		envHelp(),
	)

	docs := fmt.Sprintf(
		"%s\n\t\t%s\n\n",
		pterm.Yellow("DOCUMENTATION"),
		"https://github.com/ayoisaiah/f2/wiki",
	)
	website := fmt.Sprintf(
		"%s\n\t\thttps://github.com/ayoisaiah/f2",
		pterm.Yellow("WEBSITE"),
	)

	return description + usage + author + version + flags + options + env + docs + website + "\n"
}

func envHelp() string {
	return `
  F2_DEFAULT_OPTS: override the default options according to your preferences. 
      For example, you can enable execute mode and ignore file extensions by default:
      'export F2_DEFAULT_OPTS=--exec --ignore-ext'.

  F2_NO_COLOR, NO_COLOR: set to any value to disable coloured output.

  F2_UPDATE_NOTIFIER: set to any value to periodically check for updates.`
}

func ShortHelp(app *cli.App) string {
	heading := fmt.Sprintf(
		"F2 — Command-line bulk renaming tool [version %s]\n\n",
		app.Version,
	)

	usage := fmt.Sprintf("Usage: %s\n", app.UsageText)

	description := `
F2 helps you organise your filesystem through batch renaming.
The simplest usage is to do a basic find and replace:

$ f2 -f Screenshot -r Image
+--------------------+---------------+--------+
|       INPUT        |    OUTPUT     | STATUS |
+--------------------+---------------+--------+
| Screenshot (1).png | Image (1).png | ok     |
| Screenshot (2).png | Image (2).png | ok     |
| Screenshot (3).png | Image (3).png | ok     |
+--------------------+---------------+--------+

The argument to -f is the find string, while the one to -r is the
replacement string. The current directory is used by default, but
you can pass relative or absolute paths to other files and
directories.

F2 supports many command-line options. Use the --help flag to examine
the full list. For extensive usage examples, visit the project wiki:
https://github.com/ayoisaiah/f2/wiki`

	return heading + usage + description
}
