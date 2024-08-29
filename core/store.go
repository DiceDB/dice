package core

import (
	"path"
	"strings"
	"sync"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/server/utils"
)

type WatchEvent struct {
	Key       string
	Operation string
	Value     *Obj
}

type Store struct {
	store        map[string]*Obj
	expires      map[*Obj]uint64 // Does not need to be thread-safe as it is only accessed by a single thread.
	keypool      map[string]*string
	storeMutex   sync.RWMutex
	keypoolMutex sync.RWMutex
	WatchList    sync.Map // Maps queries to the file descriptors of clients that are watching them.

}

// WatchChannel Channel to receive updates about keys that are being watched.
var WatchChannel chan WatchEvent

func NewStore() *Store {
	WatchChannel = make(chan WatchEvent, config.KeysLimit)
	return &Store{
		store:   make(map[string]*Obj),
		expires: make(map[*Obj]uint64),
		keypool: make(map[string]*string),
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
		store.store = make(map[string]*Obj)
		store.expires = make(map[*Obj]uint64)
		store.keypool = make(map[string]*string)
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
	currentObject, ok := store.store[*ptr]
	if ok {
		if options.KeepTTL && store.expires[currentObject] > 0 {
			store.expires[obj] = store.expires[currentObject]
		}
		delete(store.expires, currentObject)
	}
	store.store[*ptr] = obj

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

		v = store.store[*ptr]
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
			ptr, ok := store.keypool[k]
			if !ok {
				response = append(response, nil)
				continue
			}
			v := store.store[*ptr]
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
		ptr, ok := store.keypool[k]
		if !ok {
			return false
		}
		return store.deleteKey(k, *ptr, store.store[*ptr])
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
func (store *Store) AddWatcher(query DSQLQuery, clientFd int) { //nolint:gocritic
	clients, _ := store.WatchList.LoadOrStore(query, &sync.Map{})
	clients.(*sync.Map).Store(clientFd, struct{}{})
}

// Function to remove a watcher from a query.
func (store *Store) RemoveWatcher(query DSQLQuery, clientFd int) { //nolint:gocritic
	if clients, ok := store.WatchList.Load(query); ok {
		clients.(*sync.Map).Delete(clientFd)
		// If no more clients for this query, remove the query from WatchList
		if countClients(clients.(*sync.Map)) == 0 {
			store.WatchList.Delete(query)
		}
	}
}

// Rename function to implement RENAME functionality using existing helpers

func (store *Store) Rename(sourceKey, destKey string) bool {
	return withLocksReturn(func() bool {
		// If source and destination are the same, do nothing and return true
		if sourceKey == destKey {
			return true
		}

		sourcePtr, sourceOk := store.keypool[sourceKey]
		if !sourceOk {
			return false
		}

		sourceObj := store.store[*sourcePtr]
		if sourceObj == nil || hasExpired(sourceObj, store) {
			if sourceObj != nil {
				store.deleteKey(sourceKey, *sourcePtr, sourceObj)
			}
			return false
		}

		// Use putHelper to handle putting the object at the destination key
		store.putHelper(destKey, sourceObj)

		// Remove the source key
		delete(store.store, *sourcePtr)
		delete(store.keypool, sourceKey)
		if KeyspaceStat[0] != nil {
			KeyspaceStat[0]["keys"]--
		}

		// Notify watchers about the deletion of the source key
		notifyWatchers(sourceKey, "DEL", sourceObj)

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
		return "string"
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
		ptr, ok := store.keypool[k]
		if !ok {
			return
		}

		v = store.store[*ptr]
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
	store.expires[obj] = uint64(utils.GetCurrentTime().UnixMilli()) + uint64(expDurationMs)
}

// setUnixTimeExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func (store *Store) setUnixTimeExpiry(obj *Obj, exUnixTimeSec int64) {
	// convert unix-time-seconds to unix-time-milliseconds
	store.expires[obj] = uint64(exUnixTimeSec * 1000)
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

func (store *Store) ensureKeyInPool(k string) *string {
	ptr, ok := store.keypool[k]
	if !ok {
		ptr = &k
		store.keypool[k] = ptr
	}
	return ptr
}

func (store *Store) deleteKey(k, ptr string, obj *Obj) bool {
	if obj != nil {
		delete(store.store, ptr)
		delete(store.expires, obj)
		delete(store.keypool, k)
		KeyspaceStat[0]["keys"]--
		notifyWatchers(k, "DEL", obj)
		return true
	}
	return false
}

func (store *Store) delByPtr(ptr string) bool {
	if obj, ok := store.store[ptr]; ok {
		key := ptr
		return store.deleteKey(key, ptr, obj)
	}
	return false
}

func notifyWatchers(k, operation string, obj *Obj) {
	WatchChannel <- WatchEvent{k, operation, obj}
}
