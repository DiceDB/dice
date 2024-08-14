package testutils

// UnorderedEqual compares two slices of interfaces and returns true if they
// contain the same elements, regardless of order.
func UnorderedEqual(expected, actual interface{}) bool {
	expectedSlice, ok := expected.([]interface{})
	if !ok {
		return false
	}

	actualSlice, ok := actual.([]interface{})
	if !ok {
		return false
	}

	if len(expectedSlice) != len(actualSlice) {
		return false
	}

	for _, e := range expectedSlice {
		found := false
		for _, a := range actualSlice {
			if e == a {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// EqualByteSlice compares two byte slices and returns true if they are equal.
func EqualByteSlice(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// EqualInt64Slice compares two int64 slices and returns true if they are equal.
func EqualInt64Slice(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
