//go:build darwin

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var DicedbDataDir = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error fetching home directory, defaulting to /tmp/dicedb for dir")
		return "/tmp/dicedb"
	}

	return filepath.Join(home, "Library/Application Support/dicedb")
}()
