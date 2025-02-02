// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoAdd(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "GEOADD with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "mygeo", "values": []interface{}{"1.2", "2.4"}}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'geoadd' command"},
		},
		{
			name: "GEOADD Commands with new member and updating it",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "mygeo", "values": []interface{}{"1.2", "2.4", "NJ"}}},
				{Command: "GEOADD", Body: map[string]interface{}{"key": "mygeo", "values": []interface{}{"1.24", "2.48", "NJ"}}},
			},
			expected: []interface{}{float64(1), float64(0)},
		},
		{
			name: "GEOADD Adding both XX and NX options together",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "mygeo", "values": []interface{}{"XX", "NX", "1.2", "2.4", "NJ"}}},
			},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name: "GEOADD Invalid Longitude",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "mygeo", "values": []interface{}{"181", "2.4", "MT"}}},
			},
			expected: []interface{}{"ERR invalid longitude"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoDist(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "GEODIST b/w existing points",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "points", "values": []interface{}{"13.361389", "38.115556", "Palermo"}}},
				{Command: "GEOADD", Body: map[string]interface{}{"key": "points", "values": []interface{}{"15.087269", "37.502669", "Catania"}}},
				{Command: "GEODIST", Body: map[string]interface{}{"key": "points", "values": []interface{}{"Palermo", "Catania"}}},
				{Command: "GEODIST", Body: map[string]interface{}{"key": "points", "values": []interface{}{"Palermo", "Catania", "km"}}},
			},
			expected: []interface{}{float64(1), float64(1), float64(166274.144), float64(166.2741)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoPos(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "GEOPOS for existing points",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "index", "values": []interface{}{"13.361389", "38.115556", "Palermo"}}},
				{Command: "GEOPOS", Body: map[string]interface{}{"key": "index", "values": []interface{}{"Palermo"}}},
			},
			expected: []interface{}{
				float64(1),
				[]interface{}{[]interface{}{float64(13.361387), float64(38.115556)}},
			},
		},
		{
			name: "GEOPOS for non-existing points",
			commands: []HTTPCommand{
				{Command: "GEOPOS", Body: map[string]interface{}{"key": "index", "values": []interface{}{"NonExisting"}}},
			},
			expected: []interface{}{[]interface{}{nil}},
		},
		{
			name: "GEOPOS for non-existing index",
			commands: []HTTPCommand{
				{Command: "GEOPOS", Body: map[string]interface{}{"key": "NonExisting", "values": []interface{}{"Palermo"}}},
			},
			expected: []interface{}{nil},
		},
		{
			name: "GEOPOS for a key not used for setting geospatial values",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "k", "value": "v"}},
				{Command: "GEOPOS", Body: map[string]interface{}{"key": "k", "values": []interface{}{"v"}}},
			},
			expected: []interface{}{
				"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %v", cmd)
			}
		})
	}
}

func TestGeoHash(t *testing.T) {
	exec := NewHTTPCommandExecutor()

	testCases := []struct {
		name     string
		commands []HTTPCommand
		expected []interface{}
	}{
		{
			name: "GEOHASH with wrong number of arguments",
			commands: []HTTPCommand{
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "geopoints"}},
			},
			expected: []interface{}{"ERR wrong number of arguments for 'geohash' command"},
		},
		{
			name: "GEOHASH with non-existent key",
			commands: []HTTPCommand{
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "nonexistent", "values": []interface{}{"Member1"}}},
			},
			expected: []interface{}{"ERR no such key"},
		},
		{
			name: "GEOHASH with existing key but missing member",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"13.361389", "38.115556", "Palermo"}}},
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"NonExistentMember"}}},
			},
			expected: []interface{}{float64(1), []interface{}{(nil)}},
		},
		{
			name: "GEOHASH with a key of wrong type",
			commands: []HTTPCommand{
				{Command: "SET", Body: map[string]interface{}{"key": "notgeo", "value": "value"}},
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "notgeo", "values": []interface{}{"Member"}}},
			},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "GEOHASH with a single member",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"-74.0060", "40.7128", "NewYork"}}},
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"NewYork"}}},
			},
			expected: []interface{}{float64(1), []interface{}{"dr5regw3pp"}},
		},
		{
			name: "GEOHASH with multiple members",
			commands: []HTTPCommand{
				{Command: "GEOADD", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"-73.935242", "40.730610", "Brooklyn"}}},
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"NewYork", "Brooklyn"}}},
			},
			expected: []interface{}{float64(1), []interface{}{"dr5regw3pp", "dr5rtwccpb"}},
		},
		{
			name: "GEOHASH with a non-existent member",
			commands: []HTTPCommand{
				{Command: "GEOHASH", Body: map[string]interface{}{"key": "geopoints", "values": []interface{}{"NonExistent"}}},
			},
			expected: []interface{}{[]interface{}{(nil)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result, _ := exec.FireCommand(cmd)
				assert.Equal(t, tc.expected[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
