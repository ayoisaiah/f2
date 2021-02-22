package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/gookit/color.v1"
)

var (
	red    = color.FgRed.Render
	green  = color.FgGreen.Render
	yellow = color.FgYellow.Render
)

type conflict int

const (
	EMPTY_FILENAME conflict = iota
	FILE_EXISTS
	OVERWRITNG_NEW_PATH
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
	paths           []Change
	matches         []Change
	replaceString   string
	startNumber     int
	exec            bool
	ignoreConflicts bool
	includeHidden   bool
	includeDir      bool
	onlyDir         bool
	ignoreCase      bool
	ignoreExt       bool
	searchRegex     *regexp.Regexp
	directories     []string
	recursive       bool
	undoFile        string
	outputFile      string
}

type mapFile struct {
	Date       string   `json:"date"`
	Operations []Change `json:"operations"`
}

// WriteToFile writes the details of a successful operation
// to the specified file so that it may be reversed if necessary
func (op *Operation) WriteToFile() error {
	// Create or truncate file
	file, err := os.Create(op.outputFile)
	if err != nil {
		return err
	}

	defer file.Close()

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

// Undo reverses the a successful renaming operation indicated
// in the specified map file
func (op *Operation) Undo() error {
	if op.undoFile == "" {
		return fmt.Errorf("Please pass a previously created map file to continue")
	}

	file, err := os.ReadFile(op.undoFile)
	if err != nil {
		return err
	}

	var mf mapFile
	err = json.Unmarshal([]byte(file), &mf)
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

	return op.Apply()
}

// PrintChanges displays the changes to be made in a
// table format
func (op *Operation) PrintChanges() {
	var data = make([][]string, len(op.matches))
	for i, v := range op.matches {
		source := filepath.Join(v.BaseDir, v.Source)
		target := filepath.Join(v.BaseDir, v.Target)
		d := []string{source, target, green("ok")}
		data[i] = d
	}

	printTable(data)
}

// Apply will check for conflicts and print the changes to be made
// or apply them directly to the filesystem if in execute mode.
// Conflicts will be ignored if indicated
func (op *Operation) Apply() error {
	if len(op.matches) == 0 {
		return fmt.Errorf("%s", red("Failed to match any files"))
	}

	if !op.ignoreConflicts {
		conflicts := op.DetectConflicts()
		if len(conflicts) > 0 {
			op.ReportConflicts(conflicts)
			fmt.Fprintln(os.Stderr, "Conflict detected! Please resolve before proceeding")
			return fmt.Errorf("Or append the %s flag to ignore all conflicts (may cause data loss)", yellow("-F"))
		}
	}

	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		if op.exec {
			// If target contains a slash, create all missing
			// directories before renaming the file
			execErr := fmt.Errorf("An error occurred while renaming '%s' to '%s'", source, target)
			if strings.Contains(ch.Target, "/") {
				// No need to check if the `dir` exists since `os.MkdirAll` handles that
				dir := filepath.Dir(ch.Target)
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					return execErr
				}
			}

			if err := os.Rename(source, target); err != nil {
				return execErr
			}
		}
	}

	if op.exec && len(op.matches) > 0 && op.outputFile != "" {
		return op.WriteToFile()
	} else if !op.exec && len(op.matches) > 0 {
		op.PrintChanges()
		fmt.Printf("Append the %s flag to apply the above changes\n", yellow("-x"))
	}

	return nil
}

// ReportConflicts prints any detected conflicts to the standard error
func (op *Operation) ReportConflicts(conflicts map[conflict][]Conflict) {
	var data [][]string
	if slice, exists := conflicts[EMPTY_FILENAME]; exists {
		for _, v := range slice {
			slice := []string{strings.Join(v.source, ""), "", red("❌ [Empty filename]")}
			data = append(data, slice)
		}
	}

	if slice, exists := conflicts[FILE_EXISTS]; exists {
		for _, v := range slice {
			slice := []string{strings.Join(v.source, ""), v.target, red("❌ [Path already exists]")}
			data = append(data, slice)
		}
	}

	if slice, exists := conflicts[OVERWRITNG_NEW_PATH]; exists {
		for _, v := range slice {
			for i, s := range v.source {
				var slice []string
				if i == 0 {
					slice = []string{s, v.target, green("ok")}
				} else {
					slice = []string{s, v.target, red("❌ [Overwriting newly renamed path]")}
				}
				data = append(data, slice)
			}
		}
	}

	printTable(data)
}

