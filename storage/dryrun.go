package storage

import (
	"fmt"
	"io"
	"sync"
)

type DryrunDriver struct {
	dataMap sync.Map
}

func (d *DryrunDriver) Init(s *StorageConfig) error {
	return nil
}

func (d *DryrunDriver) Read(path string, offset int64, data []byte) (int, error) {
	v, ok := d.dataMap.Load(path)
	if !ok {
		return 0, fmt.Errorf("no data available")
	}
	src := v.([]byte)
	src = src[offset:]
	n := copy(data, src)
	return n, nil
}

func (d *DryrunDriver) ReadStream(path string, offset int64, length int64) (io.ReadCloser, error) {
	panic("unsupport")
}

func (d *DryrunDriver) Create(path string, data []byte) error {
	_, ok := d.dataMap.Load(path)
	if ok {
		return fmt.Errorf("%s already exists", path)
	}
	d.dataMap.Store(path, data)
	return nil
}

func (d *DryrunDriver) Delete(path string) error {
	d.dataMap.Delete(path)
	return nil
}

func (d *DryrunDriver) Mkdir(path string) error {
	return nil
}

func (d *DryrunDriver) Find(path string) error {
	_, ok := d.dataMap.Load(path)
	if ok { // exists
		return nil
	}
	return fmt.Errorf("%s not exists", path)
}