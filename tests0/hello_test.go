// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/config"
)

func TestHello(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	runTestcases(t, client, []TestCase{
		{
			name:     "Hello",
			commands: []string{"HELLO"},
			expected: []interface{}{
				"proto", int64(2),
				"id", fmt.Sprintf("%s:%d", config.Config.Host, config.Config.Port),
				"mode", "standalone",
				"role", "master",
				"modules", []interface{}{},
			},
		},
	})
}
