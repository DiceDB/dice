// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestBFReserveAddInfoExists(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name    string
		cmds    []string
		expect  []interface{}
		delays  []time.Duration
		cleanUp []string
	}{
		{
			name:    "BF.RESERVE and BF.ADD",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.ADD bf item1", "BF.EXISTS bf item1"},
			expect:  []interface{}{"OK", int64(1), int64(1)},
			delays:  []time.Duration{0, 0, 0},
			cleanUp: []string{"DEL bf"},
		},
		{
			name:    "BF.EXISTS returns false for non-existing item",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.EXISTS bf item2"},
			expect:  []interface{}{"OK", int64(0)},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"DEL bf"},
		},
		{
			name:    "BF.INFO provides correct information",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.ADD bf item1", "BF.INFO bf"},
			expect:  []interface{}{"OK", int64(1), []interface{}{"Capacity", int64(1000), "Size", int64(10104), "Number of filters", int64(7), "Number of items inserted", int64(1), "Expansion rate", int64(2)}},
			delays:  []time.Duration{0, 0, 0},
			cleanUp: []string{"DEL bf"},
		},
		{
			name:    "BF.RESERVE on existent filter returns error",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.RESERVE bf 0.01 1000"},
			expect:  []interface{}{"OK", diceerrors.ErrKeyExists.Error()},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"DEL bf"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}

			for _, cmd := range tc.cleanUp {
				client.FireString(cmd)
			}
		})
	}
}

func TestBFEdgeCasesAndErrors(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name    string
		cmds    []string
		expect  []interface{}
		delays  []time.Duration
		cleanUp []string
	}{
		{
			name:    "BF.RESERVE with incorrect number of arguments",
			cmds:    []string{"BF.RESERVE bf", "BF.RESERVE bf a"},
			expect:  []interface{}{"ERR wrong number of arguments for 'bf.reserve' command", "ERR wrong number of arguments for 'bf.reserve' command"},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with zero capacity",
			cmds:    []string{"BF.RESERVE bf 0.01 0"},
			expect:  []interface{}{"ERR (capacity should be larger than 0)"},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with zero capacity",
			cmds:    []string{"BF.RESERVE bf 0.01 -1"},
			expect:  []interface{}{"ERR (capacity should be larger than 0)"},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with zero capacity",
			cmds:    []string{"BF.RESERVE bf 0.01 a"},
			expect:  []interface{}{"ERR bad capacity"},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with invalid error rate",
			cmds:    []string{"BF.RESERVE bf -0.01 1000"},
			expect:  []interface{}{"ERR (0 < error rate range < 1) "},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with invalid error rate",
			cmds:    []string{"BF.RESERVE bf a 1000"},
			expect:  []interface{}{"ERR bad error rate"},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.ADD to a Bloom filter without reserving",
			cmds:    []string{"BF.ADD bf item1"},
			expect:  []interface{}{int64(1)},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.EXISTS on an unreserved filter",
			cmds:    []string{"BF.EXISTS bf item1"},
			expect:  []interface{}{int64(0)},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.INFO on a non-existent filter",
			cmds:    []string{"BF.INFO bf"},
			expect:  []interface{}{diceerrors.ErrKeyNotFound.Error()},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with a very high error rate",
			cmds:    []string{"BF.RESERVE bf 0.99 1000"},
			expect:  []interface{}{"OK"},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with a very low error rate",
			cmds:    []string{"BF.RESERVE bf 0.000001 1000"},
			expect:  []interface{}{"OK"},
			delays:  []time.Duration{0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.ADD multiple items and check existence",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.ADD bf item1", "BF.ADD bf item2", "BF.EXISTS bf item1", "BF.EXISTS bf item2", "BF.EXISTS bf item3"},
			expect:  []interface{}{"OK", int64(1), int64(1), int64(1), int64(1), int64(0)},
			delays:  []time.Duration{0, 0, 0, 0, 0, 0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.EXISTS after BF.ADD returns false on non-existing item",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.ADD bf item1", "BF.EXISTS bf nonExistentItem"},
			expect:  []interface{}{"OK", int64(1), int64(0)},
			delays:  []time.Duration{0, 0, 0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE with duplicate filter name",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.RESERVE bf 0.01 2000"},
			expect:  []interface{}{"OK", diceerrors.ErrKeyExists.Error()},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.INFO after multiple additions",
			cmds:    []string{"BF.RESERVE bf 0.01 1000", "BF.ADD bf item1", "BF.ADD bf item2", "BF.ADD bf item3", "BF.INFO bf"},
			expect:  []interface{}{"OK", int64(1), int64(1), int64(1), []interface{}{"Capacity", int64(1000), "Size", int64(10104), "Number of filters", int64(7), "Number of items inserted", int64(3), "Expansion rate", int64(2)}},
			delays:  []time.Duration{0, 0, 0, 0, 0},
			cleanUp: []string{"del bf"},
		},
		{
			name:    "BF.RESERVE on a key holding a string value",
			cmds:    []string{"SET foo \"string_value\"", "BF.RESERVE foo 0.001 1000"},
			expect:  []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"del foo"},
		},
		{
			name:    "BF.ADD on a key holding a list",
			cmds:    []string{"LPUSH foo \"item1\"", "BF.ADD foo item2"},
			expect:  []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"del foo"},
		},
		{
			name:    "BF.INFO on a key holding a hash",
			cmds:    []string{"HSET foo field1 value1", "BF.INFO foo"},
			expect:  []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
			delays:  []time.Duration{0, 0},
			cleanUp: []string{"del foo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}

			for _, cmd := range tc.cleanUp {
				client.FireString(cmd)
			}
		})
		client.FireString("FLUSHDB")
	}
}
