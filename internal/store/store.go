// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package store

import (
	"log"
	"path"
	"sync"

	"github.com/dicedb/dice/internal/common"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
)

func NewStoreRegMap() common.ITable[string, *object.Obj] {
	return &common.RegMap[string, *object.Obj]{
		M: make(map[string]*object.Obj),
	}
}

func NewExpireRegMap() common.ITable[*object.Obj, uint64] {
	return &common.RegMap[*object.Obj, uint64]{
		M: make(map[*object.Obj]uint64),
	}
}

func NewStoreMap() common.ITable[string, *object.Obj] {
	return NewStoreRegMap()
}

func NewExpireMap() common.ITable[*object.Obj, uint64] {
	return NewExpireRegMap()
}

func NewDefaultEviction() EvictionStrategy {
	return &PrimitiveEvictionStrategy{}
}

// QueryWatchEvent represents a change in a watched key.
type QueryWatchEvent struct {
	Key       string
	Operation string
	Value     object.Obj
}

type CmdWatchEvent struct {
	Cmd         string
	AffectedKey string
}

type Store struct {
	store            common.ITable[string, *object.Obj]
	expires          common.ITable[*object.Obj, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
	numKeys          int
	cmdWatchChan     chan CmdWatchEvent
	evictionStrategy EvictionStrategy
	ShardID          int

	// Snapshot related fields
	ongoingSnapshots map[uint64]Snapshotter
	snapLock         *sync.RWMutex
}

type Snapshotter interface {
	TempAdd(string, interface{}) error
	TempGet(string) (interface{}, error)
	Store(string, interface{}) error
	Close() error
}

func NewStore(cmdWatchChan chan CmdWatchEvent, evictionStrategy EvictionStrategy, shardID int) *Store {
	store := &Store{
		store:            NewStoreRegMap(),
		expires:          NewExpireRegMap(),
		cmdWatchChan:     cmdWatchChan,
		evictionStrategy: evictionStrategy,
		ongoingSnapshots: make(map[uint64]Snapshotter),
		snapLock:         &sync.RWMutex{},
		ShardID:          shardID,
	}
	if evictionStrategy == nil {
		store.evictionStrategy = NewDefaultEviction()
	}

	return store
}

func Reset(store *Store) *Store {
	store.numKeys = 0
	store.store = NewStoreMap()
	store.expires = NewExpireMap()

	return store
}

func (store *Store) NewObj(value interface{}, expDurationMs int64, oType object.ObjectType) *object.Obj {
	obj := &object.Obj{
		Value:          value,
		Type:           oType,
		LastAccessedAt: getCurrentClock(),
	}
	if expDurationMs >= 0 {
		store.SetExpiry(obj, expDurationMs)
	}
	return obj
}

func (store *Store) ResetStore() {
	store.numKeys = 0
	store.store = NewStoreMap()
	store.expires = NewExpireMap()
}

func (store *Store) StartSnapshot(snapshotID uint64, snapshot Snapshotter) {
	store.snapLock.Lock()
	store.ongoingSnapshots[snapshotID] = snapshot
	store.snapLock.Unlock()

	// Iterate and store all the keys in the snapshot
	store.IterateAllKeysForSnapshot(snapshotID)

}

func (store *Store) StopSnapshot(snapshotID uint64) {
	store.snapLock.Lock()
	if snapshot, isPresent := store.ongoingSnapshots[snapshotID]; isPresent {
		snapshot.Close()
		delete(store.ongoingSnapshots, snapshotID)
	}
	store.snapLock.Unlock()
}

func (store *Store) IterateAllKeysForSnapshot(snapshotID uint64) (err error) {
	var (
		snapshot  Snapshotter
		isPresent bool
	)
	// If there are no ongoing snapshots, then there won't be any temp data to store
	if len(store.ongoingSnapshots) == 0 {
		return
	}
	store.snapLock.RLock()
	if snapshot, isPresent = store.ongoingSnapshots[snapshotID]; !isPresent {
		return
	}
	store.snapLock.RUnlock()

	store.store.All(func(k string, v *object.Obj) bool {
		// Check if the data is overridden
		tempVal, _ := snapshot.TempGet(k)
		if tempVal != nil {
			v = tempVal.(*object.Obj)
		}
		err = snapshot.Store(k, v)
		if err != nil {
			log.Println("Error storing data in snapshot", "error", err)
		}
		return true
	})
	store.StopSnapshot(snapshotID)
	return
}

func (store *Store) Put(k string, obj *object.Obj, opts ...PutOption) {
	store.putHelper(k, obj, opts...)
}

func (store *Store) GetKeyCount() int {
	return store.numKeys
}

func (store *Store) IncrementKeyCount() {
	store.numKeys++
}

func (store *Store) PutAll(data map[string]*object.Obj) {
	for k, obj := range data {
		store.putHelper(k, obj)
	}
}

func (store *Store) GetNoTouch(k string) *object.Obj {
	return store.getHelper(k, false)
}

func (store *Store) putHelper(k string, obj *object.Obj, opts ...PutOption) {
	options := getDefaultPutOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	obj.LastAccessedAt = getCurrentClock()
	currentObject, ok := store.store.Get(k)
	if ok {
		v, ok1 := store.expires.Get(currentObject)
		if ok1 && options.KeepTTL && v > 0 {
			v1, ok2 := store.expires.Get(currentObject)
			if ok2 {
				store.expires.Put(obj, v1)
			}
		}
		store.expires.Delete(currentObject)
	} else {
		// TODO: Inform all the io-threads and shards about the eviction.
		// TODO: Start the eviction only when all the io-thread and shards have acknowledged the eviction.
		evictCount := store.evictionStrategy.ShouldEvict(store)
		if evictCount > 0 {
			store.evict(evictCount)
		}
		store.numKeys++
	}

	// Take a snapshot of the existing data since it's going to be overridden
	for _, snapshot := range store.ongoingSnapshots {
		snapshot.TempAdd(k, obj)
	}
	store.store.Put(k, obj)
	store.evictionStrategy.OnAccess(k, obj, AccessSet)

	if store.cmdWatchChan != nil {
		store.notifyWatchManager(options.PutCmd, k)
	}
}

// getHelper is a helper function to get the object from the store. It also updates the last accessed time if touch is true.
func (store *Store) getHelper(k string, touch bool) *object.Obj {
	var obj *object.Obj
	obj, _ = store.store.Get(k)
	if obj != nil {
		if hasExpired(obj, store) {
			store.deleteKey(k, obj)
			obj = nil
		} else if touch {
			obj.LastAccessedAt = getCurrentClock()
			store.evictionStrategy.OnAccess(k, obj, AccessGet)
		}
	}
	return obj
}

func (store *Store) GetAll(keys []string) []*object.Obj {
	response := make([]*object.Obj, 0, len(keys))
	for _, k := range keys {
		v, _ := store.store.Get(k)
		if v != nil {
			if hasExpired(v, store) {
				store.deleteKey(k, v)
				response = append(response, nil)
			} else {
				v.LastAccessedAt = getCurrentClock()
				response = append(response, v)
			}
		} else {
			response = append(response, nil)
		}
	}
	return response
}

func (store *Store) Del(k string, opts ...DelOption) bool {
	v, ok := store.store.Get(k)
	if ok {
		return store.deleteKey(k, v, opts...)
	}
	return false
}

func (store *Store) DelByPtr(ptr string, opts ...DelOption) bool {
	return store.delByPtr(ptr, opts...)
}

func (store *Store) Keys(p string) ([]string, error) {
	var keys []string
	var err error

	keys = make([]string, 0, store.store.Len())

	store.store.All(func(k string, _ *object.Obj) bool {
		if found, e := path.Match(p, k); e != nil {
			err = e
			// stop iteration if any error
			return false
		} else if found {
			keys = append(keys, k)
		}
		// continue the iteration
		return true
	})

	return keys, err
}

// GetDBSize returns number of keys present in the database
func (store *Store) GetDBSize() uint64 {
	return uint64(store.store.Len())
}

// Rename function to implement RENAME functionality using existing helpers
func (store *Store) Rename(sourceKey, destKey string) bool {
	// If source and destination are the same, do nothing and return true
	if sourceKey == destKey {
		return true
	}

	sourceObj, _ := store.store.Get(sourceKey)
	if sourceObj == nil || hasExpired(sourceObj, store) {
		if sourceObj != nil {
			store.deleteKey(sourceKey, sourceObj, WithDelCmd(Rename))
		}
		return false
	}

	// Use putHelper to handle putting the object at the destination key
	store.putHelper(destKey, sourceObj, WithPutCmd(Set))

	// Remove the source key
	store.store.Delete(sourceKey)
	store.numKeys--

	if store.cmdWatchChan != nil {
		store.notifyWatchManager(Rename, sourceKey)
	}

	return true
}

func (store *Store) Get(k string) *object.Obj {
	return store.getHelper(k, true)
}

func (store *Store) GetDel(k string, opts ...DelOption) *object.Obj {
	var v *object.Obj
	v, _ = store.store.Get(k)
	if v != nil {
		expired := hasExpired(v, store)
		store.deleteKey(k, v, opts...)
		if expired {
			v = nil
		}
	}
	return v
}

// SetExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) SetExpiry(obj *object.Obj, expDurationMs int64) {
	store.expires.Put(obj, uint64(utils.GetCurrentTime().UnixMilli())+uint64(expDurationMs))
}

// SetUnixTimeExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) SetUnixTimeExpiry(obj *object.Obj, exUnixTimeSec int64) {
	// convert unix-time-seconds to unix-time-milliseconds
	store.expires.Put(obj, uint64(exUnixTimeSec*1000))
}

func (store *Store) deleteKey(k string, obj *object.Obj, opts ...DelOption) bool {
	options := getDefaultDelOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	if obj != nil {
		// Take a snapshot of the existing data since it's going to be overridden
		for _, snapshot := range store.ongoingSnapshots {
			snapshot.TempAdd(k, obj)
		}
		store.store.Delete(k)
		store.expires.Delete(obj)
		store.numKeys--

		store.evictionStrategy.OnAccess(k, obj, AccessDel)

		if store.cmdWatchChan != nil {
			store.notifyWatchManager(options.DelCmd, k)
		}

		return true
	}

	return false
}

func (store *Store) delByPtr(ptr string, opts ...DelOption) bool {
	if obj, ok := store.store.Get(ptr); ok {
		key := ptr
		return store.deleteKey(key, obj, opts...)
	}
	return false
}

func (store *Store) notifyWatchManager(cmd, affectedKey string) {
	store.cmdWatchChan <- CmdWatchEvent{cmd, affectedKey}
}

func (store *Store) GetStore() common.ITable[string, *object.Obj] {
	return store.store
}

func (store *Store) evict(evictCount int) bool {
	store.evictionStrategy.EvictVictims(store, evictCount)
	return true
}
