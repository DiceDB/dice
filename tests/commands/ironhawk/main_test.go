// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dicedb-go"
	"github.com/dicedb/dicedb-go/wire"
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

func assertEqual(t *testing.T, expected interface{}, actual *wire.Response) bool {
	var areEqual bool
	switch v := expected.(type) {
	case string:
		areEqual = v == actual.GetVStr()
	case int64:
		areEqual = v == actual.GetVInt()
	case int:
		areEqual = int64(v) == actual.GetVInt()
	case nil:
		areEqual = actual.GetVNil()
	case error:
		areEqual = v.Error() == actual.Err
	case []*structpb.Value:
		if actual.VList != nil {
			areEqual = reflect.DeepEqual(v, actual.GetVList())
		}
	}
	if !areEqual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
	return areEqual
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
