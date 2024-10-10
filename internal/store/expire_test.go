package store

import (
	"testing"

	"github.com/dicedb/dice/internal/object"
)

func TestDelExpiry(t *testing.T) {
	store := NewStore(nil, nil)
	// Initialize the test environment
	store.store = NewStoreMap()
	store.expires = NewExpireMap()

	// Define test cases
	tests := []struct {
		name     string
		obj      *object.Obj
		setup    func(*object.Obj)
		expected bool // false if we expect the key to be deleted
	}{
		{
			name: "Object with expiration",
			obj:  &object.Obj{},
			setup: func(obj *object.Obj) {
				store.expires.Put(obj, 12345) // Set some expiration time
			},
			expected: false,
		},
		{
			name: "Object without expiration",
			obj:  &object.Obj{},
			setup: func(obj *object.Obj) {
				// No setup needed as the object should not have an expiration
			},
			expected: false,
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the test case
			tc.setup(tc.obj)

			// Call DelExpiry
			DelExpiry(tc.obj, store)

			// Check if the key has been deleted from the expires map
			_, exists := store.expires.Get(tc.obj)
			if exists != tc.expected {
				t.Errorf("%s: expected key to be deleted: %v, got: %v", tc.name, tc.expected, exists)
			}
		})
	}
}
