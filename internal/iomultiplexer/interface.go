// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
