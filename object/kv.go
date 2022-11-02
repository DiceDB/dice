package object

import (
	"sync"
	"sync/atomic"
)

// var Store *DiceKVstore

// DiceKVstore represents a thread/concurrency safe KV pair
type DiceKVstore struct {
	Count *atomic.Uint64
	Store *sync.Map
}

func (dkv *DiceKVstore) Size() uint64 {
	return dkv.Count.Load()
}

func (dkv *DiceKVstore) Storage() *sync.Map {
	return dkv.Store
}

func (dkv *DiceKVstore) PurgeKVStorage() bool {
	// TODO: Implement the Purge for the KV
	return true
}

// GetKVStoreObj gets the access of the DiceKVStore object
// from the handlers so that methods that are not enforced
// by the engine interface can be accessed as well.
// For example `istorageengines` doesn't enforce to implement
// `PurgeKVStorage` but if we still need to access it from
// the handler or other palces, we can get the `DiceKVStore`
// object and access the method
func (dkv *DiceKVstore) GetKVStoreObj() *DiceKVstore {
	return dkv
}

// type IDiceKVstore interface {
// 	GetDiceKVstore() *DiceKVstore
// }

// func (dkvs *DiceKVstore) GetDiceKVstore() *DiceKVstore {
// 	return dkvs
// }
