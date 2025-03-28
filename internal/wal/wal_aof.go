// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	sync "sync"
	"time"

	"github.com/dicedb/dice/config"
)

const (
	segmentPrefix     = "seg-"
	segmentSuffix     = ".wal"
	RotationModeTime  = "time"
	RetentionModeTime = "time"
	WALModeUnbuffered = "unbuffered"
)

type AOF struct {
	logDir                 string
	currentSegmentFile     *os.File
	walMode                string
	writeMode              string
	maxSegmentSize         int
	maxSegmentCount        int
	currentSegmentIndex    int
	oldestSegmentIndex     int
	byteOffset             int
	bufferSize             int
	retentionMode          string
	recoveryMode           string
	rotationMode           string
	lastSequenceNo         uint64
	bufWriter              *bufio.Writer
	bufferSyncTicker       *time.Ticker
	segmentRotationTicker  *time.Ticker
	segmentRetentionTicker *time.Ticker
	mu                     sync.Mutex
	ctx                    context.Context
	cancel                 context.CancelFunc
}

func NewAOFWAL(directory string) (*AOF, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &AOF{
		logDir:                 directory,
		walMode:                config.Config.WALMode,
		bufferSyncTicker:       time.NewTicker(time.Duration(config.Config.WALBufferSyncIntervalMillis) * time.Millisecond),
		segmentRotationTicker:  time.NewTicker(time.Duration(config.Config.WALMaxSegmentRotationTimeSec) * time.Second),
		segmentRetentionTicker: time.NewTicker(time.Duration(config.Config.WALMaxSegmentRetentionDurationSec) * time.Second),
		writeMode:              config.Config.WALWriteMode,
		maxSegmentSize:         config.Config.WALMaxSegmentSizeMB * 1024 * 1024,
		maxSegmentCount:        config.Config.WALMaxSegmentCount,
		bufferSize:             config.Config.WALBufferSizeMB * 1024 * 1024,
		retentionMode:          config.Config.WALRetentionMode,
		recoveryMode:           config.Config.WALRecoveryMode,
		rotationMode:           config.Config.WALRotationMode,
		ctx:                    ctx,
		cancel:                 cancel,
	}, nil
}

