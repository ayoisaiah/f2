package variables

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/timeutil"
)

var transformTokens string

var (
	filenameVarRegex  *regexp.Regexp
	extensionVarRegex *regexp.Regexp
	parentDirVarRegex *regexp.Regexp
	indexVarRegex     *regexp.Regexp
	hashVarRegex      *regexp.Regexp
	transformVarRegex *regexp.Regexp
	csvVarRegex       *regexp.Regexp
	exiftoolVarRegex  *regexp.Regexp
	id3VarRegex       *regexp.Regexp
	exifVarRegex      *regexp.Regexp
	dateVarRegex      *regexp.Regexp
)

var dateTokens = map[string]string{
	"YYYY": "2006",
	"YY":   "06",
	"MMMM": "January",
	"MMM":  "Jan",
	"MM":   "01",
	"M":    "1",
	"DDDD": "Monday",
	"DDD":  "Mon",
	"DD":   "02",
	"D":    "2",
	"H":    "15",
	"hh":   "03",
	"h":    "3",
	"mm":   "04",
	"m":    "4",
	"ss":   "05",
	"s":    "5",
	"A":    "PM",
	"a":    "pm",
}

func init() {
	tokens := make([]string, 0, len(dateTokens))
	for key := range dateTokens {
		tokens = append(tokens, key)
	}

	tokenString := strings.Join(tokens, "|")

	transformTokens = fmt.Sprintf(
		"(up|lw|ti|win|mac|di|norm|(?:dt\\.(%s)))",
		tokenString,
	)

	filenameVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+f(?:\\.%s)?}+", transformTokens),
	)
	extensionVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+(2)?ext(?:\\.%s)?}+", transformTokens),
	)
	parentDirVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+(\\d+)?p(?:\\.%s)?}+", transformTokens),
	)
	indexVarRegex = regexp.MustCompile(
		`{+(\$\d+)?(\d+)?(%(\d?)+d)([borh])?(-?\d+)?(?:<(\d+(?:-\d+)?(?:;\s*\d+(?:-\d+)?)*)>)?(##)?}+`,
	)
	hashVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+hash.(sha1|sha256|sha512|md5)(?:\\.%s)?}+",
			transformTokens,
		),
	)
	transformVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+(?:<(?:(\\$\\d+)|([^\\.]+))>)?\\.%s}+", transformTokens),
	)
	csvVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+csv.(\\d+)(?:\\.%s)?}+", transformTokens),
	)
	exiftoolVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+xt\\.([0-9a-zA-Z]+)(?:\\.%s)?}+",
			transformTokens,
		),
	)
	id3VarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+id3\\.(format|type|title|album|album_artist|artist|genre|year|composer|track|disc|total_tracks|total_discs)(?:\\.%s)?}+",
			transformTokens,
		),
	)

	dateVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+("+timeutil.Mod+"|"+timeutil.Change+"|"+timeutil.Birth+"|"+timeutil.Access+"|"+timeutil.Current+")\\.("+tokenString+")(?:\\.%s)?}+",
			transformTokens,
		),
	)

	exifVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+(?:exif|x)\\.(?:(iso|et|fl|w|h|wh|make|model|lens|fnum|fl35|lat|lon|soft)|(?:(cdt)(?:\\.("+tokenString+"))?))(?:\\.%s)?}+",
			transformTokens,
		),
	)
}
