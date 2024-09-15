package eval

import (
	"fmt"
	"testing"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"gotest.tools/v3/assert"
)

func TestNewEval(t *testing.T) {
	// store := dstore.NewStore(nil)
	testScatterPING(t)
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
	input  []EvalScatterResponse
	output []byte
}

func testScatterPING(t *testing.T) {
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

			output := ScatterPING(tc.input)
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
		"single shard success": {input: []EvalScatterResponse{
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: []byte("+PONG\r\n")},
		"multi shard success": {input: []EvalScatterResponse{
			{
				Result: "PONG",
				Error:  nil,
			},
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: []byte("+PONG\r\n")},
		"multi shard success with argument": {input: []EvalScatterResponse{
			{
				Result: "ARG",
				Error:  nil,
			},
			{
				Result: "PONG",
				Error:  nil,
			},
		}, output: []byte("+ARG\r\n")},
		"single shard failure": {input: []EvalScatterResponse{
			{
				Result: nil,
				Error:  fmt.Errorf("PING"),
			},
		}, output: diceerrors.NewErrArity("PING")},
		"multi shard failure": {input: []EvalScatterResponse{
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
