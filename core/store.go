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

var store map[unsafe.Pointer]*Obj
var expires map[*Obj]uint64
var keypool map[string]unsafe.Pointer
var WatchList sync.Map // Maps queries to the file descriptors of clients that are watching them.

// Channel to receive updates about keys that are being watched.
// The Watcher goroutine will wait on this channel. When a key is updated, the
// goroutine will send the updated value and the related operation to all the
// clients that are watching the key.
var WatchChannel chan WatchEvent

func init() {
	store = make(map[unsafe.Pointer]*Obj)
	expires = make(map[*Obj]uint64)
	keypool = make(map[string]unsafe.Pointer)
	WatchChannel = make(chan WatchEvent, 100)
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

func Put(k string, obj *Obj) {
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

func Get(k string) *Obj {
	ptr, ok := keypool[k]
	if !ok {
		return nil
	}

	v := store[ptr]
	if v != nil {
		if hasExpired(v) {
			Del(k)
			return nil
		}
		v.LastAccessedAt = getCurrentClock()
	}
	return v
}

func Del(k string) bool {
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
