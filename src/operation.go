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

	"github.com/urfave/cli/v2"
)

var (
	errInvalidArgument = errors.New(
		"Invalid argument: one of `-f`, `-r` or `-u` must be present and set to a non empty string value\nUse 'f2 --help' for more information",
	)

	errConflictDetected = fmt.Errorf(
		"Resolve conflicts before proceeding or use the %s flag to auto fix all conflicts",
		printColor("yellow", "-F"),
	)
)

const (
	windows = "windows"
	darwin  = "darwin"
)

const (
	dotCharacter = 46
)

// Change represents a single filename change.
type Change struct {
	index          int
	originalSource string
	BaseDir        string `json:"base_dir"`
	Source         string `json:"source"`
	Target         string `json:"target"`
	IsDir          bool   `json:"is_dir"`
	WillOverwrite  bool   `json:"-"`
}

// renameError represents an error that occurs when
// renaming a file.
type renameError struct {
	entry Change
	err   error
}

// Operation represents a batch renaming operation.
type Operation struct {
	paths             []Change
	matches           []Change
	conflicts         map[conflictType][]Conflict
	findSlice         []string
	replacement       string
	replacementSlice  []string
	startNumber       int
	exec              bool
	fixConflicts      bool
	includeHidden     bool
	includeDir        bool
	onlyDir           bool
	ignoreCase        bool
	ignoreExt         bool
	searchRegex       *regexp.Regexp
	directories       []string
	recursive         bool
	workingDir        string
	stringLiteralMode bool
	excludeFilter     []string
	maxDepth          int
	sort              string
	reverseSort       bool
	quiet             bool
	errors            []renameError
	revert            bool
	numberOffset      []int
	replaceLimit      int
	allowOverwrites   bool
}

type backupFile struct {
	WorkingDir string   `json:"working_dir"`
	Date       string   `json:"date"`
	Operations []Change `json:"operations"`
}

// createBackupDir creates the directory for backups
// if it doesn't exist already.
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
// if the operation is successfully reverted.
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

	// Sort only in print mode
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
				printColor("yellow", path),
			)
		}
	}

	return nil
}

// printChanges displays the changes to be made in a
// table format.
func (op *Operation) printChanges() {
	var data = make([][]string, len(op.matches))

	for i, v := range op.matches {
		source := filepath.Join(v.BaseDir, v.Source)
		target := filepath.Join(v.BaseDir, v.Target)

		status := printColor("green", "ok")
		if source == target {
			status = printColor("yellow", "unchanged")
		}

		if v.WillOverwrite {
			status = printColor("yellow", "overwriting")
		}

		d := []string{source, target, status}
		data[i] = d
	}

	printTable(data)
}

// rename iterates over all the matches and renames them on the filesystem
// directories are auto-created if necessary.
// Errors are aggregated ins""tead of being reported one by one.
func (op *Operation) rename() {
	var errs []renameError

	renamed := []Change{}

	for _, ch := range op.matches {
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		// skip unchanged file names
		if source == target {
			continue
		}

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

		renamed = append(renamed, ch)
	}

	op.matches = renamed
	op.errors = errs
}

// reportErrors displays the errors that occur during a renaming operation.
func (op *Operation) reportErrors() {
	var data = make([][]string, len(op.errors)+len(op.matches))

	for i, v := range op.matches {
		source := filepath.Join(v.BaseDir, v.Source)
		target := filepath.Join(v.BaseDir, v.Target)
		d := []string{source, target, printColor("green", "success")}
		data[i] = d
	}

	for i, v := range op.errors {
		source := filepath.Join(v.entry.BaseDir, v.entry.Source)
		target := filepath.Join(v.entry.BaseDir, v.entry.Target)

		msg := v.err.Error()
		if strings.IndexByte(msg, ':') != -1 {
			msg = strings.TrimSpace(msg[strings.IndexByte(msg, ':'):])
		}

		d := []string{
			source,
			target,
			printColor("red", strings.TrimPrefix(msg, ": ")),
		}
		data[i+len(op.matches)] = d
	}

	printTable(data)
}

// handleErrors is used to report the errors and write any successful
// operations to a file.
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

	msg := fmt.Sprintf(
		"Some files could not be renamed. To revert the changes, run: %s",
		printColor("yellow", "f2 -u"),
	)
	if op.revert {
		msg = "Some files could not be reverted. See above table for the full explanation."
	}

	if err == nil && len(op.matches) > 0 {
		return fmt.Errorf(msg)
	} else if err != nil && len(op.matches) > 0 {
		return fmt.Errorf("The above files could not be renamed")
	}

	return fmt.Errorf("The renaming operation failed due to the above errors")
}

