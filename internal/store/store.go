package store

import (
	"path"

	"github.com/dicedb/dice/config"

	"github.com/dicedb/dice/internal/common"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/sql"
	"github.com/ohler55/ojg/jp"
	"github.com/xwb1989/sqlparser"
)

func NewStoreRegMap() common.ITable[string, *object.Obj] {
	return &common.RegMap[string, *object.Obj]{
		M: make(map[string]*object.Obj),
	}
}

func NewExpireRegMap() common.ITable[*object.Obj, uint64] {
	return &common.RegMap[*object.Obj, uint64]{
		M: make(map[*object.Obj]uint64),
	}
}

func NewStoreMap() common.ITable[string, *object.Obj] {
	return NewStoreRegMap()
}

func NewExpireMap() common.ITable[*object.Obj, uint64] {
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
	Value     object.Obj
}

type CmdWatchEvent struct {
	Cmd         string
	AffectedKey string
}

type Store struct {
	store            common.ITable[string, *object.Obj]
	expires          common.ITable[*object.Obj, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
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
	store.numKeys = 0
	store.store = NewStoreMap()
	store.expires = NewExpireMap()
}

func (store *Store) Put(k string, obj *object.Obj, opts ...PutOption) {
	store.putHelper(k, obj, opts...)
}

func (store *Store) GetKeyCount() int {
	return store.numKeys
}

func (store *Store) IncrementKeyCount() {
	store.numKeys++
}

func (store *Store) PutAll(data map[string]*object.Obj) {
	for k, obj := range data {
		store.putHelper(k, obj)
	}
}

func (store *Store) GetNoTouch(k string) *object.Obj {
	return store.getHelper(k, false)
}

func (store *Store) putHelper(k string, obj *object.Obj, opts ...PutOption) {
	options := getDefaultPutOptions()

	for _, optApplier := range opts {
		optApplier(options)
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
	store.evictionStrategy.OnAccess(k, obj, AccessSet)

	if store.queryWatchChan != nil {
		store.notifyQueryManager(k, Set, *obj)
	}
	if store.cmdWatchChan != nil {
		store.notifyWatchManager(options.PutCmd, k)
	}
}

// getHelper is a helper function to get the object from the store. It also updates the last accessed time if touch is true.
func (store *Store) getHelper(k string, touch bool) *object.Obj {
	var obj *object.Obj
	obj, _ = store.store.Get(k)
	if obj != nil {
		if hasExpired(obj, store) {
			store.deleteKey(k, obj)
			obj = nil
		} else if touch {
			obj.LastAccessedAt = getCurrentClock()
			store.evictionStrategy.OnAccess(k, obj, AccessGet)
		}
	}
	return obj
}

func (store *Store) GetAll(keys []string) []*object.Obj {
	response := make([]*object.Obj, 0, len(keys))
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
	if sourceObj == nil || hasExpired(sourceObj, store) {
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
	if store.queryWatchChan != nil {
		store.notifyQueryManager(sourceKey, Del, *sourceObj)
	}
	if store.cmdWatchChan != nil {
		store.notifyWatchManager(Rename, sourceKey)
	}

	return true
}

func (store *Store) Get(k string) *object.Obj {
	return store.getHelper(k, true)
}

func (store *Store) GetDel(k string, opts ...DelOption) *object.Obj {
	var v *object.Obj
	v, _ = store.store.Get(k)
	if v != nil {
		expired := hasExpired(v, store)
		store.deleteKey(k, v, opts...)
		if expired {
			v = nil
		}
	}
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

func (store *Store) deleteKey(k string, obj *object.Obj, opts ...DelOption) bool {
	options := getDefaultDelOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	if obj != nil {
		store.store.Delete(k)
		store.expires.Delete(obj)
		store.numKeys--

		store.evictionStrategy.OnAccess(k, obj, AccessDel)

		if store.queryWatchChan != nil {
			store.notifyQueryManager(k, Del, *obj)
		}
		if store.cmdWatchChan != nil {
			store.notifyWatchManager(options.DelCmd, k)
		}

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
func (store *Store) notifyQueryManager(k, operation string, obj object.Obj) {
	store.queryWatchChan <- QueryWatchEvent{k, operation, obj}
}

func (store *Store) notifyWatchManager(cmd, affectedKey string) {
	store.cmdWatchChan <- CmdWatchEvent{cmd, affectedKey}
}

func (store *Store) GetStore() common.ITable[string, *object.Obj] {
	return store.store
}

// CacheKeysForQuery scans the store for keys that match the given where clause and sends them to the cache channel.
// This allows the query manager to cache the existing keys that match the query.
func (store *Store) CacheKeysForQuery(whereClause sqlparser.Expr, cacheChannel chan *[]struct {
	Key   string
	Value *object.Obj
}) {
	shardCache := make([]struct {
		Key   string
		Value *object.Obj
	}, 0)
	store.store.All(func(k string, v *object.Obj) bool {
		matches, err := sql.EvaluateWhereClause(whereClause, sql.QueryResultRow{Key: k, Value: *v}, make(map[string]jp.Expr))
		if err != nil || !matches {
			return true
		}

		shardCache = append(shardCache, struct {
			Key   string
			Value *object.Obj
		}{Key: k, Value: v})

		return true
	})
	cacheChannel <- &shardCache
}

func (store *Store) evict(evictCount int) bool {
	store.evictionStrategy.EvictVictims(store, evictCount)
	return true
}
