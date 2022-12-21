package core

import (
	"reflect"
	"testing"
)

func TestBloomOpts(t *testing.T) {
	var testCases = []struct {
		name        string
		args        []string
		useDefaults bool
		response    *BloomOpts
		err         error
	}{
		{"default values", []string{""}, true, &BloomOpts{errorRate: defaultErrorRate, capacity: defaultCapacity}, nil},
		{"should return valid values - 1", []string{"0.01", "1000"}, false, &BloomOpts{errorRate: 0.01, capacity: 1000}, nil},
		{"should return valid values - 2", []string{"0.1", "200"}, false, &BloomOpts{errorRate: 0.1, capacity: 200}, nil},
		{"should return invalid error rate type - 1", []string{"aa", "100"}, false, nil, errInvalidErrorRateType},
		{"should return invalid error rate type - 2", []string{"0.1a", "100"}, false, nil, errInvalidErrorRateType},
		{"should return invalid error rate - 1", []string{"-0.1", "100"}, false, nil, errInvalidErrorRate},
		{"should return invalid error rate - 2", []string{"1.001", "100"}, false, nil, errInvalidErrorRate},
		{"should return invalid capacity type - 1", []string{"0.01", "aa"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity type - 2", []string{"0.01", "100a"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity type - 3", []string{"0.01", "-1"}, false, nil, errInvalidCapacityType},
		{"should return invalid capacity - 1", []string{"0.01", "0"}, false, nil, errInvalidCapacity},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts, err := newBloomOpts(tc.args, tc.useDefaults)
			// Using reflect.DeepEqual as we have pointers to struct and direct value
			// comparision is not possible because of []hash.Hash64 type.
			if !reflect.DeepEqual(opts, tc.response) {
				t.Errorf("invalid response in %s - expected %v, got %v", t.Name(), tc.response, opts)
			}

			if err != tc.err {
				t.Errorf("invalid error in %s - expected %v, got %v", t.Name(), tc.err, err)
			}
		})
	}
}

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
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := isBitSet(buf, tc.index)
			if actual != tc.expected {
				t.Errorf("error in %s - expected %t, got %t", t.Name(), tc.expected, actual)
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
		t.Run(tc.name, func(t *testing.T) {
			// Set bit first and then try to read it
			setBit(buf, tc.index)
			actual := isBitSet(buf, tc.index)
			if actual != tc.expected {
				t.Errorf("error in %s - expected %t, got %t", t.Name(), tc.expected, actual)
			}
		})
	}

	// The final values are 10111011 (=187)
	expected1, expected2 := 187, 187
	if int(buf[0]) != expected1 || int(buf[1]) != expected2 {
		t.Errorf("error in %s while comparing final buffer values - expected [%d, %d], got [%d, %d]", t.Name(), expected1, expected2, int(buf[0]), int(buf[1]))
	}
}
