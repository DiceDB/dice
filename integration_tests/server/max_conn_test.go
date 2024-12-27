// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	commands "github.com/dicedb/dice/integration_tests/commands/resp"
)

func getConnection(port int) (net.Conn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TestMaxConnection(t *testing.T) {
	var wg sync.WaitGroup
	var maxConnTestOptions = commands.TestServerOptions{
		Port:       8741,
		MaxClients: 50,
	}
	commands.RunTestServer(&wg, maxConnTestOptions)

	time.Sleep(2 * time.Second)

	var maxConnLimit = maxConnTestOptions.MaxClients + 2
	connections := make([]net.Conn, maxConnLimit)
	defer func() {
		// Ensure all connections are closed at the end of the test
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	for i := int32(0); i < maxConnLimit; i++ {
		conn, err := getConnection(maxConnTestOptions.Port)
		if err == nil {
			connections[i] = conn
		} else {
			t.Fatalf("unexpected error while getting connection %d: %v", i, err)
		}
	}
	assert.Equal(t, maxConnLimit, int32(len(connections)), "should have reached the max connection limit")

	result := commands.FireCommand(connections[0], "ABORT")
	if result != "OK" {
		t.Fatalf("Unexpected response to ABORT command: %v", result)
	} else {
		slog.Info("Closed server for max_conn_test")
	}
	wg.Wait()
}
