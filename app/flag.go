package app

import "github.com/urfave/cli/v3"

// supportedDefaultOpts contains flags whose values can be
// overridden through the `F2_DEFAULT_OPTS` environmental variable.
var supportedDefaultOpts = map[string]bool{
	"--clean": true, "-c": true,
	"--exclude": true, "-E": true,
	"--exclude-dir": true,
	"--exec":        true, "-x": true,
	"--exiftool-opts": true,
	"--fix-conflicts": true, "-F": true,
	"--fix-conflicts-pattern": true,
	"--hidden":                true, "-H": true,
	"--ignore-case": true, "-i": true,
	"--ignore-ext": true, "-e": true,
	"--include-dir": true, "-d": true,
	"--json":     true,
	"--no-color": true,
	"--quiet":    true, "-q": true,
	"--recursive": true, "-R": true,
	"--sort":                true,
	"--sortr":               true,
	"--reset-index-per-dir": true,
	"--string-mode":         true, "-s": true,
	"--verbose": true, "-V": true,
}

var (
	flagCSV = &cli.StringFlag{
		Name: "csv",
		Usage: `
		Load a CSV file, and rename according to its contents.`,
		DefaultText: "<path/to/csv/file>",
		TakesFile:   true,
	}

	flagFind = &cli.StringSliceFlag{
		Name:    "find",
		Aliases: []string{"f"},
		Usage: `
    A regular expression pattern used for matching files and directories.
    It accepts the syntax defined by the RE2 standard and defaults to .* 
		if omitted which matches the entire file/directory name.

		When -s/--string-mode is used, this pattern is treated as a literal string.`,
		DefaultText: "<pattern>",
	}

	flagReplace = &cli.StringSliceFlag{
		Name:    "replace",
		Aliases: []string{"r"},
		Usage: `
    The replacement string which replaces each match in the file name.
    It supports capture variables, built-in variables, and exiftool variables.
    If omitted, it defaults to an empty string.`,
		DefaultText: "<string>",
	}

	flagUndo = &cli.BoolFlag{
		Name:    "undo",
		Aliases: []string{"u"},
		Usage: `
		Undo the last renaming operation performed in the current working directory.`,
	}

	flagAllowOverwrites = &cli.BoolFlag{
		Name: "allow-overwrites",
		Usage: `
		Allows the renaming operation to overwrite existing files.
		Caution: Using this option can lead to unrecoverable data loss.`,
	}

	flagClean = &cli.BoolFlag{
		Name:    "clean",
		Aliases: []string{"c"},
		Usage: `
		Clean empty directories that were traversed in a renaming operation.`,
	}

	flagExclude = &cli.StringSliceFlag{
		Name:    "exclude",
		Aliases: []string{"E"},
		Usage: `
		Excludes files and directories that match the provided regular expression.
		This flag can be repeated to specify multiple exclude patterns.

		Example: 
			-E 'json' -E 'yml' (filters out JSON and YAML files)
			-E 'json|yaml' (equivalent to the above)

		Note: 
			This does not prevent recursing into matching directories (use
			--exclude-dir instead).`,

		DefaultText: "<pattern>",
	}

	flagExcludeDir = &cli.StringSliceFlag{
		Name: "exclude-dir",
		Usage: `
		Prevents F2 from recursing into directories that match the provided regular
		expression pattern.`,
		DefaultText: "<pattern>",
	}

	flagExec = &cli.BoolFlag{
		Name:    "exec",
		Aliases: []string{"x"},
		Usage: `
		Executes the renaming operation and applies the changes to the filesystem.`,
	}

	flagExiftoolOpts = &cli.StringFlag{
		Name: "exiftool-opts",
		Usage: `
		Provides options to customize Exiftool's output when using ExifTool
		variables in replacement patterns.

		Supported options:
			--api
			--charset
			--coordFormat
			--dateFormat
			--extractEmbedded

		Example:
			$ f2 -r '{xt.GPSDateTime}' --exiftool-opts '--dateFormat %Y-%m-%d'`,
	}

	flagFixConflicts = &cli.BoolFlag{
		Name:    "fix-conflicts",
		Aliases: []string{"F"},
		Usage: `
		Automatically fixes renaming conflicts using predefined rules.`,
	}

	flagFixConflictsPattern = &cli.StringFlag{
		Name: "fix-conflicts-pattern",
		Usage: `
		Specifies a custom pattern for renaming files when conflicts occur.
		The pattern should be a valid Go format string containing a single '%d'
		placeholder for the conflict index.

		Example: '_%02d'  (generates _01, _02, etc.)

		If not specified, the default pattern '(%d)' is used.`,
	}

	flagHidden = &cli.BoolFlag{
		Name:    "hidden",
		Aliases: []string{"H"},
		Usage: `
		Includes hidden files and directories in the search and renaming process.

		On Linux and macOS, hidden files are those that start with a dot character.
		On Windows, only files with the 'hidden' attribute are considered hidden.

		To match hidden directories as well, combine this with the -d/--include-dir
		flag.`,
	}

	flagInclude = &cli.StringSliceFlag{
		Name:    "include",
		Aliases: []string{"I"},
		Usage: `
		Only includes files that match the provided regular expression instead of 
		all files matched by the --find flag.

		This flag can be repeated to specify multiple include patterns.

		Example: 
			-I 'json' -I 'yml' (only include JSON and YAML files)`,
	}

	flagIncludeDir = &cli.BoolFlag{
		Name:    "include-dir",
		Aliases: []string{"d"},
		Usage: `
		Includes matching directories in the renaming operation (they are excluded
		by default).`,
	}

	flagIgnoreCase = &cli.BoolFlag{
		Name:    "ignore-case",
		Aliases: []string{"i"},
		Usage: `
		Ignores case sensitivity when searching for matches.`,
	}

	flagIgnoreExt = &cli.BoolFlag{
		Name:    "ignore-ext",
		Aliases: []string{"e"},
		Usage: `
		Ignores the file extension when searching for matches.`,
	}

	flagJSON = &cli.BoolFlag{
		Name: "json",
		Usage: `
		Produces JSON output, except for error messages which are sent to the
		standard error.`,
	}

	flagMaxDepth = &cli.UintFlag{
		Name:    "max-depth",
		Aliases: []string{"m"},
		Usage: `
		Limits the depth of recursive search. Set to 0 (default) for no limit.`,
		Value:       0,
		DefaultText: "<integer>",
	}

	flagNoColor = &cli.BoolFlag{
		Name: "no-color",
		Usage: `
		Disables colored output.`,
	}

	flagOnlyDir = &cli.BoolFlag{
		Name:    "only-dir",
		Aliases: []string{"D"},
		Usage: `
		Renames only directories, not files (implies -d/--include-dir).`,
	}

	flagPair = &cli.BoolFlag{
		Name:    "pair",
		Aliases: []string{"p"},
		Usage: `
		Enable pair renaming to rename files with the same name (but different 
		extensions) in the same directory to the same new name. In pair mode,
		file extensions are ignored.

		Example:
			Before: DSC08533.ARW DSC08533.JPG DSC08534.ARW DSC08534.JPG

			$ f2 -r "Photo_{%03d}" --pair -x

			After: Photo_001.ARW Photo_001.JPG Photo_002.ARW Photo_002.JPG`,
	}

	flagPairOrder = &cli.StringFlag{
		Name: "pair-order",
		Usage: `
		Order the paired files according to their extension. This helps you control 
		the file to be renamed first, and whose metadata should be extracted when
		using variables.

		Example:
		  --pair-order 'dng,jpg' # rename dng files before jpg
		  --pair-order 'xmp,arw' # rename xmp files before arw`,
	}

	flagQuiet = &cli.BoolFlag{
		Name:    "quiet",
		Aliases: []string{"q"},
		Usage: `
		Don't print anything to stdout. If no matches are found, f2 will exit with
	  an error code instead of the normal success code without this flag.
		Errors will continue to be written to stderr.`,
	}

	flagRecursive = &cli.BoolFlag{
		Name:    "recursive",
		Aliases: []string{"R"},
		Usage: `
		Recursively traverses directories when searching for matches.`,
	}

	flagReplaceLimit = &cli.IntFlag{
		Name:    "replace-limit",
		Aliases: []string{"l"},
		Usage: `
		Limits the number of replacements made on each matched file. 0 (default)
		means replace all matches. Negative values replace from the end of the
		filename.`,
		Value:       0,
		DefaultText: "<integer>",
	}

	flagResetIndexPerDir = &cli.BoolFlag{
		Name: "reset-index-per-dir",
		Usage: `
		Resets the auto-incrementing index when entering a new directory during a
		recursive operation.`,
	}

	flagSort = &cli.StringFlag{
		Name: "sort",
		Usage: `
		Sorts matches in ascending order based on the provided criteria.

    Allowed values:
      * 'default'    : Lexicographical order.
      * 'size'       : Sort by file size.
      * 'natural'    : Sort according to natural order.
      * 'mtime'      : Sort by file last modified time.
      * 'btime'      : Sort by file creation time.
      * 'atime'      : Sort by file last access time.
      * 'ctime'      : Sort by file metadata last change time.
      * 'time_var'   : Sort by time variable.
      * 'int_var'    : Sort by integer variable.
      * 'string_var' : Sort lexicographically by string variable.`,
		DefaultText: "<sort>",
	}

	flagSortr = &cli.StringFlag{
		Name: "sortr",
		Usage: `
		Accepts the same values as --sort but sorts matches in descending order.`,
		DefaultText: "<sort>",
	}

	flagSortPerDir = &cli.BoolFlag{
		Name: "sort-per-dir",
		Usage: `
		Ensures sorting is performed separately within each directory rather than
		globally.`,
	}

	flagSortVar = &cli.StringFlag{
		Name: "sort-var",
		Usage: `
		Active when using --sort/--sortr with time_var, int_var, or string_var.
		Provide a supported variable to sort the files based on file metadata.
		See https://f2.freshman.tech/guide/sorting for more details.`,
	}

	flagStringMode = &cli.BoolFlag{
		Name:    "string-mode",
		Aliases: []string{"s"},
		Usage: `
		Treats the search pattern (specified by -f/--find) as a literal string
		instead of a regular expression.`,
	}

	flagTargetDir = &cli.StringFlag{
		Name:    "target-dir",
		Aliases: []string{"t"},
		Usage: `
		Specify a target directory to move renamed files and reorganize your 
		filesystem.`,
	}

	flagVerbose = &cli.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"V"},
		Usage: `
		Enables verbose output during the renaming operation.`,
	}
)
