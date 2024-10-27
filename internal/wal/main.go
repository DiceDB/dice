package wal

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dicedb/dice/internal/cmd"
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
func (w *WAL) LogCommand(c *cmd.DiceDBCmd) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	repr := fmt.Sprintf("%s %s", c.Cmd, strings.Join(c.Args, " "))

	entry := &WALLogEntry{
		Command:  repr,
		Checksum: checksum(repr),
	}

	fmt.Println("entry", entry)

	// Serialize entry to protobuf binary format
	data, err := proto.Marshal(entry)
	if err != nil {
		slog.Warn("failed to serialize command", slog.Any("error", err.Error()))
	}

	// Write data to the file
	if _, err := w.file.Write(data); err != nil {
		slog.Warn("failed to write serialized command to WAL", slog.Any("error", err.Error()))
	}

	// Ensure data is flushed to disk for durability
	if err := w.file.Sync(); err != nil {
		slog.Warn("failed to sync WAL", slog.Any("error", err.Error()))
	}
}

// checksum generates a SHA-256 hash for the given command.
func checksum(command string) []byte {
	hash := sha256.Sum256([]byte(command))
	return hash[:]
}

// LoadWAL loads all WAL files from the log directory starting with the oldest file first.
func (w *WAL) LoadWAL() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	files, err := os.ReadDir(w.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %v", err)
	}

	var walFiles []os.DirEntry

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".log" {
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

		file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open WAL file %s: %v", file.Name(), err)
		}

		reader := bytes.NewBuffer(nil)
		for {
			buf := make([]byte, 4096)
			n, err := file.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to read WAL file %s: %v", file.Name(), err)
			}

			reader.Write(buf[:n])
			for {
				entry := &WALLogEntry{}
				if err := proto.Unmarshal(reader.Bytes(), entry); err != nil {
					if err == io.ErrUnexpectedEOF {
						break
					}
					return fmt.Errorf("failed to unmarshal WAL entry: %v", err)
				}

				if entry.Checksum == nil || entry.Command == "" {
					break
				}

				slog.Debug("loading log entry", slog.Any("command", entry.Command))
			}
		}

		file.Close()
	}

	return nil
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
