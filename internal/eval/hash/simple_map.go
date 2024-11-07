package hash

import (
	"path"

	"github.com/dicedb/dice/internal/server/utils"
)

type SimpleMap[K comparable, V any] struct {
	hmap    map[K]V
	expires map[K]uint64
}

// NewSimpleMap creates a new instance of a simple hash map.
func NewSimpleMap[K comparable, V any]() *SimpleMap[K, V] {
	return &SimpleMap[K, V]{
		hmap:    make(map[K]V),
		expires: make(map[K]uint64),
	}
}

// SetHelper adds or updates a value in the map and returns the old value if it exists.
func (h *SimpleMap[K, V]) SetHelper(key K, value V) (*V, bool) {
	oldVal, ok := h.hmap[key]
	h.hmap[key] = value
	delete(h.expires, key)
	return &oldVal, ok
}

// Set adds or updates a value in the map.
func (h *SimpleMap[K, V]) Set(key K, value V) (*V, bool) {
	old, ok := h.SetHelper(key, value)
	return old, ok
}

// SetAll adds or updates multiple values in the map.
func (h *SimpleMap[K, V]) SetAll(data map[K]V) int64 {
	var cnt int64
	for k, v := range data {
		_, ok := h.SetHelper(k, v)
		if !ok {
			cnt++
		}
	}
	return cnt
}

// Get retrieves the value associated with a key, checking expiration.
// Any key which is past its ttl will be deleted
func (h *SimpleMap[K, V]) Get(key K) (*V, bool) {
	isExpired := h.HasExpired(key)
	if isExpired {
		return nil, false
	}
	value, ok := h.hmap[key]
	if !ok {
		return nil, false
	}
	return &value, ok
}

// Delete removes the entry associated with the specified key from the IHashMap.
// If the key does not exist in the map, the method does nothing.
//
// Parameters:
//
//	key (K): The key of the entry to be removed.
func (h *SimpleMap[K, V]) Delete(key K) {
	delete(h.hmap, key)
	delete(h.expires, key)
}

// Len returns the actual number of valid entries in the HashMap.
// It counts only those entries that have not expired.
func (h *SimpleMap[K, V]) Len() int64 {
	validKeys := h.Keys()
	return int64(len(validKeys))
}

// ALen returns the approximate length of the HashMap.
// This counts all the entries in the map, including potentially expired ones.
func (h *SimpleMap[K, V]) ALen() int64 {
	return int64(len(h.hmap))
}

