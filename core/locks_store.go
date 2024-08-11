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

	LockH struct {
		hash      [32]*Lock
		lockCount uint8
	}
)

func NewLockH() (lockH *LockH) {
	lockH = &LockH{
		hash:      [32]*Lock{},
		lockCount: 0,
	}
	if err := lockH.setup(); err != nil {
		return
	}
	return
}

func (lockH *LockH) setup() (err error) {
	availableLocks := []LockName{
		StoreLock,
		KeypoolLock,
	}
	for _, lockName := range availableLocks {
		if _, err = lockH.addLock(lockName); err != nil {
			return
		}
	}
	return
}

func (lockH *LockH) addLock(name LockName) (lock *Lock, err error) {
	lock = &Lock{
		mutex:    &sync.RWMutex{},
		name:     name,
		refCount: 0,
	}
	if lockH.hash[uint8(lock.name)] != nil {
		err = fmt.Errorf("slot already filled for %d", lock.name)
		return
	}
	lockH.hash[uint8(lock.name)] = lock
	lockH.lockCount++
	return
}

func (lockH *LockH) getLock(name LockName) (lock *Lock, err error) {
	if lock = lockH.hash[uint8(name)]; lock == nil {
		err = fmt.Errorf("lock not found for %d", name)
		return
	}
	return
}

func (lockH *LockH) removeLock(name LockName) (err error) {
	if lockH.hash[uint8(name)] == nil {
		err = fmt.Errorf("lock not found for %d", name)
		return
	}
	lockH.hash[uint8(name)] = nil
	lockH.lockCount--
	return
}
