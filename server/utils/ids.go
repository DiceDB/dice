package utils

import (
	"sync"
	"sync/atomic"
)

var counter int32 = 0
var mu sync.Mutex

// The logic fais when the counter cycles back
// we need to look into different approaches.
// @arpitbbhayani is looking into TS-Counter based approach
// but the issue is when we reset the value of counter, we have to
// ensure it is the beginning of the second. if not then we cannot
// guarantee the order.
// @arpitbbhayani is looking into different databases and papers
// But this is a good approach to start with.
func GenerateCmdID() int32 {
	mu.Lock()
	defer mu.Unlock()
	atomic.AddInt32(&counter, 1)
	return atomic.LoadInt32(&counter)
}
