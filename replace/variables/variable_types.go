package variables

import (
	"context"
	"regexp"
	"sync"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
)

// VariableProvider defines the interface for all variable replacement providers.
type VariableProvider interface {
	Replace(ctx context.Context, conf *config.Config, change *file.Change) error
}

type hashAlgorithm string

const (
	sha1Hash   hashAlgorithm = "sha1"
	sha256Hash hashAlgorithm = "sha256"
	sha512Hash hashAlgorithm = "sha512"
	md5Hash    hashAlgorithm = "md5"
	xxh32Hash  hashAlgorithm = "xxh32"
	xxh64Hash  hashAlgorithm = "xxh64"
)

type numbersToSkip struct {
	min int
	max int
}

type indexVarMatch struct {
	regex        *regexp.Regexp
	indexFormat  string
	numberSystem string // Binary, Octal, Roman, Decimal
	skip         []numbersToSkip
	submatch     []string
	step         struct {
		isSet bool
		value int
	}
	startNumber  int
	isCaptureVar bool
}

type indexVars struct {
	mu             *sync.Mutex
	currentBaseDir string
	capturVarIndex []int
	offset         []int
	matches        []indexVarMatch
	newDirIndex    int
}

type transformVarMatch struct {
	regex      *regexp.Regexp
	token      string
	captureVar string
	inputStr   string
	timeStr    string
	val        []string
}

type transformVars struct {
	matches []transformVarMatch
}

type exiftoolVarMatch struct {
	regex          *regexp.Regexp
	attr           string
	transformToken string
	val            []string
}

type exiftoolVars struct {
	matches []exiftoolVarMatch
}

type exifVarMatch struct {
	regex          *regexp.Regexp
	attr           string
	timeStr        string
	transformToken string
	val            []string
}

type exifVars struct {
	matches []exifVarMatch
}

type id3VarMatch struct {
	regex          *regexp.Regexp
	tag            string
	transformToken string
	val            []string
}

type id3Vars struct {
	matches []id3VarMatch
}

type dateVarMatch struct {
	regex          *regexp.Regexp
	attr           string
	token          string
	transformToken string
	val            []string
}

type dateVars struct {
	matches []dateVarMatch
}

type hashVarMatch struct {
	regex          *regexp.Regexp
	hashFn         hashAlgorithm
	transformToken string
	val            []string
}

type hashVars struct {
	matches []hashVarMatch
}

type csvVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	column         int
}

type csvVars struct {
	submatches [][]string
	values     []csvVarMatch
}

type filenameVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
}

type filenameVars struct {
	matches []filenameVarMatch
}

type extVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	doubleExt      bool
}

type extVars struct {
	matches []extVarMatch
}

type parentDirVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	parent         int
}

type parentDirVars struct {
	matches []parentDirVarMatch
}

type Variables struct {
	csv       csvVars
	exif      exifVars
	filename  filenameVars
	id3       id3Vars
	hash      hashVars
	date      dateVars
	transform transformVars
	exiftool  exiftoolVars
	ext       extVars
	parentDir parentDirVars
	index     indexVars
}

func (v *Variables) IndexMatches() int {
	return len(v.index.matches)
}

//nolint:revive // intentional unexported type
func (v *Variables) HashMatches() []hashVarMatch {
	return v.hash.matches
}

//nolint:revive // intentional unexported type
func (v *Variables) ID3Matches() []id3VarMatch {
	return v.id3.matches
}

//nolint:revive // intentional unexported type
func (v *Variables) ExifMatches() []exifVarMatch {
	return v.exif.matches
}
