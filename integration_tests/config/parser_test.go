package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dicedb/dice/config"
)

// TestConfig is a test struct that mimics your actual config structure
type TestConfig struct {
	Host     string `default:"localhost"`
	Port     int    `default:"8080"`
	LogLevel string `default:"info"`
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
			wantErr: true,
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
			wantErr: true,
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
				if cfg.Host != "localhost" || cfg.Port != 8080 || cfg.LogLevel != "info" {
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
				if tt.content != "" && (cfg.Host != "customhost" || cfg.Port != 9090 || cfg.LogLevel != "debug") {
					t.Error("Config values were not properly loaded")
				}
			}
		})
	}
}
