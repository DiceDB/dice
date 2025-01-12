// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

package resp

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	// Run the test server
	// This is a synchronous method, because internally it
	// checks for available port and then forks a goroutine
	// to start the server
	opts := TestServerOptions{
		Port: 9739,
	}
	RunTestServer(&wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Run the test suite
	exitCode := m.Run()

	conn := getLocalConnection()
	if conn == nil {
		panic("Failed to connect to the test server")
	}
	defer conn.Close()
	result := FireCommand(conn, "ABORT")
	if result != "OK" {
		panic("Failed to abort the server")
	}

	wg.Wait()
	os.Exit(exitCode)
}
