package core

import (
	"testing"
)

func TestIsBitSet(t *testing.T) {
	buf := []byte{170, 43} // 10101010 00101011
	var testCases = []struct {
		name     string
		index    int
		expected bool
	}{
		{"Handle negative index", -1, false},
		{"Handle index equal to length", 16, false},
		{"Handle index more than length", 17, false},
		{"Handle start bit 1", 0, true},
		{"Handle start bit 2", 8, false},
		{"Handle mid bit 1", 3, false},
		{"Handle mid bit 2", 4, true},
		{"Handle mid bit 3", 11, false},
		{"Handle mid bit 4", 14, true},
		{"Handle end bit 1", 7, false},
		{"Handle end bit 2", 15, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(_ *testing.T) {
			t.Parallel()
			actual := isBitSet(buf, tc.index)
			if actual != tc.expected {
				t.Errorf("error in %s for case: %s - expected %t, got %t", t.Name(), tc.name, tc.expected, actual)
			}
		})
	}
}

func TestSetBit(t *testing.T) {
	buf := []byte{170, 43} // 10101010 00101011
	var testCases = []struct {
		name     string
		index    int
		expected bool
	}{
		{"Handle negative index", -1, false},
		{"Handle index equal to length", 16, false},
		{"Handle index more than length", 17, false},
		{"Handle start bit 1", 0, true}, // 10101010 00101011
		{"Handle start bit 2", 8, true}, // 10101010 10101011
		{"Handle mid bit 1", 3, true},   // 10111010 10101011
		{"Handle mid bit 2", 4, true},   // 10111010 10101011
		{"Handle mid bit 3", 11, true},  // 10111010 10111011
		{"Handle mid bit 4", 14, true},  // 10111010 10111011
		{"Handle end bit 1", 7, true},   // 10111011 10111011
		{"Handle end bit 2", 15, true},  // 10111011 10111011
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(_ *testing.T) {
			// Set bit first and then try to read it
			setBit(buf, tc.index)
			actual := isBitSet(buf, tc.index)
			if actual != tc.expected {
				t.Errorf("error in %s for case: %s - expected %t, got %t", t.Name(), tc.name, tc.expected, actual)
			}
		})
	}

	// The final values are 10111011 (=187)
	expected1, expected2 := 187, 187
	if int(buf[0]) != expected1 || int(buf[1]) != expected2 {
		t.Errorf("error in %s while comparing final buffer values - expected [%d, %d], got [%d, %d]", t.Name(), expected1, expected2, int(buf[0]), int(buf[1]))
	}
}
