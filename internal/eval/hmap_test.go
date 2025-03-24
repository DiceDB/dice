// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/store"
	"github.com/stretchr/testify/assert"
)

func TestHashMapSetAndGet(t *testing.T) {
	hmap := make(HashMap)

	hmap.Set("key1", "value1")
	val, ok := hmap.Get("key1")
	assert.True(t, ok, "Expected key1 to exist in the HashMap")
	assert.Equal(t, "value1", *val, "Expected value1 to be returned")

	oldVal, ok := hmap.Set("key1", "newValue")
	assert.True(t, ok, "Expected key1 to exist when overwriting")
	assert.Equal(t, "value1", *oldVal, "Expected the old value to be returned")

	val, ok = hmap.Get("key2")
	assert.False(t, ok, "Expected key2 to not exist in the HashMap")
	assert.Nil(t, val, "Expected nil value for non-existent key")
}

func TestHashMapBuilder(t *testing.T) {
	keyValuePairs := []string{"key1", "value1", "key2", "value2"}
	hmap, numSet, err := hashMapBuilder(keyValuePairs, nil)
	assert.Nil(t, err, "Expected no error for valid input")
	assert.Equal(t, int64(2), numSet, "Expected 2 keys to be newly set")
	val1, ok := hmap.Get("key1")
	assert.True(t, ok, "Expected key1 to exist")
	assert.Equal(t, "value1", *val1, "Expected value1 for key1")
	val2, ok := hmap.Get("key2")
	assert.True(t, ok, "Expected key2 to exist")
	assert.Equal(t, "value2", *val2, "Expected value2 for key2")

	keyValuePairs = []string{"key1", "value1", "key2"}
	hmap, numSet, err = hashMapBuilder(keyValuePairs, nil)
	assert.NotNil(t, err, "Expected error for odd number of key-value pairs")
	assert.Equal(t, int64(-1), numSet, "Expected -1 for number of keys set when error occurs")
}

func TestHashMapIncrementValue(t *testing.T) {
	hmap := make(HashMap)

	val, err := hmap.incrementValue("field1", 10)
	assert.Nil(t, err, "Expected no error when incrementing a non-existent key")
	assert.Equal(t, int64(10), val, "Expected value to be set to 10")

	val, err = hmap.incrementValue("field1", 5)
	assert.Nil(t, err, "Expected no error when incrementing an existing key")
	assert.Equal(t, int64(15), val, "Expected value to be incremented to 15")

	hmap.Set("field2", "notAnInt")
	val, err = hmap.incrementValue("field2", 1)
	assert.NotNil(t, err, "Expected error when incrementing a non-integer value")
	assert.Equal(t, errors.HashValueNotIntegerErr, err.Error(), "Expected hash value not integer error")

	hmap.Set("field3", strconv.FormatInt(math.MaxInt64, 10))
	val, err = hmap.incrementValue("field3", 1)
	assert.NotNil(t, err, "Expected error when integer overflow occurs")
	assert.Equal(t, errors.IncrDecrOverflowErr, err.Error(), "Expected increment overflow error")
}

func TestGetValueFromHashMap(t *testing.T) {
	store := store.NewStore(nil, nil)
	key := "key1"
	field := "field1"
	value := "value1"
	hmap := make(HashMap)
	hmap.Set(field, value)

	obj := &object.Obj{
		Type:           object.ObjTypeSSMap,
		Value:          hmap,
		LastAccessedAt: uint32(time.Now().Unix()),
	}

	store.Put(key, obj)

	// Test case: Fetching an existing field
	response := getValueFromHashMap(key, field, store)
	assert.Nil(t, response.Error, "Expected no error when fetching an existing value from the hashmap")
	assert.NotNil(t, response.Result, "Expected a non-nil value to be fetched for key 'key1' and field 'field1'")
	assert.Equal(t, value, response.Result, "Expected 'value1' to be fetched for key 'key1' and field 'field1'")

	// Test case: Fetching a non-existing field (should return NIL and no error)
	response = getValueFromHashMap(key, "nonfield", store)
	assert.Equal(t, NIL, response.Result, "Expected NIL for a non-existing field")
	assert.Nil(t, response.Error, "Expected no error when fetching a non-existing field from the hashmap")

	// Test case: Fetching a non-existing key (should return NIL and ErrKeyNotFound)
	response = getValueFromHashMap("nonkey", field, store)
	assert.Equal(t, NIL, response.Result, "Expected NIL for a non-existing key")
	assert.Nil(t, response.Error, "Expected no error for a non-existing key")
}

func TestHashMapIncrementFloatValue(t *testing.T) {
	hmap := make(HashMap)

	val, err := hmap.incrementFloatValue("field1", 5.5)
	assert.Nil(t, err, "Expected no error when incrementing a non-existent key")
	assert.Equal(t, "5.5", val, "Expected value to be set to 5.5")

	val, err = hmap.incrementFloatValue("field1", 4.5)
	assert.Nil(t, err, "Expected no error when incrementing an existing key")
	assert.Equal(t, "10", val, "Expected value to be incremented to 10")

	hmap.Set("field2", "notAFloat")
	val, err = hmap.incrementFloatValue("field2", 1.0)
	assert.NotNil(t, err, "Expected error when incrementing a non-float value")
	assert.Equal(t, errors.IntOrFloatErr, err.Error(), "Expected int or float error")

	inf := math.MaxFloat64

	val, err = hmap.incrementFloatValue("field1", inf+float64(1e308))
	assert.NotNil(t, err, "Expected error when incrementing a overflowing value")
	assert.Equal(t, errors.IncrDecrOverflowErr, err.Error(), "Expected overflow to be detected")

	val, err = hmap.incrementFloatValue("field1", -inf-float64(1e308))
	assert.NotNil(t, err, "Expected error when incrementing a overflowing value")
	assert.Equal(t, errors.IncrDecrOverflowErr, err.Error(), "Expected overflow to be detected")
}
