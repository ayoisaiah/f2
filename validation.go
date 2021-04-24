package f2

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

var (
	// windowsForbiddenRegex is used to match the strings that contain forbidden
	// characters in Windows' file names. This does not include also forbidden
	// forward and back slash characters because their presence will cause a new
	// directory to be created
	windowsForbiddenRegex = regexp.MustCompile(`<|>|:|"|\||\?|\*`)
)

const (
	windowsMaxLength = 260
	unixMaxBytes     = 255
)

type conflict int

const (
	emptyFilename conflict = iota
	fileExists
	overwritingNewPath
	maxLengthExceeded
	invalidCharacters
)

// Conflict represents a renaming operation conflict
// such as duplicate targets or empty filenames
type Conflict struct {
	source []string
	target string
	cause  string
}

// getNewPath returns a filename based on the target
// which is not available due to it existing on the filesystem
// or when another renamed file shares the same path.
// It appends an increasing number to the target path until it finds one
// that does not conflict with the filesystem or with another renamed
// file
func getNewPath(target, baseDir string, m map[string][]struct {
	source string
	index  int
}) string {
	f := filenameWithoutExtension(filepath.Base(target))
	re := regexp.MustCompile(`\(\d+\)$`)
	// Extract the numbered index at the end of the filename (if any)
	match := re.FindStringSubmatch(f)
	num := 2
	if len(match) == 0 {
		match = []string{"(" + strconv.Itoa(num) + ")"}
		f += " (" + strconv.Itoa(num) + ")"
	}
	// ignoring error from Sscanf. num will be set to 2 regardless
	_, _ = fmt.Sscanf(match[0], "(%d)", &num)
	for {
		newPath := re.ReplaceAllString(f, "("+strconv.Itoa(num)+")")
		newPath += filepath.Ext(target)
		fullPath := filepath.Join(baseDir, newPath)

		// Ensure the new path does not exist on the filesystem
		if _, err := os.Stat(fullPath); err != nil &&
			errors.Is(err, os.ErrNotExist) {
			for k := range m {
				if k == fullPath {
					goto out
				}
			}
			return newPath
		}
	out:
		num++
	}
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

	if slice, exists := op.conflicts[invalidCharacters]; exists {
		for _, v := range slice {
			for _, s := range v.source {
				slice := []string{
					s,
					v.target,
					red.Sprintf(
						"❌ [Invalid characters present: (%s)]",
						v.cause,
					),
				}
				data = append(data, slice)
			}
		}
	}

	if slice, exists := op.conflicts[maxLengthExceeded]; exists {
		for _, v := range slice {
			for _, s := range v.source {
				slice := []string{
					s,
					v.target,
					red.Sprintf(
						"❌ [Maximum file name exceeded: (%s)]",
						v.cause,
					),
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
			errors.Is(err, os.ErrExist) {
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

// checkForbiddenCharacters is responsible for ensuring that the file names
// do not contain forbidden characters
func checkForbiddenCharacters(path string) error {
	if runtime.GOOS == windows {
		if windowsForbiddenRegex.MatchString(path) {
			return errors.New(
				strings.Join(
					windowsForbiddenRegex.FindAllString(path, -1),
					",",
				),
			)
		}
	}

	if runtime.GOOS == darwin {
		if strings.Contains(path, ":") {
			return fmt.Errorf(
				"a file name cannot contain the colon character",
			)
		}
	}

	return nil
}

// checkPathLength is responsible for ensuring that the filename length
// does not exceed the maximum value on each supported operating system
func checkPathLength(path string) error {
	// Get the standalone filename
	filename := filepath.Base(path)

	// max length of 260 characters in windows
	if runtime.GOOS == windows &&
		len([]rune(filename)) > windowsMaxLength {
		return fmt.Errorf("%d characters", windowsMaxLength)
	} else if runtime.GOOS != windows && len([]byte(filename)) > unixMaxBytes {
		// max length of 255 bytes on Linux and other unix-based OSes
		return fmt.Errorf("%d bytes", unixMaxBytes)
	}

	return nil
}

// runChecks provides additional validation to detect OS specific problems
func (op *Operation) runChecks() {
	for i, ch := range op.matches {
		source := filepath.Join(ch.BaseDir, ch.Source)
		target := filepath.Join(ch.BaseDir, ch.Target)

		err := checkPathLength(ch.Target)
		if err != nil {
			op.conflicts[maxLengthExceeded] = append(
				op.conflicts[maxLengthExceeded],
				Conflict{
					source: []string{source},
					target: target,
					cause:  err.Error(),
				},
			)

			if op.fixConflicts {
				if runtime.GOOS == windows {
					// trim filename so that it's less than 260 characters
					filename := []rune(filepath.Base(ch.Target))
					ext := []rune(filepath.Ext(string(filename)))
					f := []rune(filenameWithoutExtension(string(filename)))
					index := windowsMaxLength - len(ext)
					f = f[:index]
					op.matches[i].Target = filepath.Join(string(f), string(ext))
					ch.Target = filepath.Join(string(f), string(ext))
				} else {
					// trim filename so that it's no more than 255 bytes
					filename := filepath.Base(ch.Target)
					ext := filepath.Ext(filename)
					f := filenameWithoutExtension(filename)
					index := unixMaxBytes - len([]byte(ext))
					for {
						if len([]byte(f)) > index {
							frune := []rune(f)
							f = string(frune[:len(frune)-1])
							continue
						}

						break
					}

					op.matches[i].Target = filepath.Join(f, ext)
					ch.Target = filepath.Join(f, ext)
				}
			}
		}

		err = checkForbiddenCharacters(ch.Target)
		if err != nil {
			op.conflicts[invalidCharacters] = append(
				op.conflicts[invalidCharacters],
				Conflict{
					source: []string{source},
					target: target,
					cause:  err.Error(),
				},
			)

			if op.fixConflicts {
				if runtime.GOOS == windows {
					op.matches[i].Target = windowsForbiddenRegex.ReplaceAllString(
						ch.Target,
						"",
					)
				}

				if runtime.GOOS == darwin {
					op.matches[i].Target = strings.ReplaceAll(
						ch.Target,
						":",
						"",
					)
				}
			}
		}
	}
}

// validate tries to prevent common renaming problems by analyzing the list
// of files and target destinations
func (op *Operation) validate() {
	op.detectConflicts()
	op.runChecks()
}
