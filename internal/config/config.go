package config

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	errInvalidArgument = errors.New(
		"Invalid argument: one of `-f`, `-r`, `-csv` or `-u` must be present and set to a non empty string value. Use 'f2 --help' for more information",
	)

	errInvalidSimpleModeArgs = errors.New(
		"At least one argument must be specified in simple mode",
	)
)

var conf *Config

type Config struct {
	date               time.Time
	stdin              io.Reader
	stderr             io.Writer
	stdout             io.Writer
	searchRegex        *regexp.Regexp
	csvFilename        string
	sort               string
	replacement        string
	workingDir         string
	findSlice          []string
	excludeFilter      []string
	replacementSlice   []string
	pathsToFilesOrDirs []string
	numberOffset       []int
	maxDepth           int
	startNumber        int
	replaceLimit       int
	recursive          bool
	ignoreCase         bool
	reverseSort        bool
	onlyDir            bool
	revert             bool
	includeDir         bool
	ignoreExt          bool
	allowOverwrites    bool
	verbose            bool
	includeHidden      bool
	quiet              bool
	fixConflicts       bool
	exec               bool
	stringLiteralMode  bool
	simpleMode         bool
	json               bool
}

// SetFindStringRegex compiles a regular expression for the
// find string of the corresponding replacement index (if any).
// Otherwise, the created regex will match the entire file name.
func (c *Config) SetFindStringRegex(replacementIndex int) error {
	// findPattern is set to match the entire file name by default
	// except if a find string for the corresponding replacement index
	// is found
	findPattern := ".*"
	if len(c.findSlice) > replacementIndex {
		findPattern = c.findSlice[replacementIndex]

		// Escape all regular expression metacharacters in string literal mode
		if c.stringLiteralMode {
			findPattern = regexp.QuoteMeta(findPattern)
		}

		if c.ignoreCase {
			findPattern = "(?i)" + findPattern
		}
	}

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return err
	}

	c.searchRegex = re

	return nil
}

func (c *Config) setOptions(ctx *cli.Context) error {
	if len(ctx.StringSlice("find")) == 0 &&
		len(ctx.StringSlice("replace")) == 0 &&
		ctx.String("csv") == "" &&
		!ctx.Bool("undo") {
		return errInvalidArgument
	}

	c.findSlice = ctx.StringSlice("find")
	c.replacementSlice = ctx.StringSlice("replace")
	c.csvFilename = ctx.String("csv")
	c.revert = ctx.Bool("undo")
	c.pathsToFilesOrDirs = ctx.Args().Slice()
	c.exec = ctx.Bool("exec")

	c.setDefaultOpts(ctx)

	// Ensure that each findString has a corresponding replacement.
	// The replacement defaults to an empty string if unset
	for len(c.findSlice) > len(c.replacementSlice) {
		c.replacementSlice = append(c.replacementSlice, "")
	}

	return c.SetFindStringRegex(0)
}

// setSimpleModeOptions is used to set the options for the
// renaming operation in simpleMode.
func (c *Config) setSimpleModeOptions(ctx *cli.Context) error {
	args := ctx.Args().Slice()

	if len(args) < 1 {
		return errInvalidSimpleModeArgs
	}

	// If a replacement string is not specified, it shoud be
	// an empty string
	if len(args) == 1 {
		args = append(args, "")
	}

	minArgs := 2

	c.simpleMode = true
	c.exec = true

	c.findSlice = []string{args[0]}
	c.replacementSlice = []string{args[1]}

	c.setDefaultOpts(ctx)

	c.includeDir = true

	if len(args) > minArgs {
		c.pathsToFilesOrDirs = args[minArgs:]
	}

	return c.SetFindStringRegex(0)
}

