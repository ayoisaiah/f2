package rename

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/ayoisaiah/f2/internal/file"
)

func createBackupFile(fileName string) (io.Writer, error) {
	backupFilePath, err := xdg.DataFile(
		filepath.Join("f2", "backups", fileName),
	)
	if err != nil {
		return nil, err
	}

	// Create or truncate backupFile
	backupFile, err := os.Create(backupFilePath)
	if err != nil {
		return nil, err
	}

	return bufio.NewWriter(backupFile), nil
}

// backupChanges records the details of a renaming operation to the specified
// writer so that it may be reverted if necessary. If a writer is not specified
// it records the changes to the filesystem.
func backupChanges(
	changes file.Changes,
	fileName string,
	w io.Writer,
) error {
	var err error

	if w == nil {
		w, err = createBackupFile(fileName)
		if err != nil {
			return err
		}
	}

	err = changes.RenderJSON(w)
	if err != nil {
		return err
	}

	if f, ok := w.(*bufio.Writer); ok {
		return f.Flush()
	}

	return nil
}
