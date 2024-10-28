package dicedbadapter

type Position struct {
	BeginIndex int
	EndIndex   int
	Step       int
}
type DiceDBAdapterMeta struct {
	Route        string
	Command      string
	Encoder      CommandEncoder
	Decoder      CommandDecoder
	ArgsOrder    []interface{}
	Flags        map[string]int
	RequiredArgs map[string]Position
	OptionalArgs map[string]int
	Subcommands  map[string]DiceDBAdapterMeta
}

var DiceCmdAdapters = map[string]DiceDBAdapterMeta{}

func init() {

	setCmdAdapterMeta := DiceDBAdapterMeta{
		Route:   "/set",
		Command: "SET",
		Encoder: setEncoder,
		Decoder: setDecoder,
		RequiredArgs: map[string]Position{"key": {
			BeginIndex: 0,
			EndIndex:   0,
			Step:       1,
		}, "value": {
			BeginIndex: 1,
			EndIndex:   1,
			Step:       1,
		}},
		Flags:        map[string]int{"nx": 0, "xx": 1, "keepttl": 2, "get": 3},
		OptionalArgs: map[string]int{"ex": 0, "px": 1, "pxat": 2, "exat": 3},
	}

	getCmdAdapterMeta := DiceDBAdapterMeta{

		Route:     "/get",
		Command:   "GET",
		Encoder:   getEncoder,
		Decoder:   getDecoder,
		ArgsOrder: []interface{}{"key"},
		RequiredArgs: map[string]Position{"key": {
			BeginIndex: 0,
			EndIndex:   0,
			Step:       1,
		}},
	}

	delCmdAdapterMeta := DiceDBAdapterMeta{

		Route:   "/del",
		Command: "DEL",
		Encoder: delEncoder,
		Decoder: delDecoder,
		RequiredArgs: map[string]Position{"keys": {
			BeginIndex: 0,
			EndIndex:   -1,
			Step:       1,
		}},
	}

	mgetCmdAdapterMeta := DiceDBAdapterMeta{
		Route:   "/mget",
		Command: "MGET",
		Encoder: mgetEncoder,
		Decoder: mgetDecoder,
		RequiredArgs: map[string]Position{
			"keys": {
				BeginIndex: 0,
				EndIndex:   -1,
				Step:       1,
			},
		}, // MGET requires keys but no specific required args
	}

	msetCmdAdapterMeta := DiceDBAdapterMeta{
		Route:   "/mset",
		Command: "MSET",
		Encoder: msetEncoder,
		Decoder: msetDecoder,
		RequiredArgs: map[string]Position{
			"key-values": {
				BeginIndex: 0,
				EndIndex:   -1,
				Step:       2,
			},
		}, // MSET requires key-value pairs but no specific required args
	}

	bitopCmdAdapterMeta := DiceDBAdapterMeta{
		Route:   "/bitop",
		Command: "BITOP",
		Encoder: bitopEncoder,
		Decoder: bitopDecoder,
		RequiredArgs: map[string]Position{
			"operation": {
				BeginIndex: 0,
				EndIndex:   0,
				Step:       1,
			},
			"destkey": {
				BeginIndex: 1,
				EndIndex:   1,
				Step:       1,
			},
			"keys": {
				BeginIndex: 2,
				EndIndex:   -1,
				Step:       1,
			},
		},
	}
	bitfieldSubcommandGetAdapterMeta := DiceDBAdapterMeta{
		RequiredArgs: map[string]Position{
			"encoding": {
				BeginIndex: 0,
				EndIndex:   0,
				Step:       1,
			},
			"offset": {
				BeginIndex: 1,
				EndIndex:   1,
				Step:       1,
			},
		},
	}

	bitfieldSubcommandSetAdapterMeta := DiceDBAdapterMeta{
		RequiredArgs: map[string]Position{
			"encoding": {
				BeginIndex: 0,
				EndIndex:   0,
				Step:       1,
			},
			"offset": {
				BeginIndex: 1,
				EndIndex:   1,
				Step:       1,
			},
			"value": {
				BeginIndex: 2,
				EndIndex:   2,
				Step:       1,
			},
		},
	}

	bitfieldCmdAdapterMeta := DiceDBAdapterMeta{
		Route:   "/bitfield",
		Command: "BITFIELD",
		Encoder: bitfieldEncoder,
		Decoder: bitfieldDecoder,
		RequiredArgs: map[string]Position{
			"key": {
				BeginIndex: 0,
				EndIndex:   0,
				Step:       1,
			},
		},
		Subcommands: map[string]DiceDBAdapterMeta{
			"get": bitfieldSubcommandGetAdapterMeta,
			"set": bitfieldSubcommandSetAdapterMeta,
			// "incrby":   bitfieldSubcommandIncrbyAdapterMeta,
			// "overflow": bitfieldSubcommandOverflowAdapterMeta,
		},
	}

	zaddCmdAdapterMeta := DiceDBAdapterMeta{
		Route:   "/zadd",
		Command: "ZADD",
		Encoder: zaddEncoder,
		Decoder: zaddDecoder,
		RequiredArgs: map[string]Position{
			"key": {
				BeginIndex: 0,
				EndIndex:   0,
				Step:       1,
			},
			"score-members": {
				BeginIndex: -1,
				EndIndex:   -1,
				Step:       2,
			},
		},
		Flags: map[string]int{"nx": 0, "xx": 1, "gt": 2, "lt": 3, "ch": 4, "incr": 5},
	}

	DiceCmdAdapters["SET"] = setCmdAdapterMeta

	DiceCmdAdapters["GET"] = getCmdAdapterMeta

	DiceCmdAdapters["DEL"] = delCmdAdapterMeta

	DiceCmdAdapters["MGET"] = mgetCmdAdapterMeta

	DiceCmdAdapters["MSET"] = msetCmdAdapterMeta

	DiceCmdAdapters["BITOP"] = bitopCmdAdapterMeta

	DiceCmdAdapters["BITFIELD"] = bitfieldCmdAdapterMeta

	DiceCmdAdapters["ZADD"] = zaddCmdAdapterMeta

}
