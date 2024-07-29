// Package report provides details about the renaming operation in table or json
// format
package report

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/file"
	"github.com/ayoisaiah/f2/internal/jsonutil"
	"github.com/ayoisaiah/f2/internal/status"
)

var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr
)

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
	table.SetAutoWrapText(false)
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
		case status.Unchanged, status.Overwriting:
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
		data[change.Position] = d
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
	conflictDetected bool,
) {
	changes(fileChanges)

	if conflictDetected {
		return
	}

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

func Report(conf *config.Config, fileChanges []*file.Change) {
	if conf.JSON {
		JSON(fileChanges)
		return
	}

	NonInteractive(fileChanges, false)
}
