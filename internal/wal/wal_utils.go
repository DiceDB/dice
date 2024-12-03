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

func (w *WALAOF) validateConfig() error {
	if w.logDir == "" {
		return fmt.Errorf("logDir cannot be empty")
	}

	if w.maxSegmentSize <= 0 {
		return fmt.Errorf("maxSegmentSize must be greater than 0")
	}

	if w.maxSegmentCount <= 0 {
		return fmt.Errorf("maxSegmentCount must be greater than 0")
	}

	if w.bufferSize <= 0 {
		return fmt.Errorf("bufferSize must be greater than 0")
	}

	if w.walMode == "buffered" && w.writeMode == "fsync" {
		return fmt.Errorf("walMode 'buffered' cannot be used with writeMode 'fsync'")
	}

	if w.walMode == "unbuffered" && w.writeMode == "default" {
		return fmt.Errorf("walMode 'unbuffered' cannot have writeMode as 'default'")
	}

	return nil
}
