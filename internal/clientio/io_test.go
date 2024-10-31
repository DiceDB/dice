package clientio

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
)

// MockReadWriter to simulate different io behaviors
type MockReadWriter struct {
	ReadChunks [][]byte
	WriteData  []byte
	ReadIndex  int
	ReadErr    error
	WriteErr   error
}

func (m *MockReadWriter) Read(p []byte) (int, error) {
	if m.ReadIndex >= len(m.ReadChunks) {
		return 0, io.EOF
	}
	n := copy(p, m.ReadChunks[m.ReadIndex])
	m.ReadIndex++
	if m.ReadErr != nil {
		return n, m.ReadErr
	}
	return n, nil
}

func (m *MockReadWriter) Write(p []byte) (int, error) {
	m.WriteData = append(m.WriteData, p...)
	if m.WriteErr != nil {
		return len(p), m.WriteErr
	}
	return len(p), nil
}

// Helper function to compare arrays recursively
func equalArrays(a, b interface{}) bool {
	switch a := a.(type) {
	case []interface{}:
		b, ok := b.([]interface{})
		if !ok || len(a) != len(b) {
			return false
		}
		for i := range a {
			if !equalArrays(a[i], b[i]) {
				return false
			}
		}
		return true
	case string:
		return a == b
	case int64:
		return a == b
	default:
		return false
	}
}

// Test cases for RESPParser

func TestDecodeMultiple(t *testing.T) {
	mockRW := &MockReadWriter{ReadChunks: [][]byte{[]byte("+OK\r\n+PONG\r\n")}}
	parser := NewRESPParser(mockRW)
	results, err := parser.DecodeMultiple()
	if err != nil {
		t.Fatalf("DecodeMultiple returned error: %v", err)
	}
	expected := []interface{}{"OK", "PONG"}
	if !equalArrays(results, expected) {
		t.Fatalf("Expected %v, got %v", expected, results)
	}
}

func TestDecodeOneCrossProtocolScripting(t *testing.T) {
	mockRW := &MockReadWriter{ReadChunks: [][]byte{[]byte("GET / HTTP/1.1\r\n\r\n")}}
	parser := NewRESPParser(mockRW)
	_, err := parser.DecodeOne()
	if err == nil || err.Error() != "possible cross protocol scripting attack detected" {
		t.Fatalf("Expected cross protocol scripting attack error, got %v", err)
	}
}

func TestDecodeOneSplitBuffers(t *testing.T) {
	// Simulate a bulk string message split across two buffers
	mockRW := &MockReadWriter{
		ReadChunks: [][]byte{
			[]byte("$6\r\nfoo"),
			[]byte("bar\r\n"),
		},
	}
	parser := NewRESPParser(mockRW)

	// Perform the decode operation to read the complete message
	result, err := parser.DecodeOne()
	if err != nil {
		t.Fatalf("DecodeOne returned error: %v", err)
	}
	if result != "foobar" {
		t.Fatalf("Expected 'foobar', got %v", result)
	}
}

func TestDecodeOneEmptyMessage(t *testing.T) {
	mockRW := &MockReadWriter{}
	parser := NewRESPParser(mockRW)
	_, err := parser.DecodeOne()
	if err == nil || err != io.EOF {
		t.Fatalf("Expected io.EOF error, got %v", err)
	}
}

func TestDecodeOneHighVolumeData(t *testing.T) {
	largeString := bytes.Repeat([]byte("a"), 10*config.DiceConfig.Network.IOBufferLength)
	mockRW := &MockReadWriter{
		ReadChunks: [][]byte{
			[]byte("$" + strconv.Itoa(len(largeString)) + "\r\n"),
			largeString,
			[]byte("\r\n"),
		},
	}
	parser := NewRESPParser(mockRW)
	result, err := parser.DecodeOne()
	if err != nil {
		t.Fatalf("DecodeOne returned error: %v", err)
	}
	if result != string(largeString) {
		t.Fatalf("Expected large string, got %v", result)
	}
}

func TestDecodeOneNestedArrays(t *testing.T) {
	mockRW := &MockReadWriter{
		ReadChunks: [][]byte{
			[]byte("*2\r\n*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$3\r\nbaz\r\n"),
		},
	}
	parser := NewRESPParser(mockRW)
	result, err := parser.DecodeOne()
	if err != nil {
		t.Fatalf("DecodeOne returned error: %v", err)
	}
	expected := []interface{}{
		[]interface{}{"foo", "bar"},
		"baz",
	}
	if !equalArrays(result, expected) {
		t.Fatalf("Expected %v, got %v", expected, result)
	}
}

func TestDecodeOnePartialMessages(t *testing.T) {
	mockRW := &MockReadWriter{
		ReadChunks: [][]byte{
			[]byte("$6\r\nfoo"),
			[]byte("bar\r\n"),
		},
	}
	parser := NewRESPParser(mockRW)

	result, err := parser.DecodeOne()
	if err != nil {
		t.Fatalf("DecodeOne returned error: %v", err)
	}
	if result != "foobar" {
		t.Fatalf("Expected 'foobar', got %v", result)
	}
}

func TestDecodeOneVeryLargeMessage(t *testing.T) {
	largeString := bytes.Repeat([]byte("a"), 10*config.DiceConfig.Network.IOBufferLength)
	mockRW := &MockReadWriter{
		ReadChunks: [][]byte{
			[]byte("$" + strconv.Itoa(len(largeString)) + "\r\n"),
			largeString,
			[]byte("\r\n"),
		},
	}
	parser := NewRESPParser(mockRW)
	result, err := parser.DecodeOne()
	if err != nil {
		t.Fatalf("DecodeOne returned error: %v", err)
	}
	if result != string(largeString) {
		t.Fatalf("Expected large string, got %v", result)
	}
}

func TestDecodeOneNoDataRead(t *testing.T) {
	mockRW := &MockReadWriter{
		ReadChunks: [][]byte{
			[]byte(utils.EmptyStr), // Empty read chunk
		},
	}
	parser := NewRESPParser(mockRW)

	_, err := parser.DecodeOne()
	assert.Equal(t, err, net.ErrClosed)
}
