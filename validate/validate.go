package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
	"github.com/ayoisaiah/f2/v2/internal/status"
)

type validationCtx struct {
	change          *file.Change
	seenPaths       map[string]int
	changeIndex     int
	autoFix         bool
	allowOverwrites bool
}

func (ctx validationCtx) updateSeenPaths() {
	if _, ok := ctx.seenPaths[ctx.change.TargetPath]; !ok {
		ctx.seenPaths[ctx.change.TargetPath] = ctx.changeIndex
	}
}

var changes file.Changes

const (
	// max filename length of 255 characters in Windows.
	windowsMaxFileCharLength = 255
	// max filename length of 255 bytes on Linux and other unix-based OSes.
	unixMaxBytes = 255
)

// newTarget appends a number to the target file name so that it
// does not conflict with an existing path on the filesystem or
// another renamed file. For example: image.png becomes image(1).png.
func newTarget(change *file.Change) string {
	conf := config.Get()

	counter := 1

	baseName := filepath.Base(change.Target)
	if !change.IsDir {
		baseName = pathutil.StripExtension(baseName)
	}

	regex := conf.FixConflictsPatternRegex

	// Extract the numbered index at the end of the filename (if any)
	match := regex.FindStringSubmatch(baseName)

	if len(match) > 0 {
		num, _ := strconv.Atoi(match[1])
		num += counter

		baseName = regex.ReplaceAllString(
			baseName,
			fmt.Sprintf(conf.FixConflictsPattern, num),
		)
	} else {
		baseName += fmt.Sprintf(conf.FixConflictsPattern, counter)
	}

	target := baseName + filepath.Ext(change.Target)

	return filepath.Join(filepath.Dir(change.Target), target)
}

// checkSourceNotFoundConflict reports if the source file is missing in an
// undo operation. It is automatically fixed by changing the status so that
// the file is skipped when renaming.
func checkSourceNotFoundConflict(
	ctx validationCtx,
) (conflictDetected bool) {
	if ctx.change.Status == status.SourceNotFound {
		conflictDetected = true

		if ctx.autoFix {
			ctx.change.Status = status.Ignored
		}
	}

	return
}

// checkEmptyFilenameConflict reports if the file renaming has resulted
// in an empty string. This conflict is automatically fixed by leaving
// the filename unchanged.
func checkEmptyFilenameConflict(
	ctx validationCtx,
) (conflictDetected bool) {
	if ctx.change.Target == "." || ctx.change.Target == "" {
		conflictDetected = true

		ctx.change.AutoFixTarget("")
		ctx.change.Status = status.EmptyFilename

		if ctx.autoFix {
			// The file is left unchanged
			ctx.change.AutoFixTarget(ctx.change.Source)
			ctx.change.Status = status.Unchanged
		}
	}

	return
}

// checkPathExistsConflict reports if the newly renamed path
// already exists on the filesystem.
func checkPathExistsConflict(
	ctx validationCtx,
) (conflictDetected bool) {
	// Report if target path exists on the filesystem
	if _, err := os.Stat(ctx.change.TargetPath); err == nil ||
		errors.Is(err, os.ErrExist) {
		// Don't report a conflict for an unchanged filename
		if ctx.change.SourcePath == ctx.change.TargetPath {
			ctx.change.Status = status.Unchanged
			return
		}

		// Case-insensitive filesystems should not report conflicts
		// if only the case of the filename is being changed.
		if strings.EqualFold(
			ctx.change.SourcePath,
			ctx.change.TargetPath,
		) {
			return
		}

		// Don't report a conflict if overwriting files are allowed
		if ctx.allowOverwrites {
			ctx.change.WillOverwrite = true
			ctx.change.Status = status.Overwriting

			return
		}

		// Don't report a conflict if target path is changing before
		// the source path is renamed
		for i := 0; i < len(changes); i++ {
			ch := changes[i]

			if ctx.change.TargetPath == ch.SourcePath &&
				!strings.EqualFold(ch.SourcePath, ch.TargetPath) &&
				ctx.changeIndex > i {
				return
			}
		}

		conflictDetected = true
		ctx.change.Status = status.PathExists

		if ctx.autoFix {
			ctx.change.AutoFixTarget(newTarget(ctx.change))
		}
	}

	return conflictDetected
}

