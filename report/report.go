// Package report provides details about the renaming operation in table or json
// format
package report

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/exp/slices"

	"github.com/ayoisaiah/f2/internal/conflict"
	"github.com/ayoisaiah/f2/internal/file"
	internaljson "github.com/ayoisaiah/f2/internal/json"
	internalsort "github.com/ayoisaiah/f2/internal/sort"
	"github.com/ayoisaiah/f2/internal/status"
)

var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

func printTable(data [][]string, writer io.Writer) {
	d := [][]string{
		{"ORIGINAL", "RENAMED", "STATUS"},
	}

	d = append(d, data...)

	table := pterm.DefaultTable
	table.HeaderRowSeparator = "*"
	table.Boxed = true

	str, err := table.WithHasHeader().WithData(d).Srender()
	if err != nil {
		pterm.Error.Printfln("Unable to print table: %s", err.Error())
		return
	}

	fmt.Fprintln(writer, str)
}

// Changes displays the changes to be made in a table or json format.
func Changes(
	changes []*file.Change,
	errs []int,
	quiet bool,
	jsonOpts *internaljson.OutputOpts,
) {
	if quiet {
		return
	}

	if jsonOpts.Print {
		o, err := internaljson.GetOutput(jsonOpts, changes, errs)
		if err != nil {
			pterm.Fprintln(Stderr, pterm.Error.Sprint(err))
		}

		pterm.Fprintln(Stdout, string(o))

		return
	}

	data := make([][]string, len(changes))

	for i := range changes {
		change := changes[i]

		source := filepath.Join(change.BaseDir, change.Source)
		target := filepath.Join(change.BaseDir, change.Target)

		changeStatus := pterm.Green(change.Status)
		if change.Status != status.OK {
			changeStatus = pterm.Yellow(change.Status)
		}

		if slices.Contains(errs, i) {
			changeStatus = pterm.Red(change.Error)
		}

		d := []string{source, target, changeStatus}
		data[i] = d
	}

	printTable(data, Stdout)
}

// Conflicts prints any detected conflicts to the standard output in table format.
func Conflicts(
	conflicts conflict.Collection,
	jsonOpts *internaljson.OutputOpts,
) {
	if jsonOpts.Print {
		o, err := internaljson.GetOutput(jsonOpts, nil, nil)
		if err != nil {
			pterm.Fprintln(Stderr, pterm.Error.Sprint(err))
		}

		pterm.Fprintln(Stdout, string(o))

		return
	}

	var data [][]string

	if slice, exists := conflicts[conflict.EmptyFilename]; exists {
		for _, v := range slice {
			slice := []string{
				strings.Join(v.Sources, ""),
				"",
				pterm.Red(status.EmptyFilename),
			}
			data = append(data, slice)
		}
	}

	if slice, exists := conflicts[conflict.TrailingPeriod]; exists {
		for _, v := range slice {
			for _, s := range v.Sources {
				slice := []string{
					s,
					v.Target,
					pterm.Red(
						status.TrailingPeriod,
					),
				}
				data = append(data, slice)
			}
		}
	}

	if slice, exists := conflicts[conflict.FileExists]; exists {
		for _, v := range slice {
			slice := []string{
				strings.Join(v.Sources, ""),
				v.Target,
				pterm.Red(status.PathExists),
			}
			data = append(data, slice)
		}
	}

	if slice, exists := conflicts[conflict.OverwritingNewPath]; exists {
		for _, v := range slice {
			for _, s := range v.Sources {
				slice := []string{
					s,
					v.Target,
					pterm.Red(status.OverwritingNewPath),
				}
				data = append(data, slice)
			}
		}
	}

	if slice, exists := conflicts[conflict.InvalidCharacters]; exists {
		for _, v := range slice {
			for _, s := range v.Sources {
				slice := []string{
					s,
					v.Target,
					pterm.Red(
						fmt.Sprintf(
							string(status.InvalidCharacters),
							v.Cause,
						),
					),
				}
				data = append(data, slice)
			}
		}
	}

	if slice, exists := conflicts[conflict.MaxFilenameLengthExceeded]; exists {
		for _, v := range slice {
			for _, s := range v.Sources {
				slice := []string{
					s,
					v.Target,
					pterm.Red(
						fmt.Sprintf(
							string(status.FilenameLengthExceeded),
							v.Cause,
						),
					),
				}
				data = append(data, slice)
			}
		}
	}

	printTable(data, Stdout)
}

func BackupFailed(err error) {
	pterm.Fprintln(Stderr,
		pterm.Warning.Sprintf(
			"Failed to backup renaming operation due to error: %s",
			err.Error(),
		),
	)
}

// NoMatches prints out a message indicating that the find string failed
// to match any files.
func NoMatches(
	jsonOpts *internaljson.OutputOpts,
) {
	msg := "Failed to match any files"

	if jsonOpts.Print {
		b, err := internaljson.GetOutput(jsonOpts, nil, nil)
		if err != nil {
			pterm.Fprintln(Stderr, err)
			return
		}

		pterm.Fprintln(Stdout, string(b))

		return
	}

	pterm.Info.Println(msg)
}

// Dry prints a report of the renaming changes to be made.
func Dry(
	changes []*file.Change,
	includeDir, quiet, revert bool,
	jsonOpts *internaljson.OutputOpts,
) {
	if includeDir {
		internalsort.FilesBeforeDirs(changes, revert)
	}

	Changes(changes, nil, quiet, jsonOpts)

	if !jsonOpts.Print {
		pterm.Info.Prefix = pterm.Prefix{
			Text:  "DRY RUN",
			Style: pterm.NewStyle(pterm.BgBlue, pterm.FgBlack),
		}

		pterm.Fprintln(
			Stdout,
			pterm.Info.Sprint(
				"Commit the above changes with the -x/--exec flag",
			),
		)
	}
}
