// Package validate is used to ensure that the renaming operation cannot result
// in conflicts before the operation is carried out. It protects against the
// following scenarios:
//
// 1. Overwriting a newly renamed path.
// 2. Target destination contains forbidden characters (varies based on the operating system).
// 3. Target destination already exists on the file system (except if
// --allow-overwrite is specified)
// 4. Target name exceeds the maximum allowed length (255 characters in windows, and 255 bytes on Linux and macOS).
// 5. Target destination contains trailing periods in any of the sub paths (Windows only).
// 6. Target destination is empty.
//
// It detects each conflicts and reports them, but it can also automatically fix
// them according to predefined rules (if -F/--fix-conflicts is specified).
package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ayoisaiah/f2/internal/conflict"
	"github.com/ayoisaiah/f2/internal/file"
	internalos "github.com/ayoisaiah/f2/internal/os"
	internalpath "github.com/ayoisaiah/f2/internal/path"
	"github.com/ayoisaiah/f2/internal/status"
)

var conflicts conflict.Collection

var changes []*file.Change

const (
	// max filename length of 255 characters in Windows.
	windowsMaxFileCharLength = 255
	// max filename length of 255 bytes on Linux and other unix-based OSes.
	unixMaxBytes = 255
)

// renamedPathsType is used to detect overwriting file paths
// after the renaming operation. The key of the map
// is the target path.and its slice value must
// have a length of 1, otherwise a conflict will be detected
// for that target path (it means 2 or more source files are
// being renamed to the same target).
type renamedPathsType map[string][]struct {
	sourcePath string
	index      int // helps keep track of source position in the changes slice
}

// newTarget appends a number to the target file name so that it
// does not conflict with an existing path on the filesystem or
// another renamed file. For example: image.png becomes image (2).png.
func newTarget(change *file.Change, renamedPaths map[string][]struct {
	sourcePath string
	index      int
},
) string {
	fileNoExt := internalpath.StripExtension(
		filepath.Base(change.Target),
	)
	regex := regexp.MustCompile(`\(\d+\)$`)
	// Extract the numbered index at the end of the filename (if any)
	match := regex.FindStringSubmatch(fileNoExt)
	num := 2

	if len(match) > 0 {
		_, _ = fmt.Sscanf(match[0], "(%d)", &num)
		num++
	} else {
		fileNoExt += " (" + strconv.Itoa(num) + ")"
	}

	for {
		target := regex.ReplaceAllString(fileNoExt, "("+strconv.Itoa(num)+")")
		target += filepath.Ext(change.Target)
		target = filepath.Join(filepath.Dir(change.Target), target)
		targetPath := filepath.Join(change.BaseDir, target)

		// Ensure the new path does not exist on the filesystem
		if _, err := os.Stat(targetPath); err != nil &&
			errors.Is(err, os.ErrNotExist) {
			for k := range renamedPaths {
				if k == targetPath {
					goto out
				}
			}

			return target
		}
	out:
		num++
	}
}

// checkEmptyFilenameConflict reports if the file renaming has resulted
// in an empty string. This conflict is automatically fixed by leaving
// the filename unchanged.
func checkEmptyFilenameConflict(
	change *file.Change,
	autoFix bool,
) (conflictDetected bool) {
	sourcePath := filepath.Join(change.BaseDir, change.Source)
	targetPath := filepath.Join(change.BaseDir, change.Target)

	if change.Target == "." || change.Target == "" {
		conflictDetected = true

		if autoFix {
			// The file is left unchanged
			change.Target = change.Source
			change.Status = status.Unchanged

			return
		}

		conflicts[conflict.EmptyFilename] = append(
			conflicts[conflict.EmptyFilename],
			conflict.Conflict{
				Sources: []string{sourcePath},
				Target:  targetPath,
			},
		)
		change.Status = status.EmptyFilename
	}

	return
}

// checkPathExistsConflict reports if the newly renamed path
// already exists on the filesystem.
func checkPathExistsConflict(
	change *file.Change,
	autoFix, allowOverwrites bool,
) (conflictDetected bool) {
	sourcePath := filepath.Join(change.BaseDir, change.Source)
	targetPath := filepath.Join(change.BaseDir, change.Target)

	// Report if target path exists on the filesystem
	if _, err := os.Stat(targetPath); err == nil ||
		errors.Is(err, os.ErrExist) {
		// Don't report a conflict for an unchanged filename
		if sourcePath == targetPath {
			change.Status = status.Unchanged
			return
		}

		// Case-insensitive filesystems should not report conflicts
		// if only the case of the filename is being changed.
		if strings.EqualFold(sourcePath, targetPath) {
			return
		}

		// Don't report a conflict if overwriting files are allowed
		if allowOverwrites {
			change.WillOverwrite = true
			change.Status = status.Overwriting

			return
		}

		// Don't report a conflict if target path is changing before
		// the source path is renamed
		for j := 0; j < len(changes); j++ {
			ch := changes[j]
			sp := filepath.Join(ch.BaseDir, ch.Source)
			tp := filepath.Join(ch.BaseDir, ch.Target)

			if targetPath == sp && !strings.EqualFold(sp, tp) &&
				change.Index > j {
				return
			}
		}

		if autoFix {
			change.Target = newTarget(change, nil)
			change.Status = status.OK

			return
		}

		conflicts[conflict.FileExists] = append(
			conflicts[conflict.FileExists],
			conflict.Conflict{
				Sources: []string{sourcePath},
				Target:  targetPath,
			},
		)

		conflictDetected = true
		change.Status = status.PathExists
	}

	return conflictDetected
}

