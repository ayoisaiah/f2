package replace

import (
	"regexp"
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
	startNumber int
}

type indexVars struct {
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

type variables struct {
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
