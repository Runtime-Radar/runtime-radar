package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
)

var _ io.Writer = (*RotateFileWriter)(nil)

// RotateFileWriter is a writer that rotates log files based on size and number of backups.
type RotateFileWriter struct {
	mu sync.Mutex

	filename   string
	basename   string
	dir        string
	maxSize    int
	maxBackups int

	file        *os.File
	size        int
	nameMatcher *regexp.Regexp
}

// NewRotateFileWriter creates a new RotateFileWriter for the given filename and
// configuration. It ensures the log directory exists, compiles a regexp to match
// rotated log files, opens the initial log file, and returns a RotateFileWriter.
func NewRotateFileWriter(filename string, maxBackups int, maxSize int) (*RotateFileWriter, error) {
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	if maxBackups <= 0 {
		return nil, fmt.Errorf("maxBackups cannot be less or equal to zero")
	}

	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize cannot be less or equal to zero")
	}

	filename = filepath.Clean(filename)
	basename := filepath.Base(filename)
	dir := filepath.Dir(filename)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("can't make directories for new logfile: %w", err)
	}

	r, err := regexp.Compile(basename + `\.[0-9]+$`)
	if err != nil {
		return nil, fmt.Errorf("can't compile regexp: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("can't open log file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("can't get info about file: %w", err)
	}

	rfw := &RotateFileWriter{
		filename:    filename,
		basename:    basename,
		dir:         dir,
		maxSize:     maxSize,
		maxBackups:  maxBackups,
		nameMatcher: r,
		file:        file,
		size:        int(info.Size()),
	}

	return rfw, nil
}

// Write implements the io.Writer interface. It writes the given bytes to the
// rolling log file, handling log rotation when the size exceeds the max.
// It locks access to the file for the duration of the write.
func (rfw *RotateFileWriter) Write(p []byte) (n int, err error) {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	if rfw.size+len(p) > rfw.maxSize {
		if err := rfw.rotate(); err != nil {
			return 0, fmt.Errorf("can't rotate log file: %w", err)
		}
	}

	n, err = rfw.file.Write(p)
	rfw.size += n

	return
}

// rotate sequentially increments old log suffixes,
// e.g. `app.log.3` -> `app.log.4`, `app.log.2` -> `app.log.3` etc
// and then closes current file and renames it to the first backup.
func (rfw *RotateFileWriter) rotate() error {
	if rfw.maxBackups == 0 {
		return nil
	}

	cnt := 0
	err := filepath.WalkDir(rfw.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't walk directory: %w", err)
		}

		if d.IsDir() && path != rfw.dir {
			return filepath.SkipDir
		}

		if rfw.nameMatcher.MatchString(d.Name()) {
			cnt++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("can't walk directory: %w", err)
	}

	for i := cnt; i > 0; i-- {
		oldName := rfw.filename + "." + strconv.Itoa(i)
		newName := rfw.filename + "." + strconv.Itoa(i+1)
		// just remove everything over `rfw.maxBackups`
		if i >= rfw.maxBackups {
			if err := os.Remove(oldName); err != nil {
				return fmt.Errorf("can't remove old log file: %w", err)
			}
			continue
		}

		if err := os.Rename(oldName, newName); err != nil {
			return fmt.Errorf("can't rename old log file: %w", err)
		}
	}

	// Try to close current file and rename it
	if err := rfw.file.Close(); err != nil {
		return fmt.Errorf("can't close current log file: %w", err)
	}

	if err := os.Rename(rfw.filename, rfw.filename+".1"); err != nil {
		return fmt.Errorf("can't rename current log file: %w", err)
	}

	file, err := os.OpenFile(rfw.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("can't open log file: %w", err)
	}
	rfw.file = file
	rfw.size = 0

	return nil
}
