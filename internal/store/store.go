package store

import (
	"path"
	"strings"
	"sync"

	"github.com/dicedb/dice/internal/server/utils"

	"github.com/cockroachdb/swiss"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/regex"
)

// WatchEvent represents a change in a watched key.
type WatchEvent struct {
	Key       string
	Operation string
	Value     *Obj
}

type DSQLQueryResultRow struct {
	Key   string
	Value Obj
}

type KeyValue struct {
	Key   string
	Value *Obj
}

type Store struct {
	store      *swiss.Map[string, *Obj]
	expires    *swiss.Map[*Obj, uint64] // Does not need to be thread-safe as it is only accessed by a single thread.
	storeMutex sync.RWMutex
	watchChan  chan WatchEvent
}

func NewStore(watchChan chan WatchEvent) *Store {
	return &Store{
		store:     swiss.New[string, *Obj](10240),
		expires:   swiss.New[*Obj, uint64](10240),
		watchChan: watchChan,
	}
}

func ResetStore(store *Store) *Store {
	store.store = swiss.New[string, *Obj](10240)
	store.expires = swiss.New[*Obj, uint64](10240)

	return store
}

func (store *Store) NewObj(value interface{}, expDurationMs int64, oType, oEnc uint8) *Obj {
	obj := &Obj{
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
	withLocks(func() {
		store.store.Clear()
		store.expires.Clear()
		store.watchChan = make(chan WatchEvent, config.KeysLimit)
	}, store, WithStoreLock())
}

type PutOptions struct {
	KeepTTL bool
}

func (store *Store) Put(k string, obj *Obj, opts ...PutOption) {
	withLocks(func() {
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

func (store *Store) PutAll(data map[string]*Obj) {
	withLocks(func() {
		for k, obj := range data {
			store.putHelper(k, obj)
		}
	}, store, WithStoreLock())
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

func (store *Store) getHelper(k string, touch bool) *Obj {
	var v *Obj
	withLocks(func() {
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

func (store *Store) GetAll(keys []string) []*Obj {
	response := make([]*Obj, 0, len(keys))
	withLocks(func() {
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

	withLocks(func() {
		keys = make([]string, 0, store.store.Len())

		store.store.All(func(k string, _ *Obj) bool {
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
	withLocks(func() {
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
			store.notifyWatchers(sourceKey, Del, nil)
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

func (store *Store) Get(k string) *Obj {
	return store.getHelper(k, true)
}

// CloneKeypool creates and returns a deep copy of the store's keypool.
// The keypool is a map that associates string keys with pointers to objects.
// The function locks the keypool for reading to ensure thread-safe access while copying.
//
// Returns:
// - clone: A new map containing the same keys and pointers as the original keypool.
func (store *Store) CloneKeypool() map[string]*string {
	// Lock the keypool for reading to ensure thread safety while copying.
	store.keypoolMutex.RLock()
	defer store.keypoolMutex.RUnlock()

	clone := make(map[string]*string, len(store.keypool))

	for k, v := range store.keypool {
		clone[k] = v
	}

	return clone
}

// scanKeys performs a scan operation on the keypool, retrieving keys that match specified criteria.
// The scan operation is controlled by a cursor and can be limited by a count, pattern, and key type.
// The function takes a snapshot of the keypool to ensure that the scan is performed on a consistent view of the data.
//
// Arguments:
// - cursor: An integer representing the current scan position within the keypool.
// - count: The maximum number of keys to return in this scan operation.
// - pattern: A string representing a glob pattern to filter keys (optional).
// - keyType: A string representing the type of keys to include in the results (optional).
//
// Returns:
// - newCursor: The updated cursor position after the scan, or 0 if the end is reached.
// - keys: A slice of strings containing the keys that match the scan criteria.
func (store *Store) scanKeys(cursor, count int, pattern, keyType string) (newCursor int, keys []string) {
	snapshot := store.CloneKeypool()

	keyList := make([]string, 0, len(snapshot))
	for k := range snapshot {
		keyList = append(keyList, k)
	}

	endCursor := cursor + count
	if endCursor > len(keyList) {
		endCursor = len(keyList)
	}

	for i := cursor; i < endCursor; i++ {
		key := keyList[i]

		objPtr := snapshot[key]

		obj, ok := store.store[*objPtr]
		if !ok || hasExpired(obj, store) {
			continue
		}

		objType := getTypeAsString(obj.TypeEncoding)

		if keyType != "" && !strings.EqualFold(objType, keyType) {
			continue
		}

		if pattern == "" || matchGlob(key, pattern) {
			keys = append(keys, key)
		}
	}

	if endCursor < len(keyList) && len(keys) >= count {
		newCursor = endCursor
	} else {
		newCursor = 0
	}

	return newCursor, keys
}

// getTypeAsString converts the encoding type of an object to a string representation.
// Arguments:
// - encoding: The type encoding of the object.
// Returns:
// - A string representing the type of the object.
func getTypeAsString(encoding uint8) string {
	typePart, _ := ExtractTypeEncoding(&Obj{TypeEncoding: encoding})

	switch typePart {
	case ObjTypeString:
		return constants.String
	case ObjTypeByteList:
		return "byte_list"
	case ObjTypeBitSet:
		return "bitset"
	case ObjTypeJSON:
		return "json"
	case ObjTypeByteArray:
		return "byte_array"
	default:
		return "unknown"
	}
}

// matchGlob checks if the string s matches the glob pattern pattern.
// Arguments:
// - s: The string to be matched.
// - pattern: The glob pattern to match against (can contain wildcards).
// Returns:
// - true if s matches the pattern; false otherwise.
func matchGlob(s, pattern string) bool {
	if pattern == "*" {
		return true // Pattern '*' matches any string
	}
	return strings.HasPrefix(s, strings.TrimSuffix(pattern, "*")) // Match based on prefix
}

func (store *Store) GetDel(k string) *Obj {
	var v *Obj
	withLocks(func() {
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
func (store *Store) SetExpiry(obj *Obj, expDurationMs int64) {
	store.expires.Put(obj, uint64(utils.GetCurrentTime().UnixMilli())+uint64(expDurationMs))
}

// SetUnixTimeExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) SetUnixTimeExpiry(obj *Obj, exUnixTimeSec int64) {
	// convert unix-time-seconds to unix-time-milliseconds
	store.expires.Put(obj, uint64(exUnixTimeSec*1000))
}

func (store *Store) deleteKey(k string, obj *Obj) bool {
	if obj != nil {
		store.store.Delete(k)
		store.expires.Delete(obj)
		KeyspaceStat[0]["keys"]--

		if store.watchChan != nil {
			store.notifyWatchers(k, Del, nil)
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

func (store *Store) notifyWatchers(k, operation string, obj *Obj) {
	store.watchChan <- WatchEvent{k, operation, obj}
}

func (store *Store) GetStore() *swiss.Map[string, *Obj] {
	return store.store
}

func (store *Store) CacheKeysForQuery(keyRegex string, cacheChannel chan *[]KeyValue) {
	shardCache := make([]KeyValue, 0)
	withLocks(func() {
		store.store.All(func(k string, v *Obj) bool {
			if regex.WildCardMatch(keyRegex, k) {
				shardCache = append(shardCache, KeyValue{k, v})
			}
			return true
		})
	}, store, WithStoreRLock())
	cacheChannel <- &shardCache
}
