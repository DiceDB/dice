package core

import (
	"fmt"
	"sync"
)

const (
	StoreLock   LockName = LockName(1)
	KeypoolLock LockName = LockName(2)
)

type (
	// Used to define different locks for different functionalities
	LockName uint8

	Lock struct {
		mutex    *sync.RWMutex
		name     LockName
		refCount uint32
	}

	LockStore struct {
		hash      [32]*Lock
		lockCount uint8
	}
)

func NewLockH() (lockSt *LockStore) {
	lockSt = &LockStore{
		hash:      [32]*Lock{},
		lockCount: 0,
	}
	if err := lockSt.setup(); err != nil {
		return
	}
	return
}

func (lockSt *LockStore) setup() (err error) {
	availableLocks := []LockName{
		StoreLock,
		KeypoolLock,
	}
	for _, lockName := range availableLocks {
		if _, err = lockSt.addLock(lockName); err != nil {
			return
		}
	}
	return
}

func (lockSt *LockStore) addLock(name LockName) (lock *Lock, err error) {
	lock = &Lock{
		mutex:    &sync.RWMutex{},
		name:     name,
		refCount: 0,
	}
	if lockSt.hash[uint8(lock.name)] != nil {
		err = fmt.Errorf("slot already filled for %d", lock.name)
		return
	}
	lockSt.hash[uint8(lock.name)] = lock
	lockSt.lockCount++
	return
}

func (lockSt *LockStore) getLock(name LockName) (lock *Lock, err error) {
	if lock = lockSt.hash[uint8(name)]; lock == nil {
		err = fmt.Errorf("lock not found for %d", name)
		return
	}
	return
}

func (lockSt *LockStore) removeLock(name LockName) (err error) {
	if lockSt.hash[uint8(name)] == nil {
		err = fmt.Errorf("lock not found for %d", name)
		return
	}
	lockSt.hash[uint8(name)] = nil
	lockSt.lockCount--
	return
}
