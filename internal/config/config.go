// Package config is responsible for setting the program config from
// command-line arguments
package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
	"github.com/pterm/pterm"
	"github.com/urfave/cli/v3"
)

const (
	EnvNoColor   = "NO_COLOR"
	EnvF2NoColor = "F2_NO_COLOR"
	EnvDebug     = "F2_DEBUG"
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
	customFixConflictsPatternRegex  = regexp.MustCompile(
		`^(\D*?(%(\d+)?d)\D*?)$`,
	)
	capturVarIndexRegex = regexp.MustCompile(
		`{+(\$\d+)(%(\d?)+d)([borh])?(-?\d+)?(?:<(\d+(?:-\d+)?(?:;\s*\d+(?:-\d+)?)*)>)?}+`,
	)
	indexVarRegex = regexp.MustCompile(
		`{+(\$\d+)?(\d+)?(%(\d?)+d)([borh])?(-?\d+)?(?:<(\d+(?:-\d+)?(?:;\s*\d+(?:-\d+)?)*)>)?(##)?}+`,
	)
	findVariableRegex = regexp.MustCompile(`{(.*)}`)
	exifToolVarRegex  = regexp.MustCompile(`{xt\..*}`)
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
	Regex    *regexp.Regexp `json:"regex"`
	FindCond *regexp.Regexp
	Index    int `json:"index"`
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
	ExifToolVarPresent       bool           `json:"-"`
	IndexPresent             bool           `json:"-"`
}

func (c *Config) setFindCond(replacementIndex int) error {
	submatches := findVariableRegex.FindAllStringSubmatch(
		c.FindSlice[replacementIndex],
		-1,
	)

	re, err := regexp.Compile(submatches[0][1])
	if err != nil {
		return err
	}

	c.Search = &Search{
		// When filtering using arbitrary condition, match the entire file name
		Regex:    regexp.MustCompile(".*"),
		FindCond: re,
		Index:    replacementIndex,
	}

	return nil
}

