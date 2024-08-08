package core

import (
	"testing"
	"unsafe"
)

func TestDelExpiry(t *testing.T) {
	// Initialize the test environment
	store = make(map[unsafe.Pointer]*Obj)
	expires = make(map[*Obj]uint64)
	keypool = make(map[string]unsafe.Pointer)

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
				expires[obj] = 12345 // Set some expiration time
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
			delExpiry(tc.obj)

			// Check if the key has been deleted from the expires map
			_, exists := expires[tc.obj]
			if exists != tc.expected {
				t.Errorf("%s: expected key to be deleted: %v, got: %v", tc.name, tc.expected, exists)
			}
		})
	}
}