// checkSourceAlreadyRenamedConflict ensures that renaming a file multiple times
// is detected to prevent data loss. It is automatically fixed by swapping the
// items around so that any renaming targets do not change later.
func checkSourceAlreadyRenamedConflict(
	ctx validationCtx,
) (conflictDetected bool) {
	seenIndex, ok := ctx.seenPaths[ctx.change.SourcePath]
	if !ok {
		return
	}

	conflictDetected = true
	ctx.change.Status = status.SourceAlreadyRenamed

	if ctx.autoFix {
		changes[seenIndex], changes[ctx.changeIndex] = changes[ctx.changeIndex], changes[seenIndex]
		ctx.change.Status = status.OK
	}

	return
}

// checkOverwritingPathConflict ensures that a newly renamed path
// is not overwritten by another renamed file. Such conflicts are solved by
// appending a number to the filename until no conflict is detected.
func checkOverwritingPathConflict(
	ctx validationCtx,
) (conflictDetected bool) {
	if _, ok := ctx.seenPaths[ctx.change.TargetPath]; ok {
		conflictDetected = true
		ctx.change.Status = status.OverwritingNewPath
	}

	if !conflictDetected {
		return
	}

	if ctx.autoFix {
		ctx.change.AutoFixTarget(newTarget(ctx.change))
	}

	return
}

