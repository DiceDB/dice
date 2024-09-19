package eval

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	dstore "github.com/dicedb/dice/internal/store"
	"gotest.tools/v3/assert"
)

func TestNewEval(t *testing.T) {
	store := dstore.NewStore(nil)
	testEvalPING(t, store)
	testGatherPING(t)

}

type evalScatterTestCase struct {
	setup           func()
	input           []string
	output          interface{}
	isErrorScenario bool
}

type evalGatherTestCase struct {
	setup  func()
	input  []EvalResponse
	output []byte
}

func testEvalPING(t *testing.T, store *dstore.Store) {
	tests := map[string]evalScatterTestCase{
		"nil value":            {input: nil, isErrorScenario: false, output: "PONG"},
		"empty args":           {input: []string{}, isErrorScenario: false, output: "PONG"},
		"one value":            {input: []string{"HEY"}, isErrorScenario: false, output: "HEY"},
		"more than one values": {input: []string{"HEY", "HELLO"}, isErrorScenario: true, output: fmt.Errorf("PING").Error()},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			output := EvalPING(tc.input, store)
			if tc.isErrorScenario {
				assert.Equal(t, output.Error.Error(), tc.output)
			} else {
				assert.Equal(t, output.Result, tc.output)
			}
		})
	}
}

func testEvalSET(t *testing.T, store *dstore.Store) {
	tests := map[string]evalScatterTestCase{
		"nil value":                       {input: nil, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"empty array":                     {input: []string{}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"one value":                       {input: []string{"KEY"}, output: []byte("-ERR wrong number of arguments for 'set' command\r\n")},
		"key val pair":                    {input: []string{"KEY", "VAL"}, output: clientio.RespOK},
		"key val pair with int val":       {input: []string{"KEY", "123456"}, output: clientio.RespOK},
		"key val pair and expiry key":     {input: []string{"KEY", "VAL", Px}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and EX no val":      {input: []string{"KEY", "VAL", Ex}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and valid EX":       {input: []string{"KEY", "VAL", Ex, "2"}, output: clientio.RespOK},
		"key val pair and invalid EX":     {input: []string{"KEY", "VAL", Ex, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and valid PX":       {input: []string{"KEY", "VAL", Px, "2000"}, output: clientio.RespOK},
		"key val pair and invalid PX":     {input: []string{"KEY", "VAL", Px, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and both EX and PX": {input: []string{"KEY", "VAL", Ex, "2", Px, "2000"}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and PXAT no val":    {input: []string{"KEY", "VAL", Pxat}, output: []byte("-ERR syntax error\r\n")},
		"key val pair and invalid PXAT":   {input: []string{"KEY", "VAL", Pxat, "invalid_expiry_val"}, output: []byte("-ERR value is not an integer or out of range\r\n")},
		"key val pair and expired PXAT":   {input: []string{"KEY", "VAL", Pxat, "2"}, output: clientio.RespOK},
		"key val pair and negative PXAT":  {input: []string{"KEY", "VAL", Pxat, "-123456"}, output: []byte("-ERR invalid expire time in 'set' command\r\n")},
		"key val pair and valid PXAT":     {input: []string{"KEY", "VAL", Pxat, strconv.FormatInt(time.Now().Add(2*time.Minute).UnixMilli(), 10)}, output: clientio.RespOK},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			output := EvalSET(tc.input, store)
			if tc.isErrorScenario {
				assert.Equal(t, output.Error.Error(), tc.output)
			} else {
				assert.Equal(t, output.Result, tc.output)
			}
		})
	}
}

func testGatherPING(t *testing.T) {
	tests := map[string]evalGatherTestCase{
		"single shard success": {input: []EvalResponse{
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: []byte("+PONG\r\n")},
		"multi shard success": {input: []EvalResponse{
			{
				Result: "PONG",
				Error:  nil,
			},
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: []byte("+PONG\r\n")},
		"multi shard success with argument": {input: []EvalResponse{
			{
				Result: "ARG",
				Error:  nil,
			},
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: []byte("+ARG\r\n")},
		"single shard failure": {input: []EvalResponse{
			{
				Result: nil,
				Error:  fmt.Errorf("PING"),
			},
		}, output: diceerrors.NewErrArity("PING")},
		"multi shard failure": {input: []EvalResponse{
			{
				Result: nil,
				Error:  fmt.Errorf("PING"),
			},
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: diceerrors.NewErrArity("PING")},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}

			output := GatherPING(tc.input...)
			assert.Equal(t, string(output), string(tc.output))
		})
	}
}
