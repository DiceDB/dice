package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	DiceDBVersion     string = "0.0.5"
	DefaultConfigName string = "dicedb.conf"
	DefaultConfigDir  string = "."

	EvictSimpleFirst   string = "simple-first"
	EvictAllKeysRandom string = "allkeys-random"
	EvictAllKeysLRU    string = "allkeys-lru"
	EvictAllKeysLFU    string = "allkeys-lfu"

	DefaultKeysLimit     int     = 200000000
	DefaultEvictionRatio float64 = 0.1
)

type Config struct {
	Version     string      `config:"version" default:"0.0.5"`
	InstanceID  string      `config:"instance_id"`
	Auth        auth        `config:"auth"`
	AsyncServer asyncServer `config:"async_server"`
	HTTP        http        `config:"http"`
	WebSocket   websocket   `config:"websocket"`
	Performance performance `config:"performance"`
	Memory      memory      `config:"memory"`
	Persistence persistence `config:"persistence"`
	Logging     logging     `config:"logging"`
	Network     network     `config:"network"`
}

type auth struct {
	UserName string `config:"username" default:"dice"`
	Password string `config:"password"`
}

type asyncServer struct {
	Addr      string `config:"addr" default:"0.0.0.0"`
	Port      int    `config:"port" default:"7379" validate:"min=1024,max=65535"`
	KeepAlive int32  `config:"keepalive" default:"300"`
	Timeout   int32  `config:"timeout" default:"300"`
	MaxConn   int32  `config:"max_conn" default:"0"`
}

type http struct {
	Enabled bool `config:"enabled" default:"true"`
	Port    int  `config:"port" default:"8082" validate:"min=1024,max=65535"`
}

type websocket struct {
	Enabled                 bool          `config:"enabled" default:"true"`
	Port                    int           `config:"port" default:"8379" validate:"min=1024,max=65535"`
	MaxWriteResponseRetries int           `config:"max_write_response_retries" default:"3" validate:"min=0"`
	WriteResponseTimeout    time.Duration `config:"write_response_timeout" default:"10s"`
}

type performance struct {
	WatchChanBufSize       int           `config:"watch_chan_buf_size" default:"20000"`
	ShardCronFrequency     time.Duration `config:"shard_cron_frequency" default:"1s"`
	MultiplexerPollTimeout time.Duration `config:"multiplexer_poll_timeout" default:"100ms"`
	MaxClients             int32         `config:"max_clients" default:"20000" validate:"min=0"`
	EnableMultiThreading   bool          `config:"enable_multithreading" default:"false"`
	StoreMapInitSize       int           `config:"store_map_init_size" default:"1024000"`
	AdhocReqChanBufSize    int           `config:"adhoc_req_chan_buf_size" default:"20"`
	EnableProfiling        bool          `config:"profiling" default:"false"`
	EnableWatch            bool          `config:"enable_watch" default:"false"`
	NumShards              int           `config:"num_shards" default:"-1" validate:"oneof=-1|min=1,lte=128"`
}

type memory struct {
	MaxMemory      int64   `config:"max_memory" default:"0"`
	EvictionPolicy string  `config:"eviction_policy" default:"allkeys-lfu" validate:"oneof=simple-first allkeys-random allkeys-lru allkeys-lfu"`
	EvictionRatio  float64 `config:"eviction_ratio" default:"0.9" validate:"min=0,lte=1"`
	KeysLimit      int     `config:"keys_limit" default:"200000000" validate:"min=0"`
	LFULogFactor   int     `config:"lfu_log_factor" default:"10" validate:"min=0"`
}

type persistence struct {
	AOFFile            string `config:"aof_file" default:"./dice-master.aof" validate:"filepath"`
	PersistenceEnabled bool   `config:"persistence_enabled" default:"true"`
	WriteAOFOnCleanup  bool   `config:"write_aof_on_cleanup" default:"false"`
	EnableWAL          bool   `config:"enable-wal" default:"true"`
	WALDir             string `config:"wal-dir" default:"./" validate:"dirpath"`
	RestoreFromWAL     bool   `config:"restore-wal" default:"false"`
	WALEngine          string `config:"wal-engine" default:"aof" validate:"oneof=sqlite aof"`
}

type logging struct {
	LogLevel string `config:"log_level" default:"info" validate:"oneof=debug info warn error"`
}

