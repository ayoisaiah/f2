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

	"github.com/urfave/cli/v2"
	"gopkg.in/gookit/color.v1"
)

var printFullPaths bool

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
	ignoreCase      bool
	ignoreExt       bool
	searchRegex     *regexp.Regexp
	directories     []string
	recursive       bool
	undoMode        bool
	outputFile      string
	mapFile         string
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

	writer := bufio.NewWriter(file)
	b, err := json.Marshal(op.matches)
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
	if op.mapFile == "" {
		return fmt.Errorf("Please pass a previously created map file to continue")
	}

	file, err := os.ReadFile(op.mapFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(file), &op.matches)
	if err != nil {
		return err
	}

	for i, v := range op.matches {
		isDir, err := isDirectory(v.Source)
		if err != nil {
			// An error may mean that the path does not exist
			// which indicates that the directory containing the file
			// was also renamed.
			if os.IsNotExist(err) {
				dir := filepath.Dir(v.Source)

				// Get the directory that is changing
				var d Change
				for _, m := range op.matches {
					if m.Target == dir {
						d = m
						break
					}
				}

				re, err := regexp.Compile(d.Target)
				if err != nil {
					return err
				}

				srcFile, srcDir := filepath.Base(v.Source), filepath.Dir(v.Source)
				targetFile, targetDir := filepath.Base(v.Target), filepath.Dir(v.Target)

				// Update the directory of the path to the current name
				// instead of the old one which no longer exists
				srcDir = re.ReplaceAllString(srcDir, d.Source)
				targetDir = re.ReplaceAllString(targetDir, d.Source)

				v.Source = filepath.Join(srcDir, srcFile)
				v.Target = filepath.Join(targetDir, targetFile)
			} else {
				return err
			}
		}

		v.IsDir = isDir
		op.matches[i] = v
	}

	op.SortMatches()

	return op.Apply()
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
			return fmt.Errorf("%s", yellow("Resolve the conflicts before proceeding or use the -F flag to ignore conflicts and rename anyway"))
		}
	}

	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		if printFullPaths {
			source = filepath.Join(ch.BaseDir, source)
			target = filepath.Join(ch.BaseDir, target)
		}

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
		} else {
			fmt.Println(source, "➟", green(target), "✅")
		}
	}

	if op.exec && len(op.matches) > 0 && op.outputFile != "" {
		return op.WriteToFile()
	} else if !op.exec && len(op.matches) > 0 {
		fmt.Printf("%s\n", yellow("*** Use the -x flag to apply the above changes ***"))
	}

	return nil
}

// ReportConflicts prints any detected conflicts to the standard error
func (op *Operation) ReportConflicts(conflicts map[conflict][]Conflict) {
	if slice, exists := conflicts[EMPTY_FILENAME]; exists {
		fmt.Fprintln(os.Stderr, color.Bold.Sprintf("Operation resulted in empty filename:"))
		for _, v := range slice {
			fmt.Fprintf(os.Stderr, "%s ➟ %s %s\n", strings.Join(v.source, ""), red("[Empty filename]"), "❌")
		}
	}

	if slice, exists := conflicts[FILE_EXISTS]; exists {
		fmt.Fprintln(os.Stderr, color.Bold.Sprintf("Overwriting existing path"))
		for _, v := range slice {
			fmt.Fprintf(os.Stderr, "%s ➟ %s %s %s\n", strings.Join(v.source, ""), red(v.target), red("[File exists]"), "❌")
		}
	}

	if slice, exists := conflicts[OVERWRITNG_NEW_PATH]; exists {
		for _, v := range slice {
			fmt.Fprintln(os.Stderr, color.Bold.Sprintf("Overwriting newly renamed path:"))
			for i, s := range v.source {
				if i == 0 {
					fmt.Fprintf(os.Stderr, "%s ➟ %s %s\n", s, green(v.target), "✅")
				} else {
					fmt.Fprintf(os.Stderr, "%s ➟ %s %s\n", s, red(v.target), "❌")
				}
			}
		}
	}
}

// DetectConflicts detects any conflicts that occur
// after renaming a file
func (op *Operation) DetectConflicts() map[conflict][]Conflict {
	conflicts := make(map[conflict][]Conflict)
	m := make(map[string][]string)

	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		if printFullPaths {
			source = filepath.Join(ch.BaseDir, source)
			target = filepath.Join(ch.BaseDir, target)
		}

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
func (op *Operation) SortMatches() {
	sort.SliceStable(op.matches, func(i, j int) bool {
		return !op.matches[i].IsDir
	})
}

// Replace replaces the matched text in each path with the
// replacement string
func (op *Operation) Replace() {
	og := regexp.MustCompile("{og}")
	ext := regexp.MustCompile("{ext}")
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

func (op *Operation) Run() error {
	if op.undoMode {
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
	if c.String("find") == "" && c.String("replace") == "" {
		return nil, fmt.Errorf("Invalid arguments: one of `-f` or `-r` must be present and set to a non empty string value\nUse 'goname --help' for more information")
	}

	op := &Operation{}
	op.outputFile = c.String("output-file")
	op.mapFile = c.String("map-file")
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
	op.undoMode = c.Bool("undo")

	if op.undoMode {
		return op, nil
	}

	findPattern := c.String("find")
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
		absolutePath, err := filepath.Abs(v)
		if err != nil {
			return nil, err
		}

		paths[absolutePath], err = os.ReadDir(v)
		if err != nil {
			return nil, err
		}
	}

	// Use current directory
	if len(paths) == 0 {
		currentDir, err := filepath.Abs(".")
		if err != nil {
			return nil, err
		}

		paths[currentDir], err = os.ReadDir(".")
		if err != nil {
			return nil, err
		}
	}

	if op.recursive {
		paths, err = walk(paths)
		if err != nil {
			return nil, err
		}
	}

	if len(paths) > 1 {
		printFullPaths = true
	}

	return op, op.setPaths(paths)
}
