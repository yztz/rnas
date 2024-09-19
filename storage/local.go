package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	// log "github.com/sirupsen/logrus"
)

// LocalDriver stores the base path for local storage
type LocalDriver struct {
	basePath string
}

// Init initializes the local storage driver by checking the path exists
func (d *LocalDriver) Init(s *StorageConfig) error {
	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		return fmt.Errorf("local path does not exist: %s", s.Path)
	}
	d.basePath = s.Path
	return nil
}

// Read reads data from a file at a specific offset
func (d *LocalDriver) Read(path string, offset int64, data []byte) (int, error) {
	fullPath := filepath.Join(d.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	file.Seek(int64(offset), 0)
	n, err := file.Read(data)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %v", err)
	}
	return n, nil
}

func (d *LocalDriver) ReadStream(path string, offset int64, length int64) (io.ReadCloser, error) {
	fullPath := filepath.Join(d.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	file.Seek(int64(offset), 0)

	return LimitReadCloser(file, length), nil
}

// Create creates a new file, returns error if file already exists
func (d *LocalDriver) Create(path string, data []byte) error {
	fullPath := filepath.Join(d.basePath, path)
	if _, err := os.Stat(fullPath); err == nil {
		// return fmt.Errorf("file %s already exists", path)
		log.Warnf("file %s has existed, will be overwritten", path)
	}
	err := os.WriteFile(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	return nil
}

// Find checks if a file or directory exists
func (d *LocalDriver) Find(path string) error {
	fullPath := filepath.Join(d.basePath, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file or directory does not exist: %s", path)
	}
	return nil
}

// Delete deletes a file
func (d *LocalDriver) Delete(path string) error {
	fullPath := filepath.Join(d.basePath, path)
	err := os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

func (d *LocalDriver) Mkdir(path string) error {
	fullPath := filepath.Join(d.basePath, path)
	err := os.MkdirAll(fullPath, 0744)
	if err != nil {
		return fmt.Errorf("failed to mkdir: %v", err)
	}
	return nil
}



type LimitedReadCloser struct {
    R io.ReadCloser
    L *io.LimitedReader
}

func (l *LimitedReadCloser) Read(p []byte) (n int, err error) {
    return l.L.Read(p)
}

func (l *LimitedReadCloser) Close() error {
    return l.R.Close()
}

func LimitReadCloser(rc io.ReadCloser, limit int64) io.ReadCloser {
    return &LimitedReadCloser{
        R: rc,
        L: &io.LimitedReader{
            R: rc,
            N: limit,
        },
    }
}