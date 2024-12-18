// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
