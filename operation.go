package f2

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

var (
	red    = color.HEX("#FF2F2F")
	green  = color.HEX("#23D160")
	yellow = color.HEX("#FFAB00")
)

var (
	errInvalidArgument = errors.New(
		"Invalid argument: one of `-f`, `-r` or `-u` must be present and set to a non empty string value\nUse 'f2 --help' for more information",
	)

	errConflictDetected = fmt.Errorf(
		"Conflict detected! Please resolve before proceeding or append the %s flag to fix conflicts automatically",
		yellow.Sprint("-F"),
	)
)

var pathSeperator = "/"

const (
	windows = "windows"
	darwin  = "darwin"
)

const (
	dotCharacter = 46
)

// Change represents a single filename change
type Change struct {
	BaseDir string `json:"base_dir"`
	Source  string `json:"source"`
	Target  string `json:"target"`
	IsDir   bool   `json:"is_dir"`
}

// renameError represents an error that occurs when
// renaming a file
type renameError struct {
	entry Change
	err   error
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
	workingDir    string
	stringMode    bool
	excludeFilter []string
	maxDepth      int
	sort          string
	reverseSort   bool
	quiet         bool
	errors        []renameError
	revert        bool
}

type backupFile struct {
	WorkingDir string   `json:"working_dir"`
	Date       string   `json:"date"`
	Operations []Change `json:"operations"`
}

