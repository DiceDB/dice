package core

import (
	"testing"
)

func TestDelExpiry(t *testing.T) {
	store := NewStore()
	// Initialize the test environment
	store.store = make(map[string]*Obj)
	store.expires = make(map[*Obj]uint64)
	store.keypool = make(map[string]*string)

	// Define test cases
	tests := []struct {
		name     string
		obj      *Obj
		setup    func(*Obj)
		expected bool // false if we expect the key to be deleted
	}{
		{
			name: "Object with expiration",
			obj:  &Obj{},
			setup: func(obj *Obj) {
				store.expires[obj] = 12345 // Set some expiration time
			},
			expected: false,
		},
		{
			name: "Object without expiration",
			obj:  &Obj{},
			setup: func(obj *Obj) {
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

			// Call delExpiry
			delExpiry(tc.obj, store)

			// Check if the key has been deleted from the expires map
			_, exists := store.expires[tc.obj]
			if exists != tc.expected {
				t.Errorf("%s: expected key to be deleted: %v, got: %v", tc.name, tc.expected, exists)
			}
		})
	}
}
