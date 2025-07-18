// Package replace substitutes each match according to the configured
// replacement directives which could be plain strings, builtin variables, or
// regex capture variables
package replace

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/eval"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
	"github.com/ayoisaiah/f2/v2/internal/sortfiles"
	"github.com/ayoisaiah/f2/v2/internal/status"
	"github.com/ayoisaiah/f2/v2/replace/variables"
	"github.com/ayoisaiah/f2/v2/report"
)

// replaceString replaces all matches in the filename
// with the replacement string.
func replaceString(conf *config.Config, originalName string) string {
	return variables.RegexReplace(
		conf.Search.Regex,
		originalName,
		conf.Replacement,
		conf.ReplaceLimit,
	)
}

// applyReplacements applies the configured replacement patterns to the source
// filename.
func applyReplacement(
	conf *config.Config,
	vars *variables.Variables,
	change *file.Change,
) error {
	originalName := change.Source
	fileExt := filepath.Ext(originalName)

	if conf.IgnoreExt && !change.IsDir {
		originalName = pathutil.StripExtension(originalName)
	}

	change.Target = replaceString(conf, originalName)

	// Replace any variables present with their corresponding values
	err := variables.Replace(conf, change, vars)
	if err != nil {
		return err
	}

	// Reattach the original extension to the new file name
	if conf.IgnoreExt && !change.IsDir {
		change.Target += fileExt
	}

	change.Target = strings.TrimSpace(filepath.Clean(change.Target))
	change.Status = status.OK
	change.TargetPath = filepath.Join(change.TargetDir, change.Target)

	return nil
}

// replaceMatches handles the replacement of matches in each file with the
// replacement string.
func replaceMatches(
	conf *config.Config,
	matches file.Changes,
) (file.Changes, error) {
	vars, err := variables.Extract(conf.Replacement)
	if err != nil {
		return nil, err
	}

	var pairs int

	for i := range matches {
		change := matches[i]

		if conf.Search.FindCond != nil && !change.MatchesFindCond {
			slog.Debug(
				"skipping match due to not meeting find condition",
				slog.Any("match", change),
			)

			continue
		}

		// Detect and rename file pairs
		if change.PrimaryPair != nil {
			ext := filepath.Ext(change.Source)
			common := pathutil.StripExtension(change.PrimaryPair.Target)
			change.Target = common + ext
			change.TargetPath = filepath.Join(
				change.TargetDir,
				change.Target,
			)
			change.Status = status.OK
			pairs++

			continue
		}

		change.Position = i - pairs

		err := applyReplacement(conf, &vars, change)
		if err != nil {
			return nil, err
		}

		matches[i] = change
	}

	slog.Debug(
		"all replacements applied",
		slog.Any("changes", matches),
	)

	return matches, nil
}

func handleReplacementChain(
	conf *config.Config,
	matches file.Changes,
) (file.Changes, error) {
	for i, v := range conf.ReplacementSlice {
		conf.Replacement = v

		slog.Debug(
			"handling replacement chain",
			slog.Int("position", i),
			slog.String("replacement", v),
		)

		var err error

		matches, err = replaceMatches(conf, matches)
		if err != nil {
			return nil, err
		}

		if len(conf.ReplacementSlice)-1 == i {
			return matches, nil
		}

		err = conf.SetFind(i + 1)
		if err != nil {
			return nil, err
		}

		err = prepNextChain(conf, matches)
		if err != nil {
			return nil, err
		}
	}

	return matches, nil
}

func prepNextChain(
	conf *config.Config,
	matches file.Changes,
) (err error) {
	var findVars variables.Variables

	if conf.Search.FindCond != nil {
		findVars, err = variables.Extract(
			conf.Search.FindCond.String(),
		)
		if err != nil {
			return err
		}
	}

	for j := range matches {
		change := matches[j]

		originalTarget := change.Target

		// Update the source to the target from the previous replacement
		// in preparation for the next replacement
		matches[j].Source = change.Target

		slog.Debug(
			"preparing for next replacement chain",
			slog.Any("change", change),
		)

		if conf.Search.FindCond == nil {
			continue
		}

		if change.PrimaryPair != nil && change.PrimaryPair.MatchesFindCond {
			change.MatchesFindCond = true
			continue
		}

		change.Target = conf.Search.FindCond.String()

		err := variables.Replace(conf, change, &findVars)
		if err != nil {
			return err
		}

		result, err := eval.Evaluate(change.Target)
		if err != nil {
			if conf.Verbose {
				report.SearchEvalFailed(change.SourcePath, change.Target, err)
			}

			matches[j].MatchesFindCond = false
		}

		if !result {
			matches[j].MatchesFindCond = false
		}

		matches[j].Target = originalTarget
	}

	return nil
}

// Replace applies the file name replacements according to the --replace
// argument.
func Replace(
	conf *config.Config,
	matches file.Changes,
) (file.Changes, error) {
	if conf.ExifToolVarPresent && matches.ShouldExtractExiftool() {
		names, indices := matches.SourceNamesWithIndices(conf.Pair)

		slog.Debug("extracting exif variables", slog.Any("paths", names))

		fileMeta, err := variables.ExtractExiftoolMetadata(
			conf,
			names...)
		if err != nil {
			return matches, err
		}

		for i := range fileMeta {
			index := indices[i]

			slog.Debug(
				"attaching exif data to file",
				slog.String("match", matches[index].SourcePath),
				slog.String("file", fileMeta[i].File),
				slog.Bool(
					"is_match",
					fileMeta[i].File == matches[index].SourcePath,
				),
			)

			matches[index].ExiftoolData = &fileMeta[i]
		}
	}

	if conf.CSVFilename != "" {
		for i := range matches {
			ch := matches[i]

			conf.Replacement = ch.Target

			vars, err := variables.Extract(conf.Replacement)
			if err != nil {
				return nil, err
			}

			err = applyReplacement(conf, &vars, ch)
			if err != nil {
				return nil, err
			}
		}
	}

	matches, err := handleReplacementChain(conf, matches)
	if err != nil {
		return nil, err
	}

	if (conf.IncludeDir || conf.CSVFilename != "") && conf.Exec {
		sortfiles.ForRenamingAndUndo(matches, conf.Revert)
	}

	slog.Debug(
		"replacement complete",
		slog.Any("changes", matches),
	)

	return matches, nil
}
