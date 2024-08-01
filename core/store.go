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

type Store struct {
	store        map[unsafe.Pointer]*Obj
	expires      map[*Obj]uint64 // Does not need to be thread-safe as it is only accessed by a single thread.
	keypool      map[string]unsafe.Pointer
	storeMutex   sync.RWMutex // Mutex to protect the store map, must be acquired before keypoolMutex if both are needed.
	keypoolMutex sync.RWMutex // Mutex to protect the keypool map, must be acquired after storeMutex if both are needed.
}

var WatchList sync.Map // Maps queries to the file descriptors of clients that are watching them.

// Channel to receive updates about keys that are being watched.
// The Watcher goroutine will wait on this channel. When a key is updated, the
// goroutine will send the updated value and the related operation to all the
// clients that are watching the key.
var WatchChannel chan WatchEvent

func init() {
	WatchChannel = make(chan WatchEvent, 100)
}

func NewStore() *Store {
	return &Store{
		store:   make(map[unsafe.Pointer]*Obj),
		expires: make(map[*Obj]uint64),
		keypool: make(map[string]unsafe.Pointer),
	}
}

func (s *Store) setExpiry(obj *Obj, expDurationMs int64) {
	s.expires[obj] = uint64(time.Now().UnixMilli()) + uint64(expDurationMs)
}

func NewObj(value interface{}, oType uint8, oEnc uint8) *Obj {
	obj := &Obj{
		Value:          value,
		TypeEncoding:   oType | oEnc,
		LastAccessedAt: getCurrentClock(),
	}

	return obj
}

func (s *Store) Put(k string, obj *Obj, expDurationMs int64) {
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()
	s.keypoolMutex.Lock()
	defer s.keypoolMutex.Unlock()

	if len(s.store) >= config.KeysLimit {
		evict()
	}
	obj.LastAccessedAt = getCurrentClock()

	ptr, ok := s.keypool[k]
	if !ok {
		s.keypool[k] = unsafe.Pointer(&k)
		ptr = unsafe.Pointer(&k)
	}

	s.store[ptr] = obj

	if expDurationMs > 0 {
		s.setExpiry(obj, expDurationMs)
	}

	if KeyspaceStat[0] == nil {
		KeyspaceStat[0] = make(map[string]int)
	}
	KeyspaceStat[0]["keys"]++

	WatchChannel <- WatchEvent{k, "SET", obj}
}

func (s *Store) Get(k string) *Obj {
	s.storeMutex.RLock()
	defer s.storeMutex.RUnlock()
	s.keypoolMutex.RLock()
	defer s.keypoolMutex.RUnlock()

	ptr, ok := s.keypool[k]
	if !ok {
		return nil
	}

	v := s.store[ptr]
	if v != nil {
		if hasExpired(v) {
			s.storeMutex.RUnlock()
			s.Del(k)
			s.storeMutex.RLock()
			return nil
		}
		v.LastAccessedAt = getCurrentClock()
	}
	return v
}

func (s *Store) Del(k string) bool {
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()
	s.keypoolMutex.Lock()
	defer s.keypoolMutex.Unlock()

	ptr, ok := s.keypool[k]
	if !ok {
		return false
	}

	if obj, ok := s.store[ptr]; ok {
		delete(s.store, ptr)
		delete(s.expires, obj)
		delete(s.keypool, k)
		KeyspaceStat[0]["keys"]--

		WatchChannel <- WatchEvent{k, "DEL", obj}
		return true
	}
	return false
}

func (s *Store) DelByPtr(ptr unsafe.Pointer) bool {
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()
	s.keypoolMutex.Lock()
	defer s.keypoolMutex.Unlock()

	if obj, ok := s.store[ptr]; ok {
		delete(s.store, ptr)
		delete(s.expires, obj)
		delete(s.keypool, *((*string)(ptr)))
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
