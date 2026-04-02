package validate

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
	"github.com/ayoisaiah/f2/v2/internal/pathutil"
	"github.com/ayoisaiah/f2/v2/internal/status"
)

type validationCtx struct {
	seenPaths           map[string]int
	fixConflictsRegex   *regexp.Regexp
	fixConflictsPattern string
	changes             file.Changes
	changeIndex         int
	autoFix             bool
	allowOverwrites     bool
}

func (ctx *validationCtx) updateSeenPaths() {
	change := ctx.changes[ctx.changeIndex]
	if _, ok := ctx.seenPaths[change.TargetPath]; !ok {
		ctx.seenPaths[change.TargetPath] = ctx.changeIndex
	}
}

const (
	// max filename length of 255 characters in Windows.
	windowsMaxFileCharLength = 255
	// max filename length of 255 bytes on Linux and other unix-based OSes.
	unixMaxBytes = 255
)

// newTarget appends a number to the target file name so that it
// does not conflict with an existing path on the filesystem or
// another renamed file. For example: image.png becomes image(1).png.
func newTarget(
	change *file.Change,
	fixConflictsRegex *regexp.Regexp,
	fixConflictsPattern string,
) string {
	counter := 1

	baseName := filepath.Base(change.Target)
	if !change.IsDir {
		baseName = pathutil.StripExtension(baseName)
	}

	// Extract the numbered index at the end of the filename (if any)
	match := fixConflictsRegex.FindStringSubmatch(baseName)

	if len(match) > 0 {
		num, _ := strconv.Atoi(match[1])
		num += counter

		baseName = fixConflictsRegex.ReplaceAllString(
			baseName,
			fmt.Sprintf(fixConflictsPattern, num),
		)
	} else {
		baseName += fmt.Sprintf(fixConflictsPattern, counter)
	}

	target := baseName + filepath.Ext(change.Target)

	return filepath.Join(filepath.Dir(change.Target), target)
}

// checkSourceNotFoundConflict reports if the source file is missing in an
// undo operation. It is automatically fixed by changing the status so that
// the file is skipped when renaming.
func checkSourceNotFoundConflict(
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	if change.Status == status.SourceNotFound {
		conflictDetected = true

		slog.Debug(
			"conflict: source file not found",
			slog.Any("match", change),
		)

		if ctx.autoFix {
			change.Status = status.Ignored
			slog.Debug(
				"auto-fix: change ignored",
				slog.Any("match", change),
			)
		}
	}

	return
}

// checkEmptyFilenameConflict reports if the file renaming has resulted
// in an empty string. This conflict is automatically fixed by leaving
// the filename unchanged.
func checkEmptyFilenameConflict(
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	if change.Target == "." || change.Target == "" {
		conflictDetected = true

		slog.Debug(
			"empty filename detected",
			slog.Any("match", change),
		)

		change.AutoFixTarget("")
		change.Status = status.EmptyFilename

		if ctx.autoFix {
			// The file is left unchanged
			change.AutoFixTarget(change.Source)

			if change.OriginalName == change.Target {
				change.Status = status.Unchanged
			}

			slog.Debug(
				"auto-fix: orignal name restored",
				slog.Any("match", change),
			)
		}
	}

	return
}

