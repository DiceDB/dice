// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZPOPMIN(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ZPOPMIN on non-existing key with/without count argument",
			commands: []string{"ZPOPMIN NON_EXISTENT_KEY"},
			expected: []interface{}{[]interface{}{}},
		},
		{
			name:     "ZPOPMIN with wrong type of key with/without count argument",
			commands: []string{"SET stringkey string_value", "ZPOPMIN stringkey", "DEL stringkey"},
			expected: []interface{}{"OK", "WRONGTYPE Operation against a key holding the wrong kind of value", int64(1)},
		},
		{
			name:     "ZPOPMIN on existing key (without count argument)",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset", "ZCOUNT myzset  1 10"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1"}, int64(2)},
		},
		{
			name:     "ZPOPMIN with normal count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 2", "ZCOUNT myzset  1 2"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "2"}, int64(0)},
		},
		{
			name:     "ZPOPMIN with count argument but multiple members have the same score",
			commands: []string{"ZADD myzset 1 member1 1 member2 1 member3", "ZPOPMIN myzset 2", "ZCOUNT myzset 1 1"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "1"}, int64(1)},
		},
		{
			name:     "ZPOPMIN with negative count argument",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset -1", "ZCOUNT myzset 0.6 3.231"},
			expected: []interface{}{int64(3), []interface{}{}, int64(3)},
		},
		{
			name:     "ZPOPMIN with invalid count argument",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset INCORRECT_COUNT_ARGUMENT", "ZCOUNT myzset 1 10"},
			expected: []interface{}{int64(1), "ERR value is not an integer or out of range", int64(1)},
		},
		{
			name:     "ZPOPMIN with count argument greater than length of sorted set",
			commands: []string{"ZADD myzset 1 member1 2 member2 3 member3", "ZPOPMIN myzset 10", "ZCOUNT myzset 1 10"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1", "member2", "2", "member3", "3"}, int64(0)},
		},
		{
			name:     "ZPOPMIN on empty sorted set",
			commands: []string{"ZADD myzset 1 member1", "ZPOPMIN myzset 1", "ZPOPMIN myzset", "ZCOUNT myzset  0 10000"},
			expected: []interface{}{int64(1), []interface{}{"member1", "1"}, []interface{}{}, int64(0)},
		},
		{
			name:     "ZPOPMIN with floating-point scores",
			commands: []string{"ZADD myzset 1.5 member1 2.7 member2 3.8 member3", "ZPOPMIN myzset", "ZCOUNT myzset 1.499 2.711"},
			expected: []interface{}{int64(3), []interface{}{"member1", "1.5"}, int64(1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
			client.FireString("DEL myzset")
		})
	}
}

