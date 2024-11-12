package eval

import (
	"github.com/dicedb/dice/internal/server/utils"
)

type Hash interface {
	// Set inserts a value into the hash map without expiry.
	// It returns a boolean indicating whether the field already existed.
	// If the field already exists, the value is updated.
	//
	// Parameters:
	//   - field: The key to insert into the hash map
	//   - value: The value to associate with the field
	//
	// Returns:
	//   - bool: true if the field already existed, false otherwise
	Set(field, value string) bool

	// SetExpiry sets an expiry time for a specific value in milliseconds.
	//	If the field does not exist, this function does nothing.
	//
	// Parameters:
	//   - field: The key to set the expiry for
	//   - expiryMs: The expiry time in milliseconds
	SetExpiry(field string, expiryMs int64)

	// SetWithExpiry inserts a value into the hash map with an expiry in milliseconds.
	// It returns a boolean indicating whether the field already existed.
	// If the field already exists, the value is updated.
	// If the expiry is less than or equal to 0, the field is deleted.
	//
	// Parameters:
	//   - field: The key to insert into the hash map
	//   - value: The value to associate with the field
	//   - expiryMs: The expiry time in milliseconds
	//
	// Returns:
	//   - bool: true if the field already existed, false otherwise
	SetWithExpiry(field, value string, expiryMs int64) bool

	// GetWithExpiry retrieves a value from the hash map with expiry information.
	// It returns the value as a HashItem and a boolean indicating whether the key exists.
	// If the key does not exist, the HashItem will be empty.
	// If the key exists but has expired, it will be deleted and the HashItem will be empty.
	//
	// Parameters:
	//   - field: The key to look up in the hash map
	//
	// Returns:
	//   - HashItem: The value associated with the field and its expiry information
	//               HashItem contains the following fields:
	//                 - Value: The value associated with the field
	//                 - Expiry: The expiry time in milliseconds
	//                 - Persist: A boolean indicating whether the value is persistent
	//   - bool: true if the field exists, false otherwise
	GetWithExpiry(field string) (*HashItem, bool)

	// Get retrieves a value from the hash map given a field key.
	// It returns the value as a string and a boolean indicating whether the key exists.
	// This is a simplified version of GetWithExpiry that only returns the value,
	// ignoring any expiry information.
	//
	// Parameters:
	//   - field: The key to look up in the hash map
	//
	// Returns:
	//   - string: The value associated with the field
	//   - bool: true if the field exists, false otherwise
	Get(field string) (*string, bool)

	// Has checks if a value exists in the hash map and is not expired.
	//
	// Parameters:
	//   - field: The key to look up in the hash map
	//
	// Returns:
	//   - bool: true if the field exists and is not expired, false otherwise
	Has(field string) bool

	// Delete removes a value from the hash map.
	// It returns a boolean indicating whether the field existed.
	//	If the field does not exist, this function does nothing.
	//
	// Parameters:
	//   - field: The key to remove from the hash map
	//
	// Returns:
	//   - bool: true if the field existed, false otherwise
	Delete(field string) bool

	// Clear removes all entries from the hash map, resetting it to an empty state.
	Clear()

	// Len returns the total number of elements in the hash map.
	//
	// Returns:
	//   - int: The number of elements in the hash map
	Len() int

	//	Keys returns all keys in the hash map.
	//
	//	Returns:
	//		- []string: A slice containing all keys in the hash map
	Keys() []string

	//	Values returns all values in the hash map.
	//
	//	Returns:
	//		- []string: A slice containing all values in the hash map
	Values() []string

	//	Items returns all valid (non-expired) elements in the hash map.
	//
	//	Returns:
	//		- [][]string: A slice containing all valid elements in the hash map
	Items() [][]string
}

type HashItem struct {
	Value   string
	Expiry  uint64
	Persist bool
}

func (i *HashItem) IsPersistent() bool {
	return i.Persist
}

// HashImpl implements the Hash interface
type HashImpl struct {
	data map[string]HashItem
}

// NewHash creates a new instance of Hash.
func NewHash() Hash {
	return &HashImpl{
		data: make(map[string]HashItem),
	}
}

// Set inserts a value into the hash without expiry.
func (h *HashImpl) Set(field, value string) bool {
	existed := h.Has(field)
	h.data[field] = HashItem{Value: value, Persist: true}
	return existed
}

// SetWithExpiry inserts a value into the hash with an expiry in milliseconds.
func (h *HashImpl) SetWithExpiry(field, value string, expiryMs int64) bool {
	existed := h.Has(field)
	if expiryMs > 0 {
		expiryUnixMs := uint64(utils.GetCurrentTime().UnixMilli() + expiryMs)
		h.data[field] = HashItem{Value: value, Expiry: expiryUnixMs, Persist: false}
	} else {
		h.Delete(field)
	}

	return existed
}

// GetWithExpiry retrieves a value from the hash with expiry information.
func (h *HashImpl) GetWithExpiry(field string) (*HashItem, bool) {
	item, exists := h.data[field]
	if !exists {
		return &HashItem{}, false
	}

	// If the item has no expiry, return it
	if item.IsPersistent() {
		return &item, true
	}
	// If the item has expired, delete it
	if item.Expiry <= uint64(utils.GetCurrentTime().UnixMilli()) {
		delete(h.data, field)
		return &HashItem{}, false
	}
	// If the item has not expired, return it
	return &item, true
}

func (h *HashImpl) Get(field string) (*string, bool) {
	item, exists := h.GetWithExpiry(field)
	return &item.Value, exists
}

// Has checks if a value exists in the hash and is not expired.
func (h *HashImpl) Has(field string) bool {
	_, exists := h.GetWithExpiry(field)
	return exists
}

// Delete removes a value from the hash.
func (h *HashImpl) Delete(field string) bool {
	exists := h.Has(field)
	if exists {
		delete(h.data, field)
	}
	return exists
}

// Clear removes all entries from the hash, resetting it to an empty state.
func (h *HashImpl) Clear() {
	h.data = make(map[string]HashItem)
}

// SetExpiry sets an expiry time for a specific value in milliseconds.
func (h *HashImpl) SetExpiry(field string, expiryMs int64) {
	item, exists := h.data[field]
	if !exists {
		return
	}
	h.SetWithExpiry(field, item.Value, expiryMs)
}

// Len returns the total number of elements in the hash.
func (h *HashImpl) Len() int {
	return len(h.data)
}

// Keys returns all keys in the hash.
func (h *HashImpl) Keys() []string {
	keys := make([]string, 0, h.Len())
	for key := range h.data {
		if h.Has(key) {
			keys = append(keys, key)
		}
	}
	return keys
}

// Values returns all values in the hash.
func (h *HashImpl) Values() []string {
	values := make([]string, 0, h.Len())
	for key := range h.data {
		if value, exists := h.Get(key); exists {
			values = append(values, *value)
		}
	}
	return values
}

// Items returns all valid (non-expired) elements in the hash.
func (h *HashImpl) Items() [][]string {
	items := make([][]string, 0, h.Len())
	for field := range h.data {
		if value, exists := h.Get(field); exists {
			items = append(items, []string{field, *value})
		}
	}
	return items
}