// checkOverwritingPathConflict ensures that a newly renamed path
// is not overwritten by another renamed file. Such conflicts are solved by
// appending a number to the filename until no conflict is detected.
func checkOverwritingPathConflict(
	renamedPaths renamedPathsType,
	autoFix bool,
) {
	// Report duplicate targets if any
	for targetPath, source := range renamedPaths {
		if len(source) > 1 {
			var sources []string

			for _, s := range source {
				sources = append(sources, s.sourcePath)
				changes[s.index].Status = status.OverwritingNewPath
			}

			if autoFix {
				for i := 0; i < len(source); i++ {
					item := source[i]

					if i == 0 {
						changes[item.index].Status = status.OK
						continue
					}

					target := newTarget(
						changes[item.index],
						renamedPaths,
					)
					pt := filepath.Join(changes[item.index].BaseDir, target)

					if _, ok := renamedPaths[pt]; !ok {
						renamedPaths[pt] = []struct {
							sourcePath string
							index      int
						}{}
						changes[item.index].Target = target
						changes[item.index].Status = status.OK
					} else {
						// repeat the last iteration to generate a new path
						changes[item.index].Target = target
						changes[item.index].Status = status.OK
						i--
						continue
					}
				}

				continue
			}

			conflicts[conflict.OverwritingNewPath] = append(
				conflicts[conflict.OverwritingNewPath],
				conflict.Conflict{
					Sources: sources,
					Target:  targetPath,
				},
			)
		}
	}
}

// checkForbiddenCharacters is responsible for ensuring that target file names
// do not contain forbidden characters for the current OS.
func checkForbiddenCharacters(path string) string {
	if runtime.GOOS == internalos.Windows {
		// partialWindowsForbiddenCharRegex is used here as forward and backward
		// slashes are used for auto creating directories
		if internalos.PartialWindowsForbiddenCharRegex.MatchString(path) {
			return strings.Join(
				internalos.PartialWindowsForbiddenCharRegex.FindAllString(
					path,
					-1,
				),
				",",
			)
		}
	}

	if runtime.GOOS == internalos.Darwin {
		if strings.Contains(path, ":") {
			return ":"
		}
	}

	return ""
}

// isTargetLengthExceeded is responsible for ensuring that the target name length
// does not exceed the maximum value on each supported rating system.
func isTargetLengthExceeded(target string) bool {
	// Get the standalone filename
	filename := filepath.Base(target)

	// max length of 255 characters in windows
	if runtime.GOOS == internalos.Windows &&
		len([]rune(filename)) > windowsMaxFileCharLength {
		return true
	}

	if runtime.GOOS != internalos.Windows &&
		len([]byte(filename)) > unixMaxBytes {
		// max length of 255 bytes on Linux and other unix-based OSes
		return true
	}

	return false
}

// checkTrailingPeriods reports if the file renaming has resulted in
// files or sub directories that end in trailing dots (Windows only).
// This conflict is automatically resolved by removing the trailing periods.
func checkTrailingPeriodConflict(
	change *file.Change,
	autoFix bool,
) (conflictDetected bool) {
	sourcePath := filepath.Join(change.BaseDir, change.Source)
	targetPath := filepath.Join(change.BaseDir, change.Target)

	if runtime.GOOS == internalos.Windows {
		pathComponents := strings.Split(change.Target, internalpath.Separator)

		for _, v := range pathComponents {
			if v != strings.TrimRight(v, ".") {
				conflictDetected = true

				break
			}
		}

		if autoFix && conflictDetected {
			for j, v := range pathComponents {
				s := strings.TrimRight(v, ".")
				pathComponents[j] = s
			}

			change.Target = strings.Join(pathComponents, internalpath.Separator)
			change.Status = status.OK
			conflictDetected = false

			return
		}

		if conflictDetected {
			conflicts[conflict.TrailingPeriod] = append(
				conflicts[conflict.TrailingPeriod],
				conflict.Conflict{
					Sources: []string{sourcePath},
					Target:  targetPath,
				},
			)

			change.Status = status.TrailingPeriod
		}
	}

	return
}

