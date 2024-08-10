package core

// LockOption represents a function that can modify a LockStrategy
type LockOption func(*LockStrategy)

// LockStrategy holds the state for our locking mechanism
type LockStrategy struct {
	storeLock    bool
	storeRLock   bool
	keypoolLock  bool
	keypoolRLock bool
}

// WithStoreLock sets the storeLock flag
func WithStoreLock() LockOption {
	return func(ls *LockStrategy) {
		ls.storeLock = true
	}
}

// WithStoreRLock sets the storeRLock flag
func WithStoreRLock() LockOption {
	return func(ls *LockStrategy) {
		ls.storeRLock = true
	}
}

// WithKeypoolLock sets the keypoolLock flag
func WithKeypoolLock() LockOption {
	return func(ls *LockStrategy) {
		ls.keypoolLock = true
	}
}

// WithKeypoolRLock sets the keypoolRLock flag
func WithKeypoolRLock() LockOption {
	return func(ls *LockStrategy) {
		ls.keypoolRLock = true
	}
}

// withLocks takes a function and a list of LockOptions and executes the function
// with the specified locks. It manages the locking and unlocking of the mutexes
// based on the LockOptions provided.
func withLocks(f func(), options ...LockOption) {
	ls := &LockStrategy{}

	for _, option := range options {
		option(ls)
	}

	if ls.storeLock {
		storeMutex.Lock()
		defer storeMutex.Unlock()
	} else if ls.storeRLock {
		storeMutex.RLock()
		defer storeMutex.RUnlock()
	}

	if ls.keypoolLock {
		keypoolMutex.Lock()
		defer keypoolMutex.Unlock()
	} else if ls.keypoolRLock {
		keypoolMutex.RLock()
		defer keypoolMutex.RUnlock()
	}

	f()
}

// Helper function for operations that return a boolean
func withLocksReturn(f func() bool, options ...LockOption) bool {
	var result bool
	withLocks(func() {
		result = f()
	}, options...)
	return result
}
