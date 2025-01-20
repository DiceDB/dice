// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package wal

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

const (
	versionTagSize          = 1 // Tag for "version" field
	versionLengthPrefixSize = 1 // Length prefix for "version"
	versionSize             = 6 // Fixed size for "v0.0.1"
	logSequenceNumberSize   = 8
	dataTagSize             = 1 // Tag for "data" field
	dataLengthPrefixSize    = 1 // Length prefix for "data"
	CRCSize                 = 4
	timestampSize           = 8
)

// Marshals
func MustMarshal(entry *WALEntry) []byte {
	marshaledEntry, err := proto.Marshal(entry)
	if err != nil {
		panic(fmt.Sprintf("marshal should never fail (%v)", err))
	}

	return marshaledEntry
}

func MustUnmarshal(data []byte, entry *WALEntry) {
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
