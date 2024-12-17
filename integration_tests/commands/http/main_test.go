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

package http

import (
	"context"
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
		Port: 8083,
	}
	ctx, cancel := context.WithCancel(context.Background())
	RunHTTPServer(ctx, &wg, opts)

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	executor := NewHTTPCommandExecutor()

	// Run the test suite
	exitCode := m.Run()

	executor.FireCommand(HTTPCommand{
		Command: "ABORT",
		Body:    map[string]interface{}{},
	})

	cancel()
	wg.Wait()
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}