// checkPathExistsConflict reports if the newly renamed path
// already exists on the filesystem.
func checkPathExistsConflict(
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	// Report if target path exists on the filesystem
	if _, err := os.Stat(change.TargetPath); err == nil ||
		errors.Is(err, os.ErrExist) {
		// Don't report a conflict for an unchanged filename
		if change.SourcePath == change.TargetPath {
			change.Status = status.Unchanged
			return
		}

		// Case-insensitive filesystems should not report conflicts
		// if only the case of the filename is being changed.
		if strings.EqualFold(
			change.SourcePath,
			change.TargetPath,
		) {
			return
		}

		// Don't report a conflict if overwriting files are allowed
		if ctx.allowOverwrites {
			change.WillOverwrite = true
			change.Status = status.Overwriting

			return
		}

		// Don't report a conflict if target path is changing before
		// the source path is renamed
		for i := 0; i < len(ctx.changes); i++ {
			ch := ctx.changes[i]

			if change.TargetPath == ch.SourcePath &&
				!strings.EqualFold(ch.SourcePath, ch.TargetPath) &&
				ctx.changeIndex > i {
				return
			}
		}

		conflictDetected = true
		change.Status = status.PathExists

		slog.Debug(
			"conflict: target path already exists",
			slog.Any("match", change),
		)

		if ctx.autoFix {
			change.AutoFixTarget(
				newTarget(
					change,
					ctx.fixConflictsRegex,
					ctx.fixConflictsPattern,
				),
			)

			slog.Debug(
				"auto-fix: new target generated",
				slog.Any("match", change),
			)
		}
	}

	return conflictDetected
}

// checkSourceAlreadyRenamedConflict ensures that renaming a file multiple times
// is detected to prevent data loss. It is automatically fixed by swapping the
// items around so that any renaming targets do not change later.
func checkSourceAlreadyRenamedConflict(
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	seenIndex, ok := ctx.seenPaths[change.SourcePath]
	if !ok {
		return
	}

	conflictDetected = true
	change.Status = status.SourceAlreadyRenamed

	slog.Debug(
		"conflict: source has already been renamed",
		slog.Any("match", change),
		slog.Int("match.index", ctx.changeIndex),
		slog.Any("prev_change", ctx.changes[seenIndex]),
		slog.Int("prev_change.index", seenIndex),
	)

	if ctx.autoFix {
		ctx.changes[seenIndex], ctx.changes[ctx.changeIndex] = ctx.changes[ctx.changeIndex], ctx.changes[seenIndex]
		change.Status = status.OK

		slog.Debug("auto-fix: swap change positions",
			slog.Any("match", ctx.changes[seenIndex]),
			slog.Int("match.index", seenIndex),
			slog.Any("prev_change", ctx.changes[ctx.changeIndex]),
			slog.Int("prev_change.index", ctx.changeIndex),
		)
	}

	return
}

