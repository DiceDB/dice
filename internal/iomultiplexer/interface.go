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

package iomultiplexer

import "time"

// IOMultiplexer defines a generic interface for platform-specific
// I/O multiplexer implementations
type IOMultiplexer interface {
	// Subscribe subscribes to the given event
	// When the event is triggered, the Poll method will return it
	Subscribe(event Event) error

	// Poll polls for all the subscribed events simultaneously
	// and returns all the events that were triggered
	// It blocks until at least one event is triggered or the timeout is reached
	Poll(timeout time.Duration) ([]Event, error)

	// Close closes the IOMultiplexer instance
	Close() error
}
