package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dicedb/dice/config"
)

// scenario 1: Create a config file if the directory is provided (-o flag)
func TestSetupConfig_CreateAndLoadDefault(t *testing.T) {
	config.ResetConfig()
	tempDir := t.TempDir()

	// Simulate the flag: -o=<dir_path>
	config.CustomConfigFilePath = tempDir
	config.SetupConfig()

	if config.DiceConfig.AsyncServer.Addr != config.DefaultHost {
		t.Fatalf("Expected server addr to be '%s', got '%s'", config.DefaultHost, config.DiceConfig.AsyncServer.Addr)
	}
	if config.DiceConfig.AsyncServer.Port != config.DefaultPort {
		t.Fatalf("Expected server port to be %d, got %d", config.DefaultPort, config.DiceConfig.AsyncServer.Port)
	}
}

// scenario 2: Load default config if no config file or directory is provided
func TestSetupConfig_DefaultConfig(t *testing.T) {
	// Simulate no flags being set (default config scenario)
	config.ResetConfig()
	config.CustomConfigFilePath = ""
	config.FileLocation = filepath.Join(config.DefaultConfigFilePath, config.DefaultConfigName)

	// Verify that the configuration was loaded from the default values
	if config.DiceConfig.AsyncServer.Addr != config.DefaultHost {
		t.Fatalf("Expected server addr to be '%s', got '%s'", config.DefaultHost, config.DiceConfig.AsyncServer.Addr) // 127.0.0.1
	}
	if config.DiceConfig.AsyncServer.Port != config.DefaultPort {
		t.Fatalf("Expected server port to be %d, got %d", 8739, config.DiceConfig.AsyncServer.Port)
	}
}

// scenario 3: Config file is present but not well-structured (Malformed)
func TestSetupConfig_InvalidConfigFile(t *testing.T) {
	config.DiceConfig = nil
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "dice.toml")

	content := `
		[asyncserver]
		addr = 127.0.0.1  // Missing quotes around string value
		port = abc        // Invalid integer
	`
	if err := os.WriteFile(configFilePath, []byte(content), 0666); err != nil {
		t.Fatalf("Failed to create invalid test config file: %v", err)
	}

	// Simulate the flag: -c=<configfile_path>
	config.CustomConfigFilePath = ""
	config.FileLocation = configFilePath

	config.SetupConfig()

	if config.DiceConfig.AsyncServer.Addr != config.DefaultHost {
		t.Fatalf("Expected server addr to be '%s' after unmarshal error, got '%s'", config.DefaultHost, config.DiceConfig.AsyncServer.Addr)
	}
	if config.DiceConfig.AsyncServer.Port != config.DefaultPort {
		t.Fatalf("Expected server port to be %d after unmarshal error, got %d", config.DefaultPort, config.DiceConfig.AsyncServer.Port)
	}
}

// scenario 4: Config file is present with partial content
func TestSetupConfig_PartialConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "dice.toml")

	content := `
        [asyncserver]
        addr = "127.0.0.1"
    `
	if err := os.WriteFile(configFilePath, []byte(content), 0666); err != nil {
		t.Fatalf("Failed to create partial test config file: %v", err)
	}

	// Simulate the flag: -c=<configfile_path>
	config.CustomConfigFilePath = ""
	config.FileLocation = configFilePath

	config.SetupConfig()

	t.Log(config.DiceConfig.AsyncServer.Port)

	if config.DiceConfig.AsyncServer.Addr != "127.0.0.1" {
		t.Fatalf("Expected server addr to be '127.0.0.1', got '%s'", config.DiceConfig.AsyncServer.Addr)
	}
	if config.DiceConfig.AsyncServer.Port != config.DefaultPort {
		t.Fatalf("Expected server port to be %d (default), got %d", config.DefaultPort, config.DiceConfig.AsyncServer.Port)
	}
}

// scenario 5: Load config from the provided file path
func TestSetupConfig_LoadFromFile(t *testing.T) {
	config.ResetConfig()
	tempDir := t.TempDir()
	configFilePath := filepath.Join(tempDir, "dice.toml")

	content := `
		[asyncserver]
		addr = "127.0.0.1"
		port = 8739
	`
	if err := os.WriteFile(configFilePath, []byte(content), 0666); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Simulate the flag: -c=<configfile_path>
	config.CustomConfigFilePath = ""
	config.FileLocation = configFilePath

	config.SetupConfig()

	if config.DiceConfig.AsyncServer.Addr != "127.0.0.1" {
		t.Fatalf("Expected server addr to be '127.0.0.1', got '%s'", config.DiceConfig.AsyncServer.Addr)
	}
	if config.DiceConfig.AsyncServer.Port != 8739 {
		t.Fatalf("Expected server port to be 8374, got %d", config.DiceConfig.AsyncServer.Port)
	}

}
