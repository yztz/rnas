package storage

import "io"

type StorageConfig struct {
	Path     string `json:"path,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type StorageDriver interface {
	// init driver
	Init(s *StorageConfig) error
	// read data from file
	Read(path string, offset int64, data []byte) (int, error)
	ReadStream(path string, offset int64, length int64) (io.ReadCloser, error)
	// create file, error when exist
	Create(path string, data []byte) error
	// file or dir exist
	Find(path string) error
	// delete file
	Delete(path string) error
	// create dir
	Mkdir(path string) error
}


var DriverInitializers = map[string]func() StorageDriver{
	"local": func() StorageDriver { return &LocalDriver{} },
	"webdav": func() StorageDriver { return &WebDAVDriver{} },
	"dryrun": func() StorageDriver { return &DryrunDriver{} },
}



