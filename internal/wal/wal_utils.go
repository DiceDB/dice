package wal

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Marshals
func MustMarshal(entry *WAL_Entry) []byte {
	marshaledEntry, err := proto.Marshal(entry)
	if err != nil {
		panic(fmt.Sprintf("marshal should never fail (%v)", err))
	}

	return marshaledEntry
}

func MustUnmarshal(data []byte, entry *WAL_Entry) {
	if err := proto.Unmarshal(data, entry); err != nil {
		panic(fmt.Sprintf("unmarshal should never fail (%v)", err))
	}
}

func getEntrySize(data []byte) int {
	return versionTagSize + versionLengthPrefixSize + versionSize + // Version field
		logSequenceNumberSize + // Log Sequence Number field
		dataTagSize + dataLengthPrefixSize + len(data) + // Data field
		CRCSize + // CRC field
		timestampSize // Timestamp field
}

func (wal *WALAOF) validateConfig() error {
	if wal.logDir == "" {
		return fmt.Errorf("logDir cannot be empty")
	}

	if wal.maxSegmentSize <= 0 {
		return fmt.Errorf("maxSegmentSize must be greater than 0")
	}

	if wal.maxSegmentCount <= 0 {
		return fmt.Errorf("maxSegmentCount must be greater than 0")
	}

	if wal.bufferSize <= 0 {
		return fmt.Errorf("bufferSize must be greater than 0")
	}

	if wal.walMode == "buffered" && wal.writeMode == "fsync" {
		return fmt.Errorf("walMode 'buffered' cannot be used with writeMode 'fsync'")
	}

	if wal.walMode == "unbuffered" && wal.writeMode == "default" {
		return fmt.Errorf("walMode 'unbuffered' cannot have writeMode as 'default'")
	}

	return nil
}