// backup creates the path where the backup file
// will be written to.
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

// noMatches prints out a message if the renaming operation
// failed to match any files.
func (op *Operation) noMatches() {
	msg := "Failed to match any files"
	if op.revert {
		msg = "No operations to undo"
	}

	if !op.quiet {
		fmt.Println(msg)
	}
}

// execute applies the renaming operation to the filesystem.
// A backup file is auto created as long as at least one file
// was renamed and it wasn't an undo operation.
func (op *Operation) execute() error {
	if op.includeDir || op.revert {
		op.sortMatches()
	}

	op.rename()

	if len(op.errors) > 0 {
		return op.handleErrors()
	}

	if len(op.matches) > 0 && !op.revert {
		return op.backup()
	}

	if !op.quiet && !op.revert {
		fmt.Println("No files were renamed")
	}

	return nil
}

// dryRun prints the changes to be made to the standard output.
func (op *Operation) dryRun() {
	if !op.quiet {
		op.printChanges()
		fmt.Printf(
			"Append the %s flag to apply the above changes\n",
			printColor("yellow", "-x"),
		)
	}
}

// apply prints the changes to be made in dry-run mode
// or commits the operation to the filesystem if in execute mode.
// If conflicts are detected, the operation is aborted and the conflicts
// are printed out so that they may be corrected by the user.
func (op *Operation) apply() error {
	if len(op.matches) == 0 {
		op.noMatches()
		return nil
	}

	op.detectConflicts()

	if len(op.conflicts) > 0 && !op.fixConflicts {
		if !op.quiet {
			op.reportConflicts()
		}

		return errConflictDetected
	}

	if op.exec {
		return op.execute()
	}

	op.dryRun()

	return nil
}

// findMatches locates matches for the search pattern
// in each filename. Hidden files and directories are exempted
// by default.
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

		matched := op.searchRegex.MatchString(f)
		if matched {
			op.matches = append(op.matches, v)
		}
	}

	return nil
}

// filterMatches excludes any files or directories that match
// the find pattern in accordance with the provided exclude pattern.
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

// setPaths creates a Change struct for each path.
func (op *Operation) setPaths(paths map[string][]os.DirEntry) {
	if op.exec {
		if !indexRegex.MatchString(op.replacement) {
			op.paths = op.sortPaths(paths, false)
			return
		}
	}

	// Don't bother sorting the paths in alphabetical order
	// if a different sort has been set that's not the default
	if op.sort != "" && op.sort != "default" {
		op.paths = op.sortPaths(paths, false)
		return
	}

	op.paths = op.sortPaths(paths, true)
}

// retrieveBackupFile retrieves the path to a previously created
// backup file for the current directory.
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

