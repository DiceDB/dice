package iomultiplexer

import (
	"fmt"
	"syscall"
	"time"
)

// Epoll implements the IOMultiplexer interface for Linux-based systems
type Epoll struct {
	// fd stores the file descriptor of the epoll instance
	fd int
	// ePollEvents acts as a buffer for the events returned by the EpollWait syscall
	ePollEvents []syscall.EpollEvent
	// diceEvents stores the events after they are converted to the generic Event type
	// and is returned to the caller
	diceEvents []Event
}

// New creates a new Epoll instance
func New(maxClients int) (*Epoll, error) {
	if maxClients < 0 {
		return nil, ErrInvalidMaxClients
	}

	fd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}

	return &Epoll{
		fd:          fd,
		ePollEvents: make([]syscall.EpollEvent, maxClients),
		diceEvents:  make([]Event, maxClients),
	}, nil
}

// Subscribe subscribes to the given event
func (ep *Epoll) Subscribe(event Event) error {
	nativeEvent := event.toNative()
	if err := syscall.EpollCtl(ep.fd, syscall.EPOLL_CTL_ADD, event.Fd, &nativeEvent); err != nil {
		return fmt.Errorf("failed to subscribe to event: %w", err)
	}
	return nil
}

// Poll polls for all the subscribed events simultaneously
// and returns all the events that were triggered
// It blocks until at least one event is triggered or the timeout is reached
func (ep *Epoll) Poll(timeout time.Duration) ([]Event, error) {
	nEvents, err := syscall.EpollWait(ep.fd, ep.ePollEvents, newTime(timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to poll events: %w", err)
	}

	for i := 0; i < nEvents; i++ {
		ep.diceEvents[i] = newEvent(ep.ePollEvents[i])
	}

	return ep.diceEvents[:nEvents], nil
}

// Close closes the Epoll instance
func (ep *Epoll) Close() error {
	return syscall.Close(ep.fd)
}
