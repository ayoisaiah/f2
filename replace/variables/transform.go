package variables

import (
	"context"
	"log/slog"
	"strings"

	"github.com/araddon/dateparse"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
)

// Replace handles string transformations like uppercase,
// lowercase, stripping characters, e.t.c.
func (tv transformVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if !transformVarRegex.MatchString(change.Target) {
		return nil
	}

	sourceName := change.Source
	if conf.IgnoreExt && !change.IsDir {
		sourceName = pathutil.StripExtension(sourceName)
	}

	matches := conf.Search.Regex.FindAllString(sourceName, -1)

	target, err := replaceTransformVars(conf, change.Target, matches, tv)
	if err != nil {
		return err
	}

	change.Target = target

	return nil
}

func transformString(conf *config.Config, source, token string) string {
	switch token {
	case "up":
		return strings.ToUpper(source)
	case "lw":
		return strings.ToLower(source)
	case "ti":
		c := cases.Title(language.English)
		return c.String(strings.ToLower(source))
	case "win":
		return RegexReplace(
			osutil.CompleteWindowsForbiddenCharRegex,
			source,
			"",
			0,
			nil,
		)
	case "mac":
		return RegexReplace(osutil.MacForbiddenCharRegex, source, "", 0, nil)
	case "di":
		return removeDiacritics(source)
	case "norm":
		result, _, err := transform.String(norm.NFKC, source)
		if err != nil {
			slog.Debug(
				"unable to perform unicode normalization",
				slog.String("source", source),
			)

			return source
		}

		return result
	}

	if strings.HasPrefix(token, "dt") {
		dateTime, err := dateparse.ParseAny(source)
		if err != nil {
			slog.Debug(
				"unable to parse datetime string",
				slog.String("source", source),
			)

			return source
		}

		if conf.Location != nil {
			dateTime = dateTime.In(conf.Location)
		}

		return replaceDateToken(dateTime, token)
	}

	return source
}

// replaceTransformVars handles string transformations like uppercase,
// lowercase, stripping characters, e.t.c.
func replaceTransformVars(
	conf *config.Config,
	target string,
	matches []string,
	tv transformVars,
) (string, error) {
	// if capture variables are present, they would have been replaced by now
	// so updated transform vars must be retrieved again
	t, err := getTransformVars(target)
	if err != nil {
		return "", err
	}

	for i := range tv.matches {
		current := tv.matches[i]
		if current.captureVar != "" {
			current = t.matches[i]
		}

		regex := current.regex

		match := current.inputStr

		// if capture variables aren't being used, transform the find matches
		if match == "" {
			for _, v := range matches {
				target = RegexReplace(
					regex,
					target,
					transformString(conf, v, current.token),
					1,
					nil,
				)
			}

			continue
		}

		target = RegexReplace(
			regex,
			target,
			transformString(conf, match, current.token),
			0,
			nil,
		)
	}

	return target, nil
}
