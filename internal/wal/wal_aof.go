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
	"google.golang.org/protobuf/proto"
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

// Log writes a command to the WAL
func (wal *AOF) Log(data []byte) error {
	wal.mu.Lock()
	wal.lastSequenceNo++
	lsn := wal.lastSequenceNo
	wal.mu.Unlock()

	// Create command payload with LSN and wire command bytes
	payload := &CommandPayload{
		Lsn:         lsn,
		WireCommand: data, // data is already the wire command bytes
	}

	// Marshal the payload to bytes using protobuf
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal command payload: %w", err)
	}

	return wal.writeEntry(payloadBytes)
}

func (wal *AOF) writeEntry(data []byte) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	// Create WAL entry with the new proto structure
	entry := &WALEntry{
		Crc32:     crc32.ChecksumIEEE(data),
		Size:      uint32(len(data)),
		Payload:   data,
		Timestamp: time.Now().UnixNano(),
		EntryType: EntryType_ENTRY_TYPE_COMMAND,
	}

	// Calculate total entry size including proto overhead
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
	// Marshal the WAL entry to bytes
	entryData, err := proto.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal WAL entry: %w", err)
	}

	// Write CRC32 (4 bytes)
	crcBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcBytes, entry.Crc32)
	if _, err := wal.bufWriter.Write(crcBytes); err != nil {
		return err
	}

	// Write size of WAL entry (4 bytes)
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(len(entryData)))
	if _, err := wal.bufWriter.Write(sizeBytes); err != nil {
		return err
	}

	// Write the actual WAL data
	_, err = wal.bufWriter.Write(entryData)
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

func (wal *AOF) Replay(callback func(any) error) error {
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
		// Format: CRC32 (4 bytes) | Size of WAL entry (4 bytes) | WAL data
		for {
			// Read CRC32 (4 bytes)
			crcBytes := make([]byte, 4)
			if _, err := io.ReadFull(reader, crcBytes); err != nil {
				if err == io.EOF {
					break
				}
				file.Close()
				return fmt.Errorf("error reading CRC32: %w", err)
			}
			crc := binary.LittleEndian.Uint32(crcBytes)

			// Read size of WAL entry (4 bytes)
			sizeBytes := make([]byte, 4)
			if _, err := io.ReadFull(reader, sizeBytes); err != nil {
				file.Close()
				return fmt.Errorf("error reading WAL entry size: %w", err)
			}
			entrySize := binary.LittleEndian.Uint32(sizeBytes)

			// Read the actual WAL data
			entryData := make([]byte, entrySize)
			if _, err := io.ReadFull(reader, entryData); err != nil {
				file.Close()
				return fmt.Errorf("error reading WAL data: %w", err)
			}

			// Unmarshal the WAL entry to get the payload
			var entry WALEntry
			if err := proto.Unmarshal(entryData, &entry); err != nil {
				file.Close()
				return fmt.Errorf("error unmarshaling WAL entry: %w", err)
			}

			// Calculate CRC32 only on the payload part
			expectedCRC := crc32.ChecksumIEEE(entry.Payload)
			if crc != expectedCRC {
				file.Close()
				return fmt.Errorf("CRC32 mismatch: expected %d, got %d", crc, expectedCRC)
			}

			// Unmarshal the payload into CommandPayload
			var cmdPayload CommandPayload
			if err := proto.Unmarshal(entry.Payload, &cmdPayload); err != nil {
				file.Close()
				return fmt.Errorf("error unmarshaling command payload: %w", err)
			}

			// Call provided replay function with parsed command
			if err := callback(&cmdPayload); err != nil {
				file.Close()
				return fmt.Errorf("error replaying command: %w", err)
			}
		}
		file.Close()
	}

	return nil
}

func (wal *AOF) Iterate(record any, callback func(any) error) error {
	// Calculate CRC32 on just the command data
	entry := record.(*WALEntry)
	expectedCRC := crc32.ChecksumIEEE(entry.Payload)
	if entry.Crc32 != expectedCRC {
		return fmt.Errorf("checksum mismatch: expected %d, got %d",
			expectedCRC, entry.Crc32)
	}

	// Unmarshal the command payload
	var payload CommandPayload
	if err := proto.Unmarshal(entry.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal command payload: %w", err)
	}

	// Update the entry's payload to contain just the wire command bytes
	entry.Payload = payload.WireCommand

	return callback(entry)
}

// wal entry should be
// crc32
// size of wal entry
// actual wal data
