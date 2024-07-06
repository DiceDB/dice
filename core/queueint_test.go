package core_test

import (
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/testutils"
)

type queueinttcase struct {
	op   byte
	val  int64
	list []int64
	err  error
}

func TestQueueInt(t *testing.T) {
	qi := core.NewQueueInt()
	for _, tc := range []queueinttcase{
		{'i', 1, []int64{1}, nil},
		{'i', 2, []int64{1, 2}, nil},
		{'i', 3, []int64{1, 2, 3}, nil},
		{'i', 4, []int64{1, 2, 3, 4}, nil},
		{'i', 5, []int64{1, 2, 3, 4, 5}, nil},
		{'i', 6, []int64{1, 2, 3, 4, 5, 6}, nil},
		{'i', 78930943, []int64{1, 2, 3, 4, 5, 6, 78930943}, nil},
		{'i', 11918629, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629}, nil},
		{'i', 68141363, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629, 68141363}, nil},
		{'i', 25324944, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944}, nil},
		{'i', 22402494, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494}, nil},
		{'i', 51881029, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029}, nil},
		{'i', 79283552, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552}, nil},
		{'i', 67459748, []int64{1, 2, 3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{2, 3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{3, 4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{4, 5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{5, 6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{6, 78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{78930943, 11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{11918629, 68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{68141363, 25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{25324944, 22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{22402494, 51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{51881029, 79283552, 67459748}, nil},
		{'r', 0, []int64{79283552, 67459748}, nil},
		{'r', 0, []int64{67459748}, nil},
		{'r', 0, []int64{}, nil},
		{'r', 0, []int64{}, core.ErrQueueEmpty},
		{'r', 0, []int64{}, core.ErrQueueEmpty},
		{'r', 0, []int64{}, core.ErrQueueEmpty},
		{'i', 1, []int64{1}, nil},
		{'i', 2, []int64{1, 2}, nil},
	} {
		var err error
		switch tc.op {
		case 'i':
			qi.Insert(tc.val)
		case 'r':
			_, err = qi.Remove()
		}

		if err != tc.err {
			t.Errorf("we expected the error: %v but got %v", tc.err, err)
		}

		obsList := qi.Iterate(50)
		if !testutils.EqualInt64Slice(tc.list, obsList) {
			t.Errorf("queueint test failed. should have been %v but found %v", tc.list, obsList)
		}
	}
}

func TestSize(t *testing.T) {
	qi := core.NewQueueInt()
	for i := 0; i < 20; i++ {
		qi.Insert(int64(i))
	}
	qsize := qi.Length
	if qsize != 20 {
		t.Errorf("queueint test failed. should have been 20 but found %v", qsize)
	}
}

func insertMany(howmany int, qi core.QueueIntI, b *testing.B) {
	for i := 0; i < howmany; i++ {
		qi.Insert(int64(i))
	}
	obsList := qi.Iterate(howmany + 10)
	if len(obsList) != howmany {
		b.Errorf("queueint test failed. should have been %d but found %v", howmany, len(obsList))
	}
}

func BenchmarkInsert20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(20, core.NewQueueInt(), b)
	}
}

func BenchmarkInsert200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(200, core.NewQueueInt(), b)
	}
}

func BenchmarkInsert2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(2000, core.NewQueueInt(), b)
	}
}

func BenchmarkInsertLL20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(20, core.NewQueueIntLL(), b)
	}
}

func BenchmarkInsertLL200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(200, core.NewQueueIntLL(), b)
	}
}

func BenchmarkInsertLL2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(2000, core.NewQueueIntLL(), b)
	}
}

func BenchmarkInsertBasic20(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(20, core.NewQueueIntBasic(), b)
	}
}

func BenchmarkInsertBasic200(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(200, core.NewQueueIntBasic(), b)
	}
}

func BenchmarkInsertBasic2000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		insertMany(2000, core.NewQueueIntBasic(), b)
	}
}

func BenchmarkInsertRemove(b *testing.B) {
	for n := 0; n < b.N; n++ {
		qi := core.NewQueueInt()
		for i := 0; i < 20; i++ {
			qi.Insert(int64(i))
		}
		for i := 0; i < 20; i++ {
			_, err := qi.Remove()
			if err != nil {
				b.Errorf("queueint test failed. should have been nil but found %v", err)
			}
		}
	}
}

func BenchmarkSize(b *testing.B) {
	for n := 0; n < b.N; n++ {
		qi := core.NewQueueInt()
		for i := 0; i < 20; i++ {
			qi.Insert(int64(i))
		}
		qsize := qi.Length
		if qsize != 20 {
			b.Errorf("queueint test failed. should have been 20 but found %v", qsize)
		}
	}
}