type network struct {
	IOBufferLengthMAX int `config:"io_buffer_length_max" default:"51200" validate:"min=0,max=1048576"` // max is 1MB'
	IOBufferLength    int `config:"io_buffer_length" default:"512" validate:"min=0"`
}

func init() {
	configFilePath := filepath.Join(DefaultConfigDir, DefaultConfigName)
	if err := loadDiceConfig(configFilePath); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
}

// DiceConfig is the global configuration object for dice
var DiceConfig = &Config{}

func CreateConfigFile(configFilePath string) error {
	// Check if the config file already exists
	if _, err := os.Stat(configFilePath); err == nil {
		if err := loadDiceConfig(configFilePath); err != nil {
			return fmt.Errorf("failed to load existing configuration: %w", err)
		}
		return nil
	}

	// Attempt to write a new config file
	if err := writeConfigFile(configFilePath); err != nil {
		slog.Warn("Failed to create config file, starting with defaults.", slog.Any("error", err))
		return nil // Continuing with defaults; may reconsider behavior.
	}

	// Load the new configuration
	if err := loadDiceConfig(configFilePath); err != nil {
		return fmt.Errorf("failed to load newly created configuration: %w", err)
	}

	slog.Info("Config file successfully created.", slog.String("path", configFilePath))
	return nil
}

// writeConfigFile writes the default configuration to the specified file path
func writeConfigFile(configFilePath string) error {
	content := `# Configuration file for Dicedb

# Version
version = "0.0.5"

# Async Server Configuration
async_server.addr = "0.0.0.0"
async_server.port = 7379
async_server.keepalive = 300
async_server.timeout = 300
async_server.max_conn = 0

# HTTP Configuration
http.enabled = false
http.port = 8082

# WebSocket Configuration
websocket.enabled = false
websocket.port = 8379
websocket.max_write_response_retries = 3
websocket.write_response_timeout = 10s

# Performance Configuration
performance.watch_chan_buf_size = 20000
performance.shard_cron_frequency = 1s
performance.multiplexer_poll_timeout = 100ms
performance.max_clients = 20000
performance.enable_multithreading = false
performance.store_map_init_size = 1024000
performance.adhoc_req_chan_buf_size = 20
performance.enable_profiling = false
performance.enable_watch = false
performance.num_shards = -1

# Memory Configuration
memory.max_memory = 0
memory.eviction_policy = "allkeys-lfu"
memory.eviction_ratio = 0.9
memory.keys_limit = 200000000
memory.lfu_log_factor = 10

# Persistence Configuration
persistence.aof_file = "./dice-master.aof"
persistence.persistence_enabled = true
persistence.write_aof_on_cleanup = false
persistence.enable-wal = true
persistence.wal-dir = "./"
persistence.restore-wal = false
persistence.wal-engine = "aof"

# Logging Configuration
logging.log_level = "info"

# Authentication Configuration
auth.username = "dice"
auth.password = ""

# Network Configuration
network.io_buffer_length = 512
network.io_buffer_length_max = 51200`

	// Check if the directory exists or not
	dir := filepath.Dir(configFilePath)
	if _, err := os.Stat(dir); err != nil {
		return err
	}

	slog.Info("creating default config file at", slog.Any("path", configFilePath))
	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	return nil
}

func loadDiceConfig(configFilePath string) error {
	parser := NewConfigParser()
	if err := parser.ParseFromFile(configFilePath); err != nil {
		slog.Warn("Failed to parse config file", slog.String("error", err.Error()), slog.String("message", "Loading default configurations"))
		return parser.ParseDefaults(DiceConfig)
	}

	return parser.Loadconfig(DiceConfig)
}

// This function returns the config file path based on the OS
// func getConfigPath() string {
// 	switch runtime.GOOS {
// 	case "windows":
// 		FileLocation = filepath.Join("C:", "ProgramData", "dice", DefaultConfigName)
// 	case "darwin", "linux":
// 		FileLocation = filepath.Join(string(filepath.Separator), "etc", "dice", DefaultConfigName)
// 	default:
// 		// Default to current directory if OS is unknown
// 		FileLocation = filepath.Join(".", DefaultConfigName)
// 	}
// 	return FileLocation
// }

// ResetConfig resets the DiceConfig to default configurations. This function is only used for testing purposes
