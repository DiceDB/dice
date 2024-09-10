package store

import (
	"github.com/dicedb/dice/internal/object"
	"path"
	"sync"

	"github.com/dicedb/dice/internal/server/utils"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/config"
)

// WatchEvent represents a change in a watched key.
type WatchEvent struct {
	Key       string
	Operation string
	Value     *object.Obj
}

type Store struct {
	store      *swiss.Map[string, *object.Obj]
	expires    *swiss.Map[*object.Obj, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
	storeMutex sync.RWMutex
	watchChan  chan WatchEvent
}

func NewStore(watchChan chan WatchEvent) *Store {
	return &Store{
		store:     swiss.New[string, *object.Obj](10240),
		expires:   swiss.New[*object.Obj, uint64](10240),
		watchChan: watchChan,
	}
}

func ResetStore(store *Store) *Store {
	store.store = swiss.New[string, *object.Obj](10240)
	store.expires = swiss.New[*object.Obj, uint64](10240)

	return store
}

func (store *Store) NewObj(value interface{}, expDurationMs int64, oType, oEnc uint8) *object.Obj {
	obj := &object.Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}
	if expDurationMs >= 0 {
		store.SetExpiry(obj, expDurationMs)
	}
	return obj
}

func (store *Store) ResetStore() {
	WithLocks(func() {
		store.store.Clear()
		store.expires.Clear()
		store.watchChan = make(chan WatchEvent, config.KeysLimit)
	}, store, WithStoreLock())
}

type PutOptions struct {
	KeepTTL bool
}

func (store *Store) Put(k string, obj *object.Obj, opts ...PutOption) {
	WithLocks(func() {
		store.putHelper(k, obj, opts...)
	}, store, WithStoreLock())
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

func (store *Store) PutAll(data map[string]*object.Obj) {
	WithLocks(func() {
		for k, obj := range data {
			store.putHelper(k, obj)
		}
	}, store, WithStoreLock())
}

func (store *Store) GetNoTouch(k string) *object.Obj {
	return store.getHelper(k, false)
}

func (store *Store) putHelper(k string, obj *object.Obj, opts ...PutOption) {
	options := getDefaultOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	if store.store.Len() >= config.KeysLimit {
		store.evict()
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
	}
	store.store.Put(k, obj)

	store.incrementKeyCount()

	if store.watchChan != nil {
		store.notifyWatchers(k, Set, obj)
	}
}

func (store *Store) getHelper(k string, touch bool) *object.Obj {
	var v *object.Obj
	WithLocks(func() {
		v, _ = store.store.Get(k)
		if v != nil {
			if hasExpired(v, store) {
				store.deleteKey(k, v)
				v = nil
			} else if touch {
				v.LastAccessedAt = getCurrentClock()
			}
		}
	}, store, WithStoreLock())
	return v
}

func (store *Store) GetAll(keys []string) []*object.Obj {
	response := make([]*object.Obj, 0, len(keys))
	WithLocks(func() {
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
	}, store, WithStoreRLock())
	return response
}

func (store *Store) Del(k string) bool {
	return withLocksReturn(func() bool {
		v, ok := store.store.Get(k)
		if ok {
			return store.deleteKey(k, v)
		}
		return false
	}, store, WithStoreLock())
}

func (store *Store) DelByPtr(ptr string) bool {
	return withLocksReturn(func() bool {
		return store.delByPtr(ptr)
	}, store, WithStoreLock())
}

func (store *Store) Keys(p string) ([]string, error) {
	var keys []string
	var err error

	WithLocks(func() {
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
	}, store, WithStoreRLock())

	return keys, err
}

// GetDbSize returns number of keys present in the database
func (store *Store) GetDBSize() uint64 {
	var noOfKeys uint64
	WithLocks(func() {
		noOfKeys = uint64(store.store.Len())
	}, store, WithStoreRLock())
	return noOfKeys
}

// Rename function to implement RENAME functionality using existing helpers

func (store *Store) Rename(sourceKey, destKey string) bool {
	return withLocksReturn(func() bool {
		// If source and destination are the same, do nothing and return true
		if sourceKey == destKey {
			return true
		}

		sourceObj, _ := store.store.Get(sourceKey)
		if sourceObj == nil || hasExpired(sourceObj, store) {
			if sourceObj != nil {
				store.deleteKey(sourceKey, sourceObj)
			}
			return false
		}

		// Use putHelper to handle putting the object at the destination key
		store.putHelper(destKey, sourceObj)

		// Remove the source key
		store.store.Delete(sourceKey)
		if KeyspaceStat[0] != nil {
			KeyspaceStat[0]["keys"]--
		}

		// Notify watchers about the deletion of the source key
		if store.watchChan != nil {
			store.notifyWatchers(sourceKey, Del, sourceObj)
		}

		return true
	}, store, WithStoreLock())
}

func (store *Store) incrementKeyCount() {
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++
}

func (store *Store) Get(k string) *object.Obj {
	return store.getHelper(k, true)
}

func (store *Store) GetDel(k string) *object.Obj {
	var v *object.Obj
	WithLocks(func() {
		v, _ = store.store.Get(k)
		if v != nil {
			expired := hasExpired(v, store)
			store.deleteKey(k, v)
			if expired {
				v = nil
			}
		}
	}, store, WithStoreLock())
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

func (store *Store) deleteKey(k string, obj *object.Obj) bool {
	if obj != nil {
		store.store.Delete(k)
		store.expires.Delete(obj)
		KeyspaceStat[0]["keys"]--

		if store.watchChan != nil {
			store.notifyWatchers(k, Del, obj)
		}

		return true
	}

	return false
}

func (store *Store) delByPtr(ptr string) bool {
	if obj, ok := store.store.Get(ptr); ok {
		key := ptr
		return store.deleteKey(key, obj)
	}
	return false
}

func (store *Store) notifyWatchers(k, operation string, obj *object.Obj) {
	store.watchChan <- WatchEvent{k, operation, obj}
}

func (store *Store) GetStore() *swiss.Map[string, *object.Obj] {
	return store.store
}
