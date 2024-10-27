package wal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

func (w *WALSQLite) Iterate() error {
	return nil
}
