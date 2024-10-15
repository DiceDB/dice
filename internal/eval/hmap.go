package eval

import (
	"path"

	"github.com/dicedb/dice/internal/server/utils"
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

// Find searches for keys that match the specified pattern p and returns them.
// This function is only available if K is a string.
func (h *HashMap[K, V]) Find(p string) []K {
	var stringKey K
	if _, ok := any(stringKey).(string); !ok {
		panic("Find is only available when K is a string")
	}

	keys := make([]K, 0, h.ApproxLength())
	for k := range h.hmap {
		if _, ok := h.Get(k); ok {
			if found, _ := path.Match(p, any(k).(string)); found {
				keys = append(keys, k)
			}
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
			values = append(values, *val)
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

// Len returns the actual number of valid entries in the HashMap.
// It counts only those entries that have not expired.
func (h *HashMap[K, V]) Len() int {
	validKeys := h.Keys()
	return len(validKeys)
}

// hasExpired checks if the key has expired.
func (h *HashMap[K, V]) HasExpired(k K) bool {
	exp, ok := h.expires[k]
	if !ok {
		return false
	}
	hasExpired := exp <= uint64(utils.GetCurrentTime().UnixMilli())
	if hasExpired {
		h.Delete(k)
	}
	return hasExpired
}

// GetExpiry returns the expiry time for a given key
// with an expiry set
// checks for expiry to invalidate the key if expired
func (h *HashMap[K, V]) GetExpiry(k K) (uint64, bool) {
	h.HasExpired(k)
	expiryMs, ok := h.expires[k]
	return expiryMs, ok
}

// GetWithExpiry returns the value and its expiry in milliseconds
// if the key does not exist, the expiry is -2
// if expire is -1, then key does not expire
// an alternative of Get() then GetExpiry() for one access
func (h *HashMap[K, V]) GetWithExpiry(k K) (value *V, expiry int64) {
	val, exist := h.Get(k)
	if !exist {
		return nil, -2
	}
	expireMs, ok := h.expires[k]
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
func (h *HashMap[K, V]) SetExpiry(k K, expiryMs int64) int64 {
	expiryUnixMs := uint64(utils.GetCurrentTime().UnixMilli() + expiryMs)
	return h.SetExpiryUnixMilli(k, expiryUnixMs)
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
func (h *HashMap[K, V]) SetExpiryUnixMilli(k K, expiryMs uint64) int64 {
	if _, exists := h.Get(k); !exists {
		return NOT_FOUND
	}
	if expiryMs <= uint64(utils.GetCurrentTime().UnixMilli()) {
		h.Delete(k)
		return PAST
	}
	h.expires[k] = expiryMs
	return EXPIRY_SET
}

// Get retrieves the value associated with a key, checking expiration.
// Any key which is past its ttl will be deleted
func (h *HashMap[K, V]) Get(k K) (*V, bool) {
	isExpired := h.HasExpired(k)
	if isExpired {
		return nil, false
	}
	value, ok := h.hmap[k]
	if !ok {
		return nil, false
	}
	return &value, ok
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
func (h *HashMap[K, V]) SetHelper(k K, v V) (*V, bool) {
	oldVal, ok := h.hmap[k]
	h.hmap[k] = v
	delete(h.expires, k)
	return &oldVal, ok
}

// Set adds or updates a value in the map.
func (h *HashMap[K, V]) Set(k K, v V) (*V, bool) {
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
func (h *HashMap[K, V]) CreateOrModify(k K, modifier func(V) (V, error)) (*V, error) {
	val, ok := h.Get(k)
	if !ok {
		val = new(V)
	}
	newVal, err := modifier(*val)
	if err != nil {
		return val, err
	}
	h.hmap[k] = newVal
	return &newVal, nil
}
