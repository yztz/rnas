package rnas

import (
	"database/sql"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

var _db *sql.DB

// Initialize the database tables
func InitDB(db *sql.DB) {
	// Create table if not exists
	createConfigTableSQL := `
	CREATE TABLE IF NOT EXISTS configs (
		id TEXT PRIMARY KEY,
		config TEXT
	);
	`
	_, err := db.Exec(createConfigTableSQL)
	if err != nil {
		log.Fatal("failed to create table:", err)
	}

	// foreign key cascad
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Error enabling foreign keys:", err)
	}


	createFileStripesTable := `
	CREATE TABLE IF NOT EXISTS file_stripes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		filepath TEXT UNIQUE,
		k INTEGER,
		m INTEGER,
		size INTEGER,
		stripe_depth INTEGER,
		min_depth INTEGER,
		config_name TEXT
	);`
	if _, err := db.Exec(createFileStripesTable); err != nil {
		log.Fatal("failed to create file_stripes table:", err)
	}

	createShardsTable := `
	CREATE TABLE IF NOT EXISTS shards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER,
		shard_index INTEGER,
		size INTEGER,
		server_id TEXT,
		is_data_shard BOOLEAN,
		shard_hashname TEXT,
		FOREIGN KEY (file_id) REFERENCES file_stripes(id) ON DELETE CASCADE,
		UNIQUE(file_id, shard_index)
	);`
	if _, err := db.Exec(createShardsTable); err != nil {
		log.Fatal("failed to create shards table:", err)
	}

	_db = db
}

// Save file stripe configuration to the database
func saveFileStripe(file *FileStripe) error {
	result, err := _db.Exec(
		`INSERT OR REPLACE INTO file_stripes 
		(filepath, k, m, config_name, size, stripe_depth, min_depth) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		file.Filepath, file.K, file.M, file.ConfigName, file.Size, file.StripeDepth, file.MinDepth)
	if err != nil {
		return fmt.Errorf("failed to insert file stripe config: %v", err)
	}

	fileID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %v", err)
	}

	file.ID = int(fileID)

	return nil
}

func getFileStripe(filepath string) (*FileStripe, error) {
	fs := &FileStripe{}
	row := _db.QueryRow(
		`SELECT id, k, m, config_name, size, stripe_depth, min_depth FROM file_stripes WHERE filepath = ?`, filepath)
	err := row.Scan(&fs.ID, &fs.K, &fs.M, &fs.ConfigName, &fs.Size, &fs.StripeDepth, &fs.MinDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to query file_strips: %v", err)
	}

	return fs, nil
}

// Save shard information to the database
func saveShard(shard *Shard) error {
	_, err := _db.Exec(`INSERT INTO shards (file_id, shard_index, server_id, shard_hashname, is_data_shard, size) VALUES (?, ?, ?, ?, ?, ?)`,
		shard.fileID, shard.shardIndex, shard.serverID, shard.shardHashname, shard.dataShard, shard.size)
	if err != nil {
		return fmt.Errorf("failed to insert shard info: %v", err)
	}
	return nil
}



// Read shard information by file ID
func getShards(fileID int) ([]Shard, error) {

	rows, err := _db.Query(`SELECT file_id, shard_index, server_id, shard_hashname, is_data_shard,size FROM shards WHERE file_id = ?`, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query shards: %v", err)
	}
	defer rows.Close()

	var shards []Shard

	for rows.Next() {
		shard := Shard{}
		if err := rows.Scan(&shard.fileID, &shard.shardIndex, &shard.serverID, &shard.shardHashname, &shard.dataShard, &shard.size); err != nil {
			return nil, fmt.Errorf("failed to scan shard row: %v", err)
		}
		shards = append(shards, shard)
	}

	return shards, nil
}

func SaveConfigToDB(config *Config) error {
	// Insert config into the table
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	_, err = _db.Exec("INSERT OR REPLACE INTO configs (id, config) VALUES (?, ?)", config.Name, string(configData))
	if err != nil {
		return fmt.Errorf("failed to insert config: %v", err)
	}

	return nil
}

func LoadConfigFromDB(configName string) (Config, error) {
	var config Config
	row := _db.QueryRow("SELECT config FROM configs WHERE id = ?", configName)
	var configData string
	err := row.Scan(&configData)
	if err != nil {
		if err == sql.ErrNoRows {
			return config, fmt.Errorf("config '%s' not found", configName)
		}
		return config, fmt.Errorf("failed to query config: %v", err)
	}

	err = json.Unmarshal([]byte(configData), &config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return config, nil
}
