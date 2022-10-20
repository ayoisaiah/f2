package f2

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

var transformTokens = "(up|lw|ti|win|mac|di)"

var (
	filenameVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+f(?:.%s)?}+", transformTokens),
	)
	extensionVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+ext(?:.%s)?}+", transformTokens),
	)
	parentDirVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+(\\d+)?p(?:.%s)?}+", transformTokens),
	)
	indexVarRegex = regexp.MustCompile(
		`(\$\d+)?(\d+)?(%(\d?)+d)([borh])?(-?\d+)?(?:<(\d+(?:-\d+)?(?:;\s*\d+(?:-\d+)?)*)>)?`,
	)
	randomVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+(\\d+)?r(?:(_l|_d|_ld)|(?:<([^>])>))?(?:.%s)?}+",
			transformTokens,
		),
	)
	hashVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+hash.(sha1|sha256|sha512|md5)(?:.%s)?}+",
			transformTokens,
		),
	)
	transformVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+(?:<(?:(\\$\\d+)|([^\\.]+))>)?\\.%s}+", transformTokens),
	)
	csvVarRegex = regexp.MustCompile(
		fmt.Sprintf("{+csv.(\\d+)(?:.%s)}+", transformTokens),
	)
	exiftoolVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+xt\\.([0-9a-zA-Z]+)(?:.%s)}+",
			transformTokens,
		),
	)
	id3VarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+id3\\.(format|type|title|album|album_artist|artist|genre|year|composer|track|disc|total_tracks|total_discs)(?:.%s)?}+",
			transformTokens,
		),
	)
	exifVarRegex *regexp.Regexp
	dateVarRegex *regexp.Regexp
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
	dateVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+("+modTime+"|"+changeTime+"|"+birthTime+"|"+accessTime+"|"+currentTime+")\\.("+tokenString+")(?:.%s)?}+",
			transformTokens,
		),
	)

	exifVarRegex = regexp.MustCompile(
		fmt.Sprintf(
			"{+(?:exif|x)\\.(?:(iso|et|fl|w|h|wh|make|model|lens|fnum|fl35|lat|lon|soft)|(?:(dt)\\.("+tokenString+")))(?:.%s)?}+",
			transformTokens,
		),
	)

	// for the sake of replacing random string variables
	rand.Seed(time.Now().UnixNano())
}
