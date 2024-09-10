package store

// LockOption represents a function that can modify a LockStrategy
type LockOption func(*LockStrategy)

// LockStrategy holds the state for our locking mechanism
type LockStrategy struct {
	storeLock  bool
	storeRLock bool
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

// WithLocks takes a function and a list of LockOptions and executes the function
// with the specified locks. It manages the locking and unlocking of the mutexes
// based on the LockOptions provided.
func WithLocks(f func(), store *Store, options ...LockOption) {
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

	f()
}

// Helper function for operations that return a boolean
func withLocksReturn(f func() bool, store *Store, options ...LockOption) bool {
	var result bool
	WithLocks(func() {
		result = f()
	}, store, options...)
	return result
}
