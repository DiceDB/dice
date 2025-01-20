// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	sync "sync"
	"time"

	"log/slog"

	"github.com/dicedb/dice/internal/cmd"
	_ "github.com/mattn/go-sqlite3"
)

type WALSQLite struct {
	logDir string
	curDB  *sql.DB
	mu     sync.Mutex
}

func NewSQLiteWAL(logDir string) (*WALSQLite, error) {
	return &WALSQLite{
		logDir: logDir,
	}, nil
}

func (w *WALSQLite) Init(t time.Time) error {
	slog.Debug("initializing WAL at", slog.Any("log-dir", w.logDir))
	if err := os.MkdirAll(w.logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := t.Format("20060102_1504")
	path := filepath.Join(w.logDir, fmt.Sprintf("wal_%s.sqlite3", timestamp))

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS wal (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL
	);`)
	if err != nil {
		return err
	}

	w.curDB = db
	return nil
}

func (w *WALSQLite) LogCommand(c *cmd.DiceDBCmd) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.curDB.Exec("INSERT INTO wal (command) VALUES (?)", c.Repr()); err != nil {
		slog.Error("failed to log command in WAL", slog.Any("error", err))
	} else {
		slog.Debug("logged command in WAL", slog.Any("command", c.Repr()))
	}
}

func (w *WALSQLite) Close() error {
	return w.curDB.Close()
}

func (w *WALSQLite) ForEachCommand(f func(c cmd.DiceDBCmd) error) error {
	files, err := os.ReadDir(w.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %v", err)
	}

	var walFiles []os.DirEntry

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sqlite3" {
			walFiles = append(walFiles, file)
		}
	}

	if len(walFiles) == 0 {
		return fmt.Errorf("no valid WAL files found in log directory")
	}

	// Sort files by timestamp in ascending order
	sort.Slice(walFiles, func(i, j int) bool {
		timestampStrI := walFiles[i].Name()[4:17]
		timestampStrJ := walFiles[j].Name()[4:17]
		timestampI, errI := time.Parse("20060102_1504", timestampStrI)
		timestampJ, errJ := time.Parse("20060102_1504", timestampStrJ)
		if errI != nil || errJ != nil {
			return false
		}
		return timestampI.Before(timestampJ)
	})

	for _, file := range walFiles {
		filePath := filepath.Join(w.logDir, file.Name())

		slog.Debug("loading WAL", slog.Any("file", filePath))

		db, err := sql.Open("sqlite3", filePath)
		if err != nil {
			return fmt.Errorf("failed to open WAL file %s: %v", file.Name(), err)
		}

		rows, err := db.Query("SELECT command FROM wal")
		if err != nil {
			return fmt.Errorf("failed to query WAL file %s: %v", file.Name(), err)
		}

		for rows.Next() {
			var command string
			if err := rows.Scan(&command); err != nil {
				return fmt.Errorf("failed to scan WAL file %s: %v", file.Name(), err)
			}

			tokens := strings.Split(command, " ")
			if err := f(cmd.DiceDBCmd{
				Cmd:  tokens[0],
				Args: tokens[1:],
			}); err != nil {
				return err
			}
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("failed to iterate WAL file %s: %v", file.Name(), err)
		}

		if err := db.Close(); err != nil {
			return fmt.Errorf("failed to close WAL file %s: %v", file.Name(), err)
		}
	}

	return nil
}
