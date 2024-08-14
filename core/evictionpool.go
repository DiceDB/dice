package core

import (
	"sort"
	"unsafe"

	"github.com/dicedb/dice/config"
)

type PoolItem struct {
	keyPtr         unsafe.Pointer
	lastAccessedAt uint32
}

// TODO: When last accessed at of object changes
// update the poolItem corresponding to that
type EvictionPool struct {
	pool   []*PoolItem
	keyset map[unsafe.Pointer]*PoolItem
}

type ByIdleTime []*PoolItem
type ByCounterAndIdleTime []*PoolItem

func (a ByIdleTime) Len() int {
	return len(a)
}

func (a ByIdleTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByIdleTime) Less(i, j int) bool {
	return getIdleTime(a[i].lastAccessedAt) > getIdleTime(a[j].lastAccessedAt)
}

func (a ByCounterAndIdleTime) Len() int {
  return len(a)
}

func (a ByCounterAndIdleTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByCounterAndIdleTime) Less(i, j int) bool {
  counterI := getLFULogCounter(a[i].lastAccessedAt) 
  counterJ := getLFULogCounter(a[j].lastAccessedAt) 

  if counterI == counterJ {
    // if access counters are same, sort by idle time
    lastAccessedAtI := a[i].lastAccessedAt & 0x00FFFFFF
    lastAccessedAtJ := a[j].lastAccessedAt & 0x00FFFFFF

    return getIdleTime(lastAccessedAtI) < getIdleTime(lastAccessedAtJ)
  }

  return counterI < counterJ
}

// TODO: Make the implementation efficient to not need repeated sorting
func (pq *EvictionPool) Push(key unsafe.Pointer, lastAccessedAt uint32) {
	_, ok := pq.keyset[key]
	if ok {
		return
	}
	item := &PoolItem{keyPtr: key, lastAccessedAt: lastAccessedAt}
	if len(pq.pool) < ePoolSizeMax {
		pq.keyset[key] = item
		pq.pool = append(pq.pool, item)

		// Performance bottleneck
    switch config.EvictionStrategy {
    case config.ALL_KEYS_LFU:
      sort.Sort(ByCounterAndIdleTime(pq.pool))
    default:
      sort.Sort(ByIdleTime(pq.pool))
    }
	} else if lastAccessedAt > pq.pool[0].lastAccessedAt {
		pq.pool = pq.pool[1:]
		pq.keyset[key] = item
		pq.pool = append(pq.pool, item)
	}
}

func (pq *EvictionPool) Pop() *PoolItem {
	if len(pq.pool) == 0 {
		return nil
	}
	item := pq.pool[0]
	pq.pool = pq.pool[1:]
	delete(pq.keyset, item.keyPtr)
	return item
}

func newEvictionPool(size int) *EvictionPool {
	return &EvictionPool{
		pool:   make([]*PoolItem, size),
		keyset: make(map[unsafe.Pointer]*PoolItem),
	}
}

var ePoolSizeMax int = 16
var ePool *EvictionPool = newEvictionPool(0)