// checkFileNameLengthConflict reports if the file renaming has resulted in a
// name that is longer than the acceptable limit (255 characters in Windows and
// 255 bytes on Unix). This conflict is automatically fixed by removing the
// excess characters/bytes until the name is under the limit.
func checkFileNameLengthConflict(
	change *file.Change,
	autoFix bool,
) (conflictDetected bool) {
	sourcePath := filepath.Join(change.BaseDir, change.Source)
	targetPath := filepath.Join(change.BaseDir, change.Target)

	exceeded := isTargetLengthExceeded(change.Target)
	if exceeded {
		if autoFix {
			if runtime.GOOS == internalos.Windows {
				// trim filename so that it's less than 255 characters
				filename := []rune(filepath.Base(change.Target))
				ext := []rune(filepath.Ext(string(filename)))
				f := []rune(
					internalpath.StripExtension(string(filename)),
				)
				index := windowsMaxFileCharLength - len(ext)
				f = f[:index]
				change.Target = string(f) + string(ext)
			} else {
				// trim filename so that it's no more than 255 bytes
				filename := filepath.Base(change.Target)
				ext := filepath.Ext(filename)
				fileNoExt := internalpath.StripExtension(filename)
				index := unixMaxBytes - len([]byte(ext))
				for {
					if len([]byte(fileNoExt)) > index {
						frune := []rune(fileNoExt)
						fileNoExt = string(frune[:len(frune)-1])
						continue
					}

					break
				}

				change.Target = fileNoExt + ext
				change.Status = status.OK
			}

			return
		}

		cause := "255 bytes"
		if runtime.GOOS == internalos.Windows {
			cause = "255 characters"
		}

		conflicts[conflict.MaxFilenameLengthExceeded] = append(
			conflicts[conflict.MaxFilenameLengthExceeded],
			conflict.Conflict{
				Sources: []string{sourcePath},
				Target:  targetPath,
				Cause:   cause,
			},
		)
		conflictDetected = true
		change.Status = status.FilenameLengthExceeded
	}

	return
}

// checkForbiddenCharactersConflict is used to detect if forbidden characters
// are present in the target path for a file or directory according to the
// naming rules of the respective OS. This detection excludes forward and
// backward slashes as their presence has a special meaning in the renaming
// ration (automatic directory creation).
// Conflicts are automatically fixed by removing the culprit characters.
func checkForbiddenCharactersConflict(
	change *file.Change,
	autoFix bool,
) (conflictDetected bool) {
	sourcePath := filepath.Join(change.BaseDir, change.Source)
	targetPath := filepath.Join(change.BaseDir, change.Target)

	forbiddenChars := checkForbiddenCharacters(change.Target)
	if forbiddenChars != "" {
		if autoFix {
			if runtime.GOOS == internalos.Windows {
				change.Target = internalos.PartialWindowsForbiddenCharRegex.ReplaceAllString(
					change.Target,
					"",
				)
			}

			if runtime.GOOS == internalos.Darwin {
				change.Target = strings.ReplaceAll(
					change.Target,
					":",
					"",
				)
			}

			change.Status = status.OK

			return
		}

		conflicts[conflict.InvalidCharacters] = append(
			conflicts[conflict.InvalidCharacters],
			conflict.Conflict{
				Sources: []string{sourcePath},
				Target:  targetPath,
				Cause:   forbiddenChars,
			},
		)

		conflictDetected = true
		change.Status = status.InvalidCharacters
	}

	return
}

// detectConflicts checks the renamed files for various conflicts and
// automatically fixes them if allowed.
func detectConflicts(autoFix, allowOverwrites bool) {
	renamedPaths := make(renamedPathsType)

	for i := 0; i < len(changes); i++ {
		change := changes[i]
		sourcePath := filepath.Join(change.BaseDir, change.Source)
		targetPath := filepath.Join(change.BaseDir, change.Target)

		detected := checkEmptyFilenameConflict(change, autoFix)
		if detected {
			// no need to check for other conflicts here since the filename
			// is empty. If auto fixed, no renaming will occur for the entry
			continue
		}

		detected = checkTrailingPeriodConflict(change, autoFix)
		if detected && autoFix {
			// going back an index allows rechecking the path for conflicts once more
			i--
			continue
		}

		detected = checkFileNameLengthConflict(change, autoFix)
		if detected && autoFix {
			i--
			continue
		}

		detected = checkForbiddenCharactersConflict(change, autoFix)
		if detected && autoFix {
			i--
			continue
		}

		detected = checkPathExistsConflict(change, autoFix, allowOverwrites)
		if detected && autoFix {
			i--
			continue
		}

		renamedPaths[targetPath] = append(renamedPaths[targetPath], struct {
			sourcePath string
			index      int
		}{
			sourcePath: sourcePath,
			index:      i,
		})
	}

	checkOverwritingPathConflict(renamedPaths, autoFix)
}

// Validate detects and reports any conflicts that can occur while renaming a
// file. Conflicts are automatically fixed if specified in the program options.
func Validate(
	matches []*file.Change,
	autoFix, allowOverwrites bool,
) conflict.Collection {
	conflicts = make(conflict.Collection)

	changes = matches

	detectConflicts(autoFix, allowOverwrites)

	return conflicts
}

func GetConflicts() conflict.Collection {
	return conflicts
}
