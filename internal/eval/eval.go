package eval

import (
	"regexp"
	"strings"
	"time"

	"github.com/maja42/goval"
	"go.withmatt.com/size"

	"github.com/ayoisaiah/f2/v2/internal/apperr"
	"github.com/ayoisaiah/f2/v2/internal/localize"
)

var functions = make(map[string]goval.ExpressionFunction)

var errInvalidArgs = &apperr.Error{
	Message: localize.T("error.eval_invalid_args"),
}

// ParseDuration parses a duration string.
// examples: "10d", "-1.5w" or "3Y4M5d".
// Add time units are "d"="D", "w"="W", "M", "y"="Y".
// Adapted from: https://gist.github.com/xhit/79c9e137e1cfe332076cdda9f5e24699?permalink_comment_id=5170854#gistcomment-5170854
func parseDuration(s string) (time.Duration, error) {
	neg := false
	if s != "" && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	re := regexp.MustCompile(`(\d*\.\d+|\d+)\D*`)
	unitMap := map[string]time.Duration{
		"d": 24,
		"w": 7 * 24,
		"M": 30 * 24,
		"y": 365 * 24,
	}

	strs := re.FindAllString(s, -1)

	var sumDur time.Duration

	for _, str := range strs {
		var _hours time.Duration = 1

		for unit, hours := range unitMap {
			if strings.Contains(str, unit) {
				str = strings.ReplaceAll(str, unit, "h")
				_hours = hours

				break
			}
		}

		dur, err := time.ParseDuration(str)
		if err != nil {
			return 0, err
		}

		sumDur += dur * _hours
	}

	if neg {
		sumDur = -sumDur
	}

	return sumDur, nil
}

func init() {
	functions["strlen"] = func(args ...any) (any, error) {
		if len(args) == 0 {
			return nil, errInvalidArgs
		}

		str, _ := args[0].(string)

		return len(str), nil
	}

	functions["dur"] = func(args ...any) (any, error) {
		if len(args) == 0 {
			return nil, errInvalidArgs
		}

		str, _ := args[0].(string)

		dur, err := parseDuration(str)
		if err != nil {
			return nil, err
		}

		return dur.Seconds(), nil
	}

	functions["contains"] = func(args ...any) (any, error) {
		if len(args) <= 1 {
			return nil, errInvalidArgs
		}

		str, _ := args[0].(string)
		substr, _ := args[1].(string)

		return strings.Contains(str, substr), nil
	}

	functions["size"] = func(args ...any) (any, error) {
		if len(args) == 0 {
			return nil, errInvalidArgs
		}

		str, _ := args[0].(string)

		// Handle Exiftool format: "26 kB" -> "26K", "1.2 MB" -> "1.2M"
		// Remove spaces between number and unit for compatibility with size.ParseCapacity
		r := strings.NewReplacer(
			" kB",
			"K",
			" MB",
			"M",
			" GB",
			"G",
			" TB",
			"T",
			" bytes",
			"",
		)
		str = r.Replace(str)

		s, err := size.ParseCapacity(str)
		if err != nil {
			return nil, err
		}

		//nolint:gosec // risk of overflow is acceptable
		return int(s.Bytes()), nil
	}

	functions["matches"] = func(args ...any) (any, error) {
		if len(args) <= 1 {
			return nil, errInvalidArgs
		}

		str, _ := args[0].(string)
		exp, _ := args[1].(string)

		reg := regexp.MustCompile(exp)

		return reg.MatchString(str), nil
	}
}

func Evaluate(expression string) (bool, error) {
	eval := goval.NewEvaluator()

	result, err := eval.Evaluate(expression, nil, functions)
	if err != nil {
		return false, err
	}

	r, _ := result.(bool)
	if !r {
		return false, nil
	}

	return true, nil
}