func (wal *AOF) Init(t time.Time) error {
	// TODO - Restore existing checkpoints to memory

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(wal.logDir, 0755); err != nil {
		return err
	}

	// Get the list of log segment files in the directory
	files, err := filepath.Glob(filepath.Join(wal.logDir, segmentPrefix+"*"+segmentSuffix))
	if err != nil {
		return err
	}

	if len(files) > 0 {
		slog.Info("Found existing log segments", slog.Any("files", files))
		// TODO - Check if we have newer WAL entries after the last checkpoint and simultaneously replay and checkpoint them
	}

	wal.lastSequenceNo = 0
	wal.currentSegmentIndex = 0
	wal.oldestSegmentIndex = 0
	wal.byteOffset = 0
	newFile, err := os.OpenFile(filepath.Join(wal.logDir, segmentPrefix+"0"+segmentSuffix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	wal.currentSegmentFile = newFile

	if _, err := wal.currentSegmentFile.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	wal.bufWriter = bufio.NewWriterSize(wal.currentSegmentFile, wal.bufferSize)

	go wal.keepSyncingBuffer()

	if wal.rotationMode == RotationModeTime {
		go wal.rotateSegmentPeriodically()
	}

	if wal.retentionMode == RetentionModeTime {
		go wal.deleteSegmentPeriodically()
	}

	return nil
}

// WriteEntry writes an entry to the WAL.
func (wal *AOF) LogCommand(data []byte) error {
	return wal.writeEntry(data)
}

func (wal *AOF) writeEntry(data []byte) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	wal.lastSequenceNo++
	entry := &WALEntry{
		LogSequenceNumber: wal.lastSequenceNo,
		Crc32:             crc32.ChecksumIEEE(append(data, byte(wal.lastSequenceNo))),
		Timestamp:         time.Now().UnixNano(),
		EntryType:         EntryType_ENTRY_TYPE_COMMAND,
		EntryData:         data,
	}

	entrySize := getEntrySize(data)
	if err := wal.rotateLogIfNeeded(entrySize); err != nil {
		return err
	}

	wal.byteOffset += entrySize

	if err := wal.writeEntryToBuffer(entry); err != nil {
		return err
	}

	// if wal-mode unbuffered immediately sync to disk
	if wal.walMode == WALModeUnbuffered {
		if err := wal.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (wal *AOF) writeEntryToBuffer(entry *WALEntry) error {
	marshaledEntry := MustMarshal(entry)

	size := int32(len(marshaledEntry))
	if err := binary.Write(wal.bufWriter, binary.LittleEndian, size); err != nil {
		return err
	}
	_, err := wal.bufWriter.Write(marshaledEntry)

	return err
}

// rotateLogIfNeeded is not thread safe
func (wal *AOF) rotateLogIfNeeded(entrySize int) error {
	if wal.byteOffset+entrySize > wal.maxSegmentSize {
		if err := wal.rotateLog(); err != nil {
			return err
		}
	}
	return nil
}

// rotateLog is not thread safe
func (wal *AOF) rotateLog() error {
	if err := wal.Sync(); err != nil {
		return err
	}

	if err := wal.currentSegmentFile.Close(); err != nil {
		return err
	}

	wal.currentSegmentIndex++

	if wal.currentSegmentIndex-wal.oldestSegmentIndex+1 > wal.maxSegmentCount {
		if err := wal.deleteOldestSegment(); err != nil {
			return err
		}
		wal.oldestSegmentIndex++
	}

	newFile, err := os.OpenFile(filepath.Join(wal.logDir, segmentPrefix+fmt.Sprintf("%d", wal.currentSegmentIndex)+segmentSuffix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	wal.byteOffset = 0

	wal.currentSegmentFile = newFile
	wal.bufWriter = bufio.NewWriter(newFile)

	return nil
}

func (wal *AOF) deleteOldestSegment() error {
	oldestSegmentFilePath := filepath.Join(wal.logDir, segmentPrefix+fmt.Sprintf("%d", wal.oldestSegmentIndex)+segmentSuffix)

	// TODO: checkpoint before deleting the file

	if err := os.Remove(oldestSegmentFilePath); err != nil {
		return err
	}
	wal.oldestSegmentIndex++

	return nil
}

// Close the WAL file. It also calls Sync() on the WAL.
func (wal *AOF) Close() error {
	wal.cancel()
	if err := wal.Sync(); err != nil {
		return err
	}
	return wal.currentSegmentFile.Close()
}

// Writes out any data in the WAL's in-memory buffer to the segment file. If
// fsync is enabled, it also calls fsync on the segment file.
func (wal *AOF) Sync() error {
	if err := wal.bufWriter.Flush(); err != nil {
		return err
	}
	if wal.writeMode == "fsync" {
		if err := wal.currentSegmentFile.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (wal *AOF) keepSyncingBuffer() {
	for {
		select {
		case <-wal.bufferSyncTicker.C:
			wal.mu.Lock()
			err := wal.Sync()
			wal.mu.Unlock()

			if err != nil {
				slog.Error("failed to sync buffer", slog.String("error", err.Error()))
			}

		case <-wal.ctx.Done():
			return
		}
	}
}

func (wal *AOF) rotateSegmentPeriodically() {
	for {
		select {
		case <-wal.segmentRotationTicker.C:
			wal.mu.Lock()
			err := wal.rotateLog()
			wal.mu.Unlock()
			if err != nil {
				slog.Error("failed to rotate segment", slog.String("error", err.Error()))
			}

		case <-wal.ctx.Done():
			return
		}
	}
}

func (wal *AOF) deleteSegmentPeriodically() {
	for {
		select {
		case <-wal.segmentRetentionTicker.C:
			wal.mu.Lock()
			err := wal.deleteOldestSegment()
			wal.mu.Unlock()
			if err != nil {
				slog.Error("failed to delete segment", slog.String("error", err.Error()))
			}
		case <-wal.ctx.Done():
			return
		}
	}
}

func (wal *AOF) segmentFiles() ([]string, error) {
	// Get all segment files matching the pattern
	files, err := filepath.Glob(filepath.Join(wal.logDir, segmentPrefix+"*"+segmentSuffix))
	if err != nil {
		return nil, err
	}

	// Sort files by numeric suffix
	sort.Slice(files, func(i, j int) bool {
		parseSuffix := func(name string) int64 {
			num, _ := strconv.ParseInt(
				strings.TrimPrefix(strings.TrimSuffix(filepath.Base(name), segmentSuffix), segmentPrefix), 10, 64)
			return num
		}
		return parseSuffix(files[i]) < parseSuffix(files[j])
	})

	return files, nil
}

func (wal *AOF) Replay(callback func(*WALEntry) error) error {
	// Get list of segment files sorted by timestamp
	segments, err := wal.segmentFiles()
	if err != nil {
		return fmt.Errorf("error getting wal-segment files: %w", err)
	}

	// Process each segment file in order
	for _, segment := range segments {
		file, err := os.Open(segment)
		if err != nil {
			return fmt.Errorf("error opening wal-segment file %s: %w", segment, err)
		}

		reader := bufio.NewReader(file)
		for {
			// Read entry size
			var entrySize int32
			if err := binary.Read(reader, binary.LittleEndian, &entrySize); err != nil {
				if err == io.EOF {
					break
				}
				file.Close()
				return fmt.Errorf("error reading wal entry size: %w", err)
			}

			// Read entry data
			entryData := make([]byte, entrySize)
			if _, err := io.ReadFull(reader, entryData); err != nil {
				file.Close()
				return fmt.Errorf("error reading wal entry data: %w", err)
			}

			// Unmarshal entry
			var entry WALEntry
			MustUnmarshal(entryData, &entry)

			// Call provided replay function with parsed command
			if err := wal.ForEachCommand(&entry, callback); err != nil {
				file.Close()
				return fmt.Errorf("error replaying command: %w", err)
			}
		}
		file.Close()
	}

	return nil
}

func (wal *AOF) ForEachCommand(entry *WALEntry, callback func(*WALEntry) error) error {
	// Get the command data from the entry

	// Calculate CRC32 on just the command data and sequence number
	expectedCRC := crc32.ChecksumIEEE(append(entry.EntryData, byte(entry.LogSequenceNumber)))
	if entry.Crc32 != expectedCRC {
		return fmt.Errorf("checksum mismatch for log sequence %d: expected %d, got %d",
			entry.LogSequenceNumber, expectedCRC, entry.Crc32)
	}

	return callback(entry)
}