// DetectConflicts detects any conflicts that occur
// after renaming a file
func (op *Operation) DetectConflicts() map[conflict][]Conflict {
	conflicts := make(map[conflict][]Conflict)
	m := make(map[string][]string)

	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		// Report if replacement operation results in
		// an empty string for the new filename
		if ch.Target == "." {
			conflicts[EMPTY_FILENAME] = append(conflicts[EMPTY_FILENAME], Conflict{
				source: []string{source},
				target: target,
			})

			continue
		}

		// Report if target file exists on the filesystem
		if _, err := os.Stat(target); err == nil || !os.IsNotExist(err) {
			conflicts[FILE_EXISTS] = append(conflicts[FILE_EXISTS], Conflict{
				source: []string{source},
				target: target,
			})
		}

		// For detecting duplicates after renaming paths
		m[target] = append(m[target], source)
	}

	// Report duplicate targets if any
	for k, v := range m {
		if len(v) > 1 {
			conflicts[OVERWRITNG_NEW_PATH] = append(conflicts[OVERWRITNG_NEW_PATH], Conflict{
				source: v,
				target: k,
			})
		}

	}

	return conflicts
}

// FindMatches locates matches for the search pattern
// in each filename. Hidden files and directories are exempted
func (op *Operation) FindMatches() {
	for _, v := range op.paths {
		filename := filepath.Base(v.Source)

		if v.IsDir && !op.includeDir {
			continue
		}

		if op.onlyDir && !v.IsDir {
			continue
		}

		// ignore dotfiles
		if !op.includeHidden && filename[0] == 46 {
			continue
		}

		var f = filename
		if op.ignoreExt {
			f = filenameWithoutExtension(f)
		}

		matched := op.searchRegex.MatchString(f)
		if matched {
			op.matches = append(op.matches, v)
		}
	}
}

// SortMatches is used to sort files before directories
// and child directories before their parents
func (op *Operation) SortMatches() {
	sort.SliceStable(op.matches, func(i, j int) bool {
		if !op.matches[i].IsDir {
			return true
		}

		return op.matches[i].BaseDir > op.matches[j].BaseDir
	})
}

// Replace replaces the matched text in each path with the
// replacement string
func (op *Operation) Replace() {
	og := regexp.MustCompile("{{f}}")
	ext := regexp.MustCompile("{{ext}}")
	index := regexp.MustCompile("%([0-9]?)+d")
	for i, v := range op.matches {
		fileName, dir := filepath.Base(v.Source), filepath.Dir(v.Source)
		fileExt := filepath.Ext(fileName)
		if op.ignoreExt {
			fileName = filenameWithoutExtension(fileName)
		}

		str := op.searchRegex.ReplaceAllString(fileName, op.replaceString)

		// replace `{og}` in the replacement string with the original
		// filename (without the extension)
		if og.Match([]byte(str)) {
			str = og.ReplaceAllString(str, filenameWithoutExtension(fileName))
		}

		// replace `{ext}` in the replacement string with the file extension
		if ext.Match([]byte(str)) {
			str = ext.ReplaceAllString(str, fileExt)
		}

		// If numbering scheme is present
		if index.Match([]byte(str)) {
			b := index.Find([]byte(str))
			r := fmt.Sprintf(string(b), op.startNumber+i)
			str = index.ReplaceAllString(str, r)
		}

		// Only perform find and replace on `dir`
		// if file is a directory to avoid conflicts
		if op.includeDir && v.IsDir {
			dir = op.searchRegex.ReplaceAllString(dir, op.replaceString)
		}

		if op.ignoreExt {
			str += fileExt
		}

		v.Target = filepath.Join(dir, str)
		op.matches[i] = v
	}
}

// setPaths creates a Change struct for each path
// and checks if its a directory or not
func (op *Operation) setPaths(paths map[string][]os.DirEntry) error {
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

	return nil
}

// Run executes the operation sequence
func (op *Operation) Run() error {
	if op.undoFile != "" {
		return op.Undo()
	}

	op.FindMatches()

	if op.includeDir {
		op.SortMatches()
	}

	op.Replace()

	return op.Apply()
}

// NewOperation returns an Operation constructed
// from command line flags & arguments
func NewOperation(c *cli.Context) (*Operation, error) {
	if c.String("find") == "" && c.String("replace") == "" && c.String("undo") == "" {
		return nil, fmt.Errorf("Invalid arguments: one of `-f`, `-r` or `-u` must be present and set to a non empty string value\nUse 'goname --help' for more information")
	}

	op := &Operation{}
	op.outputFile = c.String("output-file")
	op.replaceString = c.String("replace")
	op.exec = c.Bool("exec")
	op.ignoreConflicts = c.Bool("force")
	op.includeDir = c.Bool("include-dir")
	op.startNumber = c.Int("start-num")
	op.includeHidden = c.Bool("hidden")
	op.ignoreCase = c.Bool("ignore-case")
	op.ignoreExt = c.Bool("ignore-ext")
	op.recursive = c.Bool("recursive")
	op.directories = c.Args().Slice()
	op.undoFile = c.String("undo")
	op.onlyDir = c.Bool("only-dir")

	if op.onlyDir {
		op.includeDir = true
	}

	if op.undoFile != "" {
		return op, nil
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
		return nil, fmt.Errorf("Malformed regular expression for search pattern %s", findPattern)
	}
	op.searchRegex = re

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
		paths, err = walk(paths, op.includeHidden)
		if err != nil {
			return nil, err
		}
	}

	return op, op.setPaths(paths)
}
