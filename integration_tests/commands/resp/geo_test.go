package resp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeoAdd(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoDist(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}

func TestGeoPos(t *testing.T) {
	conn := getLocalConnection()
	defer conn.Close()

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
				result := FireCommand(conn, cmd)
				assert.Equal(t, tc.expect[i], result, "Value mismatch for cmd %s", cmd)
			}
		})
	}
}