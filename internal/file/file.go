package file

import (
	"encoding/json"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/v2/internal/config"
	"github.com/ayoisaiah/f2/v2/internal/status"
)

type Backup struct {
	Changes     Changes  `json:"changes"`
	CleanedDirs []string `json:"cleaned_dirs,omitempty"`
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

// Change represents a single renaming change.
type Change struct {
	Error         error                  `json:"error,omitempty"`
	PrimaryPair   *Change                `json:"-"`
	ExiftoolData  *exiftool.FileMetadata `json:"-"`
	BaseDir       string                 `json:"base_dir"`
	TargetDir     string                 `json:"target_dir"`
	Source        string                 `json:"source"`
	Target        string                 `json:"target"`
	OriginalName  string                 `json:"-"`
	Status        status.Status          `json:"status"`
	SourcePath    string                 `json:"-"`
	TargetPath    string                 `json:"-"`
	CSVRow        []string               `json:"-"`
	SortCriterion struct {
		TimeVar   time.Time
		Time      time.Time
		StringVar string
		IntVar    int
		Size      int64
	} `json:"-"`
	Position        int  `json:"-"`
	IsDir           bool `json:"is_dir"`
	WillOverwrite   bool `json:"-"`
	MatchesFindCond bool `json:"-"`
}

func (c *Change) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("source_path", c.SourcePath),
		slog.String("target", c.Target),
		slog.Bool("is_dir", c.IsDir),
	}

	if !c.SortCriterion.TimeVar.IsZero() {
		attrs = append(
			attrs,
			slog.Int64("time_var", c.SortCriterion.TimeVar.UnixNano()),
		)
	}

	if !c.SortCriterion.Time.IsZero() {
		attrs = append(
			attrs,
			slog.Int64("time", c.SortCriterion.Time.UnixNano()),
		)
	}

	conf := config.Get()

	if conf.Sort == config.SortStringVar {
		attrs = append(
			attrs,
			slog.String("string_var", c.SortCriterion.StringVar),
		)
	}

	if conf.Sort == config.SortIntVar {
		attrs = append(attrs, slog.Int("int_var", c.SortCriterion.IntVar))
	}

	if conf.Sort == config.SortSize {
		attrs = append(attrs, slog.Int64("size", c.SortCriterion.Size))
	}

	return slog.GroupValue(attrs...)
}

// AutoFixTarget sets the new target name.
func (c *Change) AutoFixTarget(newTarget string) {
	c.Target = newTarget
	c.TargetPath = filepath.Join(c.TargetDir, c.Target)

	// Ensure empty targets are reported as empty instead of as a dot
	if c.TargetPath == "." {
		c.TargetPath = ""
	}

	if c.Target == "" && c.TargetPath != "" {
		c.TargetPath += "/"
	}

	c.Status = status.OK
}

type Changes []*Change

func (c Changes) LogValue() slog.Value {
	vals := make([]slog.Value, len(c))

	for i := range c {
		ch := c[i]
		vals[i] = slog.GroupValue(
			slog.Int("index", i+1),
			slog.Any("file", ch.LogValue()),
		)
	}

	return slog.GroupValue(slog.Any("changes", vals))
}

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

func (c Changes) SourceNamesWithIndices(
	withPair bool,
) (names []string, indices []int) {
	for i := range c {
		ch := c[i]

		if ch.IsDir {
			continue
		}

		if withPair {
			if ch.PrimaryPair == nil {
				names = append(names, ch.SourcePath)
				indices = append(indices, i)
			}

			continue
		}

		names = append(names, ch.SourcePath)
		indices = append(indices, i)
	}

	return
}

func (c Changes) ShouldExtractExiftool() bool {
	for _, v := range c {
		if v.ExiftoolData != nil {
			return false
		}
	}

	return true
}

func (c Changes) RenderTable(w io.Writer, noColor bool) {
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

	printTable(data, w, noColor)
}

func printTable(data [][]string, w io.Writer, noColor bool) {
	// using tablewriter as pterm table rendering is too slow
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"ORIGINAL", "RENAMED", "STATUS"})
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("—")
	table.SetAutoWrapText(false)

	if !noColor {
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
		)
	}

	table.AppendBulk(data)

	table.Render()
}
