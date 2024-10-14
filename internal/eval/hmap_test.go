package eval

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHashMap(t *testing.T) {
	hmap := NewHashMap[string, string]()
	assert.NotNil(t, hmap, "Expected new HashMap instance to be non-nil")
	assert.Equal(t, 0, hmap.ApproxLength(), "Expected initial length to be 0")
}

func TestSetAndGet(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")

	val, ok := hmap.Get("key1")
	assert.True(t, ok, "Expected key1 to exist")
	assert.Equal(t, "value1", *val, "Expected value1 to be returned")

	// Test non-existent key
	val, ok = hmap.Get("key2")
	assert.False(t, ok, "Expected key2 to not exist")
	assert.Empty(t, val, "Expected empty value for non-existent key")
}

func TestSetExpiry(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")

	// Set expiry to 1 second in the future
	expiryTime := uint64(time.Now().UnixMilli() + 1000)
	hmap.SetExpiry("key1", expiryTime)

	expiry, ok := hmap.GetExpiry("key1")
	assert.True(t, ok, "Expected expiry for key1 to exist")
	assert.Equal(t, expiryTime, expiry, "Expected expiry time to match")
}

func TestGetExpiredKey(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")

	// Set expiry to 1 second in the past
	expiryTime := uint64(time.Now().UnixMilli() - 1000)
	hmap.SetExpiry("key1", expiryTime)

	val, ok := hmap.Get("key1")
	assert.False(t, ok, "Expected key1 to be expired")
	assert.Empty(t, val, "Expected empty value for expired key")
}

func TestKeys(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Set expiry for key2
	hmap.SetExpiry("key2", uint64(time.Now().UnixMilli()-1000)) // key2 expired

	keys := hmap.Keys()
	assert.Equal(t, 1, len(keys), "Expected only 1 valid key")
	assert.Contains(t, keys, "key1", "Expected key1 to be in valid keys")
	assert.NotContains(t, keys, "key2", "Expected key2 to be expired and not in valid keys")
}

func TestValues(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Set expiry for key2
	hmap.SetExpiry("key2", uint64(time.Now().UnixMilli()-1000)) // key2 expired

	values := hmap.Values()
	assert.Equal(t, 1, len(values), "Expected only 1 valid value")
	assert.Contains(t, values, "value1", "Expected value1 to be in valid values")
	assert.NotContains(t, values, "value2", "Expected value2 to be expired and not in valid values")
}

func TestClear(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")
	hmap.Clear()

	assert.Equal(t, 0, hmap.ApproxLength(), "Expected length to be 0 after clearing")
}

func TestApproxLength(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	assert.Equal(t, 2, hmap.ApproxLength(), "Expected approximate length to be 2")
}

func TestLen(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Set expiry for key2
	hmap.SetExpiry("key2", uint64(time.Now().UnixMilli()-1000)) // key2 expired

	assert.Equal(t, 1, hmap.Len(), "Expected actual valid length to be 1")
}

func TestDelete(t *testing.T) {
	hmap := NewHashMap[string, string]()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Delete key1
	deleted := hmap.Delete("key1")
	assert.True(t, deleted, "Expected key1 to be deleted successfully")
	assert.False(t, hmap.Delete("key1"), "Expected key1 to not exist after deletion")

	// Ensure key2 still exists
	val, ok := hmap.Get("key2")
	assert.True(t, ok, "Expected key2 to still exist")
	assert.Equal(t, "value2", *val, "Expected value2 for key2")
}

func TestCreateOrModify(t *testing.T) {
	hmap := NewHashMap[string, int]()

	// Increment an integer value
	modifier := func(currentValue int) (int, error) {
		return currentValue + 1, nil // Increment the current value by 1
	}

	// Key does not exist yet
	_, err := hmap.CreateOrModify("counter", modifier)
	assert.NoError(t, err, "Expected no error when modifying non-existent key")

	val, ok := hmap.Get("counter")
	assert.True(t, ok, "Expected counter to exist")
	assert.Equal(t, 1, *val, "Expected counter to be initialized to 1")

	// Increment again
	_, err = hmap.CreateOrModify("counter", modifier)
	assert.NoError(t, err, "Expected no error when modifying existing key")

	val, ok = hmap.Get("counter")
	assert.True(t, ok, "Expected counter to exist")
	assert.Equal(t, 2, *val, "Expected counter to be incremented to 2")
}
