package replace

import (
	"fmt"
	"log/slog"
	"regexp"
)

type numbersToSkip struct {
	min int
	max int
}

func (s numbersToSkip) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("min: %d, max:%d", s.min, s.max))
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

func (v indexVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"index_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("index", v.indexFormat),
		slog.String("number_system", v.numberSystem),
		slog.String("skip", fmt.Sprintf("%v", v.skip)),
		slog.Bool("step_set", v.step.isSet),
		slog.Int("step_value", v.step.value),
		slog.Int("start_number", v.startNumber),
		slog.Any("submatch", v.submatch),
	)
}

type indexVars struct {
	// stores the indices of submatches that specify a capture variable
	capturVarIndex []int
	matches        []indexVarMatch
}

type transformVarMatch struct {
	regex      *regexp.Regexp
	token      string
	captureVar string
	inputStr   string
	timeStr    string
	val        []string
}

func (v transformVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"transform_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("token", v.token),
		slog.String("capture_var", v.captureVar),
		slog.String("input_str", v.inputStr),
		slog.Any("val", v.val),
	)
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

func (v exiftoolVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"exiftool_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.String("attr", v.attr),
		slog.Any("val", v.val),
	)
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

func (v exifVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"exif_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.String("attr", v.attr),
		slog.Any("val", v.val),
	)
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

func (v id3VarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"date_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.String("tag", v.tag),
		slog.Any("val", v.val),
	)
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

func (v dateVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"date_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.String("attr", v.attr),
		slog.Any("val", v.val),
		slog.String("token", v.token),
	)
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

func (v hashVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"hash_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.Any("val", v.val),
		slog.Any("length", v.hashFn),
	)
}

type hashVars struct {
	matches []hashVarMatch
}

type csvVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	column         int
}

func (v csvVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"csv_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.Int("column", v.column),
	)
}

type csvVars struct {
	submatches [][]string
	values     []csvVarMatch
}

type filenameVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
}

func (v filenameVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"filename_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
	)
}

type filenameVars struct {
	matches []filenameVarMatch
}

type extVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
}

func (e extVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"ext_var_match",
		slog.String("regex", e.regex.String()),
		slog.String("transform_token", e.transformToken),
	)
}

type extVars struct {
	matches []extVarMatch
}

type parentDirVarMatch struct {
	regex          *regexp.Regexp
	transformToken string
	parent         int
}

func (v parentDirVarMatch) LogAttr() slog.Attr {
	return slog.Group(
		"parent_dir_var_match",
		slog.String("regex", v.regex.String()),
		slog.String("transform_token", v.transformToken),
		slog.Int("parent", v.parent),
	)
}

type parentDirVars struct {
	matches []parentDirVarMatch
}

type variables struct {
	exif      exifVars
	exiftool  exiftoolVars
	index     indexVars
	id3       id3Vars
	hash      hashVars
	date      dateVars
	transform transformVars
	csv       csvVars
	filename  filenameVars
	ext       extVars
	parentDir parentDirVars
}

func (v variables) LogValue() slog.Value {
	var slogAttr []slog.Attr

	for _, v := range v.filename.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.ext.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.parentDir.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	if len(v.csv.submatches) > 0 {
		slogAttr = append(
			slogAttr,
			slog.Any("csv_submatches", v.csv.submatches),
		)
	}

	for _, v := range v.csv.values {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.transform.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.date.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.hash.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.id3.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.index.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.exiftool.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	for _, v := range v.exif.matches {
		slogAttr = append(slogAttr, v.LogAttr())
	}

	return slog.GroupValue(
		slogAttr...,
	)
}
