// Package report provides details about the renaming operation in table or json
// format
package report

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/conflict"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/jsonutil"
	"github.com/ayoisaiah/f2/internal/status"
)

var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

// Conflicts prints any detected conflicts to the standard output in table format.
func Conflicts(conflicts conflict.Collection, jsonOut bool) {
	if jsonOut {
		o, err := jsonutil.GetOutput(nil)
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
func NoMatches(jsonOut bool) {
	msg := "Failed to match any files"

	if jsonOut {
		b, err := jsonutil.GetOutput(nil)
		if err != nil {
			pterm.Fprintln(Stderr, err)
			return
		}

		pterm.Fprintln(Stdout, string(b))

		return
	}

	pterm.Info.Println(msg)
}

func printTable(data [][]string, writer io.Writer) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"ORIGINAL", "RENAMED", "STATUS"})
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("—")
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
	)
	table.AppendBulk(data)

	table.Render()
}

// changes displays the renaming changes to be made in a table format.
func changes(
	fileChanges []*file.Change,
) {
	data := make([][]string, len(fileChanges))

	for i := range fileChanges {
		change := fileChanges[i]

		var changeStatus string

		//nolint:exhaustive // default case covers other statuses
		switch change.Status {
		case status.OK:
			changeStatus = pterm.Green(change.Status)
		case status.Unchanged:
		case status.Overwriting:
			changeStatus = pterm.Yellow(change.Status)
		default:
			changeStatus = pterm.Red(change.Status)
		}

		if change.Error != nil {
			msg := change.Error.Error()
			if strings.IndexByte(msg, ':') != -1 {
				msg = strings.TrimSpace(msg[strings.IndexByte(msg, ':'):])
			}

			changeStatus = pterm.Red(strings.TrimPrefix(msg, ": "))
		}

		d := []string{change.RelSourcePath, change.RelTargetPath, changeStatus}
		data[i] = d
	}

	printTable(data, Stdout)
}

// JSON displays the renaming changes to be made in JSON format.
func JSON(
	fileChanges []*file.Change,
) {
	o, err := jsonutil.GetOutput(fileChanges)
	if err != nil {
		pterm.Fprintln(Stderr, pterm.Error.Sprint(err))
		return
	}

	pterm.Fprintln(Stdout, string(o))
}

// Interactive prints the changes to be made and prompts the user
// to commit the changes. Blocks until user types ENTER.
func Interactive(
	fileChanges []*file.Change,
) {
	changes(fileChanges)

	reader := bufio.NewReader(os.Stdin)

	pterm.Fprint(Stderr, "\033[s")
	pterm.Info.Prefix = pterm.Prefix{
		Text:  "DRY RUN",
		Style: pterm.NewStyle(pterm.BgBlue, pterm.FgBlack),
	}

	pterm.Fprint(
		Stdout,
		pterm.Info.Sprint(
			"Press ENTER to commit the above changes",
		),
	)

	_, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		pterm.Fprintln(Stderr, pterm.Error.Print(err))
	}
}

// NonInteractive prints a report of the renaming changes to be made without
// prompting the user.
func NonInteractive(
	fileChanges []*file.Change,
) {
	changes(fileChanges)

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
