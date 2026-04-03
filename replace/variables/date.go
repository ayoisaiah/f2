package variables

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/djherbis/times"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/timeutil"
)

// Replace replaces any date variables in the target with the
// corresponding date value.
func (dv dateVars) Replace(
	_ context.Context,
	conf *config.Config,
	change *file.Change,
) error {
	if len(dv.matches) == 0 {
		return nil
	}

	target, err := replaceDateVars(conf, change.Target, change.SourcePath, dv)
	if err != nil {
		return err
	}

	change.Target = target

	return nil
}

// Helper function to select the appropriate time based on the attribute.
// nowFunc is passed for testability of the "Current" time case.
func getSelectedTime(
	attr string,
	spec times.Timespec,
	nowFunc func() time.Time,
) time.Time {
	switch attr {
	case timeutil.Mod:
		return spec.ModTime()
	case timeutil.Birth:
		if spec.HasBirthTime() {
			return spec.BirthTime()
		}

		return time.Time{}
	case timeutil.Access:
		return spec.AccessTime()
	case timeutil.Change:
		if spec.HasChangeTime() {
			return spec.ChangeTime()
		}

		return spec.ModTime()
	case timeutil.Current:
		return nowFunc()
	default:
		return time.Time{}
	}
}

func replaceDateToken(t time.Time, token string) string {
	if token == "" || token == "dt" {
		return t.Format(time.RFC3339)
	}

	token = strings.TrimPrefix(token, "dt.")

	if token == "unix" {
		return strconv.FormatInt(t.Unix(), 10)
	}

	if token == "since" {
		dur := time.Since(t)
		return strconv.FormatInt(int64(dur.Seconds()), 10)
	}

	return t.Format(dateTokens[token])
}

// replaceDateVars replaces any date variables in the target with the
// corresponding date value.
func replaceDateVars(
	conf *config.Config,
	target, sourcePath string,
	dateVarMatches dateVars,
) (string, error) {
	timeSpec, err := times.Stat(sourcePath)
	if err != nil {
		return "", err
	}

	for i := range dateVarMatches.matches {
		current := dateVarMatches.matches[i]
		regex := current.regex

		selectedTime := getSelectedTime(current.attr, timeSpec, time.Now)

		timeStr := replaceDateToken(selectedTime, current.token)

		timeStr = transformString(conf, timeStr, current.transformToken)

		target = RegexReplace(regex, target, timeStr, 0, nil)
	}

	return target, nil
}