// Keys returns all valid keys in the map.
// Any key which is past its ttl will be deleted
func (h *SimpleMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(h.hmap))
	for k := range h.hmap {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all valid values in the map.
// Any key which is past its ttl will be deleted
func (h *SimpleMap[K, V]) Values() []V {
	values := make([]V, 0, len(h.hmap))
	for _, v := range h.hmap {
		values = append(values, v)
	}
	return values
}

// Items returns all valid key-value pairs in the map.
// Any key which is past its ttl will be deleted
func (h *SimpleMap[K, V]) Items() [][]interface{} {
	items := make([][]interface{}, 0, len(h.hmap))
	for k, v := range h.hmap {
		if _, ok := h.Get(k); ok {
			items = append(items, []interface{}{k, v})
		}
	}
	return items
}

// Find searches for keys that match the specified pattern p and returns them.
// This function is only available if K is a string.
func (h *SimpleMap[K, V]) Find(pattern K) []K {
	// Ensure that K is a string
	strPattern, ok := any(pattern).(string)
	if !ok {
		panic("Find is only available when K is a string")
	}

	keys := make([]K, 0, h.ALen())
	for k := range h.hmap {
		if _, exists := h.Get(k); exists {
			if found, _ := path.Match(strPattern, any(k).(string)); found {
				keys = append(keys, k)
			}
		}
	}
	return keys
}

// Clear removes all entries from the HashMap, resetting it to an empty state.
func (h *SimpleMap[K, V]) Clear() {
	h.hmap = make(map[K]V)
	h.expires = make(map[K]uint64)
}

// hasExpired checks if the key has expired.
func (h *SimpleMap[K, V]) HasExpired(key K) bool {
	exp, ok := h.expires[key]
	if !ok {
		return false
	}
	hasExpired := exp <= uint64(utils.GetCurrentTime().UnixMilli())
	if hasExpired {
		h.Delete(key)
	}
	return hasExpired
}

// GetWithExpiry returns the value and its expiry in milliseconds
// if the key does not exist, the expiry is -2
// if expire is -1, then key does not expire
// an alternative of Get() then GetExpiry() for one access
func (h *SimpleMap[K, V]) GetExpiry(key K) (value *V, expire int64) {
	val, exist := h.Get(key)
	if !exist {
		return nil, -2
	}
	expireMs, ok := h.expires[key]
	if !ok {
		return val, -1
	}
	return val, int64(expireMs)
}

// SetExpiry sets an expiry time for a given key, specified as a duration in milliseconds from the current time.
//
// If the key does not exist in the HashMap, the function returns the constant `NOT_FOUND`.
// If the specified expiry duration is less than or equal to zero, the key will be deleted, and the function will return the constant `PAST`.
// Otherwise, the expiry time is calculated as the current time plus the provided duration, and the expiry is set using the `SetExpiryUnixMilli` method.
//
// Parameters:
//   - k: The key for which to set the expiry time.
//   - expiryMs: The duration in milliseconds from the current time to set the expiry.
//
// Returns:
//
//	An int64 status code:
//	- `NOT_FOUND` if the key does not exist.
//	- `PAST` if the expiry duration is zero or negative, resulting in key deletion.
//	- `EXPIRY_SET` if the expiry time is successfully set.
func (h *SimpleMap[K, V]) SetExpiry(key K, expiryMs int64) int64 {
	expiryUnixMs := uint64(utils.GetCurrentTime().UnixMilli() + expiryMs)
	return h.SetExpiryUnixMilli(key, expiryUnixMs)
}

// SetExpiryUnixMilli sets an expiry time for a given key, specified in Unix milliseconds.
//
// If the key does not exist in the HashMap, the function returns the constant `NOT_FOUND`.
// If the provided expiry time is in the past (i.e., less than or equal to the current time),
// the key will be deleted, and the function will return the constant `PAST`.
// Otherwise, the expiry time for the key is updated, and the function returns the constant `EXPIRY_SET`.
//
// Parameters:
//   - k: The key for which to set the expiry time.
//   - expiryMs: The expiry time in Unix milliseconds.
//
// Returns:
//
//	An int64 status code:
//	- `NOT_FOUND` if the key does not exist.
//	- `PAST` if the expiry time is in the past, resulting in key deletion.
//	- `EXPIRY_SET` if the expiry time is successfully set.
func (h *SimpleMap[K, V]) SetExpiryUnixMilli(key K, expiryMs uint64) int64 {
	if _, exists := h.Get(key); !exists {
		return NotFound
	}
	if expiryMs <= uint64(utils.GetCurrentTime().UnixMilli()) {
		h.Delete(key)
		return Past
	}
	h.expires[key] = expiryMs
	return ExpirySet
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
func (h *SimpleMap[K, V]) CreateOrModify(key K, modifier func(old V) (V, error)) (*V, error) {
	val, ok := h.Get(key)
	if !ok {
		val = new(V)
	}
	newVal, err := modifier(*val)
	if err != nil {
		return val, err
	}
	h.hmap[key] = newVal
	return &newVal, nil
}

// TODO: Implement Scan
// func (h *SimpleMap[K, V]) Scan(pattern K, count int, offset int) []KV[K, V] {
// 	if _, ok := any(pattern).(string); !ok {
// 		panic("Scan is only available when K is a string")
// 	}
// 	if h.stale {
// 		h.scan_keys = h.Find(pattern)
// 		h.cursor = 0
// 		h.stale = false
// 	}
// 	if h.cursor >= len(h.scan_keys) {
// 		return nil
// 	}
// 	keys := h.scan_keys[h.cursor:]
// 	if len(keys) > count {
// 		keys = keys[:count]
// 	}
// 	h.cursor += len(keys)
// 	return h.Items(keys)
// }
