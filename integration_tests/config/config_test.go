// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dicedb/dice/config"
)

const configFileName = "dicedb.conf"

// TestCreateConfigFile_FileExists tests the scenario when config file already exists
func TestCreateConfigFile_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, configFileName)

	if err := os.WriteFile(configPath, []byte("test config"), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config.CreateConfigFile(configPath)

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != "test config" {
		t.Error("Config file content was modified when it should have been preserved")
	}
}

// TestCreateConfigFile_NewFile tests creating a new config file
func TestCreateConfigFile_NewFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, configFileName)
	config.CreateConfigFile(configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Created config file is empty")
	}
}

// TestCreateConfigFile_InvalidPath tests creation with an invalid file path
func TestCreateConfigFile_InvalidPath(t *testing.T) {
	configPath := "/nonexistent/directory/dicedb.conf"
	config.CreateConfigFile(configPath)

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should not have been created at invalid path")
	}
}

// TestCreateConfigFile_NoPermission tests creation without write permissions
func TestCreateConfigFile_NoPermission(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	tempDir := t.TempDir()
	err := os.Chmod(tempDir, 0555) // read + execute only
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}
	defer os.Chmod(tempDir, 0755) // restore permissions

	configPath := filepath.Join(tempDir, configFileName)
	config.CreateConfigFile(configPath)

	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should not have been created without permissions")
	}
}

// TestCreateConfigFile_ExistingDirectory tests creation in existing directory
func TestCreateConfigFile_ExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	config.CreateConfigFile(configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created in existing directory")
	}
}
