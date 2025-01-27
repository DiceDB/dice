// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package observability

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// GetOrCreateInstanceID creates a file named dicedb.iid in a temp directory with a unique UUID v6.
// If the file exists, it reads the value and returns it. Otherwise, it creates the file and writes a new UUID v6 to it.
func GetOrCreateInstanceID() string {
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, "dicedb.iid")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		id := uuid.New().String()
		if err := os.WriteFile(filePath, []byte(id), 0600); err != nil {
			slog.Error("unable to create dicedb.iid hence running anon", slog.Any("error", err))
			return ""
		}
		return id
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.Error("unable to read dicedb.iid hence running anon", slog.Any("error", err))
		return ""
	}

	return string(data)
}
