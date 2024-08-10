package core

import (
	"context"
	"fmt"
	"hash"
	"hash/fnv"
	"runtime"
	"sync"
)

var (
	DefaultLockConcurrency = runtime.NumCPU() * 2
)

type (
	HashKey     uint32
	LockManager struct {
		ctx         context.Context
		concurrency uint32
		locks       map[HashKey]*LockStore
	}
	LockStore struct {
		mutex      *sync.RWMutex
		RLockCount uint32
		WLockCount uint32
	}
)

func NewLockStore() (lockStore *LockStore) {
	lockStore = &LockStore{
		mutex: &sync.RWMutex{},
	}
	return
}

func NewLockManager() (lockMgr *LockManager) {
	lockMgr = &LockManager{
		ctx:         context.Background(),
		concurrency: uint32(DefaultLockConcurrency),
		locks:       make(map[HashKey]*LockStore, DefaultLockConcurrency),
	}
	for i := 0; i < DefaultLockConcurrency; i++ {
		lockMgr.locks[HashKey(i)] = NewLockStore()
	}
	return
}

func (lockMgr *LockManager) GetStore(hashKey HashKey) (lockStore *LockStore, err error) {
	var (
		isPresent bool
		hashSlot  uint32
		hashFn    hash.Hash32
	)
	hashFn = fnv.New32a()
	if _, err = hashFn.Write([]byte(string(hashKey))); err != nil {
		return
	}
	hashSlot = hashFn.Sum32() % lockMgr.concurrency
	if lockStore, isPresent = lockMgr.locks[HashKey(hashSlot)]; !isPresent {
		err = fmt.Errorf("lock store not found for %d", hashKey)
		return
	}
	return
}

func (lockMgr *LockManager) RLock(hashKey HashKey) (err error) {
	var (
		lockStore *LockStore
	)
	if lockStore, err = lockMgr.GetStore(hashKey); err != nil {
		return
	}
	lockStore.mutex.RLock()
	lockStore.RLockCount += 1
	return
}

func (lockMgr *LockManager) RUnlock(hashKey HashKey) (err error) {
	var (
		lockStore *LockStore
	)
	if lockStore, err = lockMgr.GetStore(hashKey); err != nil {
		return
	}
	lockStore.mutex.RUnlock()
	lockStore.RLockCount -= 1
	return
}

func (lockMgr *LockManager) Lock(hashKey HashKey) (err error) {
	var (
		lockStore *LockStore
	)
	if lockStore, err = lockMgr.GetStore(hashKey); err != nil {
		return
	}
	lockStore.mutex.Lock()
	lockStore.WLockCount += 1
	return
}

func (lockMgr *LockManager) Unlock(hashKey HashKey) (err error) {
	var (
		lockStore *LockStore
	)
	if lockStore, err = lockMgr.GetStore(hashKey); err != nil {
		return
	}
	lockStore.mutex.Lock()
	lockStore.WLockCount -= 1
	return
}
