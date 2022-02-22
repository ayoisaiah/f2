## 1.8.0 (2022-02-22)

- Add a "simple mode" for quick renaming operations in the current directory.
- Ignore extension flag no longer affects directory names.
- Default options can be changed via `F2_DEFAULT_OPTS`.
- Fix duplicate paths when traversing directories to prevent unnecessary errors.
- Output is now sorted in dry-run mode.
- Update notification is now opt-in via `F2_UPDATE_NOTIFIER`.
- Version information is now more streamlined.
- Dry run table output has been improved and made more compact.

## 1.7.2 (2021-08-23)

Fixes:

- Path not specified error in Windows when running on a long path has been fixed.

## 1.7.1 (2021-08-05)

The following enhancements were made:

- Quiet mode (`--quiet`) no longer suppresses errors.
- Help output has been improved and made more succinct.
- Running F2 without arguments now shows a short help message.

## 1.7.0 (2021-08-04)

This release brings the following improvements:

- CSV support (See [renaming from a CSV file](https://github.com/ayoisaiah/f2/wiki/Renaming-from-a-CSV-file)).
- Improved no color options. You can now set the `F2_NO_COLOR` environmental variable or use the brand new `--no-color` flag to disable coloured output.
- Validation error messages are now much clearer (no longer mixing emoji and text).
- Console output has been improved by using [pterm](https://github.com/pterm/pterm). The green, red, and yellow colours are slightly different now due to this change.
- You can now specify a set of files or directory as argument to F2 (thanks to [nightson](https://github.com/nightson) for suggesting this enhancement).
- The backup directory has changed to `~/.local/share/f2/backups` on Linux, `~/Library/Application Support/f2/backups` on macOS, and `%LOCALAPPDATA%\f2\backups` on Windows. The previous backup directory (`~/.f2/backups`) is still supported (in case you have existing backups there), but new backup files will not be created there anymore. This change was made to conform to the [XDG specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) and conventions for storing application files on each operating system, and to avoid cluttering up the home directory.
- A `--verbose` option was added for outputting each renaming operation in `--exec` mode.

## 1.6.7 (2021-06-06)

Features:

- String literal mode now supports operation chaining

## 1.6.6 (2021-05-29)

Feature enhancements:

- Fix rare bug where using `--fix-conflicts` could cause an existing file to be overwritten.

## 1.6.5 (2021-05-26)

Feature enhancements:

- Add the ability to force overwriting files through the `--allow-overwrites` option.

## 1.6.4 (2021-05-24)

Feature enhancements:

- You can now chain several renaming operations by specifying the `--find` and `--replace` flags multiple times.

## 1.6.3 (2021-05-20)

Feature enhancements:

- String transformation variables have been changed to match all other built-in variables.

## 1.6.2 (2021-05-16)

Feature enhancements:

- Auto fixing conflicts is now more reliable especially when overwriting newly renamed paths.
- Trailing periods in a file or sub directory name is now detected as a conflict (Windows only).

## 1.6.1 (2021-05-08)

Feature enhancements:

- Bug fixes for exif variables (prevent potential panic).
- Improve `--help` output.
- Improve error message when reverting an operation.
- Replace slashes in `exiftool` output to prevent inadvertent directory creation.

## 1.6.0 (2021-05-07)

Feature enhancements:

- Add `exiftool` support
- Improve built-in exif variables

## 1.5.9 (2021-05-06)

Fixes:

- Bug fixes for Exif variables

## 1.5.8 (2021-05-05)

Feature enhancements:

- Add option to remove diacritics in file names so that `žůžo` becomes `zuzo`.
- Fix some minor bugs with undo mode.

## 1.5.7 (2021-05-04)

Feature enhancements:

- Respect the `NO_COLOR` environmental variable.
- Handle case insensitive filesystems correctly so that changes such as `abc.txt` -> `ABC.txt` do not produce conflicts.

## 1.5.6 (2021-05-04)

Feature enhancements:

- Replacements can now start from the end of the file name by passing a negative number to `--replace-limit`. For example, `-2` will replace the last 2 matches in the file name.

## 1.5.5 (2021-05-04)

Feature enhancements:

- Add the `-l` or `--replace-limit` option for limiting replacements.
- String literal mode now provides the correct output when used with `--ignore-case`.

## 1.5.4 (2021-05-04)

Feature enhancements:

- Updated syntax for string transformation.
- Exif variables no longer output extra text for greater control. For example: `{{exif.et}}` gives `1_10` instead of `1_10s`.

## 1.5.3 (2021-04-29)

Feature enhancements:

- Critical fix for undo mode where it wouldn't find the backup file for the current directory.
- Sorting matches in alphabetical order by default in dry-run mode.
- Improve performance when using built-in variables in the replacement string.

## 1.5.2 (2021-04-26)

Feature enhancements:

- Fix roman numeral format for numbers over 3999.
- Slight change to randomise variable syntax.

## 1.5.1 (2021-04-26)

Feature enhancements:

- Update syntax for random string variable.
- Ensure multiple instances of a random string variable is replaced correctly.
- Add support for file hash variables: `sha1`, `md5`, `sha256`, and `sha512`.

## 1.5.0 (2021-04-26)

Feature enhancements:

- Add a `--quiet` option so that F2 will not output any info to the standard output (including errors).
- Improved colours and support coloured output on Windows.
- Add support for ID3 and random string variables.
- Add sorting options. You can sort by file size and date attributes.
- Conflict detection is now much improved. F2 will now check if the filename contains invalid characters or if it exceeds the maximum allowed length.
- Backups for each operation is now automated. `--undo` no longer takes a file argument, and the `--output-file` flag is deprecated and removed.
- The backup file is now deleted automatically after a successful reversion.

## 1.4.0 (2021-04-14)

Feature enhancements:

- Auto create necessary directories when using backward slash (\) in replacement string (windows only).
- Full support for exif variables (JPEG, DNG, and most camera RAW formats).
- Add ability to limit max depth when searching recursively (`--max-depth` or `-m`).

## 1.3.0 (2021-03-27)

Feature enhancements:

- Proper support for hidden files and directories on Windows.
- Filter out matched files with the `--exclude` or `-E` flag.
- String mode now works correctly when the find pattern is empty (the whole string is replaced).

## 1.2.2 (2021-03-11)

Feature enhancements:

- String-literal mode was previously using a regex to find matches, but has now been corrected to an ordinary string.
- String-literal mode supports case insensitive mode (`-i` or `--ignore-case`).

## 1.2.1 (2021-03-11)

Fixes:

- Auto fixing conflicts is now more reliable.
- Failure to match any files no longer causes the program to exit with an error.

## 1.2.0 (2021-03-09)

Features:

- Implement date variables: (`ctime`, `atime`, `mtime`, e.t.c).
- Add support for EXIF related variables for images.
- Add string literal mode.

## 1.1.1 (2021-02-24)

- Minor fixes

## 1.1.0 (2021-02-24)

Feature enhancements:

- The `--force` flag has been renamed to `--fix-conflicts`. The short version remains `-F`.
- F2 will no longer overwrite files even if `-F` is used. Instead, it will differentiate conflicting files by appending a number suffix similar to how file managers work.

## 1.0.1 (2021-02-22)

Fixes:

- Remove unnecessary version number prefix.

## 1.0.0 (2021-02-22)

Features:

- Supports filtering files using regular expression, including capture groups.
- Ignores hidden directories and files by default.
- Dry run by default.
- Detects potential conflicts such as file collisions, or overwrites.
- Supports recursive renaming of both files and directories.
- Supports using an ascending integer for renaming (e.g 001, 002, 003, e.t.c.).
- Supports undoing an operation from a map file.

## 0.2.0 (2020-05-26)

Features:

- Undo last successful operation
- Specify starting index for numbering scheme

## 0.1.0 (2020-05-24)

Initial release
