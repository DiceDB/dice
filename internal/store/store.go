package store

import (
	"path"

	"github.com/ohler55/ojg/jp"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/sql"
	"github.com/xwb1989/sqlparser"

	"github.com/dicedb/dice/internal/server/utils"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/config"
)

// QueryWatchEvent represents a change in a watched key.
type QueryWatchEvent struct {
	Key       string
	Operation string
	Value     object.Obj
}

type Store struct {
	store     *swiss.Map[string, *object.Obj]
	expires   *swiss.Map[*object.Obj, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
	numKeys   int
	watchChan chan QueryWatchEvent
}

func NewStore(watchChan chan QueryWatchEvent) *Store {
	return &Store{
		store:     swiss.New[string, *object.Obj](config.DiceConfig.Server.StoreMapInitSize),
		expires:   swiss.New[*object.Obj, uint64](config.DiceConfig.Server.StoreMapInitSize),
		watchChan: watchChan,
	}
}

func ResetStore(store *Store) *Store {
	store.numKeys = 0
	store.store = swiss.New[string, *object.Obj](config.DiceConfig.Server.StoreMapInitSize)
	store.expires = swiss.New[*object.Obj, uint64](config.DiceConfig.Server.StoreMapInitSize)

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
	store.store.Clear()
	store.expires.Clear()
}

type PutOptions struct {
	KeepTTL bool
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
	for k, obj := range data {
		store.putHelper(k, obj)
	}
}

func (store *Store) GetNoTouch(k string) *object.Obj {
	return store.getHelper(k, false)
}

func (store *Store) putHelper(k string, obj *object.Obj, opts ...PutOption) {
	options := getDefaultOptions()

	for _, optApplier := range opts {
		optApplier(options)
	}

	if store.store.Len() >= config.DiceConfig.Server.KeysLimit {
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
	} else {
		store.numKeys++
	}
	store.store.Put(k, obj)

	if store.watchChan != nil {
		store.notifyQueryManager(k, Set, *obj)
	}
}

// getHelper is a helper function to get the object from the store. It also updates the last accessed time if touch is true.
func (store *Store) getHelper(k string, touch bool) *object.Obj {
	var v *object.Obj
	v, _ = store.store.Get(k)
	if v != nil {
		if hasExpired(v, store) {
			store.deleteKey(k, v)
			v = nil
		} else if touch {
			v.LastAccessedAt = UpdateLastAccessedAt(v.LastAccessedAt)
		}
	}
	return v
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
				v.LastAccessedAt = UpdateLastAccessedAt(v.LastAccessedAt)
				response = append(response, v)
			}
		} else {
			response = append(response, nil)
		}
	}
	return response
}

func (store *Store) Del(k string) bool {
	v, ok := store.store.Get(k)
	if ok {
		return store.deleteKey(k, v)
	}
	return false
}

func (store *Store) DelByPtr(ptr string) bool {
	return store.delByPtr(ptr)
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
			store.deleteKey(sourceKey, sourceObj)
		}
		return false
	}

	// Use putHelper to handle putting the object at the destination key
	store.putHelper(destKey, sourceObj)

	// Remove the source key
	store.store.Delete(sourceKey)
	store.numKeys--

	// Notify watchers about the deletion of the source key
	if store.watchChan != nil {
		store.notifyQueryManager(sourceKey, Del, *sourceObj)
	}

	return true
}

func (store *Store) Get(k string) *object.Obj {
	return store.getHelper(k, true)
}

func (store *Store) GetDel(k string) *object.Obj {
	var v *object.Obj
	v, _ = store.store.Get(k)
	if v != nil {
		expired := hasExpired(v, store)
		store.deleteKey(k, v)
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

func (store *Store) deleteKey(k string, obj *object.Obj) bool {
	if obj != nil {
		store.store.Delete(k)
		store.expires.Delete(obj)
		store.numKeys--

		if store.watchChan != nil {
			store.notifyQueryManager(k, Del, *obj)
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

// notifyQueryManager notifies the query manager about a key change, so that it can update the query cache if needed.
func (store *Store) notifyQueryManager(k, operation string, obj object.Obj) {
	store.watchChan <- QueryWatchEvent{k, operation, obj}
}

func (store *Store) GetStore() *swiss.Map[string, *object.Obj] {
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
