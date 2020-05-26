package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/gookit/color.v1"
)

var (
	red    = color.FgRed.Render
	green  = color.FgGreen.Render
	yellow = color.FgYellow.Render
)

const opsFile = ".goname-operation.txt"

// Change represents a single filename change
type Change struct {
	source string
	target string
	isDir  bool
}

// Operation represents a bulk rename operation
type Operation struct {
	paths           []string
	matches         []Change
	replaceString   string
	startNumber     int
	exec            bool
	ignoreConflicts bool
	templateMode    bool
	includeDir      bool
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
		return fmt.Errorf("Failed to match any files")
	}

	if !op.ignoreConflicts {
		err := op.ReportConflicts()
		if err != nil {
			return err
		}
	}

	for _, ch := range op.matches {
		if op.exec {
			if err := os.Rename(ch.source, ch.target); err != nil {
				return fmt.Errorf("An error occurred while renaming '%s' to '%s'", ch.source, ch.target)
			}
		} else {
			fmt.Println(ch.source, "➟", green(ch.target), "✅")
		}
	}

	if op.exec && len(op.matches) > 0 {
		return op.WriteToFile()
	} else if !op.exec && len(op.matches) > 0 {
		color.Style{color.FgYellow, color.OpBold}.Println("*** Use the -x flag to apply the above changes ***")
	}

	return nil
}

// ReportConflicts ensures that there are no conflicts
// after renaming a file
func (op *Operation) ReportConflicts() error {
	m := make(map[string][]string)

	var err error
	for _, ch := range op.matches {
		// Ensure file does not exist on the filesystem
		if _, err1 := os.Stat(ch.target); err1 == nil || os.IsExist(err1) {
			fmt.Printf("%s ➟ %s %s %s\n", ch.source, red(ch.target), red("[File exists]"), "❌")
			if err == nil {
				err = fmt.Errorf("%s\n%s", red("Conflict detected: overwriting existing file(s)"), yellow("Use the -F flag to ignore conflicts and rename anyway"))
			}
		}

		// Detect duplicates after renaming paths
		if _, exists := m[ch.target]; exists {
			m[ch.target] = append(m[ch.target], ch.source)
		} else {
			m[ch.target] = []string{ch.source}
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
	for _, f := range op.paths {
		var change Change
		isDir, err := isDirectory(f)
		if err != nil {
			return err
		}

		if isDir && !op.includeDir {
			continue
		}

		filename := filepath.Base(f)
		// ignore dotfiles
		if filename[0] == 46 {
			continue
		}

		matched := op.searchRegex.MatchString(filename)
		if matched {
			change.isDir = isDir
			change.source = filepath.Clean(f)
			op.matches = append(op.matches, change)
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
		var str string

		if op.templateMode {
			// Use the replacement string as a template for new name
			str = op.replaceString
		} else {
			str = op.searchRegex.ReplaceAllString(fileName, op.replaceString)
		}

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
		// if file is a directory and templateMode is off
		// to avoid conflicts
		if op.includeDir && v.isDir && !op.templateMode {
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

// NewOperation returns an Operation constructed
// from command line flags & arguments
func NewOperation(c *cli.Context) (*Operation, error) {
	if c.String("find") == "" && c.String("replace") == "" {
		return nil, fmt.Errorf("Invalid arguments: one of `-f` or `-r` must be present and set to a non empty string value\nUse 'goname --help' for more information")
	}

	op := &Operation{}
	op.paths = c.Args().Slice()
	op.replaceString = c.String("replace")
	op.exec = c.Bool("exec")
	op.ignoreConflicts = c.Bool("force")
	op.includeDir = c.Bool("include-dir")
	op.templateMode = c.Bool("template-mode")
	op.startNumber = c.Int("start-num")

	findPattern := c.String("find")

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return nil, fmt.Errorf("Malformed regular expression for search pattern %s", findPattern)
	}
	op.searchRegex = re

	// Check if a newline-separated list of paths are passed
	// to the standard input
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		op.paths = strings.Split(string(bytes), "\n")
	}

	// If paths are omitted, default to the current directory
	if len(op.paths) == 0 {
		file, err := os.Open(".")
		if err != nil {
			return nil, err
		}

		defer file.Close()

		names, err := file.Readdirnames(0)
		if err != nil {
			return nil, err
		}

		op.paths = names
	}

	return op, nil
}
