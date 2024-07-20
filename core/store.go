package core

import (
	"path"
	"sync"
	"unsafe"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server/utils"
)

type WatchEvent struct {
	Key       string
	Operation string
	Value     *Obj
}

var store map[unsafe.Pointer]*Obj
var expires map[*Obj]uint64 // Does not need to be thread-safe as it is only accessed by a single thread.
var keypool map[string]unsafe.Pointer
var WatchList sync.Map // Maps queries to the file descriptors of clients that are watching them.

// WatchChannel Channel to receive updates about keys that are being watched.
var WatchChannel chan WatchEvent

func NewStore() *Store {
	WatchChannel = make(chan WatchEvent, config.KeysLimit)
	return &Store{
		store:   make(map[unsafe.Pointer]*Obj),
		expires: make(map[*Obj]uint64),
		keypool: make(map[string]unsafe.Pointer),
	}
}

func (store *Store) NewObj(value interface{}, expDurationMs int64, oType, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}
	if expDurationMs >= 0 {
		store.setExpiry(obj, expDurationMs)
	}
	return obj
}

func (store *Store) ResetStore() {
	withLocks(func() {
		store.store = make(map[unsafe.Pointer]*Obj)
		store.expires = make(map[*Obj]uint64)
		store.keypool = make(map[string]unsafe.Pointer)
		WatchChannel = make(chan WatchEvent, config.KeysLimit)
	}, store, WithStoreLock(), WithKeypoolLock())
}

type PutOptions struct {
	KeepTTL bool
}

func (store *Store) Put(k string, obj *Obj, opts ...PutOption) {
	withLocks(func() {
		store.putHelper(k, obj, opts...)
	}, store, WithStoreLock(), WithKeypoolLock())
}

func getDefaultOptions() *PutOptions {
	return &PutOptions{
		KeepTTL: false,
	}
}

type PutOption func(*PutOptions)

func WithKeepTTL(value bool) PutOption {
	return func(po *PutOptions) {
		po.KeepTTL = value
	}
}

func (store *Store) PutAll(data map[string]*Obj) {
	withLocks(func() {
		for k, obj := range data {
			store.putHelper(k, obj)
		}
	}, store, WithStoreLock(), WithKeypoolLock())
}

func (store *Store) GetNoTouch(k string) *Obj {
	return store.getHelper(k, false)
}

func (store *Store) putHelper(k string, obj *Obj, opts ...PutOption) {
	options := getDefaultOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	if len(store.store) >= config.KeysLimit {
		store.evict()
	}
	obj.LastAccessedAt = getCurrentClock()

	ptr := store.ensureKeyInPool(k)
	currentObject, ok := store.store[ptr]
	if ok {
		if options.KeepTTL && store.expires[currentObject] > 0 {
			store.expires[obj] = store.expires[currentObject]
		}
		delete(store.expires, currentObject)
	}
	store.store[ptr] = obj

	store.incrementKeyCount()
	notifyWatchers(k, "SET", obj)
}

func (store *Store) getHelper(k string, touch bool) *Obj {
	var v *Obj
	withLocks(func() {
		ptr, ok := store.keypool[k]
		if !ok {
			return
		}

		v = store.store[ptr]
		if v != nil {
			if hasExpired(v, store) {
				store.deleteKey(k, ptr, v)
				v = nil
			} else if touch {
				v.LastAccessedAt = getCurrentClock()
			}
		}
	}, store, WithStoreLock(), WithKeypoolLock())
	return v
}

func (store *Store) GetAll(keys []string) []*Obj {
	response := make([]*Obj, 0, len(keys))
	withLocks(func() {
		for _, k := range keys {
			ptr, ok := store.keypool[k]
			if !ok {
				response = append(response, nil)
				continue
			}
			v := store.store[ptr]
			if v != nil {
				if hasExpired(v, store) {
					store.deleteKey(k, ptr, v)
					response = append(response, nil)
				} else {
					v.LastAccessedAt = getCurrentClock()
					response = append(response, v)
				}
			} else {
				response = append(response, nil)
			}
		}
	}, store, WithStoreRLock(), WithKeypoolRLock())
	return response
}

func (store *Store) Del(k string) bool {
	return withLocksReturn(func() bool {
		ptr, ok := store.keypool[k]
		if !ok {
			return false
		}
		return store.deleteKey(k, ptr, store.store[ptr])
	}, store, WithStoreLock(), WithKeypoolLock())
}

func (store *Store) DelByPtr(ptr unsafe.Pointer) bool {
	return withLocksReturn(func() bool {
		return store.delByPtr(ptr)
	}, store, WithStoreLock(), WithKeypoolLock())
}

func (store *Store) Keys(p string) ([]string, error) {
	var keys []string
	var err error

	withLocks(func() {
		keys = make([]string, 0, len(store.keypool))
		for k := range store.keypool {
			if found, e := path.Match(p, k); e != nil {
				err = e
				return
			} else if found {
				keys = append(keys, k)
			}
		}
	}, store, WithStoreRLock(), WithKeypoolRLock())

	return keys, err
}

// GetDbSize returns number of keys present in the database
func (store *Store) GetDBSize() uint64 {
	var noOfKeys uint64
	withLocks(func() {
		noOfKeys = uint64(len(store.keypool))
	}, store, WithKeypoolRLock())
	return noOfKeys
}

// Function to add a new watcher to a query.
func AddWatcher(query DSQLQuery, clientFd int) {
	clients, _ := WatchList.LoadOrStore(query, &sync.Map{})
	clients.(*sync.Map).Store(clientFd, struct{}{})
}

// Function to remove a watcher from a query.
func RemoveWatcher(query DSQLQuery, clientFd int) {
	if clients, ok := WatchList.Load(query); ok {
		clients.(*sync.Map).Delete(clientFd)
		// If no more clients for this query, remove the query from WatchList
		if clientCount := countClients(clients.(*sync.Map)); clientCount == 0 {
			WatchList.Delete(query)
		}
	}
}

// Helper function to count clients
func countClients(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// Function to add a new watcher to a query.
func AddWatcher(query DSQLQuery, clientFd int) {
	clients, _ := WatchList.LoadOrStore(query, &sync.Map{})
	clients.(*sync.Map).Store(clientFd, struct{}{})
}

// Function to remove a watcher from a query.
func RemoveWatcher(query DSQLQuery, clientFd int) {
	if clients, ok := WatchList.Load(query); ok {
		clients.(*sync.Map).Delete(clientFd)
		// If no more clients for this query, remove the query from WatchList
		if clientCount := countClients(clients.(*sync.Map)); clientCount == 0 {
			WatchList.Delete(query)
		}
	}
}

// Helper function to count clients
func countClients(clients *sync.Map) int {
	count := 0
	clients.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
