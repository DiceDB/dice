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

package netconn

import (
	"bufio"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNetConnIOHandler_RESP(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedRead     string
		writeResponse    string
		expectedWrite    string
		readErr          error
		writeErr         error
		ctxTimeout       time.Duration
		expectedReadErr  error
		expectedWriteErr error
	}{
		{
			name:          "Simple String",
			input:         "+OK\r\n",
			expectedRead:  "+OK\r\n",
			writeResponse: "+OK\r\n",
			expectedWrite: "+OK\r\n",
		},
		{
			name:          "Error",
			input:         "-Error message\r\n",
			expectedRead:  "-Error message\r\n",
			writeResponse: "-ERR unknown command 'FOOBAR'\r\n",
			expectedWrite: "-ERR unknown command 'FOOBAR'\r\n",
		},
		{
			name:          "Integer",
			input:         ":1000\r\n",
			expectedRead:  ":1000\r\n",
			writeResponse: ":1000\r\n",
			expectedWrite: ":1000\r\n",
		},
		{
			name:          "Bulk String",
			input:         "$5\r\nhello\r\n",
			expectedRead:  "$5\r\nhello\r\n",
			writeResponse: "$5\r\nworld\r\n",
			expectedWrite: "$5\r\nworld\r\n",
		},
		{
			name:          "Null Bulk String",
			input:         "$-1\r\n",
			expectedRead:  "$-1\r\n",
			writeResponse: "$-1\r\n",
			expectedWrite: "$-1\r\n",
		},
		{
			name:          "Empty Bulk String",
			input:         "$0\r\n\r\n",
			expectedRead:  "$0\r\n\r\n",
			writeResponse: "$0\r\n\r\n",
			expectedWrite: "$0\r\n\r\n",
		},
		{
			name:          "Array",
			input:         "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expectedRead:  "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			writeResponse: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expectedWrite: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
		},
		{
			name:          "Empty Array",
			input:         "*0\r\n",
			expectedRead:  "*0\r\n",
			writeResponse: "*0\r\n",
			expectedWrite: "*0\r\n",
		},
		{
			name:          "Null Array",
			input:         "*-1\r\n",
			expectedRead:  "*-1\r\n",
			writeResponse: "*-1\r\n",
			expectedWrite: "*-1\r\n",
		},
		{
			name:          "Nested Array",
			input:         "*2\r\n*2\r\n+foo\r\n+bar\r\n*2\r\n+hello\r\n+world\r\n",
			expectedRead:  "*2\r\n*2\r\n+foo\r\n+bar\r\n*2\r\n+hello\r\n+world\r\n",
			writeResponse: "*2\r\n*2\r\n+foo\r\n+bar\r\n*2\r\n+hello\r\n+world\r\n",
			expectedWrite: "*2\r\n*2\r\n+foo\r\n+bar\r\n*2\r\n+hello\r\n+world\r\n",
		},
		{
			name:          "SET command",
			input:         "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			expectedRead:  "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			writeResponse: "+OK\r\n",
			expectedWrite: "+OK\r\n",
		},
		{
			name:          "GET command",
			input:         "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
			expectedRead:  "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
			writeResponse: "$5\r\nvalue\r\n",
			expectedWrite: "$5\r\nvalue\r\n",
		},
		{
			name:          "LPUSH command",
			input:         "*4\r\n$5\r\nLPUSH\r\n$4\r\nlist\r\n$5\r\nvalue\r\n$6\r\nvalue2\r\n",
			expectedRead:  "*4\r\n$5\r\nLPUSH\r\n$4\r\nlist\r\n$5\r\nvalue\r\n$6\r\nvalue2\r\n",
			writeResponse: ":2\r\n",
			expectedWrite: ":2\r\n",
		},
		{
			name:          "HMSET command",
			input:         "*6\r\n$5\r\nHMSET\r\n$4\r\nhash\r\n$5\r\nfield\r\n$5\r\nvalue\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n",
			expectedRead:  "*6\r\n$5\r\nHMSET\r\n$4\r\nhash\r\n$5\r\nfield\r\n$5\r\nvalue\r\n$6\r\nfield2\r\n$6\r\nvalue2\r\n",
			writeResponse: "+OK\r\n",
			expectedWrite: "+OK\r\n",
		},
		{
			name:          "Partial read",
			input:         "*2\r\n$5\r\nhello\r\n$5\r\nwor",
			expectedRead:  "*2\r\n$5\r\nhello\r\n$5\r\nwor",
			writeResponse: "+OK\r\n",
			expectedWrite: "+OK\r\n",
		},
		{
			name:            "Read error",
			input:           "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			readErr:         errors.New("read error"),
			expectedReadErr: errors.New("error reading request: read error"),
		},
		{
			name:             "Write error",
			input:            "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expectedRead:     "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			writeResponse:    strings.Repeat("Hello, World!\r\n", 100),
			writeErr:         errors.New("write error"),
			expectedWriteErr: errors.New("error writing response: write error"),
		},
		{
			name:             "Write error",
			input:            "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expectedRead:     "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			writeResponse:    "Hello, World!\r\n",
			writeErr:         errors.New("write error"),
			expectedWriteErr: errors.New("error writing response: write error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConn{
				readData: []byte(tt.input),
				readErr:  tt.readErr,
				writeErr: tt.writeErr,
			}

			handler := &IOHandler{
				conn:   mock,
				reader: bufio.NewReaderSize(mock, 512),
				writer: bufio.NewWriterSize(mock, 1024),
				readPool: &sync.Pool{
					New: func() interface{} {
						b := make([]byte, ioBufferSize)
						return &b // Return pointer
					},
				},
			}

			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			// Test ReadRequest
			data, err := handler.Read(ctx)
			if tt.expectedReadErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedReadErr.Error(), err.Error())
				return
			} else {
				assert.NoError(t, err)
				assert.Equal(t, []byte(tt.expectedRead), data)
			}

			// Test WriteResponse
			err = handler.Write(ctx, []byte(tt.writeResponse))
			if tt.expectedWriteErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedWriteErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, []byte(tt.expectedWrite), mock.writeData.Bytes())
			}
		})
	}
}
