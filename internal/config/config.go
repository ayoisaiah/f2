// Package config is responsible for setting the program config from
// command-line arguments
package config

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"

	"github.com/ayoisaiah/f2/v2/internal/file"
)

const (
	EnvNoColor   = "NO_COLOR"
	EnvF2NoColor = "F2_NO_COLOR"
)

const (
	DefaultFixConflictsPattern = "(%d)"
	DefaultWorkingDir          = "."
)

var (
	Stdin  io.Reader = os.Stdin
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

var (
	sortVarRegex                    = regexp.MustCompile("^{.*}$")
	defaultFixConflictsPatternRegex = regexp.MustCompile(`\((\d+)\)$`)
	customFixConfictsPatternRegex   = regexp.MustCompile(
		`^(\D*?(%(\d+)?d)\D*?)$`,
	)
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

type Backup struct {
	Changes     file.Changes `json:"changes"`
	CleanedDirs []string     `json:"cleaned_dirs,omitempty"`
}

func (b Backup) RenderJSON(w io.Writer) error {
	jsonData, err := json.Marshal(b)
	if err != nil {
		return err
	}

	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

type Search struct {
	Regex *regexp.Regexp `json:"regex"`
	// Replacement index
	Index int `json:"index"`
}

// Config represents the program configuration.
type Config struct {
	Date                     time.Time      `json:"date"`
	BackupLocation           io.Writer      `json:"-"`
	ExcludeDirRegex          *regexp.Regexp `json:"exclude_dir_regex"`
	ExcludeRegex             *regexp.Regexp `json:"exclude_regex"`
	IncludeRegex             *regexp.Regexp `json:"include_regex"`
	Search                   *Search        `json:"search_regex"`
	FixConflictsPatternRegex *regexp.Regexp `json:"fix_conflicts_pattern_regex"`
	Replacement              string         `json:"replacement"`
	WorkingDir               string         `json:"working_dir"`
	FixConflictsPattern      string         `json:"fix_conflicts_pattern"`
	CSVFilename              string         `json:"csv_filename"`
	BackupFilename           string         `json:"backup_filename"`
	TargetDir                string         `json:"target_dir"`
	SortVariable             string         `json:"sort_variable"`
	ExiftoolOpts             ExiftoolOpts   `json:"exiftool_opts"`
	PairOrder                []string       `json:"pair_order"`
	FindSlice                []string       `json:"find_slice"`
	FilesAndDirPaths         []string       `json:"files_and_dir_paths"`
	ReplacementSlice         []string       `json:"replacement_slice"`
	ReplaceLimit             int            `json:"replace_limit"`
	StartNumber              int            `json:"start_number"`
	MaxDepth                 int            `json:"max_depth"`
	Sort                     Sort           `json:"sort"`
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
	Debug                    bool           `json:"debug"`
	Recursive                bool           `json:"recursive"`
	ResetIndexPerDir         bool           `json:"reset_index_per_dir"`
	OnlyDir                  bool           `json:"only_dir"`
	PipeOutput               bool           `json:"is_output_to_pipe"`
	ReverseSort              bool           `json:"reverse_sort"`
	AllowOverwrites          bool           `json:"allow_overwrites"`
	Pair                     bool           `json:"pair"`
	SortPerDir               bool           `json:"sort_per_dir"`
	Clean                    bool           `json:"clean"`
}

// SetFindStringRegex compiles a regular expression for the
// find string of the corresponding replacement index (if any).
// Otherwise, the created regex will match the entire file name.
// It takes into account the StringLiteralMode and IgnoreCase options.
//
// If a find string exists for the given replacementIndex, it's used as the pattern.
// Otherwise, the pattern defaults to ".*" to match the entire file name.
//
// Returns an error if the regex compilation fails.
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
	c.TargetDir = ctx.String("target-dir")
	c.SortPerDir = ctx.Bool("sort-per-dir")
	c.Pair = ctx.Bool("pair")
	c.PairOrder = strings.Split(ctx.String("pair-order"), ",")
	c.Clean = ctx.Bool("clean")
	c.SortVariable = ctx.String("sort-var")

	if c.SortVariable != "" && !sortVarRegex.MatchString(c.SortVariable) {
		return errInvalidSortVariable.Fmt(c.SortVariable)
	}

	includePattern := ctx.StringSlice("include")
	if len(includePattern) > 0 {
		includeMatchRegex, err := regexp.Compile(
			strings.Join(includePattern, "|"),
		)
		if err != nil {
			return err
		}

		c.IncludeRegex = includeMatchRegex
	}

	if c.TargetDir != "" {
		info, err := os.Stat(c.TargetDir)
		if err == nil && !info.IsDir() {
			return errInvalidTargetDir.Fmt(c.TargetDir)
		}

		if err != nil && os.IsExist(err) {
			return err
		}
	}

	if c.CSVFilename != "" {
		absPath, err := filepath.Abs(filepath.Dir(c.CSVFilename))
		if err != nil {
			return err
		}

		c.WorkingDir = absPath
	}

	if len(ctx.Args().Slice()) > 0 {
		c.FilesAndDirPaths = ctx.Args().Slice()
	}

	// Default to the current working directory if no path arguments are provided
	if len(c.FilesAndDirPaths) == 0 {
		c.FilesAndDirPaths = append(c.FilesAndDirPaths, DefaultWorkingDir)
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
	//nolint:gosec // acceptable use
	c.MaxDepth = int(ctx.Uint("max-depth"))
	c.Verbose = ctx.Bool("verbose")
	c.AllowOverwrites = ctx.Bool("allow-overwrites")
	c.ReplaceLimit = ctx.Int("replace-limit")
	c.Quiet = ctx.Bool("quiet")
	c.JSON = ctx.Bool("json")
	c.Exec = ctx.Bool("exec")
	c.FixConflictsPattern = ctx.String("fix-conflicts-pattern")
	c.ResetIndexPerDir = ctx.Bool("reset-index-per-dir")
	c.NoColor = ctx.Bool("no-color")

	if c.FixConflictsPattern == "" {
		c.FixConflictsPattern = DefaultFixConflictsPattern
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

// generateBackupFilename generates a unique filename for storing backup data
// based on the MD5 hash of the working directory path.
func generateBackupFilename(workingDir string) string {
	h := md5.New()
	h.Write([]byte(workingDir))

	return fmt.Sprintf("%x", h.Sum(nil)) + ".json"
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

// configureOutput configures the output behavior of the application based
// on environment variables and piping status. All output is suppressed in
// quiet mode.
func (c *Config) configureOutput() {
	// Disable coloured output if NO_COLOR is set
	if _, exists := os.LookupEnv(EnvNoColor); exists {
		c.NoColor = true
	}

	// Disable coloured output if F2_NO_COLOR is set
	if _, exists := os.LookupEnv(EnvF2NoColor); exists {
		c.NoColor = true
	}

	if c.PipeOutput {
		c.NoColor = true
	}

	if c.NoColor {
		pterm.DisableStyling()
	}

	if c.Quiet {
		pterm.DisableOutput()
	}
}

// Get retrieves the current configuration or panics if not initialized.
func Init(ctx *cli.Context, pipeOutput bool) (*Config, error) {
	conf = &Config{
		Date:             time.Now(),
		FilesAndDirPaths: []string{DefaultWorkingDir},
		Sort:             SortDefault,
		PipeOutput:       pipeOutput,
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

	if conf.WorkingDir == "" {
		// Get the current working directory
		conf.WorkingDir, err = filepath.Abs(DefaultWorkingDir)
		if err != nil {
			return nil, err
		}
	}

	conf.BackupFilename = generateBackupFilename(conf.WorkingDir)

	conf.configureOutput()

	return conf, nil
}
