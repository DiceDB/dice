package pool

import (
	"time"
	"github.com/panjf2000/ants/v2"
)
// Type alias for the ant pools
type DicePool = ants.PoolWithFunc

func init() {
	ants.Release()
}

const (
	// Max number of workers 256 * 1024
	DefaultPoolSize = 1 << 18

	// If we should wait when pool workers ain't available
	// false make it won't wait
	Nonblocking = false

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second
)

// NewDefaultDicePool gets a new default dice pool
func NewDefaultDicePool(job func (i interface{})) (*DicePool, error) {
	return ants.NewPoolWithFunc(DefaultPoolSize, job, 
								ants.WithNonblocking(Nonblocking),
							    ants.WithExpiryDuration(ExpiryDuration))
}

