package core_test

import (
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
)

type stackinttcase struct {
	op   byte
	val  int64
	list []int64
	err  error
}

func TestStackInt(t *testing.T) {
	si := core.NewStackInt()
	for _, tc := range []stackinttcase{
		{'i', 1, []int64{1}, nil},
		{'i', 2, []int64{2, 1}, nil},
		{'i', 3, []int64{3, 2, 1}, nil},
		{'i', 4, []int64{4, 3, 2, 1}, nil},
		{'i', 5, []int64{5, 4, 3, 2, 1}, nil},
		{'i', 6, []int64{6, 5, 4, 3, 2, 1}, nil},
		{'i', 78930943, []int64{78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 11918629, []int64{11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 9223372036854775807, []int64{9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 25324944, []int64{25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 22402494, []int64{22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 51881029, []int64{51881029, 22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 79283552, []int64{79283552, 51881029, 22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'i', 67459748, []int64{67459748, 79283552, 51881029, 22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{79283552, 51881029, 22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{51881029, 22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{22402494, 25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{25324944, 9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{9223372036854775807, 11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{11918629, 78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{78930943, 6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{6, 5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{5, 4, 3, 2, 1}, nil},
		{'r', 0, []int64{4, 3, 2, 1}, nil},
		{'r', 0, []int64{3, 2, 1}, nil},
		{'r', 0, []int64{2, 1}, nil},
		{'r', 0, []int64{1}, nil},
		{'r', 0, []int64{}, nil},
		{'r', 0, []int64{}, core.ErrStackEmpty},
		{'r', 0, []int64{}, core.ErrStackEmpty},
		{'r', 0, []int64{}, core.ErrStackEmpty},
		{'i', 1, []int64{1}, nil},
		{'i', 2, []int64{2, 1}, nil},
	} {
		var err error
		switch tc.op {
		case 'i':
			si.Push(tc.val)
		case 'r':
			_, err = si.Pop()
		}

		if err != tc.err {
			t.Errorf("we expected the error: %v but got %v", tc.err, err)
		}

		obsList := si.Iterate(50)
		if !testutils.EqualInt64Slice(tc.list, obsList) {
			t.Errorf("stackint test failed. should have been %v but found %v", tc.list, obsList)
		}
	}
}
