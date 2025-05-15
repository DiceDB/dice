// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/dicedb/dicedb-go/wire"
)

func extractValueHGETALL(res *wire.Result) interface{} {
	elements := res.GetHGETALLRes().Elements
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Key < elements[j].Key
	})

	str := ""
	for _, element := range elements {
		str += fmt.Sprintf("%s: %s\n", element.Key, element.Value)
	}
	return str
}

func TestHGETALL(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:           "Get Value for Field stored at Hash Key",
			commands:       []string{"HSET k f1 v1 f2 v2", "HGETALL k"},
			expected:       []interface{}{2, "f1: v1\nf2: v2\n"},
			valueExtractor: []ValueExtractorFn{extractValueHSET, extractValueHGETALL},
		},
		{
			name:     "HGETALL with no key argument",
			commands: []string{"HGETALL"},
			expected: []interface{}{
				errors.New("wrong number of arguments for 'HGETALL' command"),
			},
			valueExtractor: []ValueExtractorFn{nil},
		},
		{
			name:     "HGETALL with non hash key",
			commands: []string{"SET key 5", "HGETALL key"},
			expected: []interface{}{"OK",
				errors.New("wrongtype operation against a key holding the wrong kind of value"),
			},
			valueExtractor: []ValueExtractorFn{extractValueSET, nil},
		},
	}
	runTestcases(t, client, testCases)
}
