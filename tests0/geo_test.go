// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeoAdd(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

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
			expect: []interface{}{int64(1), int64(0)},
		},
		{
			name:   "GeoAdd With Adding New Member And Updating it with NX",
			cmds:   []string{"GEOADD mygeo  1.21 1.44 MD", "GEOADD mygeo 1.22 1.54 MD"},
			expect: []interface{}{int64(1), int64(0)},
		},
		{
			name:   "GEOADD with both NX and XX options",
			cmds:   []string{"GEOADD mygeo NX XX  1.21 1.44 MD"},
			expect: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:   "GEOADD invalid longitude",
			cmds:   []string{"GEOADD mygeo  181.0 1.44 MD"},
			expect: []interface{}{"ERR invalid longitude"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoDist(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "GEODIST b/w existing points",
			cmds: []string{
				"GEOADD points 13.361389 38.115556 Palermo",
				"GEOADD points 15.087269 37.502669 Catania",
				"GEODIST points Palermo Catania",
				"GEODIST points Palermo Catania km",
			},
			expect: []interface{}{int64(1), int64(1), "166274.144", "166.2741"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoPos(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "GEOPOS b/w existing points",
			cmds: []string{
				"GEOADD index 13.361389 38.115556 Palermo",
				"GEOPOS index Palermo",
			},
			expect: []interface{}{
				int64(1),
				[]interface{}{[]interface{}{"13.361387", "38.115556"}},
			},
		},
		{
			name: "GEOPOS for non existing points",
			cmds: []string{
				"GEOPOS index NonExisting",
			},
			expect: []interface{}{
				[]interface{}{"(nil)"},
			},
		},
		{
			name: "GEOPOS for non existing index",
			cmds: []string{
				"GEOPOS NonExisting Palermo",
			},
			expect: []interface{}{"(nil)"},
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
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoHash(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []interface{}
		delays []time.Duration
	}{
		{
			name: "GEOHASH with no arguments",
			cmds: []string{
				"GEOHASH points",
			},
			expect: []interface{}{"ERR wrong number of arguments for 'geohash' command"},
		},
		{
			name: "GEOHASH on non-geo key",
			cmds: []string{
				"SET key value",
				"GEOHASH key member",
			},
			expect: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name: "GEOHASH with non-existent key",
			cmds: []string{
				"GEOHASH geopoints NonExistent",
			},
			expect: []interface{}{"ERR no such key"},
		},
		{
			name: "GEOHASH with non-existent member",
			cmds: []string{
				"GEOADD points -74.0060 40.7128 NewYork",
				"GEOHASH points NonExistent",
			},
			expect: []interface{}{int64(1), []interface{}{"(nil)"}},
		},
		{
			name: "GEOHASH with a single member",
			cmds: []string{
				"GEOHASH points NewYork",
			},
			expect: []interface{}{[]interface{}{"dr5regw3pp"}},
		},
		{
			name: "GEOHASH with multiple members",
			cmds: []string{
				"GEOADD points -73.935242 40.730610 Brooklyn",
				"GEOHASH points NewYork Brooklyn NonExistent",
			},
			expect: []interface{}{int64(1), []interface{}{"dr5regw3pp", "dr5rtwccpb", "(nil)"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