// checkForbiddenCharacters is responsible for ensuring that target file names
// do not contain forbidden characters for the current OS.
func checkForbiddenCharacters(path string) string {
	if runtime.GOOS == osutil.Windows {
		// partialWindowsForbiddenCharRegex is used here as forward and backward
		// slashes are used for auto creating directories
		if osutil.PartialWindowsForbiddenCharRegex.MatchString(path) {
			return strings.Join(
				osutil.PartialWindowsForbiddenCharRegex.FindAllString(
					path,
					-1,
				),
				",",
			)
		}
	}

	if runtime.GOOS == osutil.Darwin {
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
	if runtime.GOOS == osutil.Windows &&
		len([]rune(filename)) > windowsMaxFileCharLength {
		return true
	}

	if runtime.GOOS != osutil.Windows &&
		len([]byte(filename)) > unixMaxBytes {
		// max length of 255 bytes on Linux and other unix-based OSes
		return true
	}

	return false
}

// checkTrailingPeriodConflictInWindows reports if the file renaming has
// resulted in files or sub directories that end in trailing dots.
// This conflict is automatically resolved by removing the trailing periods.
func checkTrailingPeriodConflictInWindows(
	ctx validationCtx,
) (conflictDetected bool) {
	if runtime.GOOS == osutil.Windows {
		pathComponents := strings.Split(
			ctx.change.TargetPath,
			string(os.PathSeparator),
		)

		for _, v := range pathComponents {
			if v != strings.TrimRight(v, ".") {
				conflictDetected = true

				break
			}
		}

		if conflictDetected {
			ctx.change.Status = status.TrailingPeriod
		}

		if ctx.autoFix && conflictDetected {
			for j, v := range pathComponents {
				s := strings.TrimRight(v, ".")
				pathComponents[j] = s
			}

			ctx.change.AutoFixTarget(strings.Join(
				pathComponents,
				string(os.PathSeparator),
			))

			return
		}
	}

	return
}

// checkFileNameLengthConflict reports if the file renaming has resulted in a
// name that is longer than the acceptable limit (255 characters in Windows and
// 255 bytes on Unix). This conflict is automatically fixed by removing the
// excess characters/bytes until the name is under the limit.
func checkFileNameLengthConflict(
	ctx validationCtx,
) (conflictDetected bool) {
	exceeded := isTargetLengthExceeded(ctx.change.Target)
	if exceeded {
		conflictDetected = true
		ctx.change.Status = status.FilenameLengthExceeded

		if !ctx.autoFix {
			return
		}

		if runtime.GOOS == osutil.Windows {
			// trim filename so that it's less than 255 characters
			filename := []rune(filepath.Base(ctx.change.Target))
			ext := []rune(filepath.Ext(string(filename)))
			f := []rune(
				pathutil.StripExtension(string(filename)),
			)
			index := windowsMaxFileCharLength - len(ext)
			f = f[:index]
			ctx.change.AutoFixTarget(string(f) + string(ext))

			return
		}

		// trim filename so that it's no more than 255 bytes
		filename := filepath.Base(ctx.change.Target)
		ext := filepath.Ext(filename)
		fileNoExt := pathutil.StripExtension(filename)
		index := unixMaxBytes - len([]byte(ext))

		for {
			if len([]byte(fileNoExt)) > index {
				frune := []rune(fileNoExt)
				fileNoExt = string(frune[:len(frune)-1])

				continue
			}

			break
		}

		ctx.change.AutoFixTarget(fileNoExt + ext)
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
	ctx validationCtx,
) (conflictDetected bool) {
	forbiddenChars := checkForbiddenCharacters(ctx.change.Target)
	if forbiddenChars != "" {
		conflictDetected = true
		ctx.change.Status = status.ForbiddenCharacters

		if !ctx.autoFix {
			return
		}

		newTarget := ctx.change.Target

		if runtime.GOOS == osutil.Windows {
			newTarget = osutil.PartialWindowsForbiddenCharRegex.ReplaceAllString(
				ctx.change.Target,
				"",
			)
		}

		if runtime.GOOS == osutil.Darwin {
			newTarget = strings.ReplaceAll(
				ctx.change.Target,
				":",
				"",
			)
		}

		ctx.change.AutoFixTarget(newTarget)
	}

	return
}

func checkAndHandleConflict(ctx validationCtx, loopIndex *int) (detected bool) {
	// Slice of conflict-checking functions with consistent signatures
	checks := []func(ctx validationCtx) bool{
		checkEmptyFilenameConflict,
		checkTrailingPeriodConflictInWindows,
		checkFileNameLengthConflict,
		checkForbiddenCharactersConflict,
		checkPathExistsConflict,
		checkOverwritingPathConflict,
		checkSourceNotFoundConflict,
		checkSourceAlreadyRenamedConflict, // INFO: Needs to be the last check
	}

	for i, check := range checks {
		detected = check(ctx)
		if !detected {
			continue
		}

		if !ctx.autoFix {
			ctx.updateSeenPaths()
			return detected
		}

		if i == len(checks)-1 {
			// INFO: Special handling for checkTargetFileChangingConflict
			// Restart the iteration from the beginning
			*loopIndex = -1

			clear(ctx.seenPaths)
		} else {
			*loopIndex-- // Go back an index for re-checking after fix
		}

		return detected
	}

	return detected
}

// detectConflicts checks the renamed files for various conflicts and
// automatically fixes them if configured.
func detectConflicts(autoFix, allowOverwrites bool) bool {
	ctx := validationCtx{
		autoFix:         autoFix,
		allowOverwrites: allowOverwrites,
		seenPaths:       make(map[string]int),
	}

	conflicts := make(map[int]string)

	for i := 0; i < len(changes); i++ {
		change := changes[i]

		ctx.change = change
		ctx.changeIndex = i

		detected := checkAndHandleConflict(ctx, &i)
		if detected {
			conflicts[ctx.changeIndex] = change.SourcePath
			continue
		}

		delete(conflicts, ctx.changeIndex)

		ctx.updateSeenPaths()
	}

	return len(conflicts) > 0
}

// Validate detects and reports any conflicts that can occur while renaming a
// file. Conflicts are automatically fixed if specified in the program options.
func Validate(
	matches file.Changes,
	autoFix, allowOverwrites bool,
) bool {
	changes = matches

	return detectConflicts(autoFix, allowOverwrites)
}
