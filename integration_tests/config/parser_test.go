// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
)

// TestConfig is a test struct that mimics your actual config structure
type TestConfig struct {
	Version     string      `config:"version" default:"0.1.0"`
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

func TestNewConfigParser(t *testing.T) {
	parser := config.NewConfigParser()
	if parser == nil {
		t.Fatal("NewConfigParser returned nil")
	}
}

func TestParseFromFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		setupErr bool
	}{
		{
			name: "valid config",
			content: `host=testhost
port=9090
log_level=debug`,
			wantErr: false,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
		},
		{
			name: "malformed config",
			content: `host=testhost
invalid-line
port=9090`,
			wantErr: false,
		},
		{
			name:     "non-existent file",
			setupErr: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := config.NewConfigParser()

			// Create temporary config file
			tempDir := t.TempDir()
			filename := filepath.Join(tempDir, "dicedb.conf")

			if !tt.setupErr {
				err := os.WriteFile(filename, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			err := parser.ParseFromFile(filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseFromStdin(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid input",
			input: `host=testhost
port=9090
log_level=debug`,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
		},
		{
			name: "malformed input",
			input: `host=testhost
invalid-line
port=9090`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := config.NewConfigParser()

			// Store original stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			// Create a pipe and pass the test input
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdin = r

			go func() {
				defer w.Close()
				w.Write([]byte(tt.input))
			}()

			err = parser.ParseFromStdin()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFromStdin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseDefaults(t *testing.T) {
	tests := []struct {
		name    string
		cfg     interface{}
		wantErr bool
	}{
		{
			name:    "valid struct",
			cfg:     &TestConfig{},
			wantErr: false,
		},
		{
			name:    "nil pointer",
			cfg:     nil,
			wantErr: true,
		},
		{
			name:    "non-pointer",
			cfg:     TestConfig{},
			wantErr: true,
		},
		{
			name:    "pointer to non-struct",
			cfg:     new(string),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := config.NewConfigParser()
			err := parser.ParseDefaults(tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDefaults() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.cfg != nil {
				cfg := tt.cfg.(*TestConfig)
				if cfg.AsyncServer.Addr != "0.0.0.0" || cfg.AsyncServer.Port != 7379 || cfg.Logging.LogLevel != "info" {
					t.Error("Default values were not properly set")
				}
			}
		})
	}
}

// TestLoadconfig tests the Loadconfig method
func TestLoadconfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     interface{}
		content string
		wantErr bool
	}{
		{
			name:    "nil pointer",
			cfg:     nil,
			content: "",
			wantErr: true,
		},
		{
			name:    "non-pointer",
			cfg:     TestConfig{},
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := config.NewConfigParser()

			// Create and populate config file if content is provided
			if tt.content != "" {
				tempDir := t.TempDir()
				filename := filepath.Join(tempDir, "dicedb.conf")
				err := os.WriteFile(filename, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}

				err = parser.ParseFromFile(filename)
				if err != nil {
					t.Fatalf("Failed to parse test file: %v", err)
				}
			}

			err := parser.Loadconfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Loadconfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.cfg != nil {
				cfg := tt.cfg.(*TestConfig)
				if tt.content != "" && (cfg.AsyncServer.Addr != "customhost" || cfg.AsyncServer.Port != 9090 || cfg.Logging.LogLevel != "debug") {
					t.Error("Config values were not properly loaded")
				}
			}
		})
	}
}
