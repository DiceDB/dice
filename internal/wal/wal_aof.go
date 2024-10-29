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

type WALAOF struct {
	file   *os.File
	mutex  sync.Mutex
	logDir string
}

func NewAOFWAL(logDir string) (*WALAOF, error) {
	return &WALAOF{
		logDir: logDir,
	}, nil
}

func (w *WALAOF) Init(t time.Time) error {
	slog.Debug("initializing WAL at", slog.Any("log-dir", w.logDir))
	if err := os.MkdirAll(w.logDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := t.Format("20060102_1504")
	path := filepath.Join(w.logDir, fmt.Sprintf("wal_%s.aof", timestamp))

	newFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new WAL file: %v", err)
	}

	w.file = newFile
	return nil
}

// LogCommand serializes a WALLogEntry and writes it to the current WAL file.
func (w *WALAOF) LogCommand(c *cmd.DiceDBCmd) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	repr := fmt.Sprintf("%s %s", c.Cmd, strings.Join(c.Args, " "))

	entry := &WALLogEntry{
		Command:  repr,
		Checksum: checksum(repr),
	}

	data, err := proto.Marshal(entry)
	if err != nil {
		slog.Warn("failed to serialize command", slog.Any("error", err.Error()))
	}

	if _, err := w.file.Write(data); err != nil {
		slog.Warn("failed to write serialized command to WAL", slog.Any("error", err.Error()))
	}

	if err := w.file.Sync(); err != nil {
		slog.Warn("failed to sync WAL", slog.Any("error", err.Error()))
	}
}

func (w *WALAOF) Close() error {
	if w.file == nil {
		return nil
	}
	return w.file.Close()
}

// checksum generates a SHA-256 hash for the given command.
func checksum(command string) []byte {
	hash := sha256.Sum256([]byte(command))
	return hash[:]
}

func (w *WALAOF) ForEachCommand(f func(c cmd.DiceDBCmd) error) error {
	files, err := os.ReadDir(w.logDir)
	if err != nil {
		return fmt.Errorf("failed to read log directory: %v", err)
	}

	var walFiles []os.DirEntry

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".aof" {
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