// setDefaultOpts applies the options that may be set through
// F2_DEFAULT_OPTS.
func (c *Config) setDefaultOpts(ctx *cli.Context) {
	c.fixConflicts = ctx.Bool("fix-conflicts")
	c.includeDir = ctx.Bool("include-dir")
	c.includeHidden = ctx.Bool("hidden")
	c.ignoreCase = ctx.Bool("ignore-case")
	c.ignoreExt = ctx.Bool("ignore-ext")
	c.recursive = ctx.Bool("recursive")
	c.onlyDir = ctx.Bool("only-dir")
	c.stringLiteralMode = ctx.Bool("string-mode")
	c.excludeFilter = ctx.StringSlice("exclude")
	c.maxDepth = int(ctx.Uint("max-depth"))
	c.verbose = ctx.Bool("verbose")
	c.allowOverwrites = ctx.Bool("allow-overwrites")
	c.replaceLimit = ctx.Int("replace-limit")
	c.quiet = ctx.Bool("quiet")
	c.json = ctx.Bool("json")

	// Sorting
	if ctx.String("sort") != "" {
		c.sort = ctx.String("sort")
	} else if ctx.String("sortr") != "" {
		c.sort = ctx.String("sortr")
		c.reverseSort = true
	}

	if c.onlyDir {
		c.includeDir = true
	}
}

func Init(ctx *cli.Context) (*Config, error) {
	conf = &Config{
		stdout: os.Stdout,
		stderr: os.Stderr,
		stdin:  os.Stdin,
		date:   time.Now(),
	}

	v, exists := ctx.App.Metadata["reader"]
	if exists {
		r, ok := v.(io.Reader)
		if ok {
			conf.stdin = r
		}
	}

	v, exists = ctx.App.Metadata["writer"]
	if exists {
		w, ok := v.(io.Writer)
		if ok {
			conf.stdout = w
		}
	}

	var err error

	if _, ok := ctx.App.Metadata["simple-mode"]; ok {
		err = conf.setSimpleModeOptions(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		err = conf.setOptions(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Get the current working directory
	conf.workingDir, err = filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func Get() *Config {
	return conf
}

func (c *Config) Stdin() io.Reader {
	return c.stdin
}

func (c *Config) Stderr() io.Writer {
	return c.stderr
}

func (c *Config) Stdout() io.Writer {
	return c.stdout
}

func (c *Config) ShouldRevert() bool {
	return c.revert
}

func (c *Config) PathsToFilesOrDirs() []string {
	return c.pathsToFilesOrDirs
}

func (c *Config) IsRecursive() bool {
	return c.recursive
}

func (c *Config) CSVFilename() string {
	return c.csvFilename
}

func (c *Config) IsVerbose() bool {
	return c.verbose
}

func (c *Config) IncludeHidden() bool {
	return c.includeHidden
}

func (c *Config) SortName() string {
	return c.sort
}

func (c *Config) ReverseSort() bool {
	return c.reverseSort
}

func (c *Config) ExcludeFilter() []string {
	return c.excludeFilter
}

func (c *Config) Replacement() string {
	return c.replacement
}

func (c *Config) SetReplacement(replacement string) {
	c.replacement = replacement
}

func (c *Config) ReplacementSlice() []string {
	return c.replacementSlice
}

func (c *Config) SetReplacementSlice(s []string) {
	c.replacementSlice = s
}

func (c *Config) SetFindSlice(s []string) {
	c.findSlice = s
}

func (c *Config) WorkingDir() string {
	return c.workingDir
}

func (c *Config) ShouldExec() bool {
	return c.exec
}

func (c *Config) IncludeDir() bool {
	return c.includeDir
}

func (c *Config) OnlyDir() bool {
	return c.onlyDir
}

func (c *Config) IgnoreExt() bool {
	return c.ignoreExt
}

func (c *Config) IgnoreCase() bool {
	return c.ignoreCase
}

func (c *Config) SearchRegex() *regexp.Regexp {
	return c.searchRegex
}

func (c *Config) FixConflicts() bool {
	return c.fixConflicts
}

func (c *Config) JSON() bool {
	return c.json
}

func (c *Config) SimpleMode() bool {
	return c.simpleMode
}

func (c *Config) IsQuiet() bool {
	return c.quiet
}

func (c *Config) Date() time.Time {
	return c.date
}

func (c *Config) ReplaceLimit() int {
	return c.replaceLimit
}

func (c *Config) AllowOverwrites() bool {
	return c.allowOverwrites
}

func (c *Config) StartNumber() int {
	return c.startNumber
}

func (c *Config) NumberOffset() []int {
	return c.numberOffset
}

func (c *Config) SetNumberOffset(offset []int) {
	c.numberOffset = offset
}

func (c *Config) MaxDepth() int {
	return c.maxDepth
}

func (c *Config) FindSlice() []string {
	return c.findSlice
}
