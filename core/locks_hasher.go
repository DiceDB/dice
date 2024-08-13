package core

import (
	"context"
	"fmt"
	"hash"
	"hash/fnv"
)

const (
	DefaultLockConcurrency = 1024
	DefaultLockIdentifier  = "default-lock"
)

type (
	LockHasher struct {
		ctx         context.Context
		concurrency uint32
		locks       [DefaultLockConcurrency]*LockStore
	}
)

func NewLockHasher() (lockHsh *LockHasher) {
	lockHsh = &LockHasher{
		ctx:         context.Background(),
		concurrency: uint32(DefaultLockConcurrency),
	}
	for i := 0; i < DefaultLockConcurrency; i++ {
		lockHsh.locks[i] = NewLockH()
	}
	return
}

func (lockHsh *LockHasher) GetHash(strKey string) (hashSlot uint32, err error) {
	var (
		hashFn hash.Hash32
	)
	if strKey == "" {
		strKey = DefaultLockIdentifier
	}
	hashFn = fnv.New32a()
	if _, err = hashFn.Write([]byte(strKey)); err != nil {
		return
	}
	hashSlot = hashFn.Sum32() % lockHsh.concurrency
	return
}

func (lockHsh *LockHasher) GetLockStore(strKey string) (lockH *LockStore, err error) {
	var (
		hashSlot uint32
	)
	if hashSlot, err = lockHsh.GetHash(strKey); err != nil {
		return
	}
	if lockH = lockHsh.locks[hashSlot]; lockH == nil {
		err = fmt.Errorf("lock store not found for %s", strKey)
		return
	}
	return
}
