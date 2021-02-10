package main

import (
	"bufio"
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

const opsFile = ".goname-operation.txt"

// Change represents a single filename change
type Change struct {
	baseDir string
	source  string
	target  string
	isDir   bool
}

// Operation represents a bulk rename operation
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
}

// WriteToFile writes the details of the last successful operation
// to a file so that it may be reversed if necessary
func (op *Operation) WriteToFile() error {
	// Create or truncate file
	file, err := os.Create(opsFile)
	if err != nil {
		return err
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, v := range op.matches {
		_, err = writer.WriteString(v.target + "|" + v.source + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// Undo reverses the last successful renaming operation
func (op *Operation) Undo() error {
	file, err := os.Open(opsFile)
	if err != nil {
		return err
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	// If file is empty
	if fi.Size() == 0 {
		return fmt.Errorf("No operation to undo")
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		slice := strings.Split(scanner.Text(), "|")
		if len(slice) != 2 {
			return fmt.Errorf("Corrupted data. Cannot undo")
		}
		source, target := slice[0], slice[1]
		ch := Change{}
		ch.source = source
		ch.target = target

		op.matches = append(op.matches, ch)
	}

	for i, v := range op.matches {
		isDir, err := isDirectory(v.source)
		if err != nil {
			// An error may mean that the path does not exist
			// which indicates that the directory containing the file
			// was also renamed.
			if os.IsNotExist(err) {
				dir := filepath.Dir(v.source)

				// Get the directory that is changing
				var d Change
				for _, m := range op.matches {
					if m.target == dir {
						d = m
						break
					}
				}

				re, err := regexp.Compile(d.target)
				if err != nil {
					return err
				}

				srcFile, srcDir := filepath.Base(v.source), filepath.Dir(v.source)
				targetFile, targetDir := filepath.Base(v.target), filepath.Dir(v.target)

				// Update the directory of the path to the current name
				// instead of the old one which no longer exists
				srcDir = re.ReplaceAllString(srcDir, d.source)
				targetDir = re.ReplaceAllString(targetDir, d.source)

				v.source = filepath.Join(srcDir, srcFile)
				v.target = filepath.Join(targetDir, targetFile)
			} else {
				return err
			}
		}

		v.isDir = isDir
		op.matches[i] = v
	}

	op.SortMatches()

	return op.Apply()
}

// Apply will check for conflicts and print
// the changes to be made or apply them directly
// if in execute mode. Conflicts will be ignored if
// specified
func (op *Operation) Apply() error {
	if len(op.matches) == 0 {
		return fmt.Errorf("%s", red("Failed to match any files"))
	}

	if !op.ignoreConflicts {
		err := op.ReportConflicts()
		if err != nil {
			return err
		}
	}

	for _, ch := range op.matches {
		var source, target = ch.source, ch.target
		if printFullPaths {
			source = filepath.Join(ch.baseDir, source)
			target = filepath.Join(ch.baseDir, target)
		}

		if op.exec {
			if err := os.Rename(source, target); err != nil {
				return fmt.Errorf("An error occurred while renaming '%s' to '%s'", source, target)
			}
		} else {
			fmt.Println(source, "➟", green(target), "✅")
		}
	}

	if op.exec && len(op.matches) > 0 {
		return op.WriteToFile()
	} else if !op.exec && len(op.matches) > 0 {
		fmt.Printf("%s\n", yellow("*** Use the -x flag to apply the above changes ***"))
	}

	return nil
}

// ReportConflicts ensures that there are no conflicts
// after renaming a file
func (op *Operation) ReportConflicts() error {
	m := make(map[string][]string)

	var err error
	for _, ch := range op.matches {
		var source, target = ch.source, ch.target
		if printFullPaths {
			source = filepath.Join(ch.baseDir, source)
			target = filepath.Join(ch.baseDir, target)
		}

		// Ensure file does not exist on the filesystem
		if _, err1 := os.Stat(target); err1 == nil || !os.IsNotExist(err1) {
			fmt.Printf("%s ➟ %s %s %s\n", source, red(target), red("[File exists]"), "❌")
			if err == nil {
				err = fmt.Errorf("%s\n%s", red("Conflict detected: overwriting existing file(s)"), yellow("Use the -F flag to ignore conflicts and rename anyway"))
			}
		}

		// Detect duplicates after renaming paths
		if _, exists := m[target]; exists {
			m[target] = append(m[target], source)
		} else {
			m[target] = []string{source}
		}
	}

	if err != nil {
		return err
	}

	// Report duplicates if any
	for k, v := range m {
		if len(v) > 1 {
			if err == nil {
				err = fmt.Errorf("%s\n%s", red("Conflict detected: overwriting newly renamed path"), yellow("Use the -F flag to ignore conflicts and rename anyway"))
			}

			for i, s := range v {
				if i == 0 {
					fmt.Printf("%s ➟ %s %s\n", s, green(k), "✅")
				} else {
					fmt.Printf("%s ➟ %s %s\n", s, red(k), "❌")
				}
			}
		}
	}

	return err
}

// FindMatches locates matches for the search pattern
// in each filename. Hidden files and directories are exempted
func (op *Operation) FindMatches() error {
	for _, v := range op.paths {
		filename := filepath.Base(v.source)

		if v.isDir && !op.includeDir {
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

	return nil
}

// SortMatches is used to sort files before directories
func (op *Operation) SortMatches() {
	sort.SliceStable(op.matches, func(i, j int) bool {
		return !op.matches[i].isDir
	})
}

// Replace replaces the matched text in each path with the
// replacement string
func (op *Operation) Replace() error {
	og := regexp.MustCompile("{og}")
	ext := regexp.MustCompile("{ext}")
	index := regexp.MustCompile("%([0-9]?)+d")
	for i, v := range op.matches {
		fileName, dir := filepath.Base(v.source), filepath.Dir(v.source)
		var str = op.searchRegex.ReplaceAllString(fileName, op.replaceString)

		// replace `{og}` in the replacement string with the original
		// filename (without the extension)
		if og.Match([]byte(str)) {
			str = og.ReplaceAllString(str, strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		}

		// replace `{ext}` in the replacement string with the file extension
		if ext.Match([]byte(str)) {
			str = ext.ReplaceAllString(str, filepath.Ext(fileName))
		}

		// If numbering scheme is present
		if index.Match([]byte(str)) {
			b := index.Find([]byte(str))
			r := fmt.Sprintf(string(b), op.startNumber+i)
			str = index.ReplaceAllString(str, r)
		}

		// Only perform find and replace on `dir`
		// if file is a directory to avoid conflicts
		if op.includeDir && v.isDir {
			dir = op.searchRegex.ReplaceAllString(dir, op.replaceString)
		}

		// Report error if replacement operation results in
		// an empty string for the new filename
		if str == "" {
			return fmt.Errorf("%s\n%s ➟ %s %s ", red("Error detected: Operation resulted in empty filename"), v.source, red("[Empty filename]"), "❌")
		}

		v.target = filepath.Join(dir, str)
		op.matches[i] = v
	}

	return nil
}

// setPaths creates a Change struct for each path
// and checks if its a directory or not
func (op *Operation) setPaths(paths map[string][]os.DirEntry) error {
	for k, v := range paths {
		for _, f := range v {
			var change = Change{
				baseDir: k,
				isDir:   f.IsDir(),
				source:  filepath.Clean(f.Name()),
			}

			op.paths = append(op.paths, change)
		}
	}

	return nil
}

// NewOperation returns an Operation constructed
// from command line flags & arguments
func NewOperation(c *cli.Context) (*Operation, error) {
	if c.String("find") == "" && c.String("replace") == "" {
		return nil, fmt.Errorf("Invalid arguments: one of `-f` or `-r` must be present and set to a non empty string value\nUse 'goname --help' for more information")
	}

	op := &Operation{}
	op.replaceString = c.String("replace")
	op.exec = c.Bool("exec")
	op.ignoreConflicts = c.Bool("force")
	op.includeDir = c.Bool("include-dir")
	op.startNumber = c.Int("start-num")
	op.includeHidden = c.Bool("hidden")
	op.ignoreCase = c.Bool("ignore-case")
	op.ignoreExt = c.Bool("ignore-ext")

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
	for _, v := range c.Args().Slice() {
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

	if len(paths) > 1 {
		printFullPaths = true
	}

	return op, op.setPaths(paths)
}
