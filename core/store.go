package core

import (
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

type PutOptions struct {
	KeepTTL bool
}

var store map[unsafe.Pointer]*Obj
var expires map[*Obj]uint64 // Does not need to be thread-safe as it is only accessed by a single thread.
var keypool map[string]unsafe.Pointer
var WatchList sync.Map // Maps queries to the file descriptors of clients that are watching them.

var storeMutex sync.RWMutex   // Mutex to protect the store map, must be acquired before keypoolMutex if both are needed.
var keypoolMutex sync.RWMutex // Mutex to protect the keypool map, must be acquired after storeMutex if both are needed.

// Channel to receive updates about keys that are being watched.
// The Watcher goroutine will wait on this channel. When a key is updated, the
// goroutine will send the updated value and the related operation to all the
// clients that are watching the key.
var WatchChannel chan WatchEvent

func init() {
	store = make(map[unsafe.Pointer]*Obj)
	expires = make(map[*Obj]uint64)
	keypool = make(map[string]unsafe.Pointer)
	WatchChannel = make(chan WatchEvent, config.KeysLimit)
}

func setExpiry(obj *Obj, expDurationMs int64) {
	expires[obj] = uint64(time.Now().UnixMilli()) + uint64(expDurationMs)
}

func NewObj(value interface{}, expDurationMs int64, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}
	if expDurationMs > 0 {
		setExpiry(obj, expDurationMs)
	}
	return obj
}

func Put(k string, obj *Obj, options *PutOptions) {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	keypoolMutex.Lock()
	defer keypoolMutex.Unlock()

	if len(store) >= config.KeysLimit {
		evict()
	}
	obj.LastAccessedAt = getCurrentClock()

	ptr, ok := keypool[k]
	if !ok {
		keypool[k] = unsafe.Pointer(&k)
		ptr = unsafe.Pointer(&k)
	}

	currentStoredObject, ok := store[ptr]
	if ok {
		// NOTE: In case there is an value present
		// for a given key 'k', then any updates
		// performed with the 'KEEPTTL' flag need
		// to ensure that we save the new value
		// with the same expiration time as before
		// Without the flag, the expiration time
		// stored earlier will be removed.
		_, ok = expires[currentStoredObject]
		if options != nil && options.KeepTTL && ok {
			expires[obj] = expires[currentStoredObject]
		}
		delete(expires, currentStoredObject)
	}

	store[ptr] = obj
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++

	WatchChannel <- WatchEvent{k, "SET", obj}
}

func Get(k string) *Obj {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	keypoolMutex.RLock()
	defer keypoolMutex.RUnlock()

	ptr, ok := keypool[k]
	if !ok {
		return nil
	}

	v := store[ptr]
	if v != nil {
		if hasExpired(v) {
			storeMutex.RUnlock()
			keypoolMutex.RUnlock()
			Del(k)
			storeMutex.RLock()
			keypoolMutex.RLock()
			return nil
		}
		v.LastAccessedAt = getCurrentClock()
	}
	return v
}

func Del(k string) bool {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	keypoolMutex.Lock()
	defer keypoolMutex.Unlock()

	ptr, ok := keypool[k]
	if !ok {
		return false
	}

	if obj, ok := store[ptr]; ok {
		delete(store, ptr)
		delete(expires, obj)
		delete(keypool, k)
		KeyspaceStat[0]["keys"]--

		WatchChannel <- WatchEvent{k, "DEL", obj}
		return true
	}
	return false
}

func DelByPtr(ptr unsafe.Pointer) bool {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	keypoolMutex.Lock()
	defer keypoolMutex.Unlock()

	if obj, ok := store[ptr]; ok {
		delete(store, ptr)
		delete(expires, obj)
		delete(keypool, *((*string)(ptr)))
		KeyspaceStat[0]["keys"]--
		return true
	}
	return false
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
