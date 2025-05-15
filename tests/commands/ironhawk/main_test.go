// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
	assert "github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	config.ForceInit(&config.DiceDBConfig{})
	os.Exit(m.Run())
}

type ValueExtractorFn func(result *wire.Result) interface{}

type TestCase struct {
	name           string
	commands       []string
	expected       []interface{}
	delay          []time.Duration
	valueExtractor []ValueExtractorFn
}

func assertEqualResult(t *testing.T, expected interface{}, result *wire.Result, valueExtractor ValueExtractorFn) {
	var actual interface{}
	if valueExtractor != nil {
		actual = valueExtractor(result)
	}
	switch v := expected.(type) {
	case string:
		assert.Equal(t, v, actual)
	case int64:
		assert.Equal(t, v, actual)
	case float64:
		assert.Equal(t, v, actual)
	case int:
		assert.Equal(t, int64(v), actual)
	case bool:
		assert.Equal(t, v, actual)
	case nil:
		assert.Equal(t, v, actual)
	case error:
		assert.Equal(t, v.Error(), result.Message)
	case []string:
		assert.ElementsMatch(t, v, actual)
	default:
		assert.Equal(t, v, actual)
	}
}

func runTestcases(t *testing.T, client *dicedb.Client, testCases []TestCase) {
	client.Fire(&wire.Command{
		Cmd: "FLUSHDB",
	})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				if len(tc.delay) > i && tc.delay[i] > 0 {
					time.Sleep(tc.delay[i])
				}
				result := client.Fire(&wire.Command{
					Cmd:  strings.Split(cmd, " ")[0],
					Args: strings.Split(cmd, " ")[1:],
				})

				assertEqualResult(t, tc.expected[i], result, tc.valueExtractor[i])
			}
		})
	}
}
