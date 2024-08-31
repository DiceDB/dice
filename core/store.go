package core

import (
	"path"
	"sync"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server/utils"
)

type Store struct {
	store        *swiss.Map[string, *Obj]
	expires      *swiss.Map[*Obj, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
	keypool      *swiss.Map[string, *string]
	storeMutex   sync.RWMutex
	keypoolMutex sync.RWMutex
}

func NewStore() *Store {
	WatchChan = make(chan WatchEvent, config.KeysLimit)
	WatchSubscriptionChan = make(chan WatchSubscription)

	return &Store{
		store:   swiss.New[string, *Obj](10240),
		expires: swiss.New[*Obj, uint64](10240),
		keypool: swiss.New[string, *string](10240),
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
		store.store.Clear()
		store.expires.Clear()
		store.keypool.Clear()
		WatchChan = make(chan WatchEvent, config.KeysLimit)
		WatchSubscriptionChan = make(chan WatchSubscription)
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

	if store.store.Len() >= config.KeysLimit {
		store.evict()
	}
	obj.LastAccessedAt = getCurrentClock()

	ptr := store.ensureKeyInPool(k)
	currentObject, ok := store.store.Get(*ptr)
	if ok {
		v, _ := store.expires.Get(currentObject)
		if options.KeepTTL && v > 0 {
			v1, _ := store.expires.Get(currentObject)
			store.expires.Put(obj, v1)
		}
		store.expires.Delete(currentObject)
	}
	store.store.Put(*ptr, obj)

	store.incrementKeyCount()
	notifyWatchers(k, "SET")
}

func (store *Store) getHelper(k string, touch bool) *Obj {
	var v *Obj
	withLocks(func() {
		ptr, ok := store.keypool.Get(k)
		if !ok {
			return
		}

		v, _ = store.store.Get(*ptr)
		if v != nil {
			if hasExpired(v, store) {
				store.deleteKey(k, *ptr, v)
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
			ptr, ok := store.keypool.Get(k)
			if !ok {
				response = append(response, nil)
				continue
			}
			v, _ := store.store.Get(*ptr)
			if v != nil {
				if hasExpired(v, store) {
					store.deleteKey(k, *ptr, v)
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
		ptr, ok := store.keypool.Get(k)
		if !ok {
			return false
		}
		v, _ := store.store.Get(*ptr)
		return store.deleteKey(k, *ptr, v)
	}, store, WithStoreLock(), WithKeypoolLock())
}

func (store *Store) DelByPtr(ptr string) bool {
	return withLocksReturn(func() bool {
		return store.delByPtr(ptr)
	}, store, WithStoreLock(), WithKeypoolLock())
}

func (store *Store) Keys(p string) ([]string, error) {
	var keys []string
	var err error

	withLocks(func() {
		keys = make([]string, 0, store.keypool.Len())

		store.keypool.All(func(k string, _ *string) (stop bool) {
			if found, e := path.Match(p, k); e != nil {
				err = e
				return true
			} else if found {
				keys = append(keys, k)
			}
			return false
		})
	}, store, WithStoreRLock(), WithKeypoolRLock())

	return keys, err
}

// GetDbSize returns number of keys present in the database
func (store *Store) GetDBSize() uint64 {
	var noOfKeys uint64
	withLocks(func() {
		noOfKeys = uint64(store.keypool.Len())
	}, store, WithKeypoolRLock())
	return noOfKeys
}

// Rename function to implement RENAME functionality using existing helpers

func (store *Store) Rename(sourceKey, destKey string) bool {
	return withLocksReturn(func() bool {
		// If source and destination are the same, do nothing and return true
		if sourceKey == destKey {
			return true
		}

		sourcePtr, sourceOk := store.keypool.Get(sourceKey)
		if !sourceOk {
			return false
		}

		sourceObj, _ := store.store.Get(*sourcePtr)
		if sourceObj == nil || hasExpired(sourceObj, store) {
			if sourceObj != nil {
				store.deleteKey(sourceKey, *sourcePtr, sourceObj)
			}
			return false
		}

		// Use putHelper to handle putting the object at the destination key
		store.putHelper(destKey, sourceObj)

		// Remove the source key
		store.store.Delete(*sourcePtr)
		store.keypool.Delete(sourceKey)
		if KeyspaceStat[0] != nil {
			KeyspaceStat[0]["keys"]--
		}

		// Notify watchers about the deletion of the source key
		notifyWatchers(sourceKey, "DEL")

		return true
	}, store, WithStoreLock(), WithKeypoolLock())
}

func (store *Store) incrementKeyCount() {
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++
}

func (store *Store) Get(k string) *Obj {
	return store.getHelper(k, true)
}

func (store *Store) GetDel(k string) *Obj {
	var v *Obj
	withLocks(func() {
		ptr, ok := store.keypool.Get(k)
		if !ok {
			return
		}

		v, _ = store.store.Get(*ptr)
		if v != nil {
			expired := hasExpired(v, store)
			store.deleteKey(k, *ptr, v)
			if expired {
				v = nil
			}
		}
	}, store, WithStoreLock(), WithKeypoolLock())
	return v
}

// setExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) setExpiry(obj *Obj, expDurationMs int64) {
	store.expires.Put(obj, uint64(utils.GetCurrentTime().UnixMilli())+uint64(expDurationMs))
}

// setUnixTimeExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) setUnixTimeExpiry(obj *Obj, exUnixTimeSec int64) {
	// convert unix-time-seconds to unix-time-milliseconds
	store.expires.Put(obj, uint64(exUnixTimeSec*1000))
}

func (store *Store) ensureKeyInPool(k string) *string {
	ptr, ok := store.keypool.Get(k)
	if !ok {
		ptr = &k
		store.keypool.Put(k, ptr)
	}
	return ptr
}

func (store *Store) deleteKey(k, ptr string, obj *Obj) bool {
	if obj != nil {
		store.store.Delete(ptr)
		store.expires.Delete(obj)
		store.keypool.Delete(k)
		KeyspaceStat[0]["keys"]--
		notifyWatchers(k, "DEL")
		return true
	}
	return false
}

func (store *Store) delByPtr(ptr string) bool {
	if obj, ok := store.store.Get(ptr); ok {
		key := ptr
		return store.deleteKey(key, ptr, obj)
	}
	return false
}

func notifyWatchers(k, operation string) {
	WatchChan <- WatchEvent{k, operation}
}
