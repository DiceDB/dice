package id

import "sync/atomic"

var id uint64 = 0

func NextUint64() uint64 {
	return atomic.AddUint64(&id, 1)
}