// checkOverwritingPathConflict ensures that a newly renamed path
// is not overwritten by another renamed file. Such conflicts are solved by
// appending a number to the filename until no conflict is detected.
func checkOverwritingPathConflict(
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	if i, ok := ctx.seenPaths[change.TargetPath]; ok {
		conflictDetected = true
		change.Status = status.OverwritingNewPath

		slog.Debug(
			"conflict: overwriting renamed file",
			slog.Any("match", change),
			slog.Any("prev_change", ctx.changes[i]),
		)
	}

	if !conflictDetected {
		return
	}

	if ctx.autoFix {
		change.AutoFixTarget(
			newTarget(
				change,
				ctx.fixConflictsRegex,
				ctx.fixConflictsPattern,
			),
		)

		slog.Debug(
			"auto-fix: new target generated",
			slog.Any("match", change),
		)
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
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]
	if runtime.GOOS == osutil.Windows {
		pathComponents := strings.Split(
			change.TargetPath,
			string(os.PathSeparator),
		)

		for _, v := range pathComponents {
			if v == "." || v == ".." {
				continue
			}

			if v != strings.TrimRight(v, ".") {
				conflictDetected = true

				break
			}
		}

		if conflictDetected {
			change.Status = status.TrailingPeriod

			slog.Debug(
				"conflict: trailing period detected",
				slog.Any("match", change),
			)
		}

		if ctx.autoFix && conflictDetected {
			for j, v := range pathComponents {
				if v == "." || v == ".." {
					continue
				}

				s := strings.TrimRight(v, ".")
				pathComponents[j] = s
			}

			change.AutoFixTarget(strings.Join(
				pathComponents,
				string(os.PathSeparator),
			))

			slog.Debug(
				"auto-fix: trailing periods removed",
				slog.Any("match", change),
			)

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
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	exceeded := isTargetLengthExceeded(change.Target)
	if !exceeded {
		return
	}

	conflictDetected = true
	change.Status = status.FilenameLengthExceeded

	slog.Debug(
		"conflict: filename length exceeded",
		slog.Any("match", change),
	)

	if !ctx.autoFix {
		return
	}

	if runtime.GOOS == osutil.Windows {
		// trim filename so that it's less than 255 characters
		filename := []rune(filepath.Base(change.Target))
		ext := []rune(filepath.Ext(string(filename)))
		f := []rune(
			pathutil.StripExtension(string(filename)),
		)
		index := windowsMaxFileCharLength - len(ext)
		f = f[:index]
		change.AutoFixTarget(string(f) + string(ext))

		slog.Debug(
			"auto-fix: trim file name length",
			slog.Any("match", change),
		)

		return
	}

	// trim filename so that it's no more than 255 bytes
	filename := filepath.Base(change.Target)
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

	change.AutoFixTarget(fileNoExt + ext)

	slog.Debug(
		"auto-fix: trim file name length",
		slog.Any("match", change),
	)

	return
}

// checkForbiddenCharactersConflict is used to detect if forbidden characters
// are present in the target path for a file or directory according to the
// naming rules of the respective OS. This detection excludes forward and
// backward slashes as their presence has a special meaning in the renaming
// ration (automatic directory creation).
// Conflicts are automatically fixed by removing the culprit characters.
func checkForbiddenCharactersConflict(
	ctx *validationCtx,
) (conflictDetected bool) {
	change := ctx.changes[ctx.changeIndex]

	forbiddenChars := checkForbiddenCharacters(change.Target)
	if forbiddenChars != "" {
		conflictDetected = true
		change.Status = status.ForbiddenCharacters.Append(forbiddenChars)

		slog.Debug(
			"conflict: forbidden characters detected",
			slog.Any("target", change.Target),
			slog.Any("characters", forbiddenChars),
		)

		if !ctx.autoFix {
			return
		}

		newTarget := change.Target

		if runtime.GOOS == osutil.Windows {
			newTarget = osutil.PartialWindowsForbiddenCharRegex.ReplaceAllString(
				change.Target,
				"",
			)
		}

		if runtime.GOOS == osutil.Darwin {
			newTarget = strings.ReplaceAll(
				change.Target,
				":",
				"",
			)
		}

		change.AutoFixTarget(newTarget)

		slog.Debug(
			"auto-fix: forbidden characters removed",
			slog.Any("match", change),
		)
	}

	return
}

func checkAndHandleConflict(
	ctx *validationCtx,
	loopIndex *int,
) (detected bool) {
	// Slice of conflict-checking functions with consistent signatures
	checks := []func(ctx *validationCtx) bool{
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
func detectConflicts(
	changes file.Changes,
	autoFix, allowOverwrites bool,
	fixConflictsRegex *regexp.Regexp,
	fixConflictsPattern string,
) bool {
	ctx := validationCtx{
		changes:             changes,
		autoFix:             autoFix,
		allowOverwrites:     allowOverwrites,
		fixConflictsRegex:   fixConflictsRegex,
		fixConflictsPattern: fixConflictsPattern,
		seenPaths:           make(map[string]int),
	}

	conflicts := make(map[int]string)

	for i := 0; i < len(ctx.changes); i++ {
		change := ctx.changes[i]

		slog.Debug("checking for conflicts", slog.Any("match", change))

		ctx.changeIndex = i

		detected := checkAndHandleConflict(&ctx, &i)
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
	fixConflictsRegex *regexp.Regexp,
	fixConflictsPattern string,
) bool {
	return detectConflicts(
		matches,
		autoFix,
		allowOverwrites,
		fixConflictsRegex,
		fixConflictsPattern,
	)
}