// handleReplacementChain is ensures that each find
// and replace operation (single or chained) is handled correctly.
func (op *Operation) handleReplacementChain() error {
	for i, v := range op.replacementSlice {
		op.replacement = v

		err := op.replace()
		if err != nil {
			return err
		}

		for j, ch := range op.matches {
			// Update the source to the target from the previous replacement
			// in preparation for the next replacement
			if i != len(op.replacementSlice)-1 {
				op.matches[j].Source = ch.Target
			}

			// After the last replacement, update the Source
			// back to the original
			if i > 0 && i == len(op.replacementSlice)-1 {
				op.matches[j].Source = ch.originalSource
			}
		}

		if i != len(op.replacementSlice)-1 {
			err := op.setFindStringRegex(i + 1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// run executes the operation sequence.
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

	err = op.handleReplacementChain()
	if err != nil {
		return err
	}

	return op.apply()
}

// setFindStringRegex compiles a regular expression for the
// find string of the corresponding replacement index (if any).
// Otherwise, the created regex will match the entire file name.
func (op *Operation) setFindStringRegex(replacementIndex int) error {
	// findPattern is set to match the entire file name by default
	// except if a find string for the corresponding replacement index
	// is found
	findPattern := ".*"
	if len(op.findSlice) > replacementIndex {
		findPattern = op.findSlice[replacementIndex]

		// Escape all regular expression metacharacters in string literal mode
		if op.stringLiteralMode {
			findPattern = regexp.QuoteMeta(findPattern)
		}

		if op.ignoreCase {
			findPattern = "(?i)" + findPattern
		}
	}

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return err
	}

	op.searchRegex = re

	return nil
}

// setOptions applies the command line arguments
// onto the operation.
func setOptions(op *Operation, c *cli.Context) error {
	op.findSlice = c.StringSlice("find")
	op.replacementSlice = c.StringSlice("replace")
	op.exec = c.Bool("exec")
	op.fixConflicts = c.Bool("fix-conflicts")
	op.includeDir = c.Bool("include-dir")
	op.includeHidden = c.Bool("hidden")
	op.ignoreCase = c.Bool("ignore-case")
	op.ignoreExt = c.Bool("ignore-ext")
	op.recursive = c.Bool("recursive")
	op.directories = c.Args().Slice()
	op.onlyDir = c.Bool("only-dir")
	op.stringLiteralMode = c.Bool("string-mode")
	op.excludeFilter = c.StringSlice("exclude")
	op.maxDepth = int(c.Uint("max-depth"))
	op.quiet = c.Bool("quiet")
	op.revert = c.Bool("undo")
	op.allowOverwrites = c.Bool("allow-overwrites")
	op.replaceLimit = c.Int("replace-limit")

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

	return op.setFindStringRegex(0)
}

// walk is used to navigate directories recursively
// and include their contents in the pool of paths in
// which to find matches. It respects the following properties
// set on the operation: whether hidden files should be
// included, and the maximum depth limit (0 for no limit).
// The paths argument is modified in place.
func (op *Operation) walk(paths map[string][]os.DirEntry) error {
	var recursedPaths []string

	var currentDepth int

	// currentLevel represents the current level of directories
	// and their contents
	var currentLevel = make(map[string][]os.DirEntry)

loop:
	// The goal of each iteration is to created entries for each
	// unaccounted directory in the current level
	for dir, dirContents := range paths {
		if contains(recursedPaths, dir) {
			continue
		}

		if !op.includeHidden {
			var err error
			dirContents, err = removeHidden(dirContents, dir)
			if err != nil {
				return err
			}
		}

		for _, entry := range dirContents {
			if entry.IsDir() {
				fp := filepath.Join(dir, entry.Name())
				dirEntry, err := os.ReadDir(fp)
				if err != nil {
					return err
				}

				currentLevel[fp] = dirEntry
			}
		}

		recursedPaths = append(recursedPaths, dir)
	}

	// if there are directories in the current level
	// store each directory entry and empty the
	// currentLevel so that it may be repopulated
	if len(currentLevel) > 0 {
		for dir, dirContents := range currentLevel {
			paths[dir] = dirContents

			delete(currentLevel, dir)
		}

		currentDepth++
		if !(op.maxDepth > 0 && currentDepth == op.maxDepth) {
			goto loop
		}
	}

	return nil
}

// newOperation returns an Operation constructed
// from command line flags & arguments.
func newOperation(c *cli.Context) (*Operation, error) {
	if len(c.StringSlice("find")) == 0 &&
		len(c.StringSlice("replace")) == 0 &&
		!c.Bool("undo") {
		return nil, errInvalidArgument
	}

	op := &Operation{}

	err := setOptions(op, c)
	if err != nil {
		return nil, err
	}

	// Get the current working directory
	op.workingDir, err = filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	// If reverting an operation, no need to walk through directories
	if op.revert {
		return op, nil
	}

	var paths = make(map[string][]os.DirEntry)
	for _, v := range op.directories {
		paths[v], err = os.ReadDir(v)
		if err == nil {
			continue
		}

		var f os.FileInfo

		f, err = os.Stat(v)
		if err != nil {
			return nil, err
		}

		dir := filepath.Dir(v)

		var dirEntry []os.DirEntry

		dirEntry, err = os.ReadDir(dir)
		if err != nil {
			return nil, err
		}

	entryLoop:
		for _, entry := range dirEntry {
			if entry.Name() == f.Name() {
				// Ensure that the file is not already
				// present in the directory entry
				for _, e := range paths[dir] {
					if e.Name() == f.Name() {
						break entryLoop
					}
				}

				paths[dir] = append(paths[dir], entry)

				break
			}
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
		err = op.walk(paths)
		if err != nil {
			return nil, err
		}
	}

	op.setPaths(paths)

	return op, nil
}
