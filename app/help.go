package app

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

const usageText = `f2 FLAGS [OPTIONS] [PATHS TO FILES AND DIRECTORIES...]
  command | f2 FLAGS [OPTIONS]`

func helpText(app *cli.Command) string {
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

	flagCleanHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagClean.Aliases[0]),
		pterm.Green("--", flagClean.Name),
		flagClean.GetUsage(),
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

	flagIncludeHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagInclude.Aliases[0]),
		pterm.Green("--", flagInclude.Name),
		flagInclude.GetUsage(),
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

	flagPairHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagPair.Aliases[0]),
		pterm.Green("--", flagPair.Name),
		flagPair.GetUsage(),
	)

	flagPairOrderHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagPairOrder.Name),
		flagPairOrder.GetUsage(),
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

	flagSortVarHelp := fmt.Sprintf(
		`%s %s`,
		pterm.Green("--", flagSortVar.Name),
		flagSortVar.GetUsage(),
	)

	flagStringModeHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagStringMode.Aliases[0]),
		pterm.Green("--", flagStringMode.Name),
		flagStringMode.GetUsage(),
	)

	flagTargetDirHelp := fmt.Sprintf(
		`%s, %s %s`,
		pterm.Green("-", flagTargetDir.Aliases[0]),
		pterm.Green("--", flagTargetDir.Name),
		flagTargetDir.GetUsage(),
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
    Optionally provide one or more files and directories to search for matches. 
		If omitted, it searches the current directory alone. Also, note that 
		directories are not searched recursively unless --recursive/-R is used.

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

	%s

	%s

%s
	%s

%s
  Read the manual at https://f2.freshman.tech
`,
		app.Name,
		app.Version,
		app.Authors[0],
		app.Usage,
		pterm.Bold.Sprintf("USAGE"),
		usageText,
		pterm.Bold.Sprintf("POSITIONAL ARGUMENTS"),
		pterm.Green("[PATHS TO FILES AND DIRECTORIES...]"),
		pterm.Bold.Sprintf("FLAGS"),
		flagCSVHelp,
		flagFindHelp,
		flagReplaceHelp,
		flagUndoHelp,
		pterm.Bold.Sprintf("OPTIONS"),
		flagAllowOverwritesHelp,
		flagCleanHelp,
		flagExcludeHelp,
		flagExcludeDirHelp,
		flagExiftoolOptsHelp,
		flagExecHelp,
		flagFixConflictsHelp,
		flagFixConflictsPatternHelp,
		flagHiddenHelp,
		flagIncludeHelp,
		flagIncludeDirHelp,
		flagIgnoreCaseHelp,
		flagIgnoreExtHelp,
		flagJSONHelp,
		flagMaxDepthHelp,
		flagNoColorHelp,
		flagOnlyDirHelp,
		flagPairHelp,
		flagPairOrderHelp,
		flagQuietHelp,
		flagRecursiveHelp,
		flagReplaceLimitHelp,
		flagResetIndexPerDirHelp,
		flagSortHelp,
		flagSortrHelp,
		flagSortPerDirHelp,
		flagSortVarHelp,
		flagStringModeHelp,
		flagTargetDirHelp,
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
		Set to any value to disable coloured output.`,
		pterm.Green("F2_DEFAULT_OPTS"),
		pterm.Green("F2_NO_COLOR"),
		pterm.Green("NO_COLOR"),
	)
}

func ShortHelp(_ *cli.Command) string {
	return fmt.Sprintf(
		`The batch renaming tool you'll actually enjoy using.

%s
  %s

%s
  $ f2 -f 'jpeg' -r 'jpg'
  $ f2 -r '{id3.artist}/{id3.album}/${1}_{id3.title}{ext}'

%s
  Use f2 --help to view the command-line options.
  Read the manual at https://f2.freshman.tech`,
		pterm.Bold.Sprintf("USAGE"),
		usageText,
		pterm.Bold.Sprintf("EXAMPLES"),
		pterm.Bold.Sprintf("LEARN MORE"),
	)
}
