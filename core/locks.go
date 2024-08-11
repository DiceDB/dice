package core

import "log"

var (
	LockHsh *LockHasher = NewLockHasher()
)

const (
	// ReadLockOperation is used to acquire a read lock
	ReadLockOperation = LockOperationT(1)
	// WriteLockOperation is used to acquire a read lock
	WriteLockOperation = LockOperationT(2)
)

type (
	LockOperationT uint8

	LockRequest struct {
		Name LockName
		Op   LockOperationT
	}
)

// WithStoreLock sets the storeLock flag
func WithStoreLock() *LockRequest {
	return &LockRequest{
		Name: StoreLock,
		Op:   ReadLockOperation,
	}
}

// WithStoreRLock sets the storeRLock flag
func WithStoreRLock() *LockRequest {
	return &LockRequest{
		Name: StoreLock,
		Op:   ReadLockOperation,
	}
}

// WithKeypoolLock sets the keypoolLock flag
func WithKeypoolLock() *LockRequest {
	return &LockRequest{
		Name: KeypoolLock,
		Op:   WriteLockOperation,
	}
}

// WithKeypoolRLock sets the keypoolRLock flag
func WithKeypoolRLock() *LockRequest {
	return &LockRequest{
		Name: KeypoolLock,
		Op:   ReadLockOperation,
	}
}

// withLocks takes a function and a list of LockOptions and executes the function
// with the specified locks. It manages the locking and unlocking of the mutexes
// based on the LockOptions provided.
func withLocks(id string, f func(), reqs ...*LockRequest) {
	var (
		err error
	)
	for _, req := range reqs {
		var (
			lock  *Lock
			lockH *LockH
		)
		if lockH, err = LockHsh.GetStore(id); err != nil {
			log.Println("error in fetching lockStore for id", id, err)
			return
		}
		if lock, err = lockH.getLock(req.Name); err != nil {
			log.Println("error in fetching lock for name", req.Name, err)
			return
		}
		switch req.Op {
		case ReadLockOperation:
			lock.mutex.RLock()
			defer lock.mutex.RUnlock()
		case WriteLockOperation:
			lock.mutex.Lock()
			defer lock.mutex.Unlock()
		}
	}
	f()
}

// Helper function for operations that return a boolean
func withLocksReturn(id string, f func() bool, reqs ...*LockRequest) bool {
	var result bool
	withLocks(id, func() {
		result = f()
	}, reqs...)
	return result
}
