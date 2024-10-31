// Package replace substitutes each match according to the configured
// replacement directives which could be plain strings, builtin variables, or
// regex capture variables
package replace

import (
	"path/filepath"
	"strings"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/pathutil"
	"github.com/ayoisaiah/f2/internal/sortfiles"
	"github.com/ayoisaiah/f2/internal/status"
	"github.com/ayoisaiah/f2/replace/variables"
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

	// If using indexes without an explicit sort, ensure that the files
	// are arranged hierarchically
	if vars.IndexMatches() > 0 && conf.Sort == config.SortDefault {
		sortfiles.Hierarchically(matches)
	}

	var pairs int

	for i := range matches {
		change := matches[i]

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

	return matches, nil
}

func handleReplacementChain(
	conf *config.Config,
	matches file.Changes,
) (file.Changes, error) {
	replacementSlice := conf.ReplacementSlice

	for i, v := range replacementSlice {
		conf.Replacement = v

		var err error

		matches, err = replaceMatches(conf, matches)
		if err != nil {
			return nil, err
		}

		if len(replacementSlice) == 1 ||
			(i > 0 && i == len(replacementSlice)-1) {
			return matches, nil
		}

		for j := range matches {
			change := matches[j]

			// Update the source to the target from the previous replacement
			// in preparation for the next replacement
			if i != len(replacementSlice)-1 {
				matches[j].Source = change.Target
			}
		}

		err = conf.SetFindStringRegex(i + 1)
		if err != nil {
			return nil, err
		}
	}

	return matches, nil
}

// Replace applies the file name replacements according to the --replace
// argument.
func Replace(
	conf *config.Config,
	changes file.Changes,
) (file.Changes, error) {
	var err error

	// Don't replace the extension in pair mode
	if conf.Pair {
		conf.IgnoreExt = true
	}

	if conf.CSVFilename != "" {
		for i := range changes {
			ch := changes[i]

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

	changes, err = handleReplacementChain(conf, changes)
	if err != nil {
		return nil, err
	}

	if (conf.IncludeDir || conf.CSVFilename != "") && conf.Exec {
		sortfiles.ForRenamingAndUndo(changes, conf.Revert)
	}

	return changes, nil
}
