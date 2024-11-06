package hash

// IHashMap is an interface that defines the methods that a hash map must implement.
// NewHashMap creates a new instance of IHashMap with the specified encoding.
// The encoding parameter determines the type of hash map to be created.
// Currently supported encodings:
// - 6: Returns a SimpleHashMap
// - 7: Placeholder for future implementation, currently returns a SimpleHashMap
// For any other encoding, a SimpleHashMap is returned.
//
// K: The type of keys in the hash map, must be comparable.
// V: The type of values in the hash map, can be any type.
//
// Parameters:
// - encoding: An integer representing the desired encoding type.
//
// Returns:
// - IHashMap[K, V]: An instance of a hash map with the specified encoding.
func NewHashMap[K comparable, V any](encoding int) IHashMap[K, V] {
	switch encoding {
	case 6:
		return NewSimpleMap[K, V]()
	case 7:
		// unimplemented but place holder for future
		return NewSimpleMap[K, V]()
	default:
		return NewSimpleMap[K, V]()
	}
}
