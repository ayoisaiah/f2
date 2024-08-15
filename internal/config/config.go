// Package config is responsible for setting the program config from
// command-line arguments
package config

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

var (
	defaultFixConflictsPattern      = "(%d)"
	defaultFixConflictsPatternRegex = regexp.MustCompile(`\((\d+)\)$`)
	customFixConfictsPatternRegex   = regexp.MustCompile(`^(\D?(%(\d+)?d)\D?)$`)
)

var conf *Config

// ExiftoolOpts defines supported options for customizing Exitool's output.
type ExiftoolOpts struct {
	API             string `long:"api"             json:"api"`              // corresponds to the `-api` flag
	Charset         string `long:"charset"         json:"charset"`          // corresponds to the `-charset` flag
	CoordFormat     string `long:"coordFormat"     json:"coord_format"`     // corresponds to the `-coordFormat` flag
	DateFormat      string `long:"dateFormat"      json:"date_format"`      // corresponds to the `-dateFormat` flag
	ExtractEmbedded bool   `long:"extractEmbedded" json:"extract_embedded"` // corresponds to the `-extractEmbedded` flag
}

type Search struct {
	Regex *regexp.Regexp `json:"regex"`
	// Replacement index
	Index int `json:"index"`
}

// Config represents the program configuration.
type Config struct {
	Date                     time.Time      `json:"date"`
	ExcludeDirRegex          *regexp.Regexp `json:"exclude_dir_regex"`
	ExcludeRegex             *regexp.Regexp `json:"exclude_regex"`
	Search                   *Search        `json:"search_regex"`
	FixConflictsPatternRegex *regexp.Regexp `json:"fix_conflicts_pattern_regex"`
	Sort                     Sort           `json:"sort"`
	Replacement              string         `json:"replacement"`
	WorkingDir               string         `json:"working_dir"`
	FixConflictsPattern      string         `json:"fix_conflicts_pattern"`
	CSVFilename              string         `json:"csv_filename"`
	ExiftoolOpts             ExiftoolOpts   `json:"exiftool_opts"`
	ReplacementSlice         []string       `json:"replacement_slice"`
	FilesAndDirPaths         []string       `json:"files_and_dir_paths"`
	FindSlice                []string       `json:"find_slice"`
	MaxDepth                 int            `json:"max_depth"`
	StartNumber              int            `json:"start_number"`
	ReplaceLimit             int            `json:"replace_limit"`
	AllowOverwrites          bool           `json:"allow_overwrites"`
	ReverseSort              bool           `json:"reverse_sort"`
	OnlyDir                  bool           `json:"only_dir"`
	Revert                   bool           `json:"revert"`
	IncludeDir               bool           `json:"include_dir"`
	IgnoreExt                bool           `json:"ignore_ext"`
	IgnoreCase               bool           `json:"ignore_case"`
	Verbose                  bool           `json:"verbose"`
	IncludeHidden            bool           `json:"include_hidden"`
	Quiet                    bool           `json:"quiet"`
	NoColor                  bool           `json:"no_color"`
	AutoFixConflicts         bool           `json:"auto_fix_conflicts"`
	Exec                     bool           `json:"exec"`
	StringLiteralMode        bool           `json:"string_literal_mode"`
	JSON                     bool           `json:"json"`
	Interactive              bool           `json:"interactive"`
	Print                    bool           `json:"non_interactive"`
	Debug                    bool           `json:"debug"`
	Recursive                bool           `json:"recursive"`
	ResetIndexPerDir         bool           `json:"reset_index_per_dir"`
	SortPerDir               bool           `json:"sort_per_dir"`
}

// SetFindStringRegex compiles a regular expression for the
// find string of the corresponding replacement index (if any).
// Otherwise, the created regex will match the entire file name.
func (c *Config) SetFindStringRegex(replacementIndex int) error {
	// findPattern is set to match the entire file name by default
	// except if a find string for the corresponding replacement index
	// is found
	findPattern := ".*"
	if len(c.FindSlice) > replacementIndex {
		findPattern = c.FindSlice[replacementIndex]

		// Escape all regular expression metacharacters in string literal mode
		if c.StringLiteralMode {
			findPattern = regexp.QuoteMeta(findPattern)
		}

		if c.IgnoreCase {
			findPattern = "(?i)" + findPattern
		}
	}

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return err
	}

	c.Search = &Search{
		Regex: re,
		Index: replacementIndex,
	}

	return nil
}

