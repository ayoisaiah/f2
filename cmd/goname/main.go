package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/urfave/cli/v2"
)

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

type Operation struct {
	paths         []string
	matches       []string
	newPaths      map[string]string
	replaceString string
	dryRun        bool
	searchRegex   *regexp.Regexp
}

// Apply will update the paths or print
// the changes to be made if the dry run option
// is selected
func (op *Operation) Apply() {
	for p, v := range op.newPaths {
		if op.dryRun {
			fmt.Println(p, "->", v)
		} else {
			os.Rename(p, v)
		}
	}
}

func (op *Operation) FindMatches() error {
	for _, f := range op.paths {
		isDir, err := isDirectory(f)
		if err != nil {
			return err
		}

		if isDir {
			continue
		}

		filename := filepath.Base(f)
		matched := op.searchRegex.MatchString(filename)
		if matched {
			op.matches = append(op.matches, f)
		}
	}

	return nil
}

func (op *Operation) Replace() {
	for _, f := range op.matches {
		filename, dir := filepath.Base(f), filepath.Dir(f)
		str := op.searchRegex.ReplaceAllString(filename, op.replaceString)
		op.newPaths[f] = filepath.Join(dir, str)
	}
}

func NewOperation(c *cli.Context) (*Operation, error) {
	op := &Operation{}
	op.paths = c.Args().Slice()
	op.replaceString = c.String("replace")
	op.dryRun = c.Bool("dry")
	op.newPaths = make(map[string]string)

	findPattern := c.String("find")

	re, err := regexp.Compile(findPattern)
	if err != nil {
		return nil, fmt.Errorf("Malformed regular expression for search pattern %s", findPattern)
	}

	op.searchRegex = re

	// If paths are omitted, default to the files in the
	// current directory
	if len(op.paths) == 0 {
		file, err := os.Open(".")
		if err != nil {
			return nil, err
		}

		defer file.Close()

		names, err := file.Readdirnames(0)
		if err != nil {
			return nil, err
		}

		op.paths = names
	}

	return op, nil
}

func main() {
	app := &cli.App{
		Name:  "goname",
		Usage: "Goname is a command-line utility for renaming files in bulk",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "Search pattern",
			},
			&cli.StringFlag{
				Name:    "replace",
				Aliases: []string{"r"},
				Usage:   "Replacement string",
			},
			&cli.BoolFlag{
				Name:    "dry",
				Aliases: []string{"d"},
				Usage:   "Dry run",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NumFlags() == 0 && c.NArg() == 0 {
				return fmt.Errorf("goname: not enough arguments\nTry 'goname --help' for more information.")
			}

			op, err := NewOperation(c)
			if err != nil {
				return err
			}

			op.FindMatches()
			op.Replace()
			op.Apply()

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
