package file

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/olekukonko/tablewriter"
	"github.com/pterm/pterm"

	"github.com/ayoisaiah/f2/v2/internal/localize"
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

// ID3 represents id3 data from an audio file.
type ID3 struct {
	Format      string
	FileType    string
	Title       string
	Album       string
	Artist      string
	AlbumArtist string
	Genre       string
	Composer    string
	Year        int
	Track       int
	TotalTracks int
	Disc        int
	TotalDiscs  int
}

// Change represents a single renaming change.
type Change struct {
	Error         error                  `json:"error,omitempty"`
	PrimaryPair   *Change                `json:"-"`
	ExiftoolData  *exiftool.FileMetadata `json:"-"`
	ID3Data       *ID3                   `json:"-"`
	HashData      map[string]string      `json:"-"`
	Status        status.Status          `json:"status"`
	OriginalName  string                 `json:"-"`
	Source        string                 `json:"source"`
	Target        string                 `json:"target"`
	TargetDir     string                 `json:"target_dir"`
	BaseDir       string                 `json:"base_dir"`
	SourcePath    string                 `json:"-"`
	TargetPath    string                 `json:"-"`
	Steps         []string               `json:"-"`
	CSVRow        []string               `json:"-"`
	SortCriterion struct {
		TimeVar   time.Time
		Time      time.Time
		StringVar string
		IntVar    int
		Size      int64
	} `json:"-"`
	Position        int        `json:"-"`
	Mu              sync.Mutex `json:"-"`
	IsDir           bool       `json:"is_dir"`
	WillOverwrite   bool       `json:"-"`
	MatchesFindCond bool       `json:"-"`
}

func (c *Change) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("source_path", c.SourcePath),
		slog.String("target", c.Target),
	}

	if c.IsDir {
		attrs = append(attrs, slog.Bool("is_dir", c.IsDir))
	}

	if c.Error != nil {
		attrs = append(attrs, slog.Any("error", c.Error))
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

	if c.SortCriterion.StringVar != "" {
		attrs = append(
			attrs,
			slog.String("string_var", c.SortCriterion.StringVar),
		)
	}

	if c.SortCriterion.IntVar != 0 {
		attrs = append(attrs, slog.Int("int_var", c.SortCriterion.IntVar))
	}

	if c.SortCriterion.Size != 0 {
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

	if len(c.Steps) > 0 {
		c.Steps[len(c.Steps)-1] = c.TargetPath
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

	return names, indices
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

		data[i] = append(change.Steps, changeStatus)
	}

	printTable(data, w, noColor)
}

func printTable(data [][]string, w io.Writer, noColor bool) {
	// using tablewriter as pterm table rendering is too slow
	table := tablewriter.NewWriter(w)

	headers := []string{localize.T("table.original")}

	numSteps := len(data[0])

	for i := 0; i < numSteps-3; i++ {
		headers = append(headers, fmt.Sprintf("-> %d", i+1))
	}

	headers = append(
		headers,
		localize.T("table.renamed"),
		localize.T("table.status"),
	)

	table.SetHeader(headers)
	table.SetCenterSeparator("*")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("—")
	table.SetAutoWrapText(false)

	if !noColor {
		headerColors := make([]tablewriter.Colors, 0, len(headers))
		for range headers {
			headerColors = append(
				headerColors,
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgGreenColor},
			)
		}

		table.SetHeaderColor(headerColors...)
	}

	table.AppendBulk(data)

	table.Render()
}
