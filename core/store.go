package core

import (
	"path"
	"sync"
	"time"
	"unsafe"

	"github.com/dicedb/dice/config"
)

type WatchEvent struct {
	Key       string
	Operation string
	Value     *Obj
}

var (
	store     map[unsafe.Pointer]*Obj
	expires   map[*Obj]uint64 // Does not need to be thread-safe as it is only accessed by a single thread.
	keypool   map[string]unsafe.Pointer
	WatchList sync.Map // Maps queries to the file descriptors of clients that are watching them.
)

var (
	storeMutex   sync.RWMutex // Mutex to protect the store map, must be acquired before keypoolMutex if both are needed.
	keypoolMutex sync.RWMutex // Mutex to protect the keypool map, must be acquired after storeMutex if both are needed.
)

// Channel to receive updates about keys that are being watched.
// The Watcher goroutine will wait on this channel. When a key is updated, the
// goroutine will send the updated value and the related operation to all the
// clients that are watching the key.
var WatchChannel chan WatchEvent

func init() {
	ResetStore()
}

func ResetStore() {
	withLocks(func() {
		store = make(map[unsafe.Pointer]*Obj)
		expires = make(map[*Obj]uint64)
		keypool = make(map[string]unsafe.Pointer)
		WatchChannel = make(chan WatchEvent, config.KeysLimit)
	}, WithStoreLock(), WithKeypoolLock())
}

func NewObj(value interface{}, expDurationMs int64, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}
	if expDurationMs >= 0 {
		setExpiry(obj, expDurationMs)
	}
	return obj
}

func Put(k string, obj *Obj) {
	withLocks(func() {
		putHelper(k, obj)
	}, WithStoreLock(), WithKeypoolLock())
}

// PutAll is a bulk insert function that takes a map of
// keys and values and inserts them into the store
func PutAll(data map[string]*Obj) {
	withLocks(func() {
		for k, obj := range data {
			putHelper(k, obj)
		}
	}, WithStoreLock(), WithKeypoolLock())
}

// GetNoTouch is a function to retrieve a value from the store without updating
// the last accessed time of the object.
func GetNoTouch(k string) *Obj {
	return getHelper(k, false)
}

func getHelper(k string, touch bool) *Obj {
	var v *Obj
	withLocks(func() {
		ptr, ok := keypool[k]
		if !ok {
			return
		}

		v = store[ptr]
		if v != nil {
			if hasExpired(v) {
				deleteKey(k, ptr, v)
				v = nil
			} else if touch {
				v.LastAccessedAt = getCurrentClock()
			}
		}
	}, WithStoreRLock(), WithKeypoolRLock())
	return v
}

// GetAll is a bulk retrieve function that takes array of
// keys and retrieves values for keys from the store
func GetAll(keys []string) []*Obj {
	response := make([]*Obj, 0, len(keys))
	withLocks(func() {
		for _, k := range keys {
			ptr, ok := keypool[k]
			if !ok {
				response = append(response, nil)
				continue
			}
			v := store[ptr]
			if v != nil {
				if hasExpired(v) {
					deleteKey(k, ptr, v)
					response = append(response, nil)
				} else {
					v.LastAccessedAt = getCurrentClock()
					response = append(response, v)
				}
			} else {
				response = append(response, nil)
			}
		}
	}, WithStoreRLock(), WithKeypoolRLock())
	return response
}

func Del(k string) bool {
	return withLocksReturn(func() bool {
		ptr, ok := keypool[k]
		if !ok {
			return false
		}
		return deleteKey(k, ptr, store[ptr])
	}, WithStoreLock(), WithKeypoolLock())
}

func DelByPtr(ptr unsafe.Pointer) bool {
	return withLocksReturn(func() bool {
		return delByPtr(ptr)
	}, WithStoreLock(), WithKeypoolLock())
}

