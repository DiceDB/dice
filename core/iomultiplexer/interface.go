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
