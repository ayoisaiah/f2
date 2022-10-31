// Package report provides details about the renaming operation in table or json
// format
package report

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"golang.org/x/exp/slices"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/conflict"
	"github.com/ayoisaiah/f2/internal/file"
	internaljson "github.com/ayoisaiah/f2/internal/json"
	"github.com/ayoisaiah/f2/internal/status"
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
func Changes(changes []*file.Change, errs []int) {
	conf := config.Get()

	if conf.IsQuiet() {
		return
	}

	if conf.JSON() {
		o, err := internaljson.GetOutput(changes, errs)
		if err != nil {
			pterm.Fprintln(conf.Stderr(), pterm.Error.Sprint(err))
		}

		pterm.Fprintln(conf.Stdout(), string(o))

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

	printTable(data, conf.Stdout())
}

// Conflicts prints any detected conflicts to the standard output in table format.
func Conflicts(conflicts conflict.Collection) {
	conf := config.Get()

	if conf.JSON() {
		o, err := internaljson.GetOutput(nil, nil)
		if err != nil {
			pterm.Fprintln(conf.Stderr(), pterm.Error.Sprint(err))
		}

		pterm.Fprintln(conf.Stdout(), string(o))

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

	printTable(data, conf.Stdout())
}

// NoMatches prints out a message indicating that the find string failed
// to match any files.
func NoMatches() {
	conf := config.Get()

	msg := "Failed to match any files"

	if conf.JSON() {
		b, err := internaljson.GetOutput(nil, nil)
		if err != nil {
			pterm.Fprintln(conf.Stderr(), err)
			return
		}

		pterm.Fprintln(conf.Stdout(), string(b))

		return
	}

	pterm.Info.Println(msg)
}

// Dry prints a report of the renaming changes to be made.
func Dry(changes []*file.Change) {
	conf := config.Get()

	Changes(changes, nil)

	if !conf.JSON() {
		pterm.Info.Prefix = pterm.Prefix{
			Text:  "DRY RUN",
			Style: pterm.NewStyle(pterm.BgBlue, pterm.FgBlack),
		}

		pterm.Fprintln(
			conf.Stdout(),
			pterm.Info.Sprint(
				"Commit the above changes with the -x/--exec flag",
			),
		)
	}
}
