package storage

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
)

// WebDAVDriver stores the client for interacting with the WebDAV server
type WebDAVDriver struct {
	client *gowebdav.Client
}

// Init initializes the WebDAV storage driver by setting up the client and verifying connection
func (d *WebDAVDriver) Init(s *StorageConfig) error {
	d.client = gowebdav.NewClient(s.Path, s.Username, s.Password)
	d.client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0")
	err := d.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to WebDAV server: %v", err)
	}
	return nil
}

// Read reads data from a file at a specific offset
func (d *WebDAVDriver) ReadStream(path string, offset int64, length int64) (io.ReadCloser, error) {
	fullPath := "/" + path
	rc, err := d.client.ReadStreamRange(fullPath, int64(offset), length)
	if err != nil {
		return nil, fmt.Errorf("failed to read file from WebDAV: %v", err)
	}

	return rc, nil
}

func (d *WebDAVDriver) Read(path string, offset int64, data []byte) (int, error) {
	fullPath := "/" + path
	rc, err := d.client.ReadStreamRange(fullPath, int64(offset), int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("failed to read file from WebDAV: %v", err)
	}

	n, err := io.ReadFull(rc, data)

	if err != nil {
		return 0, fmt.Errorf("failed to read file from WebDAV: %v", err)
	}
	return n, nil
	// content, err := d.client.Read(fullPath)
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to read file from WebDAV: %v", err)
	// }

	// n := copy(data, content[offset:])
	// return n, nil
}

// Create creates a new file, returns error if file already exists
func (d *WebDAVDriver) Create(path string, data []byte) error {
	fullPath := "/" + path

	err := d.Find(path)

	if err == nil {
		// return fmt.Errorf("file %s already exists", path)
		log.Warnf("file %s has existed, will be overwritten", path)
	}

	err = d.client.Write(fullPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file on WebDAV: %v", err)
	}
	return nil
}

// Find checks if a file or directory exists
func (d *WebDAVDriver) Find(path string) error {
	fullPath := "/" + path
	_, err := d.client.Stat(fullPath)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a file
func (d *WebDAVDriver) Delete(path string) error {
	fullPath := "/" + path
	err := d.client.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file from WebDAV: %v", err)
	}
	return nil
}

func (d *WebDAVDriver) Mkdir(path string) error {
	fullPath := "/" + path
	err := d.client.MkdirAll(fullPath, 0644)
	// err := d.client.Mkdir(fullPath, 0644)
	if err != nil {
		return fmt.Errorf("failed to create mkdir: %v", err)
	}
	return nil
}
