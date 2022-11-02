package object

import (
	// "fmt"
	"sync"
	"sync/atomic"
	"time"
)

var expires *DiceExpiryStore

func init() {
	expires = new(DiceExpiryStore)
	expires.store = new(sync.Map)
	expires.Count = new(atomic.Uint64)
}

type DiceExpiryStore struct {
	Count *atomic.Uint64
	store *sync.Map
}

func GetDiceExpiryStore() *DiceExpiryStore {
	return expires
}

// ---------- Member Methods ---------- //
func (des *DiceExpiryStore) Size() uint64 {
	return des.Count.Load()
}
func (des *DiceExpiryStore) Storage() *sync.Map {
	return des.store
}
func (des *DiceExpiryStore) PurgeExpiryStore() bool {
	return true
}
func (des *DiceExpiryStore) SetExpiry(obj *Obj, expDurationMs int64) {
	des.store.Store(obj, uint64(time.Now().UnixMilli())+uint64(expDurationMs))
	// expires.store.Store(obj, uint64(time.Now().UnixMilli())+uint64(expDurationMs))

}
func (des *DiceExpiryStore) GetExpiry(obj *Obj) (uint64, bool) {
	// fmt.Printf("The value of the object in GetExpiry() is %v\n", *obj)
	// fmt.Println(*expires.store)
	exp, ok := des.store.Load(obj)
	var expUnit uint64
	if ok {
		expUnit = exp.(uint64)
	}
	return expUnit, ok
}
func (des *DiceExpiryStore) HasExpired(obj *Obj) bool {
	exp, ok := des.GetExpiry(obj)
	if !ok {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}
