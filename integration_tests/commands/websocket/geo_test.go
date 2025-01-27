// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package websocket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeoAdd(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name:   "GeoAdd With Wrong Number of Arguments",
			cmds:   []string{"GEOADD mygeo 1 2"},
			expect: []interface{}{"ERR wrong number of arguments for 'geoadd' command"},
		},
		{
			name:   "GeoAdd With Adding New Member And Updating it",
			cmds:   []string{"GEOADD mygeo 1.21 1.44 NJ", "GEOADD mygeo 1.22 1.54 NJ"},
			expect: []interface{}{float64(1), float64(0)},
		},
		{
			name:   "GeoAdd With Adding New Member And Updating it with NX",
			cmds:   []string{"GEOADD mygeo NX 1.21 1.44 MD", "GEOADD mygeo 1.22 1.54 MD"},
			expect: []interface{}{float64(1), float64(0)},
		},
		{
			name:   "GEOADD with both NX and XX options",
			cmds:   []string{"GEOADD mygeo NX XX 1.21 1.44 DEL"},
			expect: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:   "GEOADD invalid longitude",
			cmds:   []string{"GEOADD mygeo 181.0 1.44 MD"},
			expect: []interface{}{"ERR invalid longitude"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoDist(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name: "GEODIST b/w existing points",
			cmds: []string{
				"GEOADD points 13.361389 38.115556 Palermo",
				"GEOADD points 15.087269 37.502669 Catania",
				"GEODIST points Palermo Catania",
				"GEODIST points Palermo Catania km",
			},
			expect: []interface{}{float64(1), float64(1), float64(166274.144), float64(166.2741)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoPos(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name: "GEOPOS b/w existing points",
			cmds: []string{
				"GEOADD index 13.361389 38.115556 Palermo",
				"GEOPOS index Palermo",
			},
			expect: []interface{}{
				float64(1),
				[]interface{}{[]interface{}{float64(13.361387), float64(38.115556)}},
			},
		},
		{
			name: "GEOPOS for non existing points",
			cmds: []string{
				"GEOPOS index NonExisting",
			},
			expect: []interface{}{[]interface{}{nil}},
		},
		{
			name: "GEOPOS for non existing index",
			cmds: []string{
				"GEOPOS NonExisting Palermo",
			},
			expect: []interface{}{nil},
		},
		{
			name: "GEOPOS for a key not used for setting geospatial values",
			cmds: []string{
				"SET k v",
				"GEOPOS k v",
			},
			expect: []interface{}{
				"OK",
				"WRONGTYPE Operation against a key holding the wrong kind of value",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoHash(t *testing.T) {
	exec := NewWebsocketCommandExecutor()
	conn := exec.ConnectToServer()
	defer conn.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
	}{
		{
			name:   "GEOHASH with wrong number of arguments",
			cmds:   []string{"GEOHASH points"},
			expect: []interface{}{"ERR wrong number of arguments for 'geohash' command"},
		},
		{
			name: "GEOHASH with non-existent key",
			cmds: []string{
				"GEOHASH nonexistent member1",
			},
			expect: []interface{}{"ERR no such key"},
		},
		{
			name: "GEOHASH with existing key but missing member",
			cmds: []string{
				"GEOADD points -74.0060 40.7128 NewYork",
				"GEOHASH points missingMember",
			},
			expect: []interface{}{float64(1), []interface{}{(nil)}},
		},
		{
			name: "GEOHASH for single member",
			cmds: []string{
				"GEOHASH points NewYork",
			},
			expect: []interface{}{[]interface{}{"dr5regw3pp"}},
		},
		{
			name: "GEOHASH for multiple members",
			cmds: []string{
				"GEOADD points -118.2437 34.0522 LosAngeles",
				"GEOHASH points NewYork LosAngeles",
			},
			expect: []interface{}{float64(1), []interface{}{"dr5regw3pp", "9q5ctr186n"}},
		},
		{
			name: "GEOHASH with a key of wrong type",
			cmds: []string{
				"SET points somevalue",
				"GEOHASH points member1",
			},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result, err := exec.FireCommandAndReadResponse(conn, cmd)
				assert.Nil(t, err, "Unexpected error for cmd: %s", cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd: %s", cmd)
			}
		})
	}
}
