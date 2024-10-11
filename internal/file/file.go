package file

import (
	"encoding/json"
	"io"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/internal/config"
	"github.com/ayoisaiah/f2/internal/status"
)

// Change represents a single renaming change.
type Change struct {
	Error error `json:"error,omitempty"`
	// The original filename which can be different from Source in
	// a multi-step renaming operation
	OriginalName string        `json:"-"`
	Status       status.Status `json:"status"`
	BaseDir      string        `json:"base_dir"`
	Source       string        `json:"source"`
	Target       string        `json:"target"`
	// SourcePath is BaseDir + Source
	SourcePath string `json:"-"`
	// TargetPath is BaseDir + Target
	TargetPath    string   `json:"-"`
	CSVRow        []string `json:"-"`
	Position      int      `json:"-"`
	IsDir         bool     `json:"is_dir"`
	WillOverwrite bool     `json:"-"`
}

// AutoFixTarget sets the new target name.
func (c *Change) AutoFixTarget(newTarget string) {
	c.Target = newTarget
	c.TargetPath = filepath.Join(c.BaseDir, c.Target)

	// Ensure empty targets is reported as empty instead of as a dot
	if c.TargetPath == "." {
		c.TargetPath = ""
	}

	if c.Target == "" && c.TargetPath != "" {
		c.TargetPath += "/"
	}

	c.Status = status.OK
}

type Changes []*Change

func (c Changes) RenderJSON(w io.Writer) error {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return err
	}

	_, err = w.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

func (c Changes) RenderTable(w io.Writer) {
	data := make([][]string, len(c))

	for i := range c {
		change := c[i]

		var changeStatus string

		//nolint:exhaustive // default case covers other statuses
		switch change.Status {
		case status.OK:
			changeStatus = pterm.Green(change.Status)
		case status.Unchanged, status.Overwriting, status.Ignored:
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

		d := []string{change.SourcePath, change.TargetPath, changeStatus}
		data[i] = d
	}

	printTable(data, w)
}

func printTable(data [][]string, w io.Writer) {
	conf := config.Get()

	// using tablewriter as pterm table rendering is too slow
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"ORIGINAL", "RENAMED", "STATUS"})
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("—")
	table.SetAutoWrapText(false)

	if !conf.NoColor {
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		)
	}

	table.AppendBulk(data)

	table.Render()
}
