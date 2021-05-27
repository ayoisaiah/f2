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
	// fullWindowsForbiddenRegex is like windowsForbiddenRegex but includes
	// forward and backslashes
	fullWindowsForbiddenRegex = regexp.MustCompile(`<|>|:|"|\||\?|\*|/|\\`)
	// macForbiddenRegex is used to match the strings that contain forbidden
	// characters in macOS' file names.
	macForbiddenRegex = regexp.MustCompile(`:`)
)

const (
	windowsMaxLength = 260
	unixMaxBytes     = 255
)

type conflictType int

const (
	emptyFilename conflictType = iota
	fileExists
	overwritingNewPath
	maxLengthExceeded
	invalidCharacters
	trailingPeriod
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
	if len(match) > 0 {
		_, _ = fmt.Sscanf(match[0], "(%d)", &num)
		num++
	} else {
		f += " (" + strconv.Itoa(num) + ")"
	}

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
				printColor("red", "❌ [Empty filename]"),
			}
			data = append(data, slice)
		}
	}

	if slice, exists := op.conflicts[trailingPeriod]; exists {
		for _, v := range slice {
			for _, s := range v.source {
				slice := []string{
					s,
					v.target,
					printColor(
						"red",
						"❌ [trailing periods are prohibited]",
					),
				}
				data = append(data, slice)
			}
		}
	}

	if slice, exists := op.conflicts[fileExists]; exists {
		for _, v := range slice {
			slice := []string{
				strings.Join(v.source, ""),
				v.target,
				printColor("red", "❌ [Path already exists]"),
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
					printColor("red", "❌ [Overwriting newly renamed path]"),
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
					printColor("red",
						fmt.Sprintf(
							"❌ [Invalid characters present: (%s)]",
							v.cause,
						),
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
					printColor("red",
						fmt.Sprintf(
							"❌ [Maximum file name exceeded: (%s)]",
							v.cause,
						),
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
	op.conflicts = make(map[conflictType][]Conflict)
	m := make(map[string][]struct {
		source string
		index  int
	})

	for i := 0; i < len(op.matches); i++ {
		ch := op.matches[i]
		var source, target = ch.Source, ch.Target
		source = filepath.Join(ch.BaseDir, source)
		target = filepath.Join(ch.BaseDir, target)

		// Report if replacement operation results in
		// an empty string for the new filename
		if ch.Target == "." || ch.Target == "" {
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

		detected := op.checkTrailingPeriodConflict(source, ch.Target, target, i)
		if detected && op.fixConflicts {
			i--
			continue
		}

		detected = op.checkPathLengthConflict(source, ch.Target, target, i)
		if detected && op.fixConflicts {
			i--
			continue
		}

		detected = op.checkForbiddenCharactersConflict(
			source,
			ch.Target,
			target,
			i,
		)
		if detected && op.fixConflicts {
			i--
			continue
		}

		detected = op.checkPathExistsConflict(source, target, ch, i)
		if detected && op.fixConflicts {
			i--
			continue
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

	op.checkOverwritingPathConflict(m)
}

// checkPathExistsConflict reports if the newly renamed path
// already exists on the filesystem
func (op *Operation) checkPathExistsConflict(
	source, target string,
	ch Change,
	i int,
) bool {
	var conflictDetected bool
	// Report if target file exists on the filesystem
	if _, err := os.Stat(target); err == nil ||
		errors.Is(err, os.ErrExist) {
		// Don't report a conflict for an unchanged filename
		// Also handles case-insensitive filesystems
		if strings.EqualFold(source, target) {
			return conflictDetected
		}

		// Don't report a conflict if overwriting files are allowed
		if op.allowOverwrites {
			op.matches[i].WillOverwrite = true
			return conflictDetected
		}

		op.conflicts[fileExists] = append(
			op.conflicts[fileExists],
			Conflict{
				source: []string{source},
				target: target,
			},
		)

		conflictDetected = true

		if op.fixConflicts {
			dir := filepath.Dir(ch.Target)
			base := filepath.Base(ch.Target)
			str := getNewPath(base, ch.BaseDir, nil)
			str = filepath.Join(dir, str)
			op.matches[i].Target = str
		}
	}

	return conflictDetected
}

// checkOverwritingPathConflict ensures that a newly renamed path
// is not overwritten
func (op *Operation) checkOverwritingPathConflict(m map[string][]struct {
	source string
	index  int
}) {
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
				for i := 0; i < len(v); i++ {
					item := v[i]
					if i == 0 {
						continue
					}

					dir := filepath.Dir(op.matches[item.index].Target)
					base := filepath.Base(op.matches[item.index].Target)
					str := getNewPath(base, op.matches[item.index].BaseDir, m)
					str = filepath.Join(dir, str)
					pt := filepath.Join(op.matches[item.index].BaseDir, str)
					if _, ok := m[pt]; !ok {
						m[pt] = []struct {
							source string
							index  int
						}{}
						op.matches[item.index].Target = str
					} else {
						// repeat the last iteration to generate a new path
						op.matches[item.index].Target = str
						i--
						continue
					}
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
			return fmt.Errorf("%s", ":")
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

// checkTrailingPeriods reports if replacement operation results
// in files or sub directories that end in trailing dots
func (op *Operation) checkTrailingPeriodConflict(
	source, target, absTarget string,
	i int,
) bool {
	var conflictDetected bool
	if runtime.GOOS == windows {
		strSlice := strings.Split(target, pathSeperator)
		for _, v := range strSlice {
			s := strings.TrimRight(v, ".")

			if s != v {
				op.conflicts[trailingPeriod] = append(
					op.conflicts[trailingPeriod],
					Conflict{
						source: []string{source},
						target: absTarget,
					},
				)
				conflictDetected = true
				break
			}
		}

		if op.fixConflicts && conflictDetected {
			for j, v := range strSlice {
				s := strings.TrimRight(v, ".")
				strSlice[j] = s
			}

			op.matches[i].Target = strings.Join(strSlice, pathSeperator)
		}
	}

	return conflictDetected
}

func (op *Operation) checkPathLengthConflict(
	source, target, absTarget string,
	i int,
) bool {
	var conflictDetected bool
	err := checkPathLength(target)
	if err != nil {
		op.conflicts[maxLengthExceeded] = append(
			op.conflicts[maxLengthExceeded],
			Conflict{
				source: []string{source},
				target: absTarget,
				cause:  err.Error(),
			},
		)
		conflictDetected = true

		if op.fixConflicts {
			if runtime.GOOS == windows {
				// trim filename so that it's less than 260 characters
				filename := []rune(filepath.Base(target))
				ext := []rune(filepath.Ext(string(filename)))
				f := []rune(filenameWithoutExtension(string(filename)))
				index := windowsMaxLength - len(ext)
				f = f[:index]
				op.matches[i].Target = filepath.Join(string(f), string(ext))
			} else {
				// trim filename so that it's no more than 255 bytes
				filename := filepath.Base(target)
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
			}
		}
	}

	return conflictDetected
}

func (op *Operation) checkForbiddenCharactersConflict(
	source, target, absTarget string,
	i int,
) bool {
	var conflictDetected bool
	err := checkForbiddenCharacters(target)
	if err != nil {
		op.conflicts[invalidCharacters] = append(
			op.conflicts[invalidCharacters],
			Conflict{
				source: []string{source},
				target: absTarget,
				cause:  err.Error(),
			},
		)

		conflictDetected = true

		if op.fixConflicts {
			if runtime.GOOS == windows {
				op.matches[i].Target = windowsForbiddenRegex.ReplaceAllString(
					target,
					"",
				)
			}

			if runtime.GOOS == darwin {
				op.matches[i].Target = strings.ReplaceAll(
					target,
					":",
					"",
				)
			}
		}
	}

	return conflictDetected
}
