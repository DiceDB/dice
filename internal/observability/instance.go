// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
