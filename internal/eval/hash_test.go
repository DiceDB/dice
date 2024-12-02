package eval

import (
	"fmt"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewHashMap(t *testing.T) {
	hmap := NewHash()
	assert.NotNil(t, hmap, "Expected new HashMap instance to be non-nil")
	assert.Equal(t, 0, hmap.Len(), "Expected initial length to be 0")
}

func TestSetAndGet(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")

	val, ok := hmap.Get("key1")
	assert.True(t, ok, "Expected key1 to exist")
	assert.Equal(t, "value1", *val, "Expected value1 to be returned")

	// Test non-existent key
	val, ok = hmap.Get("key2")
	assert.False(t, ok, "Expected key2 to not exist")
	assert.Empty(t, *val, "Expected empty value for non-existent key")
}

func TestSetExpiry(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")

	// Set expiry to 1 second in the future
	expiryTime := int64(1000) // in milliseconds
	hmap.SetExpiry("key1", expiryTime)

	item, ok := hmap.GetWithExpiry("key1")
	assert.True(t, ok, "Expected key1 to exist with expiry")
	assert.False(t, item.IsPersistent(), "Expected key1 to have an expiry time")
}

func TestGetExpiredKey(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")

	// Set expiry to 1 second in the past
	expiryTime := int64(-1000)
	hmap.SetExpiry("key1", expiryTime)

	_, ok := hmap.Get("key1")
	assert.False(t, ok, "Expected key1 to be expired")
}

func TestKeys(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Expire key2
	hmap.SetExpiry("key2", -1000)

	keys := hmap.Keys()
	assert.Equal(t, 1, len(keys), "Expected only 1 valid key")
	assert.Contains(t, keys, "key1", "Expected key1 to be in valid keys")
}

func TestValues(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Expire key2
	hmap.SetExpiry("key2", -1000)

	values := hmap.Values()
	assert.Equal(t, 1, len(values), "Expected only 1 valid value")
	assert.Contains(t, values, "value1", "Expected value1 to be in valid values")
}

func TestClear(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")
	hmap.Clear()

	assert.Equal(t, 0, hmap.Len(), "Expected length to be 0 after clearing")
}

func TestLen(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Expire key2
	hmap.SetExpiry("key2", -1000)

	assert.Equal(t, 1, hmap.Len(), "Expected length to be 1 after expiry")
}

func TestDelete(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Delete key1
	deleted := hmap.Delete("key1")
	assert.True(t, deleted, "Expected key1 to be deleted successfully")
	assert.False(t, hmap.Has("key1"), "Expected key1 to not exist after deletion")

	// Ensure key2 still exists
	val, ok := hmap.Get("key2")
	assert.True(t, ok, "Expected key2 to still exist")
	assert.Equal(t, "value2", *val, "Expected value2 for key2")
}

func TestHas(t *testing.T) {
	hmap := NewHash()
	hmap.Set("key1", "value1")

	assert.True(t, hmap.Has("key1"), "Expected key1 to exist")
	hmap.Delete("key1")
	assert.False(t, hmap.Has("key1"), "Expected key1 to not exist after deletion")
}

func TestSetWithExpiry(t *testing.T) {
	fmt.Println("INFO: TestSetWithExpiry")
	mockTime := &utils.MockClock{CurrTime: time.Now()}
	utils.CurrentTime = mockTime
	hmap := NewHash()
	hmap.SetWithExpiry("key1", "value1", 500) // 500 ms expiry

	val, ok := hmap.Get("key1")
	fmt.Println("INFO: TestSetWithExpiry 1", val, ok)
	assert.True(t, ok, "Expected key1 to exist before expiry")
	assert.Equal(t, "value1", *val, "Expected value1 for key1")
	// time.Sleep(1 * time.Second)    // Sleep for 1 second
	mockTime.SetTime(mockTime.CurrTime.Add(1 * time.Second))
	_, ok = hmap.Get("key1")
	fmt.Println("INFO: TestSetWithExpiry 2", val, ok)
	assert.False(t, ok, "Expected key1 to have expired")
}
