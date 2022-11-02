package handlers

import (
	"sync"
	"sync/atomic"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/eviction"
	"github.com/dicedb/dice/metrics"
	"github.com/dicedb/dice/object"
	"github.com/dicedb/dice/utils"
)

type DiceKVstoreHandler struct {
	object.DiceKVstore
}

func NewDiceKVstoreHandler() *DiceKVstoreHandler {
	dbHandler := new(DiceKVstoreHandler)
	dbHandler.Store = new(sync.Map)
	dbHandler.Count = new(atomic.Uint64)
	return dbHandler
}

// Put stores an entry into KV store
func (dh *DiceKVstoreHandler) Put(k string, obj *object.Obj) {
	count := dh.Count.Load()
	if count >= uint64(config.KeysLimit) {
		eviction.Evict(dh)
	}
	obj.LastAccessedAt = utils.GetCurrentClock()

	// Concurrent/Lock-free but synchronised storage
	dh.Store.Store(k, obj)
	if metrics.KeyspaceStat[0] == nil {
		metrics.KeyspaceStat[0] = make(map[string]int)
	}
	metrics.KeyspaceStat[0]["keys"]++
}

func (dh *DiceKVstoreHandler) Get(k string) *object.Obj {
	value, ok := dh.Store.Load(k)
	valueObj := value.(*object.Obj)
	if ok && value != nil {
		if object.GetDiceExpiryStore().HasExpired(valueObj) {
			dh.Del(k)
			return nil
		}
		valueObj.LastAccessedAt = utils.GetCurrentClock()
	}
	return valueObj
}

func (dh *DiceKVstoreHandler) Del(k string) bool {
	if obj, ok := dh.Store.Load(k); ok {
		dh.Store.Delete(k)
		object.GetDiceExpiryStore().Storage().Delete(obj.(*object.Obj))
		metrics.KeyspaceStat[0]["keys"]--
		return true
	}
	return false
}

func (dh *DiceKVstoreHandler) GetCount() uint64 {
	return dh.Size()
}

func (dh *DiceKVstoreHandler) GetStorage() *sync.Map {
	storage := dh.Storage()
	// We can perform custom logic for the handler
	return storage
}