func TestZCOUNT(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ZCOUNT on non-existing key",
			commands: []string{"ZCOUNT NON_EXISTENT_KEY 0 100"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZCOUNT with valid key and range",
			commands: []string{"ZADD myzset 10 member1 20 member2 30 member3", "ZCOUNT myzset 15 25"},
			expected: []interface{}{int64(3), int64(1)}, // Expecting count of 1 from ZCOUNT
		},
		{
			name:     "ZCOUNT with range including all members",
			commands: []string{"ZADD myzset 10 member1 20 member2 30 member3", "ZCOUNT myzset 0 100"},
			expected: []interface{}{int64(3), int64(3)}, // Expecting count of 3 from ZCOUNT
		},
		{
			name:     "ZCOUNT with min greater than max",
			commands: []string{"ZADD myzset 10 member1 20 member2 30 member3", "ZCOUNT myzset 25 15"},
			expected: []interface{}{int64(3), int64(0)}, // Expecting count of 0 from ZCOUNT
		},
		{
			name:     "ZCOUNT with negative range",
			commands: []string{"ZADD myzset -5 member1 0 member2 5 member3", "ZCOUNT myzset -10 -1"},
			expected: []interface{}{int64(3), int64(1)}, // Expecting count of 3 from ZCOUNT
		},
		{
			name:     "ZCOUNT on empty sorted set",
			commands: []string{"ZADD myzset 1 member1", "DEL myzset", "ZCOUNT myzset 0 100"},
			expected: []interface{}{int64(1), int64(1), int64(0)}, // Expecting count of 0 from ZCOUNT
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL myzset") // Resetting the key before each test

			// post cleanup
			t.Cleanup(func() {
				client.FireString("DEL myzset")
			})

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestZADD(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ZADD with two new members",
			commands: []string{"ZADD key 1 member1 2 member2"},
			expected: []interface{}{int64(2)},
		},
		{
			name:     "ZADD with three new members",
			commands: []string{"ZADD key 3 member3 4 member4 5 member5"},
			expected: []interface{}{int64(3)},
		},
		{
			name:     "ZADD with existing members",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5"},
			expected: []interface{}{int64(5)},
		},
		{
			name:     "ZADD with mixed new and existing members",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6"},
			expected: []interface{}{int64(6)},
		},
		{
			name:     "ZADD without any members",
			commands: []string{"ZADD key 1"},
			expected: []interface{}{"ERR wrong number of arguments for 'zadd' command"},
		},

		// *************************************** ZADD with XX options validation starts now, including XX with GT, LT, NX, INCR, CH **************************
		{
			name:     "ZADD XX option without existing key",
			commands: []string{"ZADD key XX 10 member9"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX with existing key and member2",
			commands: []string{"ZADD key XX 20 member2"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX updates existing elements scores",
			commands: []string{"ZADD key XX 15 member1 25 member2"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD GT and XX only updates existing elements when new scores are greater",
			commands: []string{"ZADD key GT XX 20 member1"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD LT and XX only updates existing elements when new scores are lower",
			commands: []string{"ZADD key LT XX 20 member1"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD NX and XX not compatible",
			commands: []string{"ZADD key NX XX 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD XX and CH compatible",
			commands: []string{"ZADD key XX CH 20 member1"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD INCR and XX compatible",
			commands: []string{"ZADD key XX INCR 20 member1"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD INCR and XX not compatible because of more than one member",
			commands: []string{"ZADD key XX INCR 20 member1 25 member2"},
			expected: []interface{}{"ERR INCR option supports a single increment-element pair"},
		},

		{
			name:     "ZADD XX, LT and GT are not compatible",
			commands: []string{"ZADD key XX LT GT 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD XX, LT, GT, CH, INCR are not compatible",
			commands: []string{"ZADD key XX LT GT INCR CH 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},

		{
			name:     "ZADD XX, GT and CH compatible",
			commands: []string{"ZADD key XX GT CH 60 member1 30 member2"},
			expected: []interface{}{int64(0)},
		},

		{
			name:     "ZADD XX, LT and CH compatible",
			commands: []string{"ZADD key XX LT CH 4 member1 1 member2"},
			expected: []interface{}{int64(0)},
		},

		//running with new members, XX wont update with new members, only existing gets updated

		{
			name:     "ZADD XX with existing key and new member",
			commands: []string{"ZADD key XX 20 member20"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX wont update as new members",
			commands: []string{"ZADD key XX 15 member18 25 member20"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX and GT wont add new member",
			commands: []string{"ZADD key GT XX 20 member18"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX and LT and new member wont update",
			commands: []string{"ZADD key LT XX 20 member18"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX and CH and new member wont work",
			commands: []string{"ZADD key XX CH 20 member18"},
			expected: []interface{}{int64(0)},
		},

		{
			name:     "ZADD XX, LT, CH, new member wont update",
			commands: []string{"ZADD key XX LT CH 50 member18 40 member20"},
			expected: []interface{}{int64(0)},
		},
		{
			name:     "ZADD XX, GT and CH, new member wont update",
			commands: []string{"ZADD key XX GT CH 60 member18 30 member20"},
			expected: []interface{}{int64(0)},
		},

		// *******************************************   ZADD with NX starts now, including GT, LT, XX, INCR, CH    ***************

		{
			name:     "ZADD NX existing key new member",
			commands: []string{"ZADD key NX 10 member9"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD NX existing key old member",
			commands: []string{"ZADD key NX 20 member2"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD NX existing key one new member and one old member",
			commands: []string{"ZADD key NX 15 member1 25 member11"},
			expected: []interface{}{int64(2)},
		},

		// NX and XX with all LT GT CH and INCR all errors
		{
			name:     "ZADD NX and XX not compatible",
			commands: []string{"ZADD key NX XX 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX CH not compatible",
			commands: []string{"ZADD key NX XX CH 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX CH INCR not compatible",
			commands: []string{"ZADD key NX XX CH INCR 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX LT not compatible",
			commands: []string{"ZADD key NX XX LT 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX GT not compatible",
			commands: []string{"ZADD key NX XX GT 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX LT CH not compatible",
			commands: []string{"ZADD key NX XX LT CH 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX LT CH INCR compatible",
			commands: []string{"ZADD key NX XX LT CH INCR 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX GT CH not compatible",
			commands: []string{"ZADD key NX XX GT CH 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX GT CH INCR not compatible",
			commands: []string{"ZADD key NX XX GT CH INCR 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX INCR not compatible",
			commands: []string{"ZADD key NX XX INCR 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX INCR LT not compatible",
			commands: []string{"ZADD key NX XX INCR LT 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX INCR GT not compatible",
			commands: []string{"ZADD key NX XX INCR GT 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX LT GT not compatible",
			commands: []string{"ZADD key NX XX LT GT 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX LT GT CH not compatible",
			commands: []string{"ZADD key NX XX LT GT CH 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX XX LT GT CH INCR not compatible",
			commands: []string{"ZADD key NX XX LT GT CH INCR 20 member1"},
			expected: []interface{}{"ERR XX and NX options at the same time are not compatible"},
		},

		// NX without XX and all LT GT CH and INCR // all are error
		{
			name:     "ZADD NX and GT incompatible",
			commands: []string{"ZADD key NX GT 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX and LT incompatible",
			commands: []string{"ZADD key NX LT 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT and GT incompatible",
			commands: []string{"ZADD key NX LT GT 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT, GT and INCR incompatible",
			commands: []string{"ZADD key NX LT GT INCR 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT, GT and CH incompatible",
			commands: []string{"ZADD key NX LT GT CH 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT, GT, CH and INCR incompatible",
			commands: []string{"ZADD key NX LT GT CH INCR 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT, CH not compatible",
			commands: []string{"ZADD key NX LT CH 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT, INCR not compatible",
			commands: []string{"ZADD key NX LT INCR 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, LT, CH, INCR not compatible",
			commands: []string{"ZADD key NX LT CH INCR 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, GT, CH not compatible",
			commands: []string{"ZADD key NX GT CH 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, GT, INCR not compatible",
			commands: []string{"ZADD key NX GT INCR 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD NX, GT, CH, INCR not compatible",
			commands: []string{"ZADD key NX GT CH INCR 20 member1"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},

		{
			name:     "ZADD NX, CH with new member returns CH based - if added or not",
			commands: []string{"ZADD key NX CH 20 member13"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD NX, CH with existing member returns CH based - if added or not",
			commands: []string{"ZADD key NX CH 10 member13"},
			expected: []interface{}{int64(1)},
		},

		// *************************************** ZADD with GT options validation starts now, including GT with XX, LT, NX, INCR, CH **************************

		{
			name:     "ZADD with GT with existing member",
			commands: []string{"ZADD key GT 15 member14"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD with GT with new member",
			commands: []string{"ZADD key GT 15 member15"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD GT and LT",
			commands: []string{"ZADD key GT LT 15 member15"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD GT LT CH",
			commands: []string{"ZADD key GT LT CH 15 member15"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD GT LT CH INCR",
			commands: []string{"ZADD key GT LT CH INCR 15 member15"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD GT LT INCR",
			commands: []string{"ZADD key GT LT INCR 15 member15"},
			expected: []interface{}{"ERR GT, LT, and/or NX options at the same time are not compatible"},
		},
		{
			name:     "ZADD GT CH with existing member score less no change hence 0",
			commands: []string{"ZADD key GT CH 10 member15"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD GT CH with existing member score more, changed score hence 1",
			commands: []string{"ZADD key GT CH 25 member15"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD GT CH with existing member score equal, nothing returned",
			commands: []string{"ZADD key GT CH 25 member15"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD GT CH with new member score",
			commands: []string{"ZADD key GT CH 5 member19"},
			expected: []interface{}{int64(1)},
		},
		{
			name:     "ZADD GT with INCR if score less than currentscore after INCR returns nil",
			commands: []string{"ZADD key GT INCR -5 member15"},
			expected: []interface{}{int64(-5)},
		},
		{
			name:     "ZADD GT with INCR updates existing member score if greater after INCR",
			commands: []string{"ZADD key GT INCR 5 member15"},
			expected: []interface{}{int64(5)},
		},

		// *************************************** ZADD with LT options validation starts now, including LT with GT, XX, NX, INCR, CH **************************

		{
			name:     "ZADD with LT with existing member score greater",
			commands: []string{"ZADD key LT 15 member14"},
			expected: []interface{}{int64(1)},
		},

		{
			name:     "ZADD with LT with new member",
			commands: []string{"ZADD key LT 15 member23"},
			expected: []interface{}{int64(1)},
		},

		{
			name:     "ZADD LT with existing member score equal",
			commands: []string{"ZADD key LT 15 member14"},
			expected: []interface{}{int64(1)},
		},

		{
			name:     "ZADD LT with existing member score less",
			commands: []string{"ZADD key LT 10 member14"},
			expected: []interface{}{int64(1)},
		},

		{
			name:     "ZADD LT with INCR not updates existing member as score is greater after INCR",
			commands: []string{"ZADD key LT INCR 5 member14"},
			expected: []interface{}{int64(5)},
		},

		{
			name:     "ZADD LT with INCR updates existing member as updatedscore after INCR is less than current",
			commands: []string{"ZADD key LT INCR -1 member14"},
			expected: []interface{}{int64(-1)},
		},

		{
			name:     "ZADD LT with CH updates existing member score if less, CH returns changed elements",
			commands: []string{"ZADD key LT CH 5 member1 2 member2"},
			expected: []interface{}{int64(2)},
		},

		// *************************************** ZADD with INCR options validation starts now, including INCR with GT, LT, NX, XX, CH **************************
		{
			name:     "ZADD INCR with new members, insert as it is ",
			commands: []string{"ZADD key INCR 15 member24"},
			expected: []interface{}{int64(15)},
		},

		{
			name:     "ZADD INCR with existing members, increase the score",
			commands: []string{"ZADD key INCR 5 member24"},
			expected: []interface{}{int64(5)},
		},

		// *************************************** ZADD with CH options validation starts now, including CH with GT, LT, NX, XX, INCR **************************
		{
			name:     "ZADD CH with one existing members update, returns count of updation",
			commands: []string{"ZADD key CH 45 member2"},
			expected: []interface{}{int64(1)},
		},

		{
			name:     "ZADD CH with multiple existing members update, returns count of updation",
			commands: []string{"ZADD key CH 50 member2 63 member3"},
			expected: []interface{}{int64(2)},
		},

		{
			name:     "ZADD CH with 1 new and 1 existing member update, returns count of updation",
			commands: []string{"ZADD key CH 50 member2 64 member32"},
			expected: []interface{}{int64(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL key") // Resetting the key before each test

			// post cleanup
			t.Cleanup(func() {
				client.FireString("DEL key")
			})

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}

func TestZRANGE(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []TestCase{
		{
			name:     "ZRANGE with mixed indices",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 0 -1"},
			expected: []interface{}{int64(6), []interface{}{"member1", "member2", "member3", "member4", "member5", "member6"}},
		},
		{
			name:     "ZRANGE with positive indices #1",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 0 2"},
			expected: []interface{}{int64(6), []interface{}{"member1", "member2", "member3"}},
		},
		{
			name:     "ZRANGE with positive indices #2",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 2 4"},
			expected: []interface{}{int64(6), []interface{}{"member3", "member4", "member5"}},
		},
		{
			name:     "ZRANGE with all positive indices",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 0 10"},
			expected: []interface{}{int64(6), []interface{}{"member1", "member2", "member3", "member4", "member5", "member6"}},
		},
		{
			name:     "ZRANGE with out of bound indices",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 10 20"},
			expected: []interface{}{int64(6), []interface{}{}},
		},
		{
			name:     "ZRANGE with positive indices and scores",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 0 10 WITHSCORES"},
			expected: []interface{}{int64(6), []interface{}{"member1", "1", "member2", "2", "member3", "3", "member4", "4", "member5", "5", "member6", "6"}},
		},
		{
			name:     "ZRANGE with positive indices and scores in reverse order",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key 0 10 REV WITHSCORES"},
			expected: []interface{}{int64(6), []interface{}{"member6", "6", "member5", "5", "member4", "4", "member3", "3", "member2", "2", "member1", "1"}},
		},
		{
			name:     "ZRANGE with negative indices",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key -1 -1"},
			expected: []interface{}{int64(6), []interface{}{"member6"}},
		},
		{
			name:     "ZRANGE with negative indices and scores",
			commands: []string{"ZADD key 1 member1 2 member2 3 member3 4 member4 5 member5 6 member6", "ZRANGE key -8 -5 WITHSCORES"},
			expected: []interface{}{int64(6), []interface{}{"member1", "1", "member2", "2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client.FireString("DEL key") // Resetting the key before each test

			// post cleanup
			t.Cleanup(func() {
				client.FireString("DEL key")
			})

			for i, cmd := range tc.commands {
				result := client.FireString(cmd)
				assert.Equal(t, tc.expected[i], result)
			}
		})
	}
}