func init() {
	if runtime.GOOS == windows {
		pathSeperator = `\`
	}
}

// createBackupDir creates the directory for backups
// if it doesn't exist already
func createBackupDir(dir string) (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return dirname, os.MkdirAll(filepath.Join(dirname, ".f2", dir), os.ModePerm)
}

// writeToFile writes the details of a successful operation
// to the specified output file, creating it if necessary.
func (op *Operation) writeToFile(outputFile string) (err error) {
	// Create or truncate file
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	defer func() {
		ferr := file.Close()
		if ferr != nil {
			err = ferr
		}
	}()

	mf := backupFile{
		WorkingDir: op.workingDir,
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
// in the specified map file. The undo file is deleted
// if the operation is successfully reverted
func (op *Operation) undo(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var bf backupFile
	err = json.Unmarshal(file, &bf)
	if err != nil {
		return err
	}
	op.matches = bf.Operations

	for i, v := range op.matches {
		ch := v
		ch.Source = v.Target
		ch.Target = v.Source

		op.matches[i] = ch
	}

	if !op.exec && op.sort != "" {
		err = op.sortBy()
		if err != nil {
			return err
		}
	}

	err = op.apply()
	if err != nil {
		return err
	}

	if op.exec {
		if err = os.Remove(path); err != nil {
			fmt.Printf(
				"Unable to remove redundant undo file '%s' after successful operation.",
				yellow.Sprint(path),
			)
		}
	}

	return nil
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
// directories are auto-created if necessary.
// Errors are aggregated instead of being reported one by one
func (op *Operation) rename() {
	var errs []renameError
	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		renameErr := renameError{
			entry: ch,
		}

		// If target contains a slash, create all missing
		// directories before renaming the file
		if strings.Contains(ch.Target, "/") ||
			strings.Contains(ch.Target, `\`) && runtime.GOOS == windows {
			// No need to check if the `dir` exists or if there are several
			// consecutive slashes since `os.MkdirAll` handles that
			dir := filepath.Dir(ch.Target)
			err := os.MkdirAll(filepath.Join(ch.BaseDir, dir), 0750)
			if err != nil {
				renameErr.err = err
				errs = append(errs, renameErr)
				continue
			}
		}

		if err := os.Rename(source, target); err != nil {
			renameErr.err = err
			errs = append(errs, renameErr)
		}
	}

	op.errors = errs
}

// reportErrors displays the errors that occur during a renaming operation
func (op *Operation) reportErrors() {
	var data = make([][]string, len(op.errors)+len(op.matches))
	for i, v := range op.matches {
		source := filepath.Join(v.BaseDir, v.Source)
		target := filepath.Join(v.BaseDir, v.Target)
		d := []string{source, target, green.Sprint("success")}
		data[i] = d
	}

	for i, v := range op.errors {
		source := filepath.Join(v.entry.BaseDir, v.entry.Source)
		target := filepath.Join(v.entry.BaseDir, v.entry.Target)

		msg := v.err.Error()
		msg = strings.TrimSpace(msg[strings.IndexByte(msg, ':'):])
		d := []string{
			source,
			target,
			red.Sprintf("%s", strings.TrimPrefix(msg, ": ")),
		}
		data[i+len(op.matches)] = d
	}

	printTable(data)
}

// handleErrors is used to report the errors and write any successful
// operations to a file
func (op *Operation) handleErrors() error {
	// first remove the error entries from the matches so they are not confused
	// with successful operations
	for _, v := range op.errors {
		target := v.entry.Target
		for j := len(op.matches) - 1; j >= 0; j-- {
			if target == op.matches[j].Target {
				op.matches = append(op.matches[:j], op.matches[j+1:]...)
			}
		}
	}

	op.reportErrors()

	var err error
	if len(op.matches) > 0 && !op.revert {
		err = op.backup()
	}

	if err == nil && len(op.matches) > 0 {
		return fmt.Errorf(
			"Some files could not be renamed. To revert the changes, run %s",
			yellow.Sprint("f2 -u"),
		)
	} else if err != nil && len(op.matches) > 0 {
		return fmt.Errorf("The above files could not be renamed")
	}

	return fmt.Errorf("The renaming operation failed due to the above errors")
}

// backup creates the path where the backup file
// will be written to
func (op *Operation) backup() error {
	workingDir := strings.ReplaceAll(op.workingDir, pathSeperator, "_")
	if runtime.GOOS == windows {
		workingDir = strings.ReplaceAll(workingDir, ":", "_")
	}

	dirname, err := createBackupDir("backups")
	if err != nil {
		return err
	}

	file := workingDir + ".json"

	return op.writeToFile(
		filepath.Join(dirname, ".f2", "backups", file),
	)
}

// apply will check for conflicts and print the changes to be made
// or apply them directly to the filesystem if in execute mode.
// Conflicts will be ignored if indicated
func (op *Operation) apply() error {
	if len(op.matches) == 0 {
		if !op.quiet {
			fmt.Println("Failed to match any files")
		}
		return nil
	}

	op.validate()
	if len(op.conflicts) > 0 && !op.fixConflicts {
		if !op.quiet {
			op.reportConflicts()
		}

		return errConflictDetected
	}

	if op.exec {
		if op.includeDir || op.revert {
			op.sortMatches()
		}

		op.rename()

		if len(op.errors) > 0 {
			return op.handleErrors()
		}

		return op.backup()
	}

	if op.quiet {
		return nil
	}
	op.printChanges()
	fmt.Printf(
		"Append the %s flag to apply the above changes\n",
		yellow.Sprint("-x"),
	)

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

// retrieveBackupFile retrieves the path to a previously created
// backup file for the current directory
func (op *Operation) retrieveBackupFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := strings.ReplaceAll(op.workingDir, pathSeperator, "_")
	if runtime.GOOS == windows {
		dir = strings.ReplaceAll(dir, ":", "_")
	}

	fullPath := filepath.Join(homeDir, ".f2", "backups", dir+".json")

	if _, err := os.Stat(fullPath); err != nil {
		return "", err
	}

	return fullPath, nil
}

// run executes the operation sequence
func (op *Operation) run() error {
	if op.revert {
		path, err := op.retrieveBackupFile()
		if err != nil {
			return fmt.Errorf(
				"Failed to retrieve backup file for the current directory: %w",
				err,
			)
		}
		return op.undo(path)
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

	if op.sort != "" {
		err = op.sortBy()
		if err != nil {
			return err
		}
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
	op.findString = c.String("find")
	op.replacement = c.String("replace")
	op.exec = c.Bool("exec")
	op.fixConflicts = c.Bool("fix-conflicts")
	op.includeDir = c.Bool("include-dir")
	op.includeHidden = c.Bool("hidden")
	op.ignoreCase = c.Bool("ignore-case")
	op.ignoreExt = c.Bool("ignore-ext")
	op.recursive = c.Bool("recursive")
	op.directories = c.Args().Slice()
	op.onlyDir = c.Bool("only-dir")
	op.stringMode = c.Bool("string-mode")
	op.excludeFilter = c.StringSlice("exclude")
	op.maxDepth = c.Int("max-depth")
	op.quiet = c.Bool("quiet")
	op.revert = c.Bool("undo")

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
	if c.String("find") == "" && c.String("replace") == "" && !c.Bool("undo") {
		return nil, errInvalidArgument
	}

	op := &Operation{}
	err := setOptions(op, c)
	if err != nil {
		return nil, err
	}

	if op.revert {
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
