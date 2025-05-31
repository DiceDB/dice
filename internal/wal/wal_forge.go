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
)

var bb []byte

func init() {
	// Pre-allocate a buffer to avoid re-allocating it
	// This will hold one WAL Forge Entry Before it is written to the buffer
	bb = make([]byte, 10*1024)
}

type walForge struct {
	// Current Segment File and its writer
	csf      *os.File
	csWriter *bufio.Writer
	csIdx    int
	csSize   uint32

	// TODO: Come up with a way to generate a LSN that is
	// monotonically increasing and even after restart it
	// resumes from the last LSN and not start from 0.
	lsn uint64

	maxSegmentSizeBytes uint32

	bufferSyncTicker      *time.Ticker
	segmentRotationTicker *time.Ticker

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func newWalForge() *walForge {
	ctx, cancel := context.WithCancel(context.Background())
	return &walForge{
		ctx:    ctx,
		cancel: cancel,

		bufferSyncTicker:      time.NewTicker(time.Duration(config.Config.WALBufferSyncIntervalMillis) * time.Millisecond),
		segmentRotationTicker: time.NewTicker(time.Duration(config.Config.WALSegmentRotationTimeSec) * time.Second),

		maxSegmentSizeBytes: uint32(config.Config.WALMaxSegmentSizeMB) * 1024 * 1024,
	}
}

func (wl *walForge) Init() error {
	// TODO: Once the checkpoint is implemented
	// Load the initial state of the database from this checkpoint
	// and then reply the WAL files that happened after this checkpoint.

	// Make sure the WAL directory exists
	if err := os.MkdirAll(config.Config.WALDir, 0755); err != nil {
		return err
	}

	// Get the list of log segment files in the WAL directory
	sfs, err := wl.segments()
	if err != nil {
		return err
	}
	slog.Debug("Loading WAL segments", slog.Any("total_segments", len(sfs)))

	// TODO: Do not assume that the first segment is always 0
	// Pick the one with the least value of the segment index
	// Maintain a metadatafile that holds the latest segment index used
	// and during rotation, it increments the segment index and uses it
	sf, err := os.OpenFile(
		filepath.Join(config.Config.WALDir, segmentPrefix+"0"+".wal"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	wl.csf = sf
	wl.csWriter = bufio.NewWriterSize(wl.csf, config.Config.WALBufferSizeMB*1024*1024)

	go wl.periodicSyncBuffer()

	switch config.Config.WALRotationMode {
	case "time":
		go wl.periodicRotateSegment()
	default:
		return nil
	}
	return nil
}

// LogCommand writes a command to the WAL with a monotonically increasing LSN.
func (wl *walForge) LogCommand(c *wire.Command) error {
	// Lock once for the entire LSN operation
	wl.mu.Lock()
	defer wl.mu.Unlock()

	// marshal the command to bytes
	b, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	// TODO: This logic changes as we change the LSN format
	wl.lsn += 1
	el := &w.Element{
		Lsn:         wl.lsn,
		Timestamp:   time.Now().UnixNano(),
		ElementType: w.ElementType_ELEMENT_TYPE_COMMAND,
		Payload:     b,
	}

	// marshal the WAL Element to bytes
	b, err = proto.Marshal(el)
	if err != nil {
		return err
	}

	// Wrap the element with Checksum and Size
	// and keep it ready to be written to the segment file through the buffer
	// We call this WAL Entry.
	entrySize := uint32(4 + 4 + len(b))
	if err := wl.rotateLogIfNeeded(entrySize); err != nil {
		return err
	}

	// If the entry size is greater than the buffer size, we need to
	// create a new buffer.
	if entrySize > uint32(cap(bb)) {
		// TODO: In this case, we can do a one time creation of a new buffer
		// and proceed rather than using the existing buffer.
		panic(fmt.Errorf("buffer too small, %d > %d", entrySize, len(bb)))
	}

	bb = bb[:8+len(b)]
	chk := crc32.ChecksumIEEE(b)

	// Write header and payload
	binary.LittleEndian.PutUint32(bb[0:4], chk)
	binary.LittleEndian.PutUint32(bb[4:8], uint32(len(b)))
	copy(bb[8:], b)

	// TODO: Check if we need to handle the error here,
	// from my initial understanding, we should not be
	// handling the error here because it would never happen.
	// Have not tested this yet.
	_, _ = wl.csWriter.Write(bb)

	wl.csSize += entrySize
	return nil
}

// rotateLogIfNeeded checks if the current segment size + the entry size is
// greater than the max segment size, and if so, it rotates the log.
// This method is not thread safe and hence should be called with the lock held.
func (wl *walForge) rotateLogIfNeeded(entrySize uint32) error {
	// If the current segment size + the entry size is greater than the max segment size,
	// we need to rotate the log.
	if wl.csSize+entrySize > wl.maxSegmentSizeBytes {
		if err := wl.rotateLog(); err != nil {
			return err
		}
	}
	return nil
}

// rotateLog rotates the log by closing the current segment file,
// incrementing the current segment index, and opening a new segment file.
func (wl *walForge) rotateLog() error {
	fmt.Println("rotating log")
	// TODO: Ideally this function should not return any error
	// Check for the conditions where it can return an error
	// and handle them gracefully.
	// I fear that we will need to do some cleanup operations in case of errors.

	// Sync the current segment file to disk
	if err := wl.sync(); err != nil {
		return err
	}

	// Close the current segment file
	if err := wl.csf.Close(); err != nil {
		return err
	}

	// Increment the current segment index
	wl.csIdx++

	// Open a new segment file
	sf, err := os.OpenFile(
		filepath.Join(config.Config.WALDir, fmt.Sprintf("%s%d.wal", segmentPrefix, wl.csIdx)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// TODO: We are panicking here because we are not handling the error
		// and we want to make sure that the WAL is not corrupted.
		// We need to handle this error gracefully.
		panic(fmt.Errorf("failed opening file: %w", err))
	}

	// Reset the trackers
	wl.csf = sf
	wl.csSize = 0
	wl.csWriter = bufio.NewWriter(sf)

	return nil
}

// Writes out any data in the WAL's in-memory buffer to the segment file.
// and syncs the segment file to disk.
func (wl *walForge) sync() error {
	// Flush the buffer to the segment file
	if err := wl.csWriter.Flush(); err != nil {
		return err
	}

	// Sync the segment file to disk to make sure
	// it is written to disk.
	if err := wl.csf.Sync(); err != nil {
		return err
	}

	// TODO: Evaluate if DIRECT_IO is needed here.
	// If we are using a file system that supports direct IO,
	// we can use it to improve the performance.
	// If we are using a file system that does not support direct IO,
	// we can use the buffer to improve the performance.
	return nil
}

func (wl *walForge) periodicSyncBuffer() {
	for {
		select {
		case <-wl.bufferSyncTicker.C:
			wl.mu.Lock()
			err := wl.sync()
			if err != nil {
				slog.Error("failed to sync buffer", slog.String("error", err.Error()))
			}
			wl.mu.Unlock()
		case <-wl.ctx.Done():
			return
		}
	}
}

func (wl *walForge) periodicRotateSegment() {
	fmt.Println("rotating segment")
	for {
		select {
		case <-wl.segmentRotationTicker.C:
			// TODO: Remove this error handling once we clean up the error handling in the rotateLog function.
			wl.mu.Lock()
			if err := wl.rotateLog(); err != nil {
				slog.Error("failed to rotate segment", slog.String("error", err.Error()))
			}
			wl.mu.Unlock()
		case <-wl.ctx.Done():
			return
		}
	}
}

func (wl *walForge) segments() ([]string, error) {
	// Get all segment files matching the pattern
	files, err := filepath.Glob(filepath.Join(config.Config.WALDir, segmentPrefix+"*"+".wal"))
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		s1, _ := strconv.Atoi(strings.Split(strings.TrimPrefix(files[i], segmentPrefix), ".")[0])
		s2, _ := strconv.Atoi(strings.Split(strings.TrimPrefix(files[i], segmentPrefix), ".")[0])
		return s1 < s2
	})

	// TODO: Check that the segment files are returned in the correct order
	// The order has to be in ascending order of the segment index.
	return files, nil
}

// ReplayCommand replays the commands from the WAL files.
// This method is thread safe.
func (wl *walForge) ReplayCommand(cb func(*wire.Command) error) error {
	var crc, entrySize uint32
	var el w.Element

	// Buffers to hold the header and the element bytes
	bb1h := make([]byte, 8)
	bb1ElementBytes := make([]byte, 10*1024)

	// Get list of segment files ordered by timestamp in ascending order
	segments, err := wl.segments()
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

		// TODO: Replace this infinite loop with a more elegant solution
		for {
			// Read CRC32 (4 bytes) + entrySize (4 bytes)
			if _, err := io.ReadFull(reader, bb1h); err != nil {
				// TODO: this terminating connection should be handled in a better way
				// and the loop should not be infinite.
				// Edge case: this EOF error can happen even in the next step when
				// we are reading the WAL element from the file.
				if err == io.EOF {
					break
				}
				file.Close()
				return fmt.Errorf("error reading WAL: %w", err)
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
				// TODO: We are reprtitively closing the file here
				// A better solution would be to move this logic to a function
				// and use defer to close the file.
				// The function. thus, in a way processes (replays) one segment at a time.
				file.Close()

				// TODO: THis is where we should trigger the WAL recovery
				// Recovery process is all about truncating the segment file
				// till this point and ignoring the rest.
				// Log appropriate messages when this happens.
				// Evaluate if this recovery mode should be a command line flag
				// that would suggest if we should truncate, ignore, or stop the boot process.
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
	}

	return nil
}

// Stop stops the WAL Forge.
// This method is thread safe.
func (wl *walForge) Stop() {
	wl.mu.Lock()
	defer wl.mu.Unlock()

	// Stop the tickers
	wl.bufferSyncTicker.Stop()
	wl.segmentRotationTicker.Stop()

	// Cancel the context
	wl.cancel()

	// Sync the current segment file to disk
	if err := wl.sync(); err != nil {
		slog.Error("failed to sync current segment file", slog.String("error", err.Error()))
	}

	wl.csf.Close()

	// TODO: See if we are missing any other cleanup operations.
}
