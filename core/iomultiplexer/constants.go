package iomultiplexer

const (
	// OP_READ represents the read operation
	OP_READ Operations = 1 << iota
	// OP_WRITE represents the write operation
	OP_WRITE
)
