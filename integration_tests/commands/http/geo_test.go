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