// setFindRegex compiles a regular expression for the
// find string of the corresponding replacement index (if any).
// Otherwise, the created regex will match the entire file name.
// It takes into account the StringLiteralMode and IgnoreCase options.
//
// If a find string exists for the given replacementIndex, it's used as the pattern.
// Otherwise, the pattern defaults to ".*" to match the entire file name.
//
// Returns an error if the regex compilation fails.
func (c *Config) setFindRegex(replacementIndex int) error {
	findPattern := c.FindSlice[replacementIndex]

	// Escape all regular expression metacharacters in string literal mode
	if c.StringLiteralMode {
		findPattern = regexp.QuoteMeta(findPattern)
	}

	if c.IgnoreCase {
		findPattern = "(?i)" + findPattern
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

func (c *Config) SetFind(replacementIndex int) error {
	if len(c.FindSlice) > replacementIndex {
		if findVariableRegex.MatchString(c.FindSlice[replacementIndex]) {
			return c.setFindCond(replacementIndex)
		}

		return c.setFindRegex(replacementIndex)
	}

	return nil
}

func (c *Config) setOptions(cmd *cli.Command) error {
	if len(cmd.StringSlice("find")) == 0 &&
		len(cmd.StringSlice("replace")) == 0 &&
		cmd.String("csv") == "" &&
		!cmd.Bool("undo") {
		return errInvalidArgument
	}

	c.FindSlice = cmd.StringSlice("find")
	c.ReplacementSlice = cmd.StringSlice("replace")

	c.CSVFilename = cmd.String("csv")
	c.Revert = cmd.Bool("undo")
	c.Debug = cmd.Bool("debug")
	c.TargetDir = cmd.String("target-dir")
	c.SortPerDir = cmd.Bool("sort-per-dir")
	c.Pair = cmd.Bool("pair")
	c.PairOrder = strings.Split(cmd.String("pair-order"), ",")
	c.Clean = cmd.Bool("clean")
	c.SortVariable = cmd.String("sort-var")

	// Don't replace the extension in pair mode
	if conf.Pair {
		conf.IgnoreExt = true
	}

	if c.SortVariable != "" && !sortVarRegex.MatchString(c.SortVariable) {
		return errInvalidSortVariable.Fmt(c.SortVariable)
	}

	includePattern := cmd.StringSlice("include")
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

	if len(cmd.Args().Slice()) > 0 {
		c.FilesAndDirPaths = cmd.Args().Slice()
	}

	// Ensure that each findString has a corresponding replacement.
	// The replacement defaults to an empty string if unset
	for len(c.FindSlice) > len(c.ReplacementSlice) {
		c.ReplacementSlice = append(c.ReplacementSlice, "")
	}

	for len(c.ReplacementSlice) > len(c.FindSlice) {
		c.FindSlice = append(c.FindSlice, ".*")
	}

	// Distinguish capture variable indices from regular indices by adding ##
	for i, v := range c.ReplacementSlice {
		if capturVarIndexRegex.MatchString(v) {
			for _, match := range capturVarIndexRegex.FindAllString(v, -1) {
				index := strings.Index(match, "}")
				if index == -1 {
					continue
				}

				captureVarIndex := match[:index] + "##" + match[index:]

				c.ReplacementSlice[i] = strings.ReplaceAll(
					v,
					match,
					captureVarIndex,
				)
			}
		}
	}

	return c.SetFind(0)
}

// setDefaultOpts applies any options that may be set through
// F2_DEFAULT_OPTS.
func (c *Config) setDefaultOpts(cmd *cli.Command) error {
	c.AutoFixConflicts = cmd.Bool("fix-conflicts")
	c.IncludeDir = cmd.Bool("include-dir")
	c.IncludeHidden = cmd.Bool("hidden")
	c.IgnoreCase = cmd.Bool("ignore-case")
	c.IgnoreExt = cmd.Bool("ignore-ext")
	c.Recursive = cmd.Bool("recursive")
	c.OnlyDir = cmd.Bool("only-dir")
	c.StringLiteralMode = cmd.Bool("string-mode")
	//nolint:gosec // acceptable use
	c.MaxDepth = int(cmd.Uint("max-depth"))
	c.Verbose = cmd.Bool("verbose")
	c.AllowOverwrites = cmd.Bool("allow-overwrites")
	c.ReplaceLimit = cmd.Int("replace-limit")
	c.Quiet = cmd.Bool("quiet")
	c.JSON = cmd.Bool("json")
	c.Exec = cmd.Bool("exec")
	c.FixConflictsPattern = cmd.String("fix-conflicts-pattern")
	c.ResetIndexPerDir = cmd.Bool("reset-index-per-dir")
	c.NoColor = cmd.Bool("no-color")

	if c.FixConflictsPattern == "" {
		c.FixConflictsPattern = DefaultFixConflictsPattern
		c.FixConflictsPatternRegex = defaultFixConflictsPatternRegex
	} else {
		if !customFixConflictsPatternRegex.MatchString(c.FixConflictsPattern) {
			return errParsingFixConflictsPattern.Fmt(c.FixConflictsPattern)
		}

		r := regexp.MustCompile(`%(\d+)?d`)
		c.FixConflictsPatternRegex = regexp.MustCompile(
			r.ReplaceAllString(conf.FixConflictsPattern, `(\d+)`),
		)
	}

	excludePattern := cmd.StringSlice("exclude")
	if len(excludePattern) > 0 {
		excludeMatchRegex, err := regexp.Compile(
			strings.Join(excludePattern, "|"),
		)
		if err != nil {
			return err
		}

		c.ExcludeRegex = excludeMatchRegex
	}

	excludeDirPattern := cmd.StringSlice("exclude-dir")
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
	if cmd.String("sort") != "" {
		c.Sort, err = parseSortArg(cmd.String("sort"))
		if err != nil {
			return err
		}
	} else if cmd.String("sortr") != "" {
		c.Sort, err = parseSortArg(cmd.String("sortr"))
		if err != nil {
			return err
		}

		c.ReverseSort = true
	}

	if cmd.String("exiftool-opts") != "" {
		args, err := shellquote.Split(cmd.String("exiftool-opts"))
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

func (c *Config) checkIfExifToolVarIsPresent() bool {
	return slices.ContainsFunc(c.FindSlice, exifToolVarRegex.MatchString) ||
		slices.ContainsFunc(c.ReplacementSlice, exifToolVarRegex.MatchString) ||
		exifToolVarRegex.MatchString(c.SortVariable)
}

// Get retrieves the current configuration or panics if not initialized.
func Init(cmd *cli.Command, pipeOutput bool) (*Config, error) {
	conf = &Config{
		Date:             time.Now(),
		FilesAndDirPaths: []string{DefaultWorkingDir},
		Sort:             SortDefault,
		PipeOutput:       pipeOutput,
		Search: &Search{
			Regex: regexp.MustCompile(".*"),
		},
	}

	var err error

	err = conf.setDefaultOpts(cmd)
	if err != nil {
		return nil, err
	}

	err = conf.setOptions(cmd)
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

	conf.ExifToolVarPresent = conf.checkIfExifToolVarIsPresent()
	conf.IndexPresent = slices.ContainsFunc(
		conf.ReplacementSlice,
		indexVarRegex.MatchString,
	)

	return conf, nil
}
