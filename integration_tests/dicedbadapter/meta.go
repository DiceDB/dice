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
	Flags        map[string]string
	RequiredArgs map[string]Position
	OptionalArgs map[string]string
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
		Flags:        map[string]string{"nx": "nx", "xx": "xx", "keepttl": "keepttl", "get": "get"},
		OptionalArgs: map[string]string{"ex": "ex", "px": "px", "pxat": "pxat", "exat": "exat"},
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
		Flags:        map[string]string{}, // No flags for MGET
		OptionalArgs: map[string]string{}, // No optional args for MGET
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
		Flags:        map[string]string{}, // No flags for MSET
		OptionalArgs: map[string]string{}, // No optional args for MSET
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
		Flags:        map[string]string{},
		OptionalArgs: map[string]string{},
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
		Flags:        map[string]string{},
		OptionalArgs: map[string]string{},
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
		Flags:        map[string]string{},
		OptionalArgs: map[string]string{},
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

	DiceCmdAdapters["SET"] = setCmdAdapterMeta

	DiceCmdAdapters["GET"] = getCmdAdapterMeta

	DiceCmdAdapters["DEL"] = delCmdAdapterMeta

	DiceCmdAdapters["MGET"] = mgetCmdAdapterMeta

	DiceCmdAdapters["MSET"] = msetCmdAdapterMeta

	DiceCmdAdapters["BITOP"] = bitopCmdAdapterMeta

	DiceCmdAdapters["BITFIELD"] = bitfieldCmdAdapterMeta

}
