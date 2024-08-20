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
func withLocks(f func(), store *Store, options ...LockOption) {
	ls := &LockStrategy{}

	for _, option := range options {
		option(ls)
	}

	if ls.storeLock {
		store.storeMutex.Lock()
		defer store.storeMutex.Unlock()
	} else if ls.storeRLock {
		store.storeMutex.RLock()
		defer store.storeMutex.RUnlock()
	}

	if ls.keypoolLock {
		store.keypoolMutex.Lock()
		defer store.keypoolMutex.Unlock()
	} else if ls.keypoolRLock {
		store.keypoolMutex.RLock()
		defer store.keypoolMutex.RUnlock()
	}

	f()
}

// Helper function for operations that return a boolean
func withLocksReturn(f func() bool, store *Store, options ...LockOption) bool {
	var result bool
	withLocks(func() {
		result = f()
	}, store, options...)
	return result
}
