package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
)

const (
	DiceDBVersion     = "0.1.0"
	DefaultConfigName = "dicedb.conf"
	DefaultConfigDir  = "."

	EvictSimpleFirst   = "simple-first"
	EvictAllKeysRandom = "allkeys-random"
	EvictAllKeysLRU    = "allkeys-lru"
	EvictAllKeysLFU    = "allkeys-lfu"
	EvictBatchKeysLRU  = "batch_keys_lru"

	DefaultKeysLimit     int     = 200000000
	DefaultEvictionRatio float64 = 0.1

	defaultConfigTemplate = `# Configuration file for Dicedb

# Version
version = "0.1.0"

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
persistence.enabled = true
persistence.aof_file = "./dice-master.aof"
persistence.persistence_enabled = true
persistence.write_aof_on_cleanup = false
persistence.wal-dir = "./"
persistence.restore-wal = false
persistence.wal-engine = "aof"

# Logging Configuration
logging.log_level = "info"
logging.log_dir = "/tmp/dicedb"

# Authentication Configuration
auth.username = "dice"
auth.password = ""

# Network Configuration
network.io_buffer_length = 512
network.io_buffer_length_max = 51200`
)

var (
	CustomConfigFilePath = utils.EmptyStr
	CustomConfigDirPath  = utils.EmptyStr
)

type Config struct {
	Version     string      `config:"version" default:"0.1.0"`
	InstanceID  string      `config:"instance_id"`
	Auth        auth        `config:"auth"`
	RespServer  respServer  `config:"async_server"`
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

type respServer struct {
	Addr      string `config:"addr" default:"0.0.0.0" validate:"ipv4"`
	Port      int    `config:"port" default:"7379" validate:"number,gte=0,lte=65535"`
	KeepAlive int32  `config:"keepalive" default:"300"`
	Timeout   int32  `config:"timeout" default:"300"`
	MaxConn   int32  `config:"max_conn" default:"0"`
}

type http struct {
	Enabled bool `config:"enabled" default:"true"`
	Port    int  `config:"port" default:"8082" validate:"number,gte=0,lte=65535"`
}

type websocket struct {
	Enabled                 bool          `config:"enabled" default:"true"`
	Port                    int           `config:"port" default:"8379" validate:"number,gte=0,lte=65535"`
	MaxWriteResponseRetries int           `config:"max_write_response_retries" default:"3" validate:"min=0"`
	WriteResponseTimeout    time.Duration `config:"write_response_timeout" default:"10s"`
}

type performance struct {
	WatchChanBufSize       int           `config:"watch_chan_buf_size" default:"20000"`
	ShardCronFrequency     time.Duration `config:"shard_cron_frequency" default:"1s"`
	MultiplexerPollTimeout time.Duration `config:"multiplexer_poll_timeout" default:"100ms"`
	MaxClients             int32         `config:"max_clients" default:"20000" validate:"min=0"`
	StoreMapInitSize       int           `config:"store_map_init_size" default:"1024000"`
	AdhocReqChanBufSize    int           `config:"adhoc_req_chan_buf_size" default:"20"`
	EnableProfiling        bool          `config:"profiling" default:"false"`
	EnableWatch            bool          `config:"enable_watch" default:"false"`
	NumShards              int           `config:"num_shards" default:"-1" validate:"oneof=-1|min=1,lte=128"`
}

type memory struct {
	MaxMemory      int64   `config:"max_memory" default:"0" validate:"min=0"`
	EvictionPolicy string  `config:"eviction_policy" default:"allkeys-lfu" validate:"oneof=simple-first allkeys-random allkeys-lru allkeys-lfu"`
	EvictionRatio  float64 `config:"eviction_ratio" default:"0.9" validate:"min=0,lte=1"`
	KeysLimit      int     `config:"keys_limit" default:"200000000" validate:"min=0"`
	LFULogFactor   int     `config:"lfu_log_factor" default:"10" validate:"min=0"`
}

type persistence struct {
	Enabled           bool   `config:"enabled" default:"false"`
	AOFFile           string `config:"aof_file" default:"./dice-master.aof" validate:"filepath"`
	WriteAOFOnCleanup bool   `config:"write_aof_on_cleanup" default:"false"`
	WALDir            string `config:"wal-dir" default:"./" validate:"dirpath"`
	RestoreFromWAL    bool   `config:"restore-wal" default:"false"`
	WALEngine         string `config:"wal-engine" default:"aof" validate:"oneof=sqlite aof"`
}

type logging struct {
	LogLevel string `config:"log_level" default:"info" validate:"oneof=debug info warn error"`
	LogDir   string `config:"log_dir" default:"/tmp/dicedb" validate:"dirpath"`
}

type network struct {
	IOBufferLengthMAX int `config:"io_buffer_length_max" default:"51200" validate:"min=0,max=1048576"` // max is 1MB'
	IOBufferLength    int `config:"io_buffer_length" default:"512" validate:"min=0"`
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

	if _, err := file.WriteString(defaultConfigTemplate); err != nil {
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

func MergeFlags(flags *Config) {
	flagset := flag.CommandLine
	flagset.Visit(func(f *flag.Flag) {
		// updating values for flags that were explicitly set by the user
		switch f.Name {
		case "host":
			DiceConfig.RespServer.Addr = flags.RespServer.Addr
		case "port":
			DiceConfig.RespServer.Port = flags.RespServer.Port
		case "enable-http":
			DiceConfig.HTTP.Enabled = flags.HTTP.Enabled
		case "http-port":
			DiceConfig.HTTP.Port = flags.HTTP.Port
		case "enable-websocket":
			DiceConfig.WebSocket.Enabled = flags.WebSocket.Enabled
		case "websocket-port":
			DiceConfig.WebSocket.Port = flags.WebSocket.Port
		case "num-shards":
			DiceConfig.Performance.NumShards = flags.Performance.NumShards
		case "enable-watch":
			DiceConfig.Performance.EnableWatch = flags.Performance.EnableWatch
		case "enable-profiling":
			DiceConfig.Performance.EnableProfiling = flags.Performance.EnableProfiling
		case "log-level":
			DiceConfig.Logging.LogLevel = flags.Logging.LogLevel
		case "log-dir":
			DiceConfig.Logging.LogDir = flags.Logging.LogDir
		case "enable-persistence":
			DiceConfig.Persistence.Enabled = flags.Persistence.Enabled
		case "restore-from-wal":
			DiceConfig.Persistence.RestoreFromWAL = flags.Persistence.RestoreFromWAL
		case "wal-engine":
			DiceConfig.Persistence.WALEngine = flags.Persistence.WALEngine
		case "require-pass":
			DiceConfig.Auth.Password = flags.Auth.Password
		case "keys-limit":
			DiceConfig.Memory.KeysLimit = flags.Memory.KeysLimit
		case "eviction-ratio":
			DiceConfig.Memory.EvictionRatio = flags.Memory.EvictionRatio
		}
	})
}
