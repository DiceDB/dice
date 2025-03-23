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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMain(m *testing.M) {
	config.ForceInit(&config.DiceDBConfig{})
	os.Exit(m.Run())
}

type TestCase struct {
	name     string
	commands []string
	expected []interface{}
	delay    []time.Duration
}

func assertEqual(t *testing.T, expected interface{}, actual *wire.Response) {
	switch v := expected.(type) {
	case string:
		assert.Equal(t, v, actual.GetVStr())
	case int64:
		assert.Equal(t, v, actual.GetVInt())
	case float64:
		assert.Equal(t, v, actual.GetVFloat())
	case int:
		assert.Equal(t, int64(v), actual.GetVInt())
	case nil:
		assert.Nil(t, actual.GetVNil())
	case error:
		assert.Equal(t, v.Error(), actual.Err)
	case []*structpb.Value:
		if actual.VList != nil {
			assert.ElementsMatch(t, v, actual.GetVList())
		}
	case []interface{}:
		expected := expected.([]interface{})

		if !assert.Equal(t, len(expected), len(actual.GetVList())) {
			return
		}

		var actualArray []interface{}
		for i, v := range actual.GetVList() {
			//TODO: handle structpb.Value_StructValue & structpb.Value_ListValue
			switch expected[i].(type) {
			case string:
				actualArray = append(actualArray, v.GetStringValue())
			case float64:
				actualArray = append(actualArray, v.GetNumberValue())
			case int64:
				actualArray = append(actualArray, v.GetNumberValue())
			case int:
				actualArray = append(actualArray, int64(v.GetNumberValue()))
			case nil:
				actualArray = append(actualArray, nil)
			}
		}
		assert.ElementsMatch(t, expected, actualArray)
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
				assertEqual(t, tc.expected[i], result)
			}
		})
	}
}
