package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"gopkg.in/gookit/color.v1"
)

// Operation represents a bulk rename operation
type Operation struct {
	paths           []string
	matches         []string
	newPaths        map[string]string
	replaceString   string
	templateString  string
	exec            bool
	ignoreConflicts bool
	searchRegex     *regexp.Regexp
}

// Apply will check for conflicts and print
// the changes to be made or apply them directly
// if in execute mode. Conflicts may be ignored
func (op *Operation) Apply() error {
	if !op.ignoreConflicts {
		err := op.CheckForConflicts()
		if err != nil {
			return err
		}
	}

	green := color.FgGreen.Render
	for p, v := range op.newPaths {
		if op.exec {
			os.Rename(p, v)
		} else {
			fmt.Println(p, "➟", green(v), "✅")
		}
	}

	return nil
}

// FindMatches locates matches for the find pattern
// in each filename. Directory names are exempted
func (op *Operation) FindMatches() error {
	for _, f := range op.paths {
		isDir, err := isDirectory(f)
		if err != nil {
			return err
		}

		if isDir {
			continue
		}

		filename := filepath.Base(f)
		matched := op.searchRegex.MatchString(filename)
		if matched {
			op.matches = append(op.matches, f)
		}
	}

	return nil
}

// CheckForConflicts ensures that there are no conflicts
// after renaming a file
func (op *Operation) CheckForConflicts() error {
	m := make(map[string][]string)
	for k, v := range op.newPaths {
		// Ensure file does not exist on the filesystem
		if _, err := os.Stat(v); err == nil {
			return fmt.Errorf("Unable to rename '%s' to '%s'. File exists", k, v)
		} else if os.IsExist(err) {
			return fmt.Errorf("Unable to rename '%s' to '%s'. File exists", k, v)
		}

		// Ensure that duplicates are not present after renaming
		if _, exists := m[v]; exists {
			m[v] = append(m[v], k)
		} else {
			m[v] = []string{k}
		}
	}

	red := color.FgRed.Render
	green := color.FgGreen.Render
	var err error
	for k, v := range m {
		if len(v) > 1 {
			if err == nil {
				err = fmt.Errorf("%s", red("Potential errors detected: overwriting newly renamed path"))
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

// UseTemplate renames files using a template
func (op *Operation) UseTemplate() {
	for _, f := range op.matches {
		fileName, dir := filepath.Base(f), filepath.Dir(f)
		var slice []string
		slice = append(slice, strings.Split(op.templateString, "|")...)
		for i, str := range slice {
			if str == "og" {
				slice[i] = strings.TrimSuffix(fileName, filepath.Ext(fileName))
			}
		}
		str := strings.Join(slice, "")
		op.newPaths[f] = filepath.Join(dir, str)
	}
}

// Replace replaces the matched text in each path with the
// replacement string
func (op *Operation) Replace() {
	for _, f := range op.matches {
		fileName, dir := filepath.Base(f), filepath.Dir(f)
		str := op.searchRegex.ReplaceAllString(fileName, op.replaceString)
		op.newPaths[f] = filepath.Join(dir, str)
	}
}

// NewOperation returns an Operation constructed
// from command line arguments
func NewOperation(c *cli.Context) (*Operation, error) {
	op := &Operation{}
	op.paths = c.Args().Slice()
	op.replaceString = c.String("replace")
	op.exec = c.Bool("exec")
	op.ignoreConflicts = c.Bool("force")
	op.newPaths = make(map[string]string)
	op.templateString = c.String("template")

	findPattern := c.String("find")

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return nil, fmt.Errorf("Malformed regular expression for search pattern %s", findPattern)
	}

	op.searchRegex = re

	// If paths are omitted, default to the file names in the
	// current directory
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
