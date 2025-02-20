// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/dicedb/dicedb-go/wire"
)

func TestGet(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []string
		expect []any
		delays []time.Duration
	}{
		{
			name:   "Get with expiration",
			cmds:   []string{"SET k v EX 4", "GET k", "GET k"},
			expect: []any{
				&wire.Response_VStr{VStr: "OK"},
				&wire.Response_VStr{VStr: "v"},
				&wire.Response_VNil{VNil: true},
			},
			delays: []time.Duration{0, 0, 5 * time.Second},
		},
		{
			name:   "Get with non existent key",
			cmds:   []string{"GET nek"},
			expect: []any{
				&wire.Response_VNil{VNil: true},
			},
			delays: []time.Duration{0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := client.FireString(cmd)
				assert.Equal(t, tc.expect[i], result.GetValue(), "Value mismatch for cmd %s", cmd)
			}
		})
	}
}
