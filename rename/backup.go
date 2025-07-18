package rename

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

	"github.com/ayoisaiah/f2/v2/internal/file"
	"github.com/ayoisaiah/f2/v2/internal/osutil"
)

func createBackupFile(fileName string) (io.Writer, error) {
	backupFilePath := filepath.Join(
		os.TempDir(),
		"f2",
		"backups",
		fileName,
	)

	err := os.MkdirAll(filepath.Dir(backupFilePath), osutil.DirPermission)
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
	cleanedDirs []string,
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

	b := file.Backup{
		Changes:     changes,
		CleanedDirs: cleanedDirs,
	}

	err = b.RenderJSON(w)
	if err != nil {
		return err
	}

	if f, ok := w.(*bufio.Writer); ok {
		return f.Flush()
	}

	return nil
}
