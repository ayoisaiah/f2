package variables

import (
	"context"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
)

func (pv parentDirVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(pv.matches) == 0 {
		return nil
	}

	abspath, err := filepath.Abs(change.SourcePath)
	if err != nil {
		return err
	}

	target := replaceParentDirVars(conf, change.Target, abspath, pv)

	change.Target = target

	return nil
}

func (fv filenameVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(fv.matches) == 0 {
		return nil
	}

	sourceName := filepath.Base(change.OriginalName)
	if !change.IsDir {
		sourceName = pathutil.StripExtension(sourceName)
	}

	target := replaceFilenameVars(conf, change.Target, sourceName, fv)

	change.Target = target

	return nil
}

func (ev extVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(ev.matches) == 0 {
		return nil
	}

	target := replaceExtVars(conf, change, ev)

	change.Target = target

	return nil
}

func replaceParentDirVars(
	conf *config.Config,
	target, absSourcePath string,
	pv parentDirVars,
) string {
	for i := range pv.matches {
		current := pv.matches[i]

		var parentDir string

		var count int

		sp := absSourcePath

		for {
			count++

			sp = filepath.Dir(sp)

			parentDir = filepath.Base(sp)

			if current.parent == count {
				break
			}

			// break if we get to the root
			if parentDir == "/" || parentDir == "\\" {
				parentDir = ""
				break
			}
		}

		source := transformString(conf, parentDir, current.transformToken)

		target = RegexReplace(current.regex, target, source, 0, nil)
	}

	return target
}

func replaceFilenameVars(
	conf *config.Config,
	target, sourceName string,
	fv filenameVars,
) string {
	for i := range fv.matches {
		current := fv.matches[i]

		source := transformString(conf, sourceName, current.transformToken)

		target = RegexReplace(current.regex, target, source, 0, nil)
	}

	return target
}

func getDoubleExtension(filename string) string {
	ext := filepath.Ext(filename)
	ext2 := filepath.Ext(pathutil.StripExtension(filename))

	return ext2 + ext
}

func replaceExtVars(
	conf *config.Config,
	change *file.Change,
	ev extVars,
) (target string) {
	for i := range ev.matches {
		fileExt := filepath.Ext(change.OriginalName)

		if change.IsDir {
			fileExt = "" // Directory names do not have extensions
		}

		current := ev.matches[i]

		if current.doubleExt {
			fileExt = getDoubleExtension(change.OriginalName)
		}

		source := transformString(conf, fileExt, current.transformToken)

		target = RegexReplace(current.regex, change.Target, source, 0, nil)
	}

	return target
}

func getExtVars(replacementInput string) (extVars, error) {
	var evMatches extVars

	if !extensionVarRegex.MatchString(replacementInput) {
		return evMatches, nil
	}

	submatches := extensionVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 3

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return evMatches, errInvalidSubmatches
		}

		var match extVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return evMatches, err
		}

		match.regex = regex

		if submatch[1] != "" {
			match.doubleExt = true
		}

		match.transformToken = submatch[2]

		evMatches.matches = append(evMatches.matches, match)
	}

	return evMatches, nil
}

func getParentDirVars(replacementInput string) (parentDirVars, error) {
	var pvMatches parentDirVars

	if !parentDirVarRegex.MatchString(replacementInput) {
		return pvMatches, nil
	}

	submatches := parentDirVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 3

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return pvMatches, errInvalidSubmatches
		}

		var match parentDirVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return pvMatches, err
		}

		match.regex = regex
		match.parent = 1

		if submatch[1] != "" {
			match.parent, err = strconv.Atoi(submatch[1])
			if err != nil {
				return pvMatches, err
			}
		}

		match.transformToken = submatch[2]

		pvMatches.matches = append(pvMatches.matches, match)
	}

	return pvMatches, nil
}

func getFilenameVars(replacementInput string) (filenameVars, error) {
	var fvMatches filenameVars

	if !filenameVarRegex.MatchString(replacementInput) {
		return fvMatches, nil
	}

	submatches := filenameVarRegex.FindAllStringSubmatch(replacementInput, -1)

	expectedLength := 2

	for _, submatch := range submatches {
		if len(submatch) < expectedLength {
			return fvMatches, errInvalidSubmatches
		}

		var match filenameVarMatch

		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return fvMatches, err
		}

		match.regex = regex

		match.transformToken = submatch[1]

		fvMatches.matches = append(fvMatches.matches, match)
	}

	return fvMatches, nil
}
