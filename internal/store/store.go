package store

import (
	"path"

	"github.com/dicedb/dice/config"

	"github.com/dicedb/dice/internal/common"
	ds "github.com/dicedb/dice/internal/datastructures"
	"github.com/dicedb/dice/internal/server/utils"
)

func NewStoreRegMap() common.ITable[string, ds.DSInterface] {
	return &common.RegMap[string, ds.DSInterface]{
		M: make(map[string]ds.DSInterface),
	}
}

func NewExpireRegMap() common.ITable[*ds.DSInterface, uint64] {
	return &common.RegMap[*ds.DSInterface, uint64]{
		M: make(map[*ds.DSInterface]uint64),
	}
}

func NewStoreMap() common.ITable[string, ds.DSInterface] {
	return NewStoreRegMap()
}

func NewExpireMap() common.ITable[*ds.DSInterface, uint64] {
	return NewExpireRegMap()
}

func NewDefaultEviction() EvictionStrategy {
	return &BatchEvictionLRU{
		maxKeys:       config.DefaultKeysLimit,
		evictionRatio: config.DefaultEvictionRatio,
	}
}

// QueryWatchEvent represents a change in a watched key.
type QueryWatchEvent struct {
	Key       string
	Operation string
	Value     ds.DSInterface
}

type CmdWatchEvent struct {
	Cmd         string
	AffectedKey string
}

