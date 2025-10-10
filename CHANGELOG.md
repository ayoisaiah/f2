# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/).

## [Unreleased](https://github.com/ayoisaiah/f2/compare/2.2.1...HEAD)

## [2.2.1] - 2025-09-14

### Fixed

- Ensure regex searches don't incorrectly trigger find expressions.

## [2.2.0] - 2025-09-08

### Added

- Ability to select files using
  [find expressions](https://f2.freshman.tech/guide/find-expressions).
- `--dt` to make entering date tokens much easier.
- `--timezone` for shifting datetime variables to a specific timezone.
- `--replace-range` to enable even more targeted replacements.
- xxHash algorithm for file hashes.

### Changed

- Performance is improved by ~2x when using `exiftool` variables.
- Fully translated to French, German, Russian, Chinese, Portuguese, and Spanish.

### Fixed

- Applied fix for upstream bug which led to unintentional space trimming.

## [2.1.2] - 2025-05-27

### Fixed

- `--target-dir` now works with relative paths in Windows.

## [2.1.1] - 2025-05-16

### Fixed

- UNC path bug in Windows.

## [2.1.0] - 2025-05-01

### Added

- `--include` flag for targeting specific files.
- Unicode normalization transform variable (`{.norm}`).

### Changed

- Improve diacritics transformation (`{.di}`).
- Improve target changing error message.

### Fixed

- Natural sort bug.
- Crash when using capture variable indices.

## [2.0.3] - 2024-11-23

### Fixed

- Bug caused by space trimming in `--find`, `--replace`, `--exclude`, and
  `--exclude-dir`.

## [2.0.2] - 2024-11-22

### Fixed

- Using commas correctly in find or replace strings.

## [2.0.1] - 2024-11-06

### Changed

- Patch release to update the Go module to v2. No new features.

## [2.0.0] - 2024-11-03

### Added

- `natural` sort option for sorting filenames containing numbers naturally.
- Ability to extract double extensions like `.tar.gz` using `{2ext}`.
- `--exiftool-opts` to customize Exiftool's output.
- `--exclude-dir` to exclude whole directories when matching files.
- Custom pattern for fixing conflicts with `--fix-conflicts-pattern`.
- Support for [file pair renaming](https://f2.freshman.tech/guide/pair-renaming)
  via `--pair` and `--pair-order`.
- `--target-dir` to specify a target directory for renamed files.
- `--clean` for cleaning up empty directories after renaming.
- Arbitrary-input sorting via `--sort` and `--sort-var`.
- Per-directory sorting with `--sort-per-dir`.
- Reset index per directory with `--reset-index-per-dir`.

### Changed

- Significant performance improvements (up to ~3× faster).
- Improved conflict detection with validations:
  - When the source file is not found.
  - When the target name changes later in the same operation.
- Cleaner output reporting.
- Improvements to `--undo`.
- Improved CSV renaming implementation.
- Improved help messages and documentation.

### Fixed

- Status reporting for unchanged files.
- Dotfiles incorrectly regarded as hidden in Windows.
- Piping file arguments from external commands.
- Windows-specific bugs with CSV renaming.

### Removed

- Random variables.
- Simple mode.

## [1.9.1] - 2023-02-09

### Changed

- Improved performance in dry-run mode (table rendering).

### Fixed

- Files could be overwritten when `--fix-conflicts` is used.

## [1.9.0] - 2023-02-02

### Added

- Capture variables with indexing.
- JSON support.
- Ability to extract date from arbitrary strings.

### Changed

- Simple mode now prompts before executing.
- Syntax for string transformation improved and simplified.
- Variables syntax simplified.
- Improved support for case-insensitive filesystems.
- Indexing fixes and syntax updates.

## [1.8.0] - 2022-02-22

### Added

- "Simple mode" for quick renaming operations in the current directory.
- Default options can be changed via `F2_DEFAULT_OPTS`.

### Changed

- Ignore extension flag no longer affects directory names.
- Fixed duplicate paths when traversing directories to prevent unnecessary
  errors.
- Output is now sorted in dry-run mode.
- Update notification is now opt-in via `F2_UPDATE_NOTIFIER`.
- Version information streamlined.
- Dry-run table output improved and made more compact.

## [1.7.2] - 2021-08-23

### Fixed

- "Path not specified" error in Windows when running on a long path.

## [1.7.1] - 2021-08-05

### Changed

- Quiet mode (`--quiet`) no longer suppresses errors.
- Help output improved and made more succinct.
- Running F2 without arguments now shows a short help message.

## [1.7.0] - 2021-08-04

### Added

- CSV support (see
  [renaming from a CSV file](https://github.com/ayoisaiah/f2/wiki/Renaming-from-a-CSV-file)).
- `--verbose` option for outputting each renaming operation in `--exec` mode.

### Changed

- Improved no-color options: set `F2_NO_COLOR` or use `--no-color`.
- Validation error messages are now clearer (no longer mixing emoji and text).
- Console output improved via [pterm](https://github.com/pterm/pterm). Colors
  slightly adjusted.
- You can now specify a set of files or directories as arguments to F2.
- Backup directory changed to:
  - Linux: `~/.local/share/f2/backups`
  - macOS: `~/Library/Application Support/f2/backups`
  - Windows: `%LOCALAPPDATA%\f2\backups`  
    Previous backup directory (`~/.f2/backups`) is still supported for reads.

## [1.6.7] - 2021-06-06

### Added

- String literal mode now supports operation chaining.

## [1.6.6] - 2021-05-29

### Fixed

- Rare bug where `--fix-conflicts` could cause an existing file to be
  overwritten.

## [1.6.5] - 2021-05-26

### Added

- `--allow-overwrites` to force overwriting files.

## [1.6.4] - 2021-05-24

### Added

- Chain several renaming operations by specifying `--find` and `--replace`
  multiple times.

## [1.6.3] - 2021-05-20

### Changed

- String transformation variables updated to match all other built-in variables.

## [1.6.2] - 2021-05-16

### Changed

- Auto-fixing conflicts is more reliable, especially when overwriting newly
  renamed paths.

### Fixed

- Trailing periods in a file or subdirectory name detected as a conflict
  (Windows only).

## [1.6.1] - 2021-05-08

### Changed

- Improved `--help` output.
- Improved error message when reverting an operation.
- Replace slashes in `exiftool` output to prevent inadvertent directory
  creation.

### Fixed

- Bug fixes for EXIF variables (prevent potential panic).

## [1.6.0] - 2021-05-07

### Added

- `exiftool` support.
- Improved built-in EXIF variables.

## [1.5.9] - 2021-05-06

### Fixed

- Bug fixes for EXIF variables.

## [1.5.8] - 2021-05-05

### Added

- Option to remove diacritics in file names (e.g., `žůžo` → `zuzo`).

### Fixed

- Minor bugs with undo mode.

## [1.5.7] - 2021-05-04

### Changed

- Respect the `NO_COLOR` environment variable.
- Handle case-insensitive filesystems correctly (e.g., `abc.txt` → `ABC.txt`
  without conflicts).

## [1.5.6] - 2021-05-04

### Added

- Negative values for `--replace-limit` to start replacements from the end
  (e.g., `-2` replaces the last two matches).

## [1.5.5] - 2021-05-04

### Added

- `-l` / `--replace-limit` to limit replacements.

### Fixed

- String literal mode provides correct output with `--ignore-case`.

## [1.5.4] - 2021-05-04

### Changed

- Updated syntax for string transformation.
- EXIF variables no longer output extra text (e.g., `{{exif.et}}` now `1_10`
  instead of `1_10s`).

## [1.5.3] - 2021-04-29

### Changed

- Sort matches alphabetically by default in dry-run mode.
- Improved performance when using built-in variables in the replacement string.

### Fixed

- Critical fix for undo mode finding the backup file for the current directory.

## [1.5.2] - 2021-04-26

### Fixed

- Roman numeral format for numbers over 3999.
- Slight change to randomize variable syntax.

## [1.5.1] - 2021-04-26

### Added

- Support for file hash variables: `sha1`, `md5`, `sha256`, `sha512`.

### Changed

- Update syntax for random string variable.

### Fixed

- Ensure multiple instances of a random string variable are replaced correctly.

## [1.5.0] - 2021-04-26

### Added

- `--quiet` to suppress all output (including errors).
- Support for ID3 and random string variables.
- Sorting options: by file size and date attributes.

### Changed

- Improved colors; support colored output on Windows.
- Conflict detection improved: invalid characters and max-length checks.
- Backups automated; `--undo` no longer takes a file argument; `--output-file`
  removed.
- Backup file is deleted automatically after a successful reversion.

## [1.4.0] - 2021-04-14

### Added

- Auto-create necessary directories when using backslash (`\`) in replacement
  string (Windows only).
- Full support for EXIF variables (JPEG, DNG, and most camera RAW formats).
- Limit max depth when searching recursively (`--max-depth` / `-m`).

## [1.3.0] - 2021-03-27

### Added

- Proper support for hidden files and directories on Windows.
- Filter out matched files with `--exclude` / `-E`.
- String mode works when find pattern is empty (whole string is replaced).

## [1.2.2] - 2021-03-11

### Changed

- String-literal mode now uses ordinary string matching (previously regex).
- String-literal mode supports case-insensitive matches (`-i` /
  `--ignore-case`).

## [1.2.1] - 2021-03-11

### Fixed

- Auto-fixing conflicts is more reliable.
- Failure to match any files no longer exits with an error.

## [1.2.0] - 2021-03-09

### Added

- Date variables (`ctime`, `atime`, `mtime`, etc.).
- EXIF-related variables for images.
- String-literal mode.

## [1.1.1] - 2021-02-24

### Fixed

- Minor fixes.

## [1.1.0] - 2021-02-24

### Changed

- `--force` renamed to `--fix-conflicts` (short `-F` unchanged).
- F2 no longer overwrites files even with `-F`; conflicting files get number
  suffixes like common file managers.

## [1.0.1] - 2021-02-22

### Fixed

- Remove unnecessary version number prefix.

## [1.0.0] - 2021-02-22

### Added

- Filter files using regular expressions, including capture groups.
- Ignore hidden directories and files by default.
- Dry run by default.
- Detect potential conflicts (collisions/overwrites).
- Recursive renaming of files and directories.
- Ascending integer renaming (e.g., `001`, `002`, `003`, …).
- Undo an operation from a map file.

## [0.2.0] - 2020-05-26

### Added

- Undo last successful operation.
- Specify starting index for numbering scheme.

## [0.1.0] - 2020-05-24

### Added

- Initial release.

[2.2.1]: https://github.com/ayoisaiah/f2/compare/v2.2.0...v2.2.1
[2.2.0]: https://github.com/ayoisaiah/f2/compare/v2.1.2...v2.2.0
[2.1.2]: https://github.com/ayoisaiah/f2/compare/v2.1.1...v2.1.2
[2.1.1]: https://github.com/ayoisaiah/f2/compare/v2.1.0...v2.1.1
[2.1.0]: https://github.com/ayoisaiah/f2/compare/v2.0.3...v2.1.0
[2.0.3]: https://github.com/ayoisaiah/f2/compare/v2.0.2...v2.0.3
[2.0.2]: https://github.com/ayoisaiah/f2/compare/v2.0.1...v2.0.2
[2.0.1]: https://github.com/ayoisaiah/f2/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/ayoisaiah/f2/compare/v1.9.1...v2.0.0
[1.9.1]: https://github.com/ayoisaiah/f2/compare/v1.9.0...v1.9.1
[1.9.0]: https://github.com/ayoisaiah/f2/compare/v1.8.0...v1.9.0
[1.8.0]: https://github.com/ayoisaiah/f2/compare/v1.7.2...v1.8.0
[1.7.2]: https://github.com/ayoisaiah/f2/compare/v1.7.1...v1.7.2
[1.7.1]: https://github.com/ayoisaiah/f2/compare/v1.7.0...v1.7.1
[1.7.0]: https://github.com/ayoisaiah/f2/compare/v1.6.7...v1.7.0
[1.6.7]: https://github.com/ayoisaiah/f2/compare/v1.6.6...v1.6.7
[1.6.6]: https://github.com/ayoisaiah/f2/compare/v1.6.5...v1.6.6
[1.6.5]: https://github.com/ayoisaiah/f2/compare/v1.6.4...v1.6.5
[1.6.4]: https://github.com/ayoisaiah/f2/compare/v1.6.3...v1.6.4
[1.6.3]: https://github.com/ayoisaiah/f2/compare/v1.6.2...v1.6.3
[1.6.2]: https://github.com/ayoisaiah/f2/compare/v1.6.1...v1.6.2
[1.6.1]: https://github.com/ayoisaiah/f2/compare/v1.6.0...v1.6.1
[1.6.0]: https://github.com/ayoisaiah/f2/compare/v1.5.9...v1.6.0
[1.5.9]: https://github.com/ayoisaiah/f2/compare/v1.5.8...v1.5.9
[1.5.8]: https://github.com/ayoisaiah/f2/compare/v1.5.7...v1.5.8
[1.5.7]: https://github.com/ayoisaiah/f2/compare/v1.5.6...v1.5.7
[1.5.6]: https://github.com/ayoisaiah/f2/compare/v1.5.5...v1.5.6
[1.5.5]: https://github.com/ayoisaiah/f2/compare/v1.5.4...v1.5.5
[1.5.4]: https://github.com/ayoisaiah/f2/compare/v1.5.3...v1.5.4
[1.5.3]: https://github.com/ayoisaiah/f2/compare/v1.5.2...v1.5.3
[1.5.2]: https://github.com/ayoisaiah/f2/compare/v1.5.1...v1.5.2
[1.5.1]: https://github.com/ayoisaiah/f2/compare/v1.5.0...v1.5.1
[1.5.0]: https://github.com/ayoisaiah/f2/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/ayoisaiah/f2/compare/v1.3.0...v1.4.0
[1.3.0]: https://github.com/ayoisaiah/f2/compare/v1.2.2...v1.3.0
[1.2.2]: https://github.com/ayoisaiah/f2/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/ayoisaiah/f2/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/ayoisaiah/f2/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/ayoisaiah/f2/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/ayoisaiah/f2/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/ayoisaiah/f2/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/ayoisaiah/f2/compare/v0.2.0...v1.0.0
[0.2.0]: https://github.com/ayoisaiah/f2/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ayoisaiah/f2/releases/tag/v0.1.0
