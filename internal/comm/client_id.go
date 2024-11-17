package comm

import "sync/atomic"

var id uint64 = 0

func nextClientID() uint64 {
	return atomic.AddUint64(&id, 1)
}
