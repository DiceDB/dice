package id

import "sync/atomic"

var id uint64 = 0

// Global generator function
func NextClientID() uint64 {
	return atomic.AddUint64(&id, 1)
}
