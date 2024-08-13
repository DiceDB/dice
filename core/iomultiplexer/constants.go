package iomultiplexer

const (
	// OpRead represents the read operation
	OpRead Operations = 1 << iota
	// OpWrite represents the write operation
	OpWrite
)
