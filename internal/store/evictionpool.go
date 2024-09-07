package store

import (
	"sort"

	"github.com/dicedb/dice/config"
)

type PoolItem struct {
	keyPtr         string
	lastAccessedAt uint32
}

// TODO: When last accessed at of object changes
// update the poolItem corresponding to that
type EvictionPool struct {
	pool   []*PoolItem
	keyset map[string]*PoolItem
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
	return GetIdleTime(a[i].lastAccessedAt) > GetIdleTime(a[j].lastAccessedAt)
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

		return GetIdleTime(lastAccessedAtI) < GetIdleTime(lastAccessedAtJ)
	}

	return counterI < counterJ
}

// TODO: Make the implementation efficient to not need repeated sorting
func (pq *EvictionPool) Push(key string, lastAccessedAt uint32) {
	_, ok := pq.keyset[key]
	if ok {
		return
	}
	item := &PoolItem{keyPtr: key, lastAccessedAt: lastAccessedAt}
	if len(pq.pool) < ePoolSizeMax {
		pq.keyset[key] = item
		pq.pool = append(pq.pool, item)

		// Performance bottleneck
		if config.EvictionStrategy == config.AllKeysLFU {
			sort.Sort(ByCounterAndIdleTime(pq.pool))
		} else {
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
		keyset: make(map[string]*PoolItem),
	}
}

var ePoolSizeMax int = 16
var ePool *EvictionPool = newEvictionPool(0)
