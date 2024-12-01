package tests

import (
	"strconv"
	"time"
)

var expiryTime = strconv.FormatInt(time.Now().Add(1*time.Minute).UnixMilli(), 10)

var testCases = []Meta{
	{
		Name:    "Set and Get Simple Value",
		Input:   []string{"SET k v", "GET k"},
		Output:  []interface{}{"OK", "v"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Set and Get Integer Value",
		Input:   []string{"SET k 123456789", "GET k"},
		Output:  []interface{}{"OK", int64(123456789)},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Overwrite Existing Key",
		Input:   []string{"SET k v1", "SET k 5", "GET k"},
		Output:  []interface{}{"OK", "OK", int64(5)},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Set with EX option",
		Input:   []string{"SET k v EX 2", "GET k", "SLEEP 3", "GET k"},
		Output:  []interface{}{"OK", "v", "OK", "(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Set with PX option",
		Input:   []string{"SET k v PX 2000", "GET k", "SLEEP 3", "GET k"},
		Output:  []interface{}{"OK", "v", "OK", "(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Set with EX and PX option",
		Input:   []string{"SET k v EX 2 PX 2000"},
		Output:  []interface{}{"ERR syntax error"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "XX on non-existing key",
		Input:   []string{"DEL k", "SET k v XX", "GET k"},
		Output:  []interface{}{int64(0), "(nil)", "(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "NX on non-existing key",
		Input:   []string{"SET k v NX", "GET k"},
		Output:  []interface{}{"OK", "v"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "NX on existing key",
		Input:   []string{"DEL k", "SET k v NX", "GET k", "SET k v NX"},
		Output:  []interface{}{int64(0), "OK", "v", "(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "PXAT option",
		Input:   []string{"SET k v PXAT " + expiryTime, "GET k"},
		Output:  []interface{}{"OK", "v"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "PXAT option with delete",
		Input:   []string{"SET k1 v1 PXAT " + expiryTime, "GET k1", "SLEEP 2", "DEL k1"},
		Output:  []interface{}{"OK", "v1", "OK", int64(1)},
		Cleanup: []string{"k1"},
	},
	{
		Name:    "PXAT option with invalid unix time ms",
		Input:   []string{"SET k2 v2 PXAT 123123", "GET k2"},
		Output:  []interface{}{"OK", "(nil)"},
		Cleanup: []string{"k2"},
	},
	{
		Name:    "XX on existing key",
		Input:   []string{"SET k v1", "SET k v2 XX", "GET k"},
		Output:  []interface{}{"OK", "OK", "v2"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "Multiple XX operations",
		Input:   []string{"SET k v1", "SET k v2 XX", "SET k v3 XX", "GET k"},
		Output:  []interface{}{"OK", "OK", "OK", "v3"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "EX option",
		Input:   []string{"SET k v EX 1", "GET k", "SLEEP 2", "GET k"},
		Output:  []interface{}{"OK", "v", "OK", "(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "XX option",
		Input:   []string{"SET k v XX EX 1", "GET k", "SLEEP 2", "GET k", "SET k v XX EX 1", "GET k"},
		Output:  []interface{}{"(nil)", "(nil)", "OK", "(nil)", "(nil)", "(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "GET with Existing Value",
		Input:   []string{"SET k v", "SET k vv GET"},
		Output:  []interface{}{"OK", "v"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "GET with Non-Existing Value",
		Input:   []string{"SET k vv GET"},
		Output:  []interface{}{"(nil)"},
		Cleanup: []string{"k"},
	},
	{
		Name:    "GET with wrong type of value",
		Input:   []string{"sadd k v", "SET k vv GET"},
		Output:  []interface{}{int64(1), "WRONGTYPE Operation against a key holding the wrong kind of value"},
		Cleanup: []string{"k"},
	},
}

func init() {
	RegisterTests(testCases)
}
