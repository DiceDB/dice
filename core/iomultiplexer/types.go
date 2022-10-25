package iomultiplexer

// Event is a platform independent representation of a subscribe event
// For linux platform, this is translated to syscall.EpollEvent
// For darwin platform, this is translated to syscall.Kevent_t
type Event struct {
	// Fd denotes the file descriptor
	Fd int
	Op Operations
}

// Operations is a platform independent representation of the operations
// that can be performed on a file descriptor
type Operations uint32
