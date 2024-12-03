package tests

import (
	"time"
)

var getTestCases = []Meta{
	{
		Name:   "Get on non-existing key",
		Input:  []string{"GET k"},
		Output: []interface{}{"(nil)"},
	},
	{
		Name:    "Get on existing key",
		Input:   []string{"SET k v", "GET k"},
		Output:  []interface{}{"OK", "v"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Get with expiration",
		Input:   []string{"SET k v EX 2", "GET k", "GET k"},
		Output:  []interface{}{"OK", "v", "(nil)"},
		Delays:  []time.Duration{0, 0, 3 * time.Second},
		Cleanup: []string{"k"},
	},
}

func init() {
	RegisterTests(getTestCases)
}