// List all keys in the store by given pattern
func Keys(p string) ([]string, error) {
	var keys []string
	var err error

	withLocks(func() {
		keys = make([]string, 0, len(keypool))
		for k := range keypool {
			if found, e := path.Match(p, k); e != nil {
				err = e
				return
			} else if found {
				keys = append(keys, k)
			}
		}
	}, WithStoreRLock(), WithKeypoolRLock())

	return keys, err
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
		if countClients(clients.(*sync.Map)) == 0 {
			WatchList.Delete(query)
		}
	}
}

// Rename function to implement RENAME functionality using existing helpers
func Rename(sourceKey string, destKey string) bool {
	return withLocksReturn(func() bool {
		// If source and destination are the same, do nothing and return true
		if sourceKey == destKey {
			return true
		}

		sourcePtr, sourceOk := keypool[sourceKey]
		if !sourceOk {
			return false
		}

		sourceObj := store[sourcePtr]
		if sourceObj == nil || hasExpired(sourceObj) {
			if sourceObj != nil {
				deleteKey(sourceKey, sourcePtr, sourceObj)
			}
			return false
		}

		// Use putHelper to handle putting the object at the destination key
		putHelper(destKey, sourceObj)

		// Remove the source key
		delete(store, sourcePtr)
		delete(keypool, sourceKey)
		if KeyspaceStat[0] != nil {
			KeyspaceStat[0]["keys"]--
		}

		// Notify watchers about the deletion of the source key
		notifyWatchers(sourceKey, "DEL", sourceObj)

		return true
	}, WithStoreLock(), WithKeypoolLock())
}

// Helper functions

// putHelper is a helper function to insert a key-value pair into the store.
// It also increments the key count in the KeyspaceStat map and notifies watchers.
// This method is not thread-safe. It should be called within a lock.
func putHelper(k string, obj *Obj) {
	if len(store) >= config.KeysLimit {
		evict()
	}
	obj.LastAccessedAt = getCurrentClock()

	ptr := ensureKeyInPool(k)
	store[ptr] = obj

	incrementKeyCount()
	notifyWatchers(k, "SET", obj)
}

func Get(k string) *Obj {
	return getHelper(k, true)
}

func GetDel(k string) *Obj {
	var v *Obj
	withLocks(func() {
		ptr, ok := keypool[k]
		if !ok {
			return
		}

		v = store[ptr]
		if v != nil {
			expired := hasExpired(v)
			deleteKey(k, ptr, v)
			if expired {
				v = nil
			}
		}
	}, WithStoreLock(), WithKeypoolLock())
	return v
}

// setExpiry sets the expiry time for an object.
// This method is not thread-safe. It should be called within a lock.
func setExpiry(obj *Obj, expDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(expDurationMs)
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

// This method is not thread-safe. It should be called within a lock.
func ensureKeyInPool(k string) unsafe.Pointer {
	ptr, ok := keypool[k]
	if !ok {
		ptr = unsafe.Pointer(&k)
		keypool[k] = ptr
	}
	return ptr
}

// incrementKeyCount increments the key count in the KeyspaceStat map.
// This method is not thread-safe. It should be called within a lock.
func incrementKeyCount() {
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++
}

// notifyWatchers sends a WatchEvent to the WatchChannel.
// This function is called whenever a key is updated.
func notifyWatchers(k string, operation string, obj *Obj) {
	WatchChannel <- WatchEvent{k, operation, obj}
}

// deleteKey deletes a key from the store, keypool, and expires maps. it also
// decrements the key count in the KeyspaceStat map and notifies watchers about
// the deletion of the key.
// This method is not thread-safe. It should be called within a lock.
func deleteKey(k string, ptr unsafe.Pointer, obj *Obj) bool {
	if obj != nil {
		delete(store, ptr)
		delete(expires, obj)
		delete(keypool, k)
		KeyspaceStat[0]["keys"]--
		notifyWatchers(k, "DEL", obj)
		return true
	}
	return false
}

// delByPtr deletes a key from the store, keypool, and expires maps using a pointer.
// This method is not thread-safe. It should be called within a lock.
func delByPtr(ptr unsafe.Pointer) bool {
	if obj, ok := store[ptr]; ok {
		key := *((*string)(ptr))
		return deleteKey(key, ptr, obj)
	}
	return false
}
