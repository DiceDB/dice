package wal

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

type WAL struct {
	file   *os.File
	mutex  sync.Mutex
	logDir string
	ticker *time.Ticker
	stopCh chan struct{}
}

func NewWAL(logDir string) (*WAL, error) {
	wal := &WAL{
		logDir: logDir,
		ticker: time.NewTicker(1 * time.Minute),
		stopCh: make(chan struct{}),
	}

	// Ensure the log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	if err := wal.rotateLogFile(); err != nil {
		return nil, fmt.Errorf("failed to create initial log file: %v", err)
	}

	go wal.rotateLogPeriodically()
	return wal, nil
}

// rotateLogFile closes the current WAL file and opens a new one with a timestamped name.
func (w *WAL) rotateLogFile() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		w.file.Close()
	}

	// Create new file with minute-level timestamp suffix
	timestamp := time.Now().Format("20060102_1504")
	filePath := filepath.Join(w.logDir, fmt.Sprintf("wal_%s.log", timestamp))
	newFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new WAL file: %v", err)
	}

	w.file = newFile
	return nil
}

func (w *WAL) rotateLogPeriodically() {
	for {
		select {
		case <-w.ticker.C:
			if err := w.rotateLogFile(); err != nil {
				fmt.Printf("Error rotating log file: %v\n", err)
			}
		case <-w.stopCh:
			return
		}
	}
}

// LogCommand serializes a WALLogEntry and writes it to the current WAL file.
func (w *WAL) LogCommand(command string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	entry := &WALLogEntry{
		Command:  command,
		Checksum: checksum(command),
	}

	// Serialize entry to protobuf binary format
	data, err := proto.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to serialize command: %v", err)
	}

	// Write data to the file
	if _, err := w.file.Write(data); err != nil {
		return fmt.Errorf("failed to write serialized command to WAL: %v", err)
	}

	// Ensure data is flushed to disk for durability
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync WAL: %v", err)
	}

	return nil
}

// checksum generates a SHA-256 hash for the given command.
func checksum(command string) string {
	hash := sha256.Sum256([]byte(command))
	return hex.EncodeToString(hash[:])
}

// LoadWAL reads all valid WAL entries from a specified WAL file for recovery.
func LoadWAL(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL file for reading: %v", err)
	}
	defer file.Close()

	var commands []string
	reader := bufio.NewReader(file)

	for {
		data, err := io.ReadAll(reader)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading WAL file: %v", err)
		}

		var entry WALLogEntry
		if err := proto.Unmarshal(data, &entry); err != nil {
			fmt.Printf("Skipping corrupted entry: %v\n", err)
			continue
		}

		if entry.Checksum != checksum(entry.Command) {
			fmt.Printf("Skipping entry with checksum mismatch: %s\n", entry.Command)
			continue
		}

		commands = append(commands, entry.Command)
	}

	return commands, nil
}

// CloseWAL stops log rotation and closes the current WAL file.
func (w *WAL) CloseWAL() error {
	close(w.stopCh) // Stop rotation goroutine
	w.ticker.Stop() // Stop the ticker

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
