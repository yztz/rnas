package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/yztz/rnas"
)

var _Dryrun = false
var Dryrun *bool = &_Dryrun

func main() {

	log.SetLevel(log.DebugLevel)
	// log.SetReportCaller(true)

	// Define CLI flags
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	putCmd := flag.NewFlagSet("put", flag.ExitOnError)
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	// putCmd := flag.NewFlagSet("put", flag.ExitOnError)
	// getCmd := flag.NewFlagSet("get", flag.ExitOnError)

	// Define flags for `create` command
	createConfig := createCmd.String("config", "", "Path to the configuration file")
	testConfig := testCmd.String("config", "default", "Name of configuration")
	putConfig := putCmd.String("config", "default", "Name of configuration")
	Dryrun = putCmd.Bool("dryrun", false, "Dryrun")
	getConfig := getCmd.String("config", "default", "Name of configuration")
	
	// configName := createCmd.String("name", "default", "Name of configuration")

	// Define flags for `put` and `get` commands
	// putKey := putCmd.String("key", "", "Key for the `put` operation")
	// getKey := getCmd.String("key", "", "Key for the `get` operation")

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command> [options]")
		fmt.Println("Commands: create, put, get")
		return
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v\n", err)
	}
	defer db.Close()

	rnas.InitDB(db)

	switch os.Args[1] {
	case "create":
		createCmd.Parse(os.Args[2:])
		handleCreate(*createConfig)
	case "test":
		testCmd.Parse(os.Args[2:])
		handleTest(*testConfig)
	case "put":
		putCmd.Parse(os.Args[2:])
		filepath := putCmd.Arg(0)
		targetPath := putCmd.Arg(1)
		handlePut(*putConfig, filepath, targetPath)
	case "get":
		getCmd.Parse(os.Args[2:])
		filepath := getCmd.Arg(0)
		targetPath := getCmd.Arg(1)
		handleGet(*getConfig, filepath, targetPath)
	// case "put":
	// 	putCmd.Parse(os.Args[2:])
	// 	handlePut(*putKey)
	// case "get":
	// 	getCmd.Parse(os.Args[2:])
	// 	handleGet(*getKey)
	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println("Usage: go run main.go <command> [options]")
		fmt.Println("Commands: create, put, get")
	}
}

func handleCreate(configPath string) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v\n", err)
	}

	var config rnas.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v\n", err)
	}

	config.Init()

	// Save configuration to database
	err = rnas.SaveConfigToDB(&config)
	if err != nil {
		log.Fatalf("Failed to save config to database: %v\n", err)
	}

	log.Println("Configuration created and saved to database.")
}

func handleTest(configName string) {
	config,err := rnas.LoadConfigFromDB(configName)

	if err != nil {
		log.Fatal(err)
	}
	config.Init()
	config.TestAll()
	
}

func getFileSize(filepath string) (int64, error) {
	// 使用 os.Stat 获取文件信息
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}

	// 返回文件大小（单位：字节）
	return fileInfo.Size(), nil
}


func handlePut(configName, filepath, targetPath string) {
	config,err := rnas.LoadConfigFromDB(configName)
	if err != nil {
		log.Fatal(err)
	}
	size, err := getFileSize(filepath)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	config.Init()
	err = config.Put(targetPath, size, f)
	if err != nil {
		log.Fatal(err)
	}
}

func handleGet(configName, filepath, targetPath string) {
	config,err := rnas.LoadConfigFromDB(configName)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	config.Init()
	reader, err := config.ReadStream(filepath)
	if err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	w, err := io.Copy(f, reader)
	end := time.Since(now)
	fmt.Printf("%s has been retrived, took %v, speed %.2fB/s\n", filepath, end, float64(w) / float64(end.Seconds()))

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("done, %d bytes written.\n", w)
}