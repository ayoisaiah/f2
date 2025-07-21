package app

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"

	"github.com/ayoisaiah/f2/v2/internal/localize"
)

func helpText(app *cli.Command) string {
	usageText := localize.T("usageText")

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

	positionalArgumentsHelp := localize.T("positionalArgumentsHelp")

	learnMoreHelp := localize.T("learnMoreHelp")

	return fmt.Sprintf(`%s %s
%s

%s

Project repository: https://github.com/ayoisaiah/f2

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

	%s

	%s

%s
  %s

%s
	%s
`,
		app.Name,
		app.Version,
		app.Authors[0],
		app.Usage,
		pterm.Bold.Sprintf("%s", localize.T("usage")),
		usageText,
		pterm.Bold.Sprintf("%s", localize.T("positionalArguments")),
		pterm.Green("[PATHS TO FILES AND DIRECTORIES...]"),
		positionalArgumentsHelp,
		pterm.Bold.Sprintf("%s", localize.T("flags")),
		flagCSVHelp,
		flagFindHelp,
		flagReplaceHelp,
		flagUndoHelp,
		pterm.Bold.Sprintf("%s", localize.T("options")),
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
		pterm.Bold.Sprintf("%s", localize.T("environmentalVariables")),
		envHelp(),
		pterm.Bold.Sprintf("%s", localize.T("learnMore")),
		learnMoreHelp,
	)
}

func envHelp() string {
	envHelp := localize.T("envHelp")

	colorHelp := localize.T("colorHelp")

	return fmt.Sprintf(`%s
		%s

	%s, %s
		%s`,
		pterm.Green("F2_DEFAULT_OPTS"),
		envHelp,
		pterm.Green("F2_NO_COLOR"),
		pterm.Green("NO_COLOR"),
		colorHelp,
	)
}

func ShortHelp(_ *cli.Command) string {
	usageText := localize.T("usageText")

	usage := pterm.Bold.Sprintf("%s", localize.T("usage"))

	examples := pterm.Bold.Sprintf("%s", localize.T("examples"))

	learnMore := pterm.Bold.Sprintf("%s", localize.T("learnMore"))

	return localize.TWithOpts(&i18n.LocalizeConfig{
		MessageID: "shortHelp",
		TemplateData: map[string]string{
			"Usage":     usage,
			"UsageText": usageText,
			"Examples":  examples,
			"LearnMore": learnMore,
		},
	})
}
