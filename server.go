package rnas

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/yztz/rnas/storage"
)

const fakeSuffix = "jpg" 

// Server struct for storing server details
type Server struct {
	Type     string `json:"type"`
	Id       string `json:"id"`
	UploadBandwidth float64 `json:"uploadBandwidth,omitempty"`
	DownloadBandwidth float64 `json:"downloadBandwidth,omitempty"`
	storage.StorageConfig
	
	driver storage.StorageDriver
	mu sync.Mutex
	config *Config
	reachable bool
}

func (server *Server) Init(config *Config) {
	var t string

	initFunc, ok := storage.DriverInitializers[server.Type]
	if !ok {
		log.Fatalf("Unsupported driver type: %s\n", t)
	}
	server.config = config

	driver := initFunc()
	err := driver.Init(&server.StorageConfig)
	if err != nil {
		log.Errorf("Failed to initialize driver for server: %s", server.Id)
		server.reachable = false
		return
	}
	server.driver = driver
	server.reachable = true

	err = server.driver.Mkdir(config.Name)
	// err := server.TestSpeed()
	if err != nil {
		log.Fatalf("Failed to init config sub-folder, because: %v", err)
	}

	
}

// testServer performs upload and download speed tests based on server type
func (server *Server) TestSpeed() error {
	server.mu.Lock()
	defer server.mu.Unlock()

	testFileName := "testfile.bin"
	fileSize := 2 * 1024 * 1024 // 2 MB

	// Generate random test file content
	randomData := generateRandomData(fileSize)
	
	// Upload file based on server type
	startUpload := time.Now()
	err := server.driver.Create(testFileName, randomData)
	if err != nil {
		return err
	}
	elapsedUpload := time.Since(startUpload)
	log.Infof("- Upload to server %s took %v\n", server.Id, elapsedUpload)
	server.UploadBandwidth = float64(fileSize) / float64(elapsedUpload.Seconds())

	// Download file based on server type
	startDownload := time.Now()
	_, err = server.driver.Read(testFileName, 0, randomData)
	if err != nil {
		server.driver.Delete(testFileName)
		return err
	}
	elapsedDownload := time.Since(startDownload)
	log.Infof("- Download from server %s took %v\n", server.Id, elapsedDownload)
	server.DownloadBandwidth = float64(fileSize) / float64(elapsedDownload.Seconds())

	err = server.driver.Delete(testFileName)
	if err != nil {
		log.Warn("failed to delete speedtest file")
	}
	return nil
}


func (server *Server) PutShard(shard *Shard, data []byte) error {
	// server.mu.Lock()
	// defer server.mu.Unlock()

	prefix := filepath.Join(server.config.Name, strconv.Itoa(shard.fileID))
	shardName := fmt.Sprintf("%s.%s", shard.shardHashname, fakeSuffix)
	err := server.driver.Mkdir(prefix)
	if err != nil {
		return err
	}

	now := time.Now()
	log.Debugf("- start to transfer shard %d with size %d to server[%s]", shard.shardIndex, len(data),shard.serverID)
	err = server.driver.Create(filepath.Join(prefix, shardName), data)
	if err != nil {
		return err
	}
	end := time.Since(now)
	log.Debugf("- transfer shard %d done, took %v", shard.shardIndex, end)
	return nil
}

func (server *Server) GetShard(shard *Shard, data []byte) (int, error) {
	// server.mu.Lock()
	// defer server.mu.Unlock()

	prefix := filepath.Join(server.config.Name, strconv.Itoa(shard.fileID))
	shardName := fmt.Sprintf("%s.%s", shard.shardHashname, fakeSuffix)

	start := time.Now()
	n,err := server.driver.Read(filepath.Join(prefix, shardName), 0, data)
	end := time.Since(start)
	if err != nil {
		return n, err
	}

	log.Debugf("- transfer shard %d done, size %d/%d, took %v", shard.shardIndex, n, len(data), end)

	return n,err
}


func (server *Server) GetShardStream(shard *Shard, offset int64, length int64) (io.ReadCloser, error) {
	// server.mu.Lock()
	// defer server.mu.Unlock()

	prefix := filepath.Join(server.config.Name, strconv.Itoa(shard.fileID))
	shardName := fmt.Sprintf("%s.%s", shard.shardHashname, fakeSuffix)

	n,err := server.driver.ReadStream(filepath.Join(prefix, shardName), offset, length)

	if err != nil {
		return nil, err
	}

	return n,err
}


type Servers []*Server

func(s Servers) Len() int {
	return len(s)
}

func(s Servers) Less(i, j int) bool {
	// more opt can do
	return s[i].DownloadBandwidth > s[j].DownloadBandwidth
}
  
func(s Servers) Swap(i, j int) {
	tmp := s[i]
	s[i] = s[j]
	s[j] = tmp
}