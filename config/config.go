// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
)

var (
	CustomConfigFilePath = utils.EmptyStr
	CustomConfigDirPath  = utils.EmptyStr
)

type Config struct {
	Version    string     `config:"version" default:"0.1.0"`
	InstanceID string     `config:"instance_id"`
	RespServer respServer `config:"async_server"`
	HTTP       http       `config:"http"`
	WebSocket  websocket  `config:"websocket"`
	WAL        WALConfig  `config:"WAL"`
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

type WALConfig struct {
	// Directory where WAL log files will be stored
	LogDir string `config:"log_dir" default:"tmp/dicedb-wal"`
	// WAL buffering mode: 'buffered' (writes buffered in memory) or 'unbuffered' (immediate disk writes)
	WalMode string `config:"wal_mode" default:"buffered" validate:"oneof=buffered unbuffered"`
	// Write mode: 'default' (OS handles syncing) or 'fsync' (explicit fsync after writes)
	WriteMode string `config:"write_mode" default:"default" validate:"oneof=default fsync"`
	// Size of the write buffer in megabytes
	BufferSizeMB int `config:"buffer_size_mb" default:"1" validate:"min=1"`
	// How WAL rotation is triggered: 'segment-size' (based on file size) or 'time' (based on duration)
	RotationMode string `config:"rotation_mode" default:"segemnt-size" validate:"oneof=segment-size time"`
	// Maximum size of a WAL segment file in megabytes before rotation
	MaxSegmentSizeMB int `config:"max_segment_size_mb" default:"16" validate:"min=1"`
	// Time interval in seconds after which WAL segment is rotated when using time-based rotation
	MaxSegmentRotationTime time.Duration `config:"max_segment_rotation_time" default:"60s" validate:"min=1s"`
	// Time interval in Milliseconds after which buffered WAL data is synced to disk
	BufferSyncInterval time.Duration `config:"buffer_sync_interval" default:"200ms" validate:"min=1ms"`
	// How old segments are removed: 'num-segments' (keep N latest), 'time' (by age), or 'checkpoint' (after checkpoint)
	RetentionMode string `config:"retention_mode" default:"num-segments" validate:"oneof=num-segments time checkpoint"`
	// Maximum number of WAL segment files to retain when using num-segments retention
	MaxSegmentCount int `config:"max_segment_count" default:"10" validate:"min=1"`
	// Time interval in Seconds till which WAL segments are retained when using time-based retention
	MaxSegmentRetentionDuration time.Duration `config:"max_segment_retention_duration" default:"600s" validate:"min=1s"`
	// How to handle WAL corruption on recovery: 'strict' (fail), 'truncate' (truncate at corruption), 'ignore' (skip corrupted)
	RecoveryMode string `config:"recovery_mode" default:"strict" validate:"oneof=strict truncate ignore"`
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
		}
	})
}

type DiceDBConfig struct {
	Host       string `mapstructure:"host" default:"0.0.0.0" description:"the host address to bind to"`
	Port       int    `mapstructure:"port" default:"7379" description:"the port to bind to"`
	EnableHTTP bool   `mapstructure:"enable-http" default:"false" description:"enable http server"`

	Username string `mapstructure:"username" default:"dicedb" description:"the username to use for authentication"`
	Password string `mapstructure:"password" default:"" description:"the password to use for authentication"`

	LogLevel string `mapstructure:"log-level" default:"info" description:"the log level"`

	EnableWAL bool   `mapstructure:"enable-wal" default:"true" description:"enable write-ahead logging"`
	WALEngine string `mapstructure:"wal-engine" default:"aof" description:"wal engine to use, values: sqlite, aof"`

	EnableWatch bool `mapstructure:"enable-watch" default:"false" description:"enable support for .WATCH commands and real-time reactivity"`
	MaxClients  int  `mapstructure:"max-clients" default:"20000" description:"the maximum number of clients to accept"`
	NumShards   int  `mapstructure:"num-shards" default:"-1" description:"number of shards to create. defaults to number of cores"`
}

var GlobalDiceDBConfig *DiceDBConfig

func Init(flags *pflag.FlagSet) {
	viper.SetConfigName("dicedb")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/dicedb")

	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok && err != nil {
		panic(err)
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		viper.Set(flag.Name, flag.Value.String())
	})

	if err := viper.Unmarshal(&GlobalDiceDBConfig); err != nil {
		panic(err)
	}
}
