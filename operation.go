package f2

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/urfave/cli/v2"
	"gopkg.in/djherbis/times.v1"
)

var (
	red    = color.HEX("#FF2F2F")
	green  = color.HEX("#23D160")
	yellow = color.HEX("#FFAB00")
)

var (
	filenameRegex  = regexp.MustCompile("{{f}}")
	extensionRegex = regexp.MustCompile("{{ext}}")
	parentDirRegex = regexp.MustCompile("{{p}}")
	indexRegex     = regexp.MustCompile(`%(\d?)+d`)
	exifRegex      *regexp.Regexp
	dateRegex      *regexp.Regexp
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

type conflict int

const (
	dotCharacter = 46
)

const (
	emptyFilename conflict = iota
	fileExists
	overwritingNewPath
)

const (
	modTime     = "mtime"
	accessTime  = "atime"
	birthTime   = "btime"
	changeTime  = "ctime"
	currentTime = "now"
)

// Conflict represents a renaming operation conflict
// such as duplicate targets or empty filenames
type Conflict struct {
	source []string
	target string
}

// Change represents a single filename change
type Change struct {
	BaseDir string `json:"base_dir"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	IsDir   bool   `json:"is_dir"`
}

// Operation represents a batch renaming operation
type Operation struct {
	paths         []Change
	matches       []Change
	conflicts     map[conflict][]Conflict
	findString    string
	replacement   string
	startNumber   int
	exec          bool
	fixConflicts  bool
	includeHidden bool
	includeDir    bool
	onlyDir       bool
	ignoreCase    bool
	ignoreExt     bool
	searchRegex   *regexp.Regexp
	directories   []string
	recursive     bool
	undoFile      string
	outputFile    string
	workingDir    string
	stringMode    bool
	excludeFilter []string
	maxDepth      int
	sort          string
	reverseSort   bool
}

type mapFile struct {
	Date       string   `json:"date"`
	Operations []Change `json:"operations"`
}

// Exif represents exif information from an image file
type Exif struct {
	ISOSpeedRatings  []int
	DateTimeOriginal string
	Make             string
	Model            string
	ExposureTime     []string
	FocalLength      []string
	FNumber          []string
	ImageWidth       []int
	ImageLength      []int // the image height
	LensModel        string
}

func init() {
	tokens := make([]string, 0, len(dateTokens))
	for key := range dateTokens {
		tokens = append(tokens, key)
	}

	tokenString := strings.Join(tokens, "|")
	dateRegex = regexp.MustCompile(
		"{{(" + modTime + "|" + changeTime + "|" + birthTime + "|" + accessTime + "|" + currentTime + ")\\.(" + tokenString + ")}}",
	)

	exifRegex = regexp.MustCompile(
		"{{exif\\.(iso|et|fl|w|h|wh|make|model|lens|fnum)?(?:(dt)\\.(" + tokenString + "))?}}",
	)
}

// WriteToFile writes the details of a successful operation
// to the specified file so that it may be reversed if necessary
func (op *Operation) WriteToFile() (err error) {
	// Create or truncate file
	file, err := os.Create(op.outputFile)
	if err != nil {
		return err
	}

	defer func() {
		ferr := file.Close()
		if ferr != nil {
			err = ferr
		}
	}()

	mf := mapFile{
		Date:       time.Now().Format(time.RFC3339),
		Operations: op.matches,
	}

	writer := bufio.NewWriter(file)
	b, err := json.MarshalIndent(mf, "", "    ")
	if err != nil {
		return err
	}
	_, err = writer.Write(b)
	if err != nil {
		return err
	}

	return writer.Flush()
}

// undo reverses a successful renaming operation indicated
// in the specified map file
func (op *Operation) undo() error {
	if op.undoFile == "" {
		return fmt.Errorf(
			"specify a previously created map file to continue",
		)
	}

	file, err := os.ReadFile(op.undoFile)
	if err != nil {
		return err
	}

	var mf mapFile
	err = json.Unmarshal(file, &mf)
	if err != nil {
		return err
	}
	op.matches = mf.Operations

	for i, v := range op.matches {
		ch := v
		ch.Source = v.Target
		ch.Target = v.Source

		op.matches[i] = ch
	}

	// sort parent directories before child directories
	sort.SliceStable(op.matches, func(i, j int) bool {
		return op.matches[i].BaseDir < op.matches[j].BaseDir
	})

	return op.apply()
}

// sortBySize sorts the matches according to their file size
func (op *Operation) sortBySize() (err error) {
	sort.SliceStable(op.matches, func(i, j int) bool {
		ipath := filepath.Join(op.matches[i].BaseDir, op.matches[i].Source)
		jpath := filepath.Join(op.matches[j].BaseDir, op.matches[j].Source)

		var ifile, jfile fs.FileInfo
		ifile, err = os.Stat(ipath)
		jfile, err = os.Stat(jpath)

		isize := ifile.Size()
		jsize := jfile.Size()

		if op.reverseSort {
			return isize < jsize
		}

		return isize > jsize
	})

	return err
}

// sortByTime sorts the matches by the specified file attribute
// (mtime, atime, btime or ctime)
func (op *Operation) sortByTime() (err error) {
	sort.SliceStable(op.matches, func(i, j int) bool {
		ipath := filepath.Join(op.matches[i].BaseDir, op.matches[i].Source)
		jpath := filepath.Join(op.matches[j].BaseDir, op.matches[j].Source)

		var ifile, jfile times.Timespec
		ifile, err = times.Stat(ipath)
		jfile, err = times.Stat(jpath)

		var itime, jtime time.Time
		switch op.sort {
		case modTime:
			itime = ifile.ModTime()
			jtime = jfile.ModTime()
		case birthTime:
			itime = ifile.ModTime()
			jtime = jfile.ModTime()
			if ifile.HasBirthTime() {
				itime = ifile.BirthTime()
			}
			if jfile.HasBirthTime() {
				jtime = jfile.BirthTime()
			}
		case accessTime:
			itime = ifile.AccessTime()
			jtime = jfile.AccessTime()
		case changeTime:
			itime = ifile.ModTime()
			jtime = jfile.ModTime()
			if ifile.HasChangeTime() {
				itime = ifile.ChangeTime()
			}
			if jfile.HasChangeTime() {
				jtime = jfile.ChangeTime()
			}
		}

		it, jt := itime.UnixNano(), jtime.UnixNano()

		if op.reverseSort {
			return it < jt
		}

		return it > jt
	})

	return err
}

// sortBy delegates the sorting of matches to the appropriate method
func (op *Operation) sortBy() (err error) {
	sortOptions := []string{"size", accessTime, modTime, birthTime, changeTime}

	switch op.sort {
	case "size":
		return op.sortBySize()
	case accessTime, modTime, birthTime, changeTime:
		return op.sortByTime()
	}

	return fmt.Errorf(
		"invalid sort option '%s'. valid options include: %s",
		red.Sprint(op.sort),
		green.Sprint(strings.Join(sortOptions, ", ")),
	)
}

// printChanges displays the changes to be made in a
// table format
func (op *Operation) printChanges() {
	var data = make([][]string, len(op.matches))
	for i, v := range op.matches {
		source := filepath.Join(v.BaseDir, v.Source)
		target := filepath.Join(v.BaseDir, v.Target)
		d := []string{source, target, green.Sprint("ok")}
		data[i] = d
	}

	printTable(data)
}

// rename iterates over all the matches and renames them on the filesystem
// directories are auto-created if necessary
func (op *Operation) rename() error {
	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		renameErr := fmt.Errorf(
			"an error occurred while renaming '%s' to '%s'",
			source,
			target,
		)
		// If target contains a slash, create all missing
		// directories before renaming the file
		if strings.Contains(ch.Target, "/") ||
			strings.Contains(ch.Target, `\`) && runtime.GOOS == "windows" {
			// No need to check if the `dir` exists since `os.MkdirAll` handles that
			dir := filepath.Dir(ch.Target)
			err := os.MkdirAll(filepath.Join(ch.BaseDir, dir), 0750)
			if err != nil {
				return renameErr
			}
		}

		if err := os.Rename(source, target); err != nil {
			return renameErr
		}
	}

	return nil
}

// apply will check for conflicts and print the changes to be made
// or apply them directly to the filesystem if in execute mode.
// Conflicts will be ignored if indicated
func (op *Operation) apply() error {
	if len(op.matches) == 0 {
		fmt.Println("Failed to match any files")
		return nil
	}

	op.detectConflicts()
	if len(op.conflicts) > 0 && !op.fixConflicts {
		op.reportConflicts()
		fmt.Fprintln(
			os.Stderr,
			"conflict detected! please resolve before proceeding",
		)
		return fmt.Errorf(
			"or append the %s flag to fix conflicts automatically",
			yellow.Sprint("-F"),
		)
	}

	if op.exec {
		err := op.rename()
		if err != nil {
			return err
		}

		if op.outputFile != "" {
			return op.WriteToFile()
		}
	} else {
		if op.sort != "" {
			err := op.sortBy()
			if err != nil {
				return err
			}
		}
		op.printChanges()
		fmt.Printf("Append the %s flag to apply the above changes\n", yellow.Sprint("-x"))
	}

	return nil
}

// reportConflicts prints any detected conflicts to the standard error
func (op *Operation) reportConflicts() {
	var data [][]string
	if slice, exists := op.conflicts[emptyFilename]; exists {
		for _, v := range slice {
			slice := []string{
				strings.Join(v.source, ""),
				"",
				red.Sprint("❌ [Empty filename]"),
			}
			data = append(data, slice)
		}
	}

	if slice, exists := op.conflicts[fileExists]; exists {
		for _, v := range slice {
			slice := []string{
				strings.Join(v.source, ""),
				v.target,
				red.Sprint("❌ [Path already exists]"),
			}
			data = append(data, slice)
		}
	}

	if slice, exists := op.conflicts[overwritingNewPath]; exists {
		for _, v := range slice {
			for _, s := range v.source {
				slice := []string{
					s,
					v.target,
					red.Sprint("❌ [Overwriting newly renamed path]"),
				}
				data = append(data, slice)
			}
		}
	}

	printTable(data)
}

// detectConflicts detects any conflicts that occur
// after renaming a file. Conflicts are automatically
// fixed if specified
func (op *Operation) detectConflicts() {
	op.conflicts = make(map[conflict][]Conflict)
	m := make(map[string][]struct {
		source string
		index  int
	})

	for i, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		// Report if replacement operation results in
		// an empty string for the new filename
		if ch.Target == "." {
			op.conflicts[emptyFilename] = append(
				op.conflicts[emptyFilename],
				Conflict{
					source: []string{source},
					target: target,
				},
			)

			if op.fixConflicts {
				// The file is left unchanged
				op.matches[i].Target = ch.Source
			}

			continue
		}

		// Report if target file exists on the filesystem
		if _, err := os.Stat(target); err == nil ||
			!errors.Is(err, os.ErrNotExist) {
			op.conflicts[fileExists] = append(
				op.conflicts[fileExists],
				Conflict{
					source: []string{source},
					target: target,
				},
			)

			if op.fixConflicts {
				str := getNewPath(target, ch.BaseDir, nil)
				fullPath := filepath.Join(ch.BaseDir, str)
				op.matches[i].Target = str
				target = fullPath
			}
		}

		// For detecting duplicates after renaming paths
		m[target] = append(m[target], struct {
			source string
			index  int
		}{
			source: source,
			index:  i,
		})
	}

	// Report duplicate targets if any
	for k, v := range m {
		if len(v) > 1 {
			var sources []string
			for _, s := range v {
				sources = append(sources, s.source)
			}

			op.conflicts[overwritingNewPath] = append(
				op.conflicts[overwritingNewPath],
				Conflict{
					source: sources,
					target: k,
				},
			)

			if op.fixConflicts {
				for i, item := range v {
					if i == 0 {
						continue
					}

					str := getNewPath(k, op.matches[item.index].BaseDir, m)
					pt := filepath.Join(op.matches[item.index].BaseDir, str)
					if _, ok := m[pt]; !ok {
						m[pt] = []struct {
							source string
							index  int
						}{}
					}
					op.matches[item.index].Target = str
				}
			}
		}
	}
}

// sortMatches is used to sort files before directories
// and child directories before their parents
func (op *Operation) sortMatches() {
	sort.SliceStable(op.matches, func(i, j int) bool {
		if !op.matches[i].IsDir {
			return true
		}

		return len(op.matches[i].BaseDir) > len(op.matches[j].BaseDir)
	})
}

func replaceDateVariables(file, input string) (string, error) {
	t, err := times.Stat(file)
	if err != nil {
		return "", err
	}

	submatches := dateRegex.FindAllStringSubmatch(input, -1)
	for _, submatch := range submatches {
		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return "", err
		}

		var timeStr string
		switch submatch[1] {
		case modTime:
			modTime := t.ModTime()
			timeStr = modTime.Format(dateTokens[submatch[2]])
		case birthTime:
			birthTime := t.ModTime()
			if t.HasBirthTime() {
				birthTime = t.BirthTime()
			}
			timeStr = birthTime.Format(dateTokens[submatch[2]])
		case accessTime:
			accessTime := t.AccessTime()
			timeStr = accessTime.Format(dateTokens[submatch[2]])
		case changeTime:
			changeTime := t.ModTime()
			if t.HasChangeTime() {
				changeTime = t.ChangeTime()
			}
			timeStr = changeTime.Format(dateTokens[submatch[2]])
		case currentTime:
			currentTime := time.Now()
			timeStr = currentTime.Format(dateTokens[submatch[2]])
		}

		input = regex.ReplaceAllString(input, timeStr)
	}

	return input, nil
}

func getExifData(file string) (*Exif, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer func() {
		ferr := f.Close()
		if ferr != nil {
			err = ferr
		}
	}()

	exifData := &Exif{}
	// Errors in decoding the exif data are ignored intentionally
	// The corresponding exif variable will be replaced by an empty
	// string
	x, err := exif.Decode(f)
	if err == nil {
		b, err := x.MarshalJSON()
		if err == nil {
			_ = json.Unmarshal(b, exifData)
		}
	}

	return exifData, nil
}

func replaceExifVariables(exifData *Exif, input string) (string, error) {
	submatches := exifRegex.FindAllStringSubmatch(input, -1)
	for _, submatch := range submatches {
		regex, err := regexp.Compile(submatch[0])
		if err != nil {
			return "", err
		}

		if strings.Contains(submatch[0], "exif.dt") {
			submatch = append(submatch[:1], submatch[1+1:]...)
		}

		switch submatch[1] {
		case "dt":
			date := exifData.DateTimeOriginal
			arr := strings.Split(date, " ")
			var dt time.Time
			d := strings.ReplaceAll(arr[0], ":", "-")
			t := arr[1]
			var err error
			dt, err = time.Parse(time.RFC3339, d+"T"+t+"Z")
			if err != nil {
				return "", err
			}

			timeStr := dt.Format(dateTokens[submatch[2]])
			input = regex.ReplaceAllString(input, timeStr)
		case "model":
			cmodel := exifData.Model
			cmodel = strings.ReplaceAll(cmodel, "/", "_")
			input = regex.ReplaceAllString(input, cmodel)
		case "lens":
			lens := exifData.LensModel
			lens = strings.ReplaceAll(lens, "/", "_")
			input = regex.ReplaceAllString(input, lens)
		case "make":
			cmake := exifData.Make
			input = regex.ReplaceAllString(input, cmake)
		case "iso":
			var iso string
			if len(exifData.ISOSpeedRatings) > 0 {
				iso = strconv.Itoa(exifData.ISOSpeedRatings[0])
			}
			input = regex.ReplaceAllString(input, "ISO"+iso)
		case "et":
			var et string
			if len(exifData.ExposureTime) > 0 {
				et = exifData.ExposureTime[0]
				et = strings.ReplaceAll(et, "/", "_")
			}
			input = regex.ReplaceAllString(input, et+"s")
		case "fnum":
			v := exifDivision(exifData.FNumber)
			input = regex.ReplaceAllString(input, "f"+v)
		case "fl":
			v := exifDivision(exifData.FocalLength)
			input = regex.ReplaceAllString(input, v+"mm")
		case "wh":
			var wh string
			if len(exifData.ImageLength) > 0 && len(exifData.ImageWidth) > 0 {
				h, w := exifData.ImageLength[0], exifData.ImageWidth[0]
				wh = strconv.Itoa(w) + "x" + strconv.Itoa(h)
			}
			input = regex.ReplaceAllString(input, wh)
		case "h":
			var h string
			if len(exifData.ImageLength) > 0 {
				h = strconv.Itoa(exifData.ImageLength[0])
			}
			input = regex.ReplaceAllString(input, h)
		case "w":
			var w string
			if len(exifData.ImageWidth) > 0 {
				w = strconv.Itoa(exifData.ImageWidth[0])
			}
			input = regex.ReplaceAllString(input, w)
		}
	}

	return input, nil
}

func (op *Operation) handleVariables(str string, ch Change) (string, error) {
	fileName := filepath.Base(ch.Source)
	fileExt := filepath.Ext(fileName)
	parentDir := filepath.Base(ch.BaseDir)
	if parentDir == "." {
		// Set to base folder of current working directory
		parentDir = filepath.Base(op.workingDir)
	}

	// replace `{{f}}` in the replacement string with the original
	// filename (without the extension)
	if filenameRegex.Match([]byte(str)) {
		str = filenameRegex.ReplaceAllString(
			str,
			filenameWithoutExtension(fileName),
		)
	}

	// replace `{{ext}}` in the replacement string with the file extension
	if extensionRegex.Match([]byte(str)) {
		str = extensionRegex.ReplaceAllString(str, fileExt)
	}

	// replace `{{p}}` in the replacement string with the parent directory name
	if parentDirRegex.Match([]byte(str)) {
		str = parentDirRegex.ReplaceAllString(str, parentDir)
	}

	// handle date variables (e.g {{mtime.DD}})
	if dateRegex.Match([]byte(str)) {
		source := filepath.Join(ch.BaseDir, ch.Source)
		out, err := replaceDateVariables(source, str)
		if err != nil {
			return "", err
		}
		str = out
	}

	if exifRegex.Match([]byte(str)) {
		source := filepath.Join(ch.BaseDir, ch.Source)
		exifData, err := getExifData(source)
		if err != nil {
			return "", err
		}

		out, err := replaceExifVariables(exifData, str)
		if err != nil {
			return "", err
		}
		str = out
	}

	return str, nil
}

func (op *Operation) replaceString(fileName string) (str string) {
	findString := op.findString
	if op.stringMode {
		if findString == "" {
			findString = fileName
		}

		if op.ignoreCase {
			str = op.searchRegex.ReplaceAllString(
				fileName,
				op.replacement,
			)
		} else {
			str = strings.ReplaceAll(fileName, findString, op.replacement)
		}
	} else {
		str = op.searchRegex.ReplaceAllString(fileName, op.replacement)
	}

	return str
}

// replace replaces the matched text in each path with the
// replacement string
func (op *Operation) replace() error {
	for i, v := range op.matches {
		fileName, dir := filepath.Base(v.Source), filepath.Dir(v.Source)
		fileExt := filepath.Ext(fileName)
		if op.ignoreExt {
			fileName = filenameWithoutExtension(fileName)
		}

		str := op.replaceString(fileName)

		// handle variables
		str, err := op.handleVariables(str, v)
		if err != nil {
			return err
		}

		// If numbering scheme is present
		if indexRegex.Match([]byte(str)) {
			b := indexRegex.Find([]byte(str))
			r := fmt.Sprintf(string(b), op.startNumber+i)
			str = indexRegex.ReplaceAllString(str, r)
		}

		if op.ignoreExt {
			str += fileExt
		}

		v.Target = filepath.Join(dir, str)
		op.matches[i] = v
	}

	return nil
}

// findMatches locates matches for the search pattern
// in each filename. Hidden files and directories are exempted
func (op *Operation) findMatches() error {
	for _, v := range op.paths {
		filename := filepath.Base(v.Source)

		if v.IsDir && !op.includeDir {
			continue
		}

		if op.onlyDir && !v.IsDir {
			continue
		}

		// ignore dotfiles on unix and hidden files on windows
		if !op.includeHidden {
			r, err := isHidden(filename, v.BaseDir)
			if err != nil {
				return err
			}
			if r {
				continue
			}
		}

		var f = filename
		if op.ignoreExt {
			f = filenameWithoutExtension(f)
		}

		if op.stringMode {
			findStr := op.findString

			if op.ignoreCase {
				f = strings.ToLower(f)
				findStr = strings.ToLower(findStr)
			}

			if strings.Contains(f, findStr) {
				op.matches = append(op.matches, v)
			}
			continue
		}

		matched := op.searchRegex.MatchString(f)
		if matched {
			op.matches = append(op.matches, v)
		}
	}

	return nil
}

// filterMatches excludes any files or directories that match
// the find pattern in accordance with the provided exclude pattern
func (op *Operation) filterMatches() error {
	var filtered []Change
	filters := strings.Join(op.excludeFilter, "|")
	regex, err := regexp.Compile(filters)
	if err != nil {
		return err
	}

	for _, m := range op.matches {
		if !regex.MatchString(m.Source) {
			filtered = append(filtered, m)
		}
	}

	op.matches = filtered
	return nil
}

// setPaths creates a Change struct for each path
// and checks if its a directory or not
func (op *Operation) setPaths(paths map[string][]os.DirEntry) {
	for k, v := range paths {
		for _, f := range v {
			var change = Change{
				BaseDir: k,
				IsDir:   f.IsDir(),
				Source:  filepath.Clean(f.Name()),
			}

			op.paths = append(op.paths, change)
		}
	}
}

// run executes the operation sequence
func (op *Operation) run() error {
	if op.undoFile != "" {
		return op.undo()
	}

	err := op.findMatches()
	if err != nil {
		return err
	}

	if len(op.excludeFilter) != 0 {
		err = op.filterMatches()
		if err != nil {
			return err
		}
	}

	if op.includeDir {
		op.sortMatches()
	}

	err = op.replace()
	if err != nil {
		return err
	}

	return op.apply()
}

// setOptions applies the command line arguments
// onto the operation
func setOptions(op *Operation, c *cli.Context) error {
	op.outputFile = c.String("output-file")
	op.findString = c.String("find")
	op.replacement = c.String("replace")
	op.exec = c.Bool("exec")
	op.fixConflicts = c.Bool("fix-conflicts")
	op.includeDir = c.Bool("include-dir")
	op.startNumber = c.Int("start-num")
	op.includeHidden = c.Bool("hidden")
	op.ignoreCase = c.Bool("ignore-case")
	op.ignoreExt = c.Bool("ignore-ext")
	op.recursive = c.Bool("recursive")
	op.directories = c.Args().Slice()
	op.undoFile = c.String("undo")
	op.onlyDir = c.Bool("only-dir")
	op.stringMode = c.Bool("string-mode")
	op.excludeFilter = c.StringSlice("exclude")
	op.maxDepth = c.Int("max-depth")

	// Sorting
	if c.String("sort") != "" {
		op.sort = c.String("sort")
	} else if c.String("sortr") != "" {
		op.sort = c.String("sortr")
		op.reverseSort = true
	}

	if op.onlyDir {
		op.includeDir = true
	}

	findPattern := c.String("find")
	// Match entire string if find pattern is empty
	if findPattern == "" {
		findPattern = ".*"
	}

	if op.ignoreCase {
		findPattern = "(?i)" + findPattern
	}

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return err
	}
	op.searchRegex = re

	return nil
}

// newOperation returns an Operation constructed
// from command line flags & arguments
func newOperation(c *cli.Context) (*Operation, error) {
	if c.String("find") == "" && c.String("replace") == "" &&
		c.String("undo") == "" {
		return nil, fmt.Errorf(
			"invalid argument: one of `-f`, `-r` or `-u` must be present and set to a non empty string value\nUse 'f2 --help' for more information",
		)
	}

	op := &Operation{}
	err := setOptions(op, c)
	if err != nil {
		return nil, err
	}

	if op.undoFile != "" {
		return op, nil
	}

	var paths = make(map[string][]os.DirEntry)
	for _, v := range op.directories {
		paths[v], err = os.ReadDir(v)
		if err != nil {
			return nil, err
		}
	}

	// Use current directory
	if len(paths) == 0 {
		paths["."], err = os.ReadDir(".")
		if err != nil {
			return nil, err
		}
	}

	if op.recursive {
		paths, err = walk(paths, op.includeHidden, op.maxDepth)
		if err != nil {
			return nil, err
		}
	}

	// Get the current working directory
	op.workingDir, err = filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	op.setPaths(paths)
	return op, nil
}
