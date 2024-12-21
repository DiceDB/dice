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
	sync "sync"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/cmd"
)

const (
	segmentPrefix     = "seg-"
	defaultVersion    = "v0.0.1"
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
		walMode:                config.DiceConfig.WAL.Mode,
		bufferSyncTicker:       time.NewTicker(config.DiceConfig.WAL.BufferSyncInterval),
		segmentRotationTicker:  time.NewTicker(config.DiceConfig.WAL.MaxSegmentRotationTime),
		segmentRetentionTicker: time.NewTicker(config.DiceConfig.WAL.MaxSegmentRetentionDuration),
		writeMode:              config.DiceConfig.WAL.WriteMode,
		maxSegmentSize:         config.DiceConfig.WAL.MaxSegmentSizeMB * 1024 * 1024,
		maxSegmentCount:        config.DiceConfig.WAL.MaxSegmentCount,
		bufferSize:             config.DiceConfig.WAL.BufferSizeMB * 1024 * 1024,
		retentionMode:          config.DiceConfig.WAL.RetentionMode,
		recoveryMode:           config.DiceConfig.WAL.RecoveryMode,
		rotationMode:           config.DiceConfig.WAL.RotationMode,
		ctx:                    ctx,
		cancel:                 cancel,
	}, nil
}

func (wal *AOF) Init(t time.Time) error {
	// TODO - Restore existing checkpoints to memory

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(wal.logDir, 0755); err != nil {
		return nil
	}

	// Get the list of log segment files in the directory
	files, err := filepath.Glob(filepath.Join(wal.logDir, segmentPrefix+"*"))
	if err != nil {
		return nil
	}

	if len(files) > 0 {
		slog.Info("Found existing log segments", slog.Any("files", files))
		// TODO - Check if we have newer WAL entries after the last checkpoint and simultaneously replay and checkpoint them
	}

	wal.lastSequenceNo = 0
	wal.currentSegmentIndex = 0
	wal.oldestSegmentIndex = 0
	wal.byteOffset = 0
	newFile, err := os.OpenFile(filepath.Join(wal.logDir, "seg-0"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
		Version:           defaultVersion,
		LogSequenceNumber: wal.lastSequenceNo,
		Data:              data,
		Crc32:             crc32.ChecksumIEEE(append(data, byte(wal.lastSequenceNo))),
		Timestamp:         time.Now().UnixNano(),
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
	if wal.walMode == WALModeUnbuffered { //nolint:goconst
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

	newFile, err := os.OpenFile(filepath.Join(wal.logDir, segmentPrefix+fmt.Sprintf("-%d", wal.currentSegmentIndex)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	wal.byteOffset = 0

	wal.currentSegmentFile = newFile
	wal.bufWriter = bufio.NewWriter(newFile)

	return nil
}

func (wal *AOF) deleteOldestSegment() error {
	oldestSegmentFilePath := filepath.Join(wal.logDir, segmentPrefix+fmt.Sprintf("%d", wal.oldestSegmentIndex))

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
	if wal.writeMode == "fsync" { //nolint:goconst
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
				log.Printf("Error while performing sync: %v", err)
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
				log.Printf("Error while performing sync: %v", err)
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
				log.Printf("Error while deleting segment: %v", err)
			}
		case <-wal.ctx.Done():
			return
		}
	}
}

func (wal *AOF) ForEachCommand(f func(c cmd.DiceDBCmd) error) error {
	// TODO: implement this method
	return nil
}
