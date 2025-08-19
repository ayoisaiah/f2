package app

import (
	"github.com/urfave/cli/v3"

	"github.com/ayoisaiah/f2/v2/internal/localize"
)

// supportedDefaultOpts contains flags whose values can be
// overridden through the `F2_DEFAULT_OPTS` environmental variable.
var supportedDefaultOpts = map[string]bool{
	"--clean": true, "-c": true,
	"--exclude": true, "-E": true,
	"--exclude-dir": true,
	"--exec":        true, "-x": true,
	"--exiftool-opts": true,
	"--fix-conflicts": true, "-F": true,
	"--fix-conflicts-pattern": true,
	"--hidden":                true, "-H": true,
	"--ignore-case": true, "-i": true,
	"--ignore-ext": true, "-e": true,
	"--include-dir": true, "-d": true,
	"--json":     true,
	"--no-color": true,
	"--quiet":    true, "-q": true,
	"--recursive": true, "-R": true,
	"--sort":                true,
	"--sortr":               true,
	"--reset-index-per-dir": true,
	"--string-mode":         true, "-s": true,
	"--verbose": true, "-V": true,
}

var (
	flagCSV = &cli.StringFlag{
		Name:        "csv",
		Usage:       localize.T("flag.csv.usage"),
		DefaultText: "<path/to/csv/file>",
		TakesFile:   true,
	}

	flagFind = &cli.StringSliceFlag{
		Name:        "find",
		Aliases:     []string{"f"},
		Usage:       localize.T("flag.find.usage"),
		DefaultText: "<pattern>",
	}

	flagReplace = &cli.StringSliceFlag{
		Name:        "replace",
		Aliases:     []string{"r"},
		Usage:       localize.T("flag.replace.usage"),
		DefaultText: "<string>",
	}

	flagUndo = &cli.BoolFlag{
		Name:    "undo",
		Aliases: []string{"u"},
		Usage:   localize.T("flag.undo.usage"),
	}

	flagAllowOverwrites = &cli.BoolFlag{
		Name:  "allow-overwrites",
		Usage: localize.T("flag.allowOverwrites.usage"),
	}

	flagClean = &cli.BoolFlag{
		Name:    "clean",
		Aliases: []string{"c"},
		Usage:   localize.T("flag.clean.usage"),
	}

	flagExclude = &cli.StringSliceFlag{
		Name:        "exclude",
		Aliases:     []string{"E"},
		Usage:       localize.T("flag.exclude.usage"),
		DefaultText: "<pattern>",
	}

	flagExcludeDir = &cli.StringSliceFlag{
		Name:        "exclude-dir",
		Usage:       localize.T("flag.excludeDir.usage"),
		DefaultText: "<pattern>",
	}

	flagExec = &cli.BoolFlag{
		Name:    "exec",
		Aliases: []string{"x"},
		Usage:   localize.T("flag.exec.usage"),
	}

	flagExiftoolOpts = &cli.StringFlag{
		Name:  "exiftool-opts",
		Usage: localize.T("flag.exiftoolOpts.usage"),
	}

	flagFixConflicts = &cli.BoolFlag{
		Name:    "fix-conflicts",
		Aliases: []string{"F"},
		Usage:   localize.T("flag.fixConflicts.usage"),
	}

	flagFixConflictsPattern = &cli.StringFlag{
		Name:  "fix-conflicts-pattern",
		Usage: localize.T("flag.fixConflictsPattern.usage"),
	}

	flagHidden = &cli.BoolFlag{
		Name:    "hidden",
		Aliases: []string{"H"},
		Usage:   localize.T("flag.hidden.usage"),
	}

	flagInclude = &cli.StringSliceFlag{
		Name:    "include",
		Aliases: []string{"I"},
		Usage:   localize.T("flag.include.usage"),
	}

	flagIncludeDir = &cli.BoolFlag{
		Name:    "include-dir",
		Aliases: []string{"d"},
		Usage:   localize.T("flag.includeDir.usage"),
	}

	flagIgnoreCase = &cli.BoolFlag{
		Name:    "ignore-case",
		Aliases: []string{"i"},
		Usage:   localize.T("flag.ignoreCase.usage"),
	}

	flagIgnoreExt = &cli.BoolFlag{
		Name:    "ignore-ext",
		Aliases: []string{"e"},
		Usage:   localize.T("flag.ignoreExt.usage"),
	}

	flagJSON = &cli.BoolFlag{
		Name:  "json",
		Usage: localize.T("flag.json.usage"),
	}

	flagMaxDepth = &cli.UintFlag{
		Name:        "max-depth",
		Aliases:     []string{"m"},
		Usage:       localize.T("flag.maxDepth.usage"),
		Value:       0,
		DefaultText: "<integer>",
	}

	flagNoColor = &cli.BoolFlag{
		Name:  "no-color",
		Usage: localize.T("flag.noColor.usage"),
	}

	flagOnlyDir = &cli.BoolFlag{
		Name:    "only-dir",
		Aliases: []string{"D"},
		Usage:   localize.T("flag.onlyDir.usage"),
	}

	flagPair = &cli.BoolFlag{
		Name:    "pair",
		Aliases: []string{"p"},
		Usage:   localize.T("flag.pair.usage"),
	}

	flagPairOrder = &cli.StringFlag{
		Name:  "pair-order",
		Usage: localize.T("flag.pairOrder.usage"),
	}

	flagQuiet = &cli.BoolFlag{
		Name:    "quiet",
		Aliases: []string{"q"},
		Usage:   localize.T("flag.quiet.usage"),
	}

	flagRecursive = &cli.BoolFlag{
		Name:    "recursive",
		Aliases: []string{"R"},
		Usage:   localize.T("flag.recursive.usage"),
	}

	flagReplaceLimit = &cli.IntFlag{
		Name:        "replace-limit",
		Aliases:     []string{"l"},
		Usage:       localize.T("flag.replaceLimit.usage"),
		Value:       0,
		DefaultText: "<integer>",
	}

	flagReplaceRange = &cli.StringFlag{
		Name:    "replace-range",
		Aliases: []string{"L"},
		Usage:   localize.T("flag.replaceRange.usage"),
	}

	flagResetIndexPerDir = &cli.BoolFlag{
		Name:  "reset-index-per-dir",
		Usage: localize.T("flag.resetIndexPerDir.usage"),
	}

	flagSort = &cli.StringFlag{
		Name:        "sort",
		Usage:       localize.T("flag.sort.usage"),
		DefaultText: "<sort>",
	}

	flagSortr = &cli.StringFlag{
		Name:        "sortr",
		Usage:       localize.T("flag.sortr.usage"),
		DefaultText: "<sort>",
	}

	flagSortPerDir = &cli.BoolFlag{
		Name:  "sort-per-dir",
		Usage: localize.T("flag.sortPerDir.usage"),
	}

	flagSortVar = &cli.StringFlag{
		Name:  "sort-var",
		Usage: localize.T("flag.sortVar.usage"),
	}

	flagStringMode = &cli.BoolFlag{
		Name:    "string-mode",
		Aliases: []string{"s"},
		Usage:   localize.T("flag.stringMode.usage"),
	}

	flagTargetDir = &cli.StringFlag{
		Name:    "target-dir",
		Aliases: []string{"t"},
		Usage:   localize.T("flag.targetDir.usage"),
	}

	flagVerbose = &cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"V"},
		Usage:   localize.T("flag.verbose.usage"),
	}
)
