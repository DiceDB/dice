package eval

import (
	"github.com/dicedb/dice/internal/server/utils"
)

type Hash interface {
	Set(field, value string) bool
	SetExpiry(field string, expiryMs int64)
	SetWithExpiry(field, value string, expiryMs int64) bool
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
	Has(field string) bool
	Delete(field string) bool
	Clear()
	Len() int
	Keys() []string
	Values() []string
	Items() [][]string
}

type HashItem struct {
	Value  string
	Expiry uint64 // MSB = 1 if the element has no expiry
}

func (i *HashItem) IsPersistent() bool {
	return (i.Expiry>>63)&1 == 1
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
	h.data[field] = HashItem{Value: value, Expiry: 1 << 63}
	return existed
}

// SetWithExpiry inserts a value into the hash with an expiry in milliseconds.
func (h *HashImpl) SetWithExpiry(field, value string, expiryMs int64) bool {
	existed := h.Has(field)
	if expiryMs > 0 {
		expiryUnixMs := uint64(utils.GetCurrentTime().UnixMilli() + expiryMs)
		h.data[field] = HashItem{Value: value, Expiry: expiryUnixMs}
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
	if (item.Expiry>>63)&1 == 1 {
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
