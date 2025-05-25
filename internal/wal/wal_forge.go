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
	"sync"
	"time"

	"github.com/dicedb/dice/config"
	"google.golang.org/protobuf/proto"

	w "github.com/dicedb/dicedb-go/wal"
	"github.com/dicedb/dicedb-go/wire"
)

const (
	segmentPrefix = "seg-"
	segmentSuffix = ".wal"
)

var bb []byte

func init() {
	// TODO: Pre-allocate a buffer to avoid re-allocating it
	// This will hold one WAL Forge Entry Before it is written to the buffer
	bb = make([]byte, 10*1024)
}

var (
	rotTicker *time.Ticker
)

type walForge struct {
	logDir                 string
	currentSegmentFile     *os.File
	writeMode              string
	maxSegmentSize         uint32
	maxSegmentCount        int
	currentSegmentIndex    int
	currentSegmentSize     uint32
	oldestSegmentIndex     int
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

func newWalForge() *walForge {
	ctx, cancel := context.WithCancel(context.Background())
	return &walForge{
		logDir:                 config.Config.WALDir,
		bufferSyncTicker:       time.NewTicker(time.Duration(config.Config.WALBufferSyncIntervalMillis) * time.Millisecond),
		segmentRotationTicker:  time.NewTicker(time.Duration(config.Config.WALMaxSegmentRotationTimeSec) * time.Second),
		segmentRetentionTicker: time.NewTicker(time.Duration(config.Config.WALMaxSegmentRetentionDurationSec) * time.Second),
		writeMode:              config.Config.WALWriteMode,
		maxSegmentSize:         uint32(config.Config.WALMaxSegmentSizeMB) * 1024 * 1024,
		maxSegmentCount:        config.Config.WALMaxSegmentCount,
		bufferSize:             config.Config.WALBufferSizeMB * 1024 * 1024,
		retentionMode:          config.Config.WALRetentionMode,
		recoveryMode:           config.Config.WALRecoveryMode,
		rotationMode:           config.Config.WALRotationMode,
		ctx:                    ctx,
		cancel:                 cancel,
	}
}

func (wl *walForge) Init() error {
	rotTicker = time.NewTicker(time.Duration(config.Config.WALRotationTimeSec) * time.Second)

	// TODO - Restore existing checkpoints to memory

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(wl.logDir, 0755); err != nil {
		return err
	}

	// Get the list of log segment files in the directory
	files, err := filepath.Glob(filepath.Join(wl.logDir, segmentPrefix+"*"+segmentSuffix))
	if err != nil {
		return err
	}

	if len(files) > 0 {
		slog.Debug("Found existing log segments", slog.Any("total_files", len(files)))
		// TODO - Check if we have newer WAL entries after the last checkpoint and simultaneously replay and checkpoint them
	}

	sf, err := os.OpenFile(
		filepath.Join(wl.logDir, segmentPrefix+"0"+segmentSuffix),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	wl.currentSegmentFile = sf
	wl.bufWriter = bufio.NewWriterSize(wl.currentSegmentFile, wl.bufferSize)

	go wl.keepSyncingBuffer()
	go wl.startAsyncJobs()

	switch wl.rotationMode {
	case "time":
		go wl.rotateSegmentPeriodically()
		go wl.deleteSegmentPeriodically()
	default:
		return nil
	}
	return nil
}

// Log writes a command to the WAL with a monotonically increasing sequence number.
// The sequence number is assigned atomically and the command is written to the wl.
func (wl *walForge) LogCommand(c *wire.Command) error {
	// Lock once for the entire sequence number operation
	wl.mu.Lock()
	defer wl.mu.Unlock()

	b, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	wl.lastSequenceNo += 1
	el := &w.Element{
		Lsn:         wl.lastSequenceNo,
		Timestamp:   time.Now().UnixNano(),
		ElementType: w.ElementType_ELEMENT_TYPE_COMMAND,
		Payload:     b,
	}

	b, err = proto.Marshal(el)
	if err != nil {
		return err
	}

	entrySize := uint32(4 + 4 + len(b))
	if err := wl.rotateLogIfNeeded(entrySize); err != nil {
		return err
	}

	// If the entry size is greater than the buffer size, we need to
	// create a new buffer.
	if entrySize > uint32(cap(bb)) {
		// TODO: In this case, we can do a one time creation
		// of a new buffer and proceed rather than using the
		// existing buffer.
		panic(fmt.Errorf("buffer too small, %d > %d", entrySize, len(bb)))
	}

	bb = bb[:8+len(b)]
	// Calculate CRC32 only on the payload
	chk := crc32.ChecksumIEEE(b)

	// Write header and payload
	binary.LittleEndian.PutUint32(bb[0:4], chk)
	binary.LittleEndian.PutUint32(bb[4:8], uint32(len(b)))
	copy(bb[8:], b)

	_, _ = wl.bufWriter.Write(bb)

	wl.currentSegmentSize += entrySize

	return nil
}

// rotateLogIfNeeded is not thread safe
func (wl *walForge) rotateLogIfNeeded(entrySize uint32) error {
	if wl.currentSegmentSize+entrySize > wl.maxSegmentSize {
		if err := wl.rotateLog(); err != nil {
			return err
		}
	}
	return nil
}

// rotateLog is not thread safe
func (wl *walForge) rotateLog() error {
	if err := wl.Sync(); err != nil {
		return err
	}

	if err := wl.currentSegmentFile.Close(); err != nil {
		return err
	}

	wl.currentSegmentIndex++
	if wl.currentSegmentIndex-wl.oldestSegmentIndex+1 > wl.maxSegmentCount {
		if err := wl.deleteOldestSegment(); err != nil {
			return err
		}
		wl.oldestSegmentIndex++
	}

	sf, err := os.OpenFile(filepath.Join(wl.logDir, segmentPrefix+fmt.Sprintf("%d", wl.currentSegmentIndex)+segmentSuffix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	wl.currentSegmentSize = 0
	wl.currentSegmentFile = sf
	wl.bufWriter = bufio.NewWriter(sf)

	return nil
}

func (wl *walForge) deleteOldestSegment() error {
	oldestSegmentFilePath := filepath.Join(wl.logDir, segmentPrefix+fmt.Sprintf("%d", wl.oldestSegmentIndex)+segmentSuffix)

	// TODO: checkpoint before deleting the file
	if err := os.Remove(oldestSegmentFilePath); err != nil {
		return err
	}
	wl.oldestSegmentIndex++
	return nil
}

// Close the WAL file. It also calls Sync() on the wl.
func (wl *walForge) Close() error {
	wl.cancel()
	if err := wl.Sync(); err != nil {
		return err
	}
	return wl.currentSegmentFile.Close()
}

// Writes out any data in the WAL's in-memory buffer to the segment file. If
// fsync is enabled, it also calls fsync on the segment file.
func (wl *walForge) Sync() error {
	if err := wl.bufWriter.Flush(); err != nil {
		return err
	}
	if wl.writeMode == "fsync" {
		if err := wl.currentSegmentFile.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (wl *walForge) keepSyncingBuffer() {
	for {
		select {
		case <-wl.bufferSyncTicker.C:
			wl.mu.Lock()
			err := wl.Sync()
			wl.mu.Unlock()

			if err != nil {
				slog.Error("failed to sync buffer", slog.String("error", err.Error()))
			}

		case <-wl.ctx.Done():
			return
		}
	}
}

func (wl *walForge) rotateSegmentPeriodically() {
	for {
		select {
		case <-wl.segmentRotationTicker.C:
			wl.mu.Lock()
			err := wl.rotateLog()
			wl.mu.Unlock()
			if err != nil {
				slog.Error("failed to rotate segment", slog.String("error", err.Error()))
			}

		case <-wl.ctx.Done():
			return
		}
	}
}

func (wl *walForge) deleteSegmentPeriodically() {
	for {
		select {
		case <-wl.segmentRetentionTicker.C:
			wl.mu.Lock()
			err := wl.deleteOldestSegment()
			wl.mu.Unlock()
			if err != nil {
				slog.Error("failed to delete segment", slog.String("error", err.Error()))
			}
		case <-wl.ctx.Done():
			return
		}
	}
}

func (wl *walForge) segmentFiles() ([]string, error) {
	// Get all segment files matching the pattern
	files, err := filepath.Glob(filepath.Join(wl.logDir, segmentPrefix+"*"+segmentSuffix))
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

func (wl *walForge) ReplayCommand(cb func(*wire.Command) error) error {
	var crc uint32
	var entrySize uint32
	var el w.Element
	bb1h := make([]byte, 8)
	bb1ElementBytes := make([]byte, 10*1024)

	// Get list of segment files sorted by timestamp
	segments, err := wl.segmentFiles()
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
			// Read CRC32 (4 bytes) + entrySize (4 bytes)
			if _, err := io.ReadFull(reader, bb1h); err != nil {
				if err == io.EOF {
					break
				}
				file.Close()
				return fmt.Errorf("error reading CRC32: %w", err)
			}
			crc = binary.LittleEndian.Uint32(bb1h[0:4])
			entrySize = binary.LittleEndian.Uint32(bb1h[4:8])

			if _, err := io.ReadFull(reader, bb1ElementBytes[:entrySize]); err != nil {
				file.Close()
				return fmt.Errorf("error reading WAL data: %w", err)
			}

			// Calculate CRC32 only on the payload
			expectedCRC := crc32.ChecksumIEEE(bb1ElementBytes[:entrySize])
			if crc != expectedCRC {
				file.Close()
				return fmt.Errorf("CRC32 mismatch: expected %d, got %d", crc, expectedCRC)
			}

			// Unmarshal the WAL entry to get the payload
			if err := proto.Unmarshal(bb1ElementBytes[:entrySize], &el); err != nil {
				file.Close()
				return fmt.Errorf("error unmarshaling WAL entry: %w", err)
			}

			var c wire.Command
			if err := proto.Unmarshal(el.Payload, &c); err != nil {
				file.Close()
				return fmt.Errorf("error unmarshaling command: %w", err)
			}

			// Call provided replay function with parsed command
			if err := cb(&c); err != nil {
				file.Close()
				return fmt.Errorf("error replaying command: %w", err)
			}
		}
		file.Close()
	}

	return nil
}

func (wl *walForge) Stop() {
	rotTicker.Stop()
}

func (wl *walForge) rotateWAL() {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	if err := wl.Close(); err != nil {
		slog.Warn("error closing the WAL", slog.Any("error", err))
	}

	if err := wl.Init(); err != nil {
		slog.Warn("error creating a new WAL", slog.Any("error", err))
	}
}

func (wl *walForge) periodicRotate() {
	for {
		select {
		case <-rotTicker.C:
			wl.rotateWAL()
		case <-stopCh:
			return
		}
	}
}

func (wl *walForge) startAsyncJobs() {
	go wl.periodicRotate()
}
