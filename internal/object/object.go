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

package object

// Obj represents a basic object structure that includes metadata about the object
// as well as its value. This struct is designed to store the core data (value)
// of the object along with additional properties such as encoding type and access
// time. The structure is intended to support different types of objects, including
// simple types (e.g., int, string) and more complex data structures (e.g., hashes, sets).
//
// Fields:
//
//   - Type: A uint8 field used to store the type of the object. This
//     helps in identifying how the object is encoded or serialized (e.g., as a string,
//     number, or more complex type). It is crucial for determining how the object should
//     be interpreted or processed when retrieved from storage.
//
//   - LastAccessedAt: A uint32 field that stores the timestamp (in seconds or milliseconds)
//     representing the last time the object was accessed. This field helps in tracking the
//     freshness of the object and can be used for cache expiry or eviction policies.
//     in DiceDB we use 32 bits due to Go's lack of native support for bitfields,
//     and to simplify management by not combining
//     `Type` and `LastAccessedAt` into a single integer.
//
//   - Value: An `interface{}` type that holds the actual data of the object. This could
//     represent any type of data, allowing flexibility to store different kinds of
//     objects (e.g., strings, numbers, complex data structures like lists or maps).
type Obj struct {
	// Type holds the type of the object (e.g., string, int, complex structure)
	Type ObjectType

	// LastAccessedAt stores the last access timestamp of the object.
	// It helps track when the object was last accessed and may be used for cache eviction or freshness tracking.
	LastAccessedAt uint32

	// Value holds the actual content or data of the object, which can be of any type.
	// This allows flexibility in storing various kinds of objects (simple or complex).
	Value interface{}
}

// ExtendedObj is an extension of the `Obj` struct, designed to add extra
// metadata for more advanced use cases. It includes a reference to an `Obj`
// and an additional field `ExDuration` to store a time-based property (e.g.,
// the duration the object is valid or how long it should be stored).
//
// This struct allows for greater flexibility in managing objects that have
// additional time-sensitive or context-dependent attributes. The `ExDuration`
// field can represent the lifespan, expiration time, or any other time-related
// characteristic associated with the object.
//
// **Caution**: `InternalObj` should be used selectively and with caution.
// It is intended for commands that require the additional metadata provided by
// the `ExDuration` field. Using this for all objects can complicate logic and
// may lead to unnecessary overhead. It is recommended to use `ExtendedObj` only
// for commands that specifically require it and when additional metadata is
// necessary for the functionality of the command.
//
// Fields:
//
//   - Obj: A pointer to the underlying `Obj` that contains the core object data
//     (such as encoding type, last accessed timestamp, and the actual value of the object).
//     This allows the extended object to retain all of the functionality of the basic object
//     while adding more flexibility through the extra metadata.
//
//   - ExDuration: Represents a time-based property, such as the duration for which
//     the object is valid or the time it should be stored or processed. This allows for
//     managing the object with respect to time-based characteristics, such as expiration.
type InternalObj struct {
	// Obj holds the core object structure with the basic metadata and value
	Obj *Obj

	// ExDuration represents a time-based property, such as the duration for which
	// the object is valid or the time it should be stored or processed.
	ExDuration int64
}

// ObjectType represents the type of a DiceDB object
type ObjectType uint8

// Define object types as constants
const (
	ObjTypeString ObjectType = iota
	_                        // skip 1 and 2 to maintain compatibility
	_
	ObjTypeJSON
	ObjTypeByteArray
	ObjTypeInt
	ObjTypeSet
	ObjTypeHashMap
	ObjTypeSortedSet
	ObjTypeCountMinSketch
	ObjTypeBF
	ObjTypeDequeue
	ObjTypeHLL
)
