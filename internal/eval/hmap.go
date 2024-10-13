package eval

import (
	"time"
)

// HashMap is a generic implementation of a hashmap with expiration support.
// hmap stores the actual values
// expires stores the expiry values
type HashMap[K comparable, V any] struct {
	hmap    map[K]V
	expires map[K]uint64
}

// NewHashMap creates a new instance of HashMap.
func NewHashMap[K comparable, V any]() *HashMap[K, V] {
	return &HashMap[K, V]{
		hmap:    make(map[K]V),
		expires: make(map[K]uint64),
	}
}

// Keys returns all valid keys in the map.
// Any key which is past its ttl will be deleted
func (h *HashMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(h.hmap))
	for k := range h.hmap {
		if _, ok := h.Get(k); ok {
			keys = append(keys, k)
		}
	}
	return keys
}

// Values returns all valid values in the map.
// Any key which is past its ttl will be deleted
func (h *HashMap[K, V]) Values() []V {
	values := make([]V, 0, len(h.hmap))
	for k := range h.hmap {
		if val, ok := h.Get(k); ok {
			values = append(values, val)
		}
	}
	return values
}

// Items returns all valid key-value pairs in the map.
// Any key which is past its ttl will be deleted
func (h *HashMap[K, V]) Items() [][]interface{} {
	items := make([][]interface{}, 0, len(h.hmap))
	for k, v := range h.hmap {
		if _, ok := h.Get(k); ok {
			items = append(items, []interface{}{k, v})
		}
	}
	return items
}

// Clear removes all entries from the HashMap, resetting it to an empty state.
func (h *HashMap[K, V]) Clear() {
	h.hmap = make(map[K]V)
	h.expires = make(map[K]uint64)
}

// ApproxLength returns the approximate length of the HashMap.
// This counts all the entries in the map, including potentially expired ones.
// This method can be used for creating slices using make()
func (h *HashMap[K, V]) ApproxLength() int {
	return len(h.hmap)
}

// ActualValidLength returns the actual number of valid entries in the HashMap.
// It counts only those entries that have not expired.
func (h *HashMap[K, V]) Len() int {
	validKeys := h.Keys()
	return len(validKeys)
}

// hasExpired checks if the key has expired.
func (h *HashMap[K, V]) hasExpired(k K) bool {
	exp, ok := h.expires[k]
	if !ok {
		return false
	}
	return exp <= uint64(time.Now().UnixMilli())
}

// GetExpiry returns the expiry time for a given key\
// with an expiry set
func (h *HashMap[K, V]) GetExpiry(k K) (uint64, bool) {
	expiryMs, ok := h.expires[k]
	return expiryMs, ok
}

// SetExpiry sets an expiry time for a given key.
func (h *HashMap[K, V]) SetExpiry(k K, expiryMs uint64) {
	h.expires[k] = expiryMs
}

// Get retrieves the value associated with a key, checking expiration.
// Any key which is past its ttl will be deleted
func (h *HashMap[K, V]) Get(k K) (V, bool) {
	value, ok := h.hmap[k]
	if ok && h.hasExpired(k) {
		h.Delete(k)
		return *new(V), false
	}
	return value, ok
}

// Delete removes a key and its expiration from the map.
func (h *HashMap[K, V]) Delete(k K) bool {
	_, ok := h.hmap[k]
	if ok {
		delete(h.hmap, k)
		delete(h.expires, k)
	}
	return ok
}

// SetHelper sets a value in the map and removes expiration.
func (h *HashMap[K, V]) SetHelper(k K, v V) (V, bool) {
	oldVal, ok := h.hmap[k]
	h.hmap[k] = v
	delete(h.expires, k)
	return oldVal, ok
}

// Set adds or updates a value in the map.
func (h *HashMap[K, V]) Set(k K, v V) (V, bool) {
	return h.SetHelper(k, v)
}

// SetAll adds or updates multiple values in the map.
func (h *HashMap[K, V]) SetAll(data map[K]V) int64 {
	var cnt int64
	for k, v := range data {
		_, ok := h.hmap[k]
		if !ok {
			cnt++
		}
		h.hmap[k] = v
	}
	return cnt
}

// ModifyOrCreate modifies the value associated with the given key using the provided modifier function.
// If the key does not exist, a new value will be created based on the modifier's logic.
// The modifier function takes the current value of type V and returns a modified value and an optional error.
//
// Example 1: Incrementing an Integer Value
//
//	hMap := NewHashMap[string, int]()
//	modifier := func(currentValue int) (int, error) {
//	    return currentValue + 1, nil // Increment the current value by 1
//	}
//	err := hMap.ModifyOrCreate("counter", modifier)
//	if err != nil {
//	    fmt.Println("Error modifying value:", err)
//	}
//
// Example 2: Updating a Field in a Struct
//
//	type User struct {
//	    Name  string
//	    Age   int
//	    Email string
//	}
//
//	hMap := NewHashMap[string, User]()
//	modifier := func(currentValue User) (User, error) {
//	    currentValue.Age += 1 // Increment the user's age by 1
//	    return currentValue, nil
//	}
//	err := hMap.ModifyOrCreate("john_doe", modifier)
//	if err != nil {
//	    fmt.Println("Error modifying value:", err)
//	}
//
// In this example, if the key "john_doe" exists, the "Age" field of the User struct is incremented by 1.
// If the key does not exist, the modifier will use the zero value for the User struct, where all fields are initialized
// to their zero values (empty string for Name and Email, 0 for Age), and increment the Age to 1.
//
// The ModifyOrCreate method allows for flexible modifications of complex data types based on the provided logic.
func (h *HashMap[K, V]) CreateOrModify(k K, modifier func(V) (V, error)) (V, error) {
	val, ok := h.hmap[k]
	if !ok {
		val = *new(V)
	}
	newVal, err := modifier(val)
	if err != nil {
		return val, err
	}
	h.hmap[k] = newVal
	return newVal, nil
}
