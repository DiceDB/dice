package hash

import (
	"testing"

	"github.com/dicedb/dice/internal/server/utils"
	"github.com/stretchr/testify/assert"
)

func testNewHashMap(t *testing.T, hmap IHashMap[string, string]) {
	assert.NotNil(t, hmap, "Expected new HashMap instance to be non-nil")
	assert.Equal(t, int64(0), hmap.ALen(), "Expected initial length to be 0")
}

func testSetAndGet(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")

	val, ok := hmap.Get("key1")
	assert.True(t, ok, "Expected key1 to exist")
	assert.Equal(t, "value1", *val, "Expected value1 to be returned")

	// Test non-existent key
	val, ok = hmap.Get("key2")
	assert.False(t, ok, "Expected key2 to not exist")
	assert.Empty(t, val, "Expected empty value for non-existent key")
}

func testSetExpiry(t *testing.T, hmap IHashMap[string, string]){
	hmap.Set("key1", "value1")

	// Set expiry to 1 second in the future
	expiryTime := 1000
	hmap.SetExpiry("key1", int64(expiryTime))

	_, expiry := hmap.GetExpiry("key1")
	assert.True(t,expiry>0, "Expected expiry for key1 to exist")
	expiryMs := uint64(expiry) - uint64(utils.GetCurrentTime().UnixMilli())
	assert.True(t, expiryMs > 0, "Expected expiry time to be greater than 0")
}

func testGetExpiredKey(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")

	// Set expiry to 1 second in the past
	expiryTime := -1000
	hmap.SetExpiry("key1", int64(expiryTime))

	val, ok := hmap.Get("key1")
	assert.False(t, ok, "Expected key1 to be expired")
	assert.Empty(t, val, "Expected empty value for expired key")
}

func testKeys(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Set expiry for key2
	hmap.SetExpiry("key2", -1000) // key2 expired

	keys := hmap.Keys()
	assert.Equal(t, 1, len(keys), "Expected only 1 valid key")
	assert.Contains(t, keys, "key1", "Expected key1 to be in valid keys")
	assert.NotContains(t, keys, "key2", "Expected key2 to be expired and not in valid keys")
}

func testValues(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Set expiry for key2
	hmap.SetExpiry("key2", -1000) // key2 expire

	values := hmap.Values()
	assert.Equal(t, 1, len(values), "Expected only 1 valid value")
	assert.Contains(t, values, "value1", "Expected value1 to be in valid values")
	assert.NotContains(t, values, "value2", "Expected value2 to be expired and not in valid values")
}

func testClear(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")
	hmap.Clear()

	assert.Equal(t, int64(0), hmap.ALen(), "Expected length to be 0 after clearing")
}

func testApproxLength(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	assert.Equal(t, int64(2), hmap.ALen(), "Expected approximate length to be 2")
}

func testLen(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")

	// Set expiry for key2
	hmap.SetExpiry("key2", -1000) // key2 expired

	assert.Equal(t, int64(1), hmap.Len(), "Expected actual valid length to be 1")
}

func testDelete(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("key1", "value1")
	hmap.Set("key2", "value2")
	hmap.Delete("key1")
	val, ok := hmap.Get("key1")
	assert.False(t, ok, "Expected key1 to be deleted")

	// Ensure key2 still exists
	val, ok = hmap.Get("key2")
	assert.True(t, ok, "Expected key2 to still exist")
	assert.Equal(t, "value2", *val, "Expected value2 for key2")
}

func TestCreateOrModify(t *testing.T) {
	hmap := NewHashMap[string, int](6)

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

func testFind(t *testing.T, hmap IHashMap[string, string]) {
	hmap.Set("apple", "fruit")
	hmap.Set("banana", "fruit")
	hmap.Set("carrot", "vegetable")
	hmap.Set("apricot", "fruit")
	hmap.Set("avocado", "fruit")
	assert.Equal(t, int64(5), hmap.Len())
	result := hmap.Find("a*")
	assert.ElementsMatch(t, result, []string{"avocado", "apple", "apricot"})
	result = hmap.Find("*r*")
	assert.ElementsMatch(t, result, []string{"carrot", "apricot"})
}

func TestAllHashMap(t *testing.T) {
	encodings := map[int]string{
		6: "SimpleHashMap",
		7: "ComplexHashMap (future)",
	}

	for encoding, name := range encodings {
		t.Run(name, func(t *testing.T) {
			hmap := NewHashMap[string, string](encoding)
			testNewHashMap(t, hmap)
			hmap.Clear()
			testSetAndGet(t, hmap)
			hmap.Clear()
			testSetExpiry(t, hmap)
			hmap.Clear()
			testGetExpiredKey(t, hmap)
			hmap.Clear()
			testKeys(t, hmap)
			hmap.Clear()
			testValues(t, hmap)
			hmap.Clear()
			testClear(t, hmap)
			hmap.Clear()
			testApproxLength(t, hmap)
			hmap.Clear()
			testLen(t, hmap)
			hmap.Clear()
			testDelete(t, hmap)
			hmap.Clear()
			testFind(t, hmap)
			hmap.Clear()
		})
	}
}