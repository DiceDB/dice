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
	storeMutex.Lock()
	defer storeMutex.Unlock()
	keypoolMutex.Lock()
	defer keypoolMutex.Unlock()

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
	if expDurationMs >= 0 {
		setExpiry(obj, expDurationMs)
	}
	return obj
}

func Put(k string, obj *Obj) {
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

	store[ptr] = obj
	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++

	WatchChannel <- WatchEvent{k, "SET", obj}
}

// PutAll is a bulk insert function that takes a map of
// keys and values and inserts them into the store
func PutAll(data map[string]*Obj) {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	keypoolMutex.Lock()
	defer keypoolMutex.Unlock()

	for k, obj := range data {
		if len(store) >= config.KeysLimit {
			evict()
		}
		obj.LastAccessedAt = getCurrentClock()

		ptr, ok := keypool[k]
		if !ok {
			// we need to create a new string for each key, ensuring that each key in
			// the keypool map has its own unique memory address. Reusing the same
			// memory address (&k) for multiple keys will cause the keypool map to
			// have incorrect entries, because the address of k remains the same
			// throughout the loop iterations, but its value changes.
			// As a result, all entries in the keypool map end up pointing to the same
			// memory location, which contains the last value of k after the loop
			// finishes.
			keyCopy := string([]byte(k))
			keypool[k] = unsafe.Pointer(&keyCopy)
			ptr = unsafe.Pointer(&keyCopy)
		}

		store[ptr] = obj
		if KeyspaceStat[0] == nil {
			KeyspaceStat[0] = make(map[string]int)
		}
		KeyspaceStat[0]["keys"]++

		WatchChannel <- WatchEvent{k, "SET", obj}
	}
}

func Get(k string) *Obj {
	storeMutex.Lock()
	defer storeMutex.Unlock()
	keypoolMutex.Lock()
	defer keypoolMutex.Unlock()

	ptr, ok := keypool[k]
	if !ok {
		return nil
	}

	v := store[ptr]
	if v != nil {
		if hasExpired(v) {
			delete(store, ptr)
			delete(expires, v)
			delete(keypool, k)
			KeyspaceStat[0]["keys"]--
			WatchChannel <- WatchEvent{k, "DEL", v}

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

		key := *((*string)(ptr))
		WatchChannel <- WatchEvent{key, "DEL", obj}
		return true
	}
	return false
}

// List all keys in the store by given pattern
func Keys(p string) ([]string, error) {
	storeMutex.RLock()
	defer storeMutex.RUnlock()
	keypoolMutex.RLock()
	defer keypoolMutex.RUnlock()

	keyCap := 0
	if p == "*" {
		keyCap = len(keypool)
	}

	keys := make([]string, 0, keyCap)
	for k := range keypool {
		if found, err := path.Match(p, k); err != nil {
			return nil, err
		} else if found {
			keys = append(keys, k)
		}
	}

	return keys, nil
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

func GetNoTouch(k string) *Obj {
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
			return nil
		}
	}
	return v
}