func (c *Config) setOptions(ctx *cli.Context) error {
	if len(ctx.StringSlice("find")) == 0 &&
		len(ctx.StringSlice("replace")) == 0 &&
		ctx.String("csv") == "" &&
		!ctx.Bool("undo") {
		return errInvalidArgument
	}

	c.FindSlice = ctx.StringSlice("find")
	c.ReplacementSlice = ctx.StringSlice("replace")
	c.CSVFilename = ctx.String("csv")
	c.Revert = ctx.Bool("undo")
	c.Debug = ctx.Bool("debug")
	c.FilesAndDirPaths = ctx.Args().Slice()
	c.Print = ctx.Bool("print")

	if len(ctx.Args().Slice()) > 0 {
		c.FilesAndDirPaths = ctx.Args().Slice()
	}

	// Default to the current working directory if no path arguments are provided
	if len(c.FilesAndDirPaths) == 0 {
		c.FilesAndDirPaths = append(c.FilesAndDirPaths, ".")
	}

	// Ensure that each findString has a corresponding replacement.
	// The replacement defaults to an empty string if unset
	for len(c.FindSlice) > len(c.ReplacementSlice) {
		c.ReplacementSlice = append(c.ReplacementSlice, "")
	}

	return c.SetFindStringRegex(0)
}

// setDefaultOpts applies any options that may be set through
// F2_DEFAULT_OPTS.
func (c *Config) setDefaultOpts(ctx *cli.Context) error {
	c.AutoFixConflicts = ctx.Bool("fix-conflicts")
	c.IncludeDir = ctx.Bool("include-dir")
	c.IncludeHidden = ctx.Bool("hidden")
	c.IgnoreCase = ctx.Bool("ignore-case")
	c.IgnoreExt = ctx.Bool("ignore-ext")
	c.Recursive = ctx.Bool("recursive")
	c.OnlyDir = ctx.Bool("only-dir")
	c.StringLiteralMode = ctx.Bool("string-mode")
	c.MaxDepth = int(ctx.Uint("max-depth"))
	c.Verbose = ctx.Bool("verbose")
	c.AllowOverwrites = ctx.Bool("allow-overwrites")
	c.ReplaceLimit = ctx.Int("replace-limit")
	c.Quiet = ctx.Bool("quiet")
	c.JSON = ctx.Bool("json")
	c.Exec = ctx.Bool("exec")
	c.Interactive = ctx.Bool("interactive")
	c.FixConflictsPattern = ctx.String("fix-conflicts-pattern")
	c.ResetIndexPerDir = ctx.Bool("reset-index-per-dir")
	c.SortPerDir = ctx.Bool("sort-per-dir")
	c.NoColor = ctx.Bool("no-color")

	if c.FixConflictsPattern == "" {
		c.FixConflictsPattern = defaultFixConflictsPattern
		c.FixConflictsPatternRegex = defaultFixConflictsPatternRegex
	} else if !customFixConfictsPatternRegex.MatchString(c.FixConflictsPattern) {
		return errParsingFixConflictsPattern.Fmt(c.FixConflictsPattern)
	}

	excludePattern := ctx.StringSlice("exclude")
	if len(excludePattern) > 0 {
		excludeMatchRegex, err := regexp.Compile(
			strings.Join(excludePattern, "|"),
		)
		if err != nil {
			return err
		}

		c.ExcludeRegex = excludeMatchRegex
	}

	excludeDirPattern := ctx.StringSlice("exclude-dir")
	if len(excludeDirPattern) > 0 {
		excludeDirMatchRegex, err := regexp.Compile(
			strings.Join(excludeDirPattern, "|"),
		)
		if err != nil {
			return err
		}

		c.ExcludeDirRegex = excludeDirMatchRegex
	}

	if c.JSON {
		c.Interactive = false
	}

	if c.Interactive {
		c.Exec = true
	}

	if c.OnlyDir {
		c.IncludeDir = true
	}

	// Sorting
	var err error
	if ctx.String("sort") != "" {
		c.Sort, err = parseSortArg(ctx.String("sort"))
		if err != nil {
			return err
		}
	} else if ctx.String("sortr") != "" {
		c.Sort, err = parseSortArg(ctx.String("sortr"))
		if err != nil {
			return err
		}

		c.ReverseSort = true
	}

	if ctx.String("exiftool-opts") != "" {
		args, err := shellquote.Split(ctx.String("exiftool-opts"))
		if err != nil {
			return err
		}

		_, err = flags.ParseArgs(&c.ExiftoolOpts, args)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsATTY checks if the given file descriptor is associated with a terminal.
func IsATTY(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// Get retrives an already set config or panics if the configuration
// has not yet been initialized.
func Get() *Config {
	if conf == nil {
		panic("config has not been initialized")
	}

	return conf
}

// Get retrieves the current configuration or panics if not initialized.
func Init(ctx *cli.Context) (*Config, error) {
	conf = &Config{
		Date:             time.Now(),
		FilesAndDirPaths: []string{"."},
		Sort:             SortDefault,
	}

	var err error

	err = conf.setDefaultOpts(ctx)
	if err != nil {
		return nil, err
	}

	err = conf.setOptions(ctx)
	if err != nil {
		return nil, err
	}

	// Get the current working directory
	conf.WorkingDir, err = filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	if conf.NoColor {
		pterm.DisableStyling()
	}

	if conf.Quiet {
		pterm.DisableOutput()
	}

	return conf, nil
}
