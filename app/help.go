package app

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

const usageText = `f2 FLAGS [OPTIONS] [PATHS TO FILES AND DIRECTORIES...]
  command | f2 FLAGS [OPTIONS]
  command | f2 FIND [REPLACE]`

func helpText(app *cli.App) string {
	flagCSVHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagCSV.Name),
		flagCSV.GetUsage(),
	)

	flagFindHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagFind.Aliases[0]),
		pterm.Green("--", flagFind.Name),
		flagFind.GetUsage(),
	)

	flagReplaceHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagReplace.Aliases[0]),
		pterm.Green("--", flagReplace.Name),
		flagReplace.GetUsage(),
	)

	flagUndoHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagUndo.Aliases[0]),
		pterm.Green("--", flagUndo.Name),
		flagUndo.GetUsage(),
	)

	flagAllowOverwritesHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagAllowOverwrites.Name),
		flagAllowOverwrites.GetUsage(),
	)

	flagDebugHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagDebug.Name),
		flagDebug.GetUsage(),
	)

	flagExcludeHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagExclude.Aliases[0]),
		pterm.Green("--", flagExclude.Name),
		flagExclude.GetUsage(),
	)

	flagExcludeDirHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagExcludeDir.Name),
		flagExcludeDir.GetUsage(),
	)

	flagExiftoolOptsHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagExiftoolOpts.Name),
		flagExiftoolOpts.GetUsage(),
	)

	flagExecHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagExec.Aliases[0]),
		pterm.Green("--", flagExec.Name),
		flagExec.GetUsage(),
	)

	flagFixConflictsHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagFixConflicts.Aliases[0]),
		pterm.Green("--", flagFixConflicts.Name),
		flagFixConflicts.GetUsage(),
	)

	flagFixConflictsPatternHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagFixConflictsPattern.Name),
		flagFixConflictsPattern.GetUsage(),
	)

	flagHiddenHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagHidden.Aliases[0]),
		pterm.Green("--", flagHidden.Name),
		flagHidden.GetUsage(),
	)

	flagIncludeDirHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagIncludeDir.Aliases[0]),
		pterm.Green("--", flagIncludeDir.Name),
		flagIncludeDir.GetUsage(),
	)

	flagIgnoreCaseHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagIgnoreCase.Aliases[0]),
		pterm.Green("--", flagIgnoreCase.Name),
		flagIgnoreCase.GetUsage(),
	)

	flagIgnoreExtHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagIgnoreExt.Aliases[0]),
		pterm.Green("--", flagIgnoreExt.Name),
		flagIgnoreExt.GetUsage(),
	)

	flagJSONHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagJSON.Name),
		flagJSON.GetUsage(),
	)

	flagMaxDepthHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagMaxDepth.Aliases[0]),
		pterm.Green("--", flagMaxDepth.Name),
		flagMaxDepth.GetUsage(),
	)

	flagNoColorHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagNoColor.Name),
		flagNoColor.GetUsage(),
	)

	flagOnlyDirHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagOnlyDir.Aliases[0]),
		pterm.Green("--", flagOnlyDir.Name),
		flagOnlyDir.GetUsage(),
	)

	flagQuietHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagQuiet.Name),
		flagQuiet.GetUsage(),
	)

	flagRecursiveHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagRecursive.Aliases[0]),
		pterm.Green("--", flagRecursive.Name),
		flagRecursive.GetUsage(),
	)

	flagReplaceLimitHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagReplaceLimit.Aliases[0]),
		pterm.Green("--", flagReplaceLimit.Name),
		flagReplaceLimit.GetUsage(),
	)

	flagResetIndexPerDirHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagResetIndexPerDir.Name),
		flagResetIndexPerDir.GetUsage(),
	)

	flagSortHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagSort.Name),
		flagSort.GetUsage(),
	)

	flagSortrHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagSortr.Name),
		flagSortr.GetUsage(),
	)

	flagSortPerDirHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagSortPerDir.Name),
		flagSortPerDir.GetUsage(),
	)

	flagStringModeHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagStringMode.Aliases[0]),
		pterm.Green("--", flagStringMode.Name),
		flagStringMode.GetUsage(),
	)

	flagVerboseHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagVerbose.Aliases[0]),
		pterm.Green("--", flagVerbose.Name),
		flagVerbose.GetUsage(),
	)

	return fmt.Sprintf(`%s %s
%s

%s

Project repository: https://github.com/ayoisaiah/f2

%s
  %s

%s
  %s
    A regular expression pattern used for matching files and directories.
    It accepts the syntax defined by the RE2 standard.

  %s
    The replacement string which replaces each match in the file name.
    It supports capture variables, built-in variables, and exiftool variables.
    If omitted, it defaults to an empty string.

  %s
    Optionally provide one or more files and directories to search for matches. 
		If omitted, it searches the current directory alone. Also, note that 
		directories are not searched recursively.

%s
  %s

  %s

	%s

	%s

%s
	%s
	
	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

	%s

%s
	%s

%s
  Read the manual at https://github.com/ayoisaiah/f2/wiki`,
		app.Name,
		app.Version,
		app.Authors[0].String(),
		app.Usage,
		pterm.Bold.Sprintf("USAGE"),
		usageText,
		pterm.Bold.Sprintf("POSITIONAL ARGUMENTS"),
		pterm.Green("<FIND>"),
		pterm.Green("[REPLACE]"),
		pterm.Green("[PATHS]"),
		pterm.Bold.Sprintf("FLAGS"),
		flagCSVHelp,
		flagFindHelp,
		flagReplaceHelp,
		flagUndoHelp,
		pterm.Bold.Sprintf("OPTIONS"),
		flagAllowOverwritesHelp,
		flagDebugHelp,
		flagExcludeHelp,
		flagExcludeDirHelp,
		flagExiftoolOptsHelp,
		flagExecHelp,
		flagFixConflictsHelp,
		flagFixConflictsPatternHelp,
		flagHiddenHelp,
		flagIncludeDirHelp,
		flagIgnoreCaseHelp,
		flagIgnoreExtHelp,
		flagJSONHelp,
		flagMaxDepthHelp,
		flagNoColorHelp,
		flagOnlyDirHelp,
		flagQuietHelp,
		flagRecursiveHelp,
		flagReplaceLimitHelp,
		flagResetIndexPerDirHelp,
		flagSortHelp,
		flagSortrHelp,
		flagSortPerDirHelp,
		flagStringModeHelp,
		flagVerboseHelp,
		pterm.Bold.Sprintf("ENVIRONMENTAL VARIABLES"),
		envHelp(),
		pterm.Bold.Sprintf("LEARN MORE"),
	)
}

func envHelp() string {
	return fmt.Sprintf(`%s
		Override the default options according to your preferences. For example, 
		you can enable execute mode and ignore file extensions by default:

		export F2_DEFAULT_OPTS=--exec --ignore-ext

	%s, %s
		Set to any value to disable coloured output.

	%s
		Enable debug mode.`,
		pterm.Green("F2_DEFAULT_OPTS"),
		pterm.Green("F2_NO_COLOR"),
		pterm.Green("NO_COLOR"),
		pterm.Green("F2_DEBUG"),
	)
}

func ShortHelp(app *cli.App) string {
	return fmt.Sprintf(
		`The batch renaming tool you'll actually enjoy using.

%s
  %s

%s
  $ f2 -f 'jpeg' -r 'jpg'
  $ f2 js ts
  $ f2 -r '{id3.artist}/{id3.album}/${1}_{id3.title}{ext}'

%s
  Use f2 --help to view the command-line options.
  Read the manual at https://github.com/ayoisaiah/f2/wiki`,
		pterm.Bold.Sprintf("USAGE"),
		usageText,
		pterm.Bold.Sprintf("EXAMPLES"),
		pterm.Bold.Sprintf("LEARN MORE"),
	)
}
