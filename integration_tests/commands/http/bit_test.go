package http

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitPos(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	exec.FireCommand(HTTPCommand{Command: "FLUSHDB"})
	defer exec.FireCommand(HTTPCommand{Command: "FLUSHDB"}) // clean up after all test cases
	testcases := []struct {
		name         string
		val          interface{}
		inCmd        HTTPCommand
		out          interface{}
		setCmdSETBIT bool
	}{
		{
			name:  "FindsFirstZeroBit",
			val:   []byte("\xff\xf0\x00"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(12),
		},
		{
			name:  "FindsFirstOneBit",
			val:   []byte("\x00\x0f\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 1}},
			out:   float64(12),
		},
		{
			name:  "NoZeroBitFound",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "value": 0}},
			out:   float64(24),
		},
		{
			name:  "NoZeroBitFoundWithRangeStartPos",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 2}}},
			out:   float64(24),
		},
		{
			name:  "NoZeroBitFoundWithOOBRangeStartPos",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 4}}},
			out:   float64(-1),
		},
		{
			name:  "NoZeroBitFoundWithRange",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 2, 2}}},
			out:   float64(-1),
		},
		{
			name:  "NoZeroBitFoundWithRangeAndRangeType",
			val:   []byte("\xff\xff\xff"),
			inCmd: HTTPCommand{Command: "BITPOS", Body: map[string]interface{}{"key": "testkey", "values": []interface{}{0, 2, 2, "BIT"}}},
			out:   float64(-1),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var setCmd HTTPCommand
			if tc.setCmdSETBIT {
				setCmd = HTTPCommand{
					Command: "SETBIT",
					Body:    map[string]interface{}{"key": "testkeysb", "value": fmt.Sprintf("%s", tc.val.(string))},
				}
			} else {
				switch v := tc.val.(type) {
				case []byte:
					setCmd = HTTPCommand{
						Command: "SET",
						Body:    map[string]interface{}{"key": "testkey", "value": v, "isByteEncodedVal": true},
					}
				case string:
					setCmd = HTTPCommand{
						Command: "SET",
						Body:    map[string]interface{}{"key": "testkey", "value": fmt.Sprintf("%s", v)},
					}
				case int:
					setCmd = HTTPCommand{
						Command: "SET",
						Body:    map[string]interface{}{"key": "testkey", "value": fmt.Sprintf("%d", v)},
					}
				default:
					// For test cases where we don't set a value (e.g., error cases)
					setCmd = HTTPCommand{Command: ""}
				}
			}

			if setCmd.Command != "" {
				_, _ = exec.FireCommand(setCmd)
			}

			result, _ := exec.FireCommand(tc.inCmd)
			assert.Equal(t, tc.out, result, "Mismatch for cmd %s\n", tc.inCmd)
		})
	}
}