type Store struct {
	store            common.ITable[string, ds.DSInterface]
	expires          common.ITable[*ds.DSInterface, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
	numKeys          int
	queryWatchChan   chan QueryWatchEvent
	cmdWatchChan     chan CmdWatchEvent
	evictionStrategy EvictionStrategy
}

func NewStore(queryWatchChan chan QueryWatchEvent, cmdWatchChan chan CmdWatchEvent, evictionStrategy EvictionStrategy) *Store {
	store := &Store{
		store:            NewStoreRegMap(),
		expires:          NewExpireRegMap(),
		queryWatchChan:   queryWatchChan,
		cmdWatchChan:     cmdWatchChan,
		evictionStrategy: evictionStrategy,
	}
	if evictionStrategy == nil {
		store.evictionStrategy = NewDefaultEviction()
	}

	return store
}

func ResetStore(store *Store) *Store {
	store.numKeys = 0
	store.store = NewStoreMap()
	store.expires = NewExpireMap()

	return store
}

func (store *Store) ResetStore() {
	store.numKeys = 0
	store.store = NewStoreMap()
	store.expires = NewExpireMap()
}

func (store *Store) NewObj(value ds.DSInterface, exDurationMs int64) ds.DSInterface {
	if exDurationMs > 0 {
		store.SetExpiry(&value, exDurationMs)
	}
	return value
}

func (store *Store) Put(k string, obj ds.DSInterface, opts ...PutOption) {
	store.putHelper(k, obj, opts...)
}

func (store *Store) GetKeyCount() int {
	return store.numKeys
}

func (store *Store) IncrementKeyCount() {
	store.numKeys++
}

func (store *Store) PutAll(data map[string]ds.DSInterface) {
	for k, obj := range data {
		store.putHelper(k, obj)
	}
}

func (store *Store) GetNoTouch(k string) ds.DSInterface {
	return store.getHelper(k, false)
}

func (store *Store) putHelper(k string, obj ds.DSInterface, opts ...PutOption) {
	options := getDefaultPutOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	obj.UpdateLastAccessedAt()
	currentObject, ok := store.store.Get(k)
	if ok {
		v, ok1 := store.expires.Get(&currentObject)
		if ok1 && options.KeepTTL && v > 0 {
			v1, ok2 := store.expires.Get(&currentObject)
			if ok2 {
				store.expires.Put(&obj, v1)
			}
		}
		store.expires.Delete(&currentObject)
	} else {
		// TODO: Inform all the io-threads and shards about the eviction.
		// TODO: Start the eviction only when all the io-thread and shards have acknowledged the eviction.
		evictCount := store.evictionStrategy.ShouldEvict(store)
		if evictCount > 0 {
			store.evict(evictCount)
		}
		store.numKeys++
	}

	store.store.Put(k, obj)
	store.evictionStrategy.OnAccess(k, &obj, AccessSet)

	// if store.queryWatchChan != nil {
	// 	store.notifyQueryManager(k, Set, *obj)
	// }
	// if store.cmdWatchChan != nil {
	// 	store.notifyWatchManager(options.PutCmd, k)
	// }
}

// getHelper is a helper function to get the object from the store. It also updates the last accessed time if touch is true.
func (store *Store) getHelper(k string, touch bool) ds.DSInterface {
	obj, _ := store.store.Get(k)
	if obj != nil {
		if hasExpired(&obj, store) {
			store.deleteKey(k, obj)
			obj = nil
		} else if touch {
			obj.UpdateLastAccessedAt()
			store.evictionStrategy.OnAccess(k, &obj, AccessGet)
		}
	}
	return obj
}

func (store *Store) GetAll(keys []string) []ds.DSInterface {
	response := make([]ds.DSInterface, 0, len(keys))
	for _, k := range keys {
		v, _ := store.store.Get(k)
		if v != nil {
			if hasExpired(&v, store) {
				store.deleteKey(k, v)
				response = append(response, nil)
			} else {
				v.UpdateLastAccessedAt()
				response = append(response, v)
			}
		} else {
			response = append(response, nil)
		}
	}
	return response
}

func (store *Store) Del(k string, opts ...DelOption) bool {
	v, ok := store.store.Get(k)
	if ok {
		return store.deleteKey(k, v, opts...)
	}
	return false
}

func (store *Store) DelByPtr(ptr string, opts ...DelOption) bool {
	return store.delByPtr(ptr, opts...)
}

func (store *Store) Keys(p string) ([]string, error) {
	var keys []string
	var err error

	keys = make([]string, 0, store.store.Len())

	store.store.All(func(k string, _ ds.DSInterface) bool {
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

	return keys, err
}

// GetDBSize returns number of keys present in the database
func (store *Store) GetDBSize() uint64 {
	return uint64(store.store.Len())
}

// Rename function to implement RENAME functionality using existing helpers
func (store *Store) Rename(sourceKey, destKey string) bool {
	// If source and destination are the same, do nothing and return true
	if sourceKey == destKey {
		return true
	}

	sourceObj, _ := store.store.Get(sourceKey)
	if sourceObj == nil || hasExpired(&sourceObj, store) {
		if sourceObj != nil {
			store.deleteKey(sourceKey, sourceObj, WithDelCmd(Rename))
		}
		return false
	}

	// Use putHelper to handle putting the object at the destination key
	store.putHelper(destKey, sourceObj, WithPutCmd(Set))

	// Remove the source key
	store.store.Delete(sourceKey)
	store.numKeys--

	// Notify watchers about the deletion of the source key
	// if store.queryWatchChan != nil {
	// 	store.notifyQueryManager(sourceKey, Del, *sourceObj)
	// }
	// if store.cmdWatchChan != nil {
	// 	store.notifyWatchManager(Rename, sourceKey)
	// }

	return true
}

func (store *Store) Get(k string) ds.DSInterface {
	return store.getHelper(k, true)
}

func (store *Store) GetDel(k string, opts ...DelOption) ds.DSInterface {
	v, _ := store.store.Get(k)
	if v != nil {
		expired := hasExpired(&v, store)
		store.deleteKey(k, v, opts...)
		if expired {
			v = nil
		}
	}
	return v
}

// SetExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) SetExpiry(obj *ds.DSInterface, expDurationMs int64) {
	store.expires.Put(obj, uint64(utils.GetCurrentTime().UnixMilli())+uint64(expDurationMs))
}

// SetUnixTimeExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) SetUnixTimeExpiry(obj *ds.DSInterface, exUnixTimeSec int64) {
	// convert unix-time-seconds to unix-time-milliseconds
	store.expires.Put(obj, uint64(exUnixTimeSec*1000))
}

func (store *Store) deleteKey(k string, obj ds.DSInterface, opts ...DelOption) bool {
	options := getDefaultDelOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	if obj != nil {
		store.store.Delete(k)
		store.expires.Delete(&obj)
		store.numKeys--

		store.evictionStrategy.OnAccess(k, &obj, AccessDel)

		// if store.queryWatchChan != nil {
		// 	store.notifyQueryManager(k, Del, *obj)
		// }
		// if store.cmdWatchChan != nil {
		// 	store.notifyWatchManager(options.DelCmd, k)
		// }

		return true
	}

	return false
}

func (store *Store) delByPtr(ptr string, opts ...DelOption) bool {
	if obj, ok := store.store.Get(ptr); ok {
		key := ptr
		return store.deleteKey(key, obj, opts...)
	}
	return false
}

// notifyQueryManager notifies the query manager about a key change, so that it can update the query cache if needed.
// func (store *Store) notifyQueryManager(k, operation string, obj T) {
// 	store.queryWatchChan <- QueryWatchEvent{k, operation, obj}
// }

// func (store *Store) notifyWatchManager(cmd, affectedKey string) {
// 	store.cmdWatchChan <- CmdWatchEvent{cmd, affectedKey}
// }

func (store *Store) GetStore() common.ITable[string, ds.DSInterface] {
	return store.store
}

// CacheKeysForQuery scans the store for keys that match the given where clause and sends them to the cache channel.
// This allows the query manager to cache the existing keys that match the query.
// func (store *Store) CacheKeysForQuery(whereClause sqlparser.Expr, cacheChannel chan *[]struct {
// 	Key   string
// 	Value *object.Obj
// }) {
// 	shardCache := make([]struct {
// 		Key   string
// 		Value *object.Obj
// 	}, 0)
// 	store.store.All(func(k string, v *T) bool {
// 		matches, err := sql.EvaluateWhereClause(whereClause, sql.QueryResultRow{Key: k, Value: *v}, make(map[string]jp.Expr))
// 		if err != nil || !matches {
// 			return true
// 		}

// 		shardCache = append(shardCache, struct {
// 			Key   string
// 			Value *object.Obj
// 		}{Key: k, Value: v})

// 		return true
// 	})
// 	cacheChannel <- &shardCache
// }

func (store *Store) evict(evictCount int) bool {
	store.evictionStrategy.EvictVictims(store, evictCount)
	return true
}
