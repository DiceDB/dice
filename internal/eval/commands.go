// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package eval

import (
	"strings"

	"github.com/dicedb/dice/internal/cmd"
	dstore "github.com/dicedb/dice/internal/store"
)

type DiceCmdMeta struct {
	Name  string
	Info  string
	Eval  func([]string, *dstore.Store) []byte
	Arity int // number of arguments, it is possible to use -N to say >= N
	KeySpecs
	SubCommands []string // list of sub-commands supported by the command

	// IsMigrated indicates whether a command has been migrated to a new evaluation
	// mechanism. If true, the command uses the newer evaluation logic represented by
	// the NewEval function. This allows backward compatibility for commands that have
	// not yet been migrated, ensuring they continue to use the older Eval function.
	// As part of the transition process, commands can be flagged with IsMigrated to
	// signal that they are using the updated execution path.
	IsMigrated bool

	// NewEval is the newer evaluation function for commands. It follows an updated
	// execution model that returns an EvalResponse struct, offering more structured
	// and detailed results, including metadata such as errors and additional info,
	// instead of just raw bytes. Commands that have been migrated to this new model
	// will utilize this function for evaluation, allowing for better handling of
	// complex command execution scenarios and improved response consistency.
	NewEval func([]string, *dstore.Store) *EvalResponse

	// StoreObjectEval is a specialized evaluation function for commands that operate on an object.
	// It is designed for scenarios where the command and subsequent dependent command requires
	// an object as part of its execution. This function processes the command,
	// evaluates it based on the provided object, and returns an EvalResponse struct
	// Commands that involve object manipulation, is not recommended for general use.
	// Only commands that really requires full object definition to pass across multiple shards
	// should implement this function. e.g. COPY, RENAME etc
	StoreObjectEval func(*cmd.DiceDBCmd, *dstore.Store) *EvalResponse
}

type KeySpecs struct {
	BeginIndex int
	Step       int
	LastKey    int
}

var (
	PreProcessing = map[string]func([]string, *dstore.Store) *EvalResponse{}
	DiceCmds      = map[string]DiceCmdMeta{}
)

// Custom Commands:
// This command type allows for flexibility in defining and executing specific,
// non-standard operations. Each command has metadata that specifies its behavior
// and execution logic (Eval function). While the RESP
// server supports custom logic for these commands and treats them as CUSTOM commands,
// their implementation for HTTP and WebSocket protocols is still pending.
// As a result, their Eval functions remain defined but not yet migrated.
var (
	helloCmdMeta = DiceCmdMeta{
		Name:  "HELLO",
		Info:  `HELLO always replies with a list of current server and connection properties, such as: versions, modules loaded, client ID, replication role and so forth`,
		Eval:  evalHELLO,
		Arity: -1,
	}
	authCmdMeta = DiceCmdMeta{
		Name: "AUTH",
		Info: `AUTH returns with an encoded "OK" if the user is authenticated.
		If the user is not authenticated, it returns with an encoded error message`,
		Eval: nil,
	}
	abortCmdMeta = DiceCmdMeta{
		Name:  "ABORT",
		Info:  "Quit the server",
		Eval:  nil,
		Arity: 1,
	}
	sleepCmdMeta = DiceCmdMeta{
		Name: "SLEEP",
		Info: `SLEEP sets db to sleep for the specified number of seconds.
		The sleep time should be the only param in args.
		Returns error response if the time param in args is not of integer format.
		SLEEP returns RespOK after sleeping for mentioned seconds`,
		Eval:  evalSLEEP,
		Arity: 1,
	}
)

// Multi Shard or All Shard Commands:
// This command type allows to do operations across multiple shards and gather the results
// While the RESP server supports scatter-gather logic for these commands and
// treats them as MultiShard or commands,
// their implementation for HTTP and WebSocket protocols is still pending.
// As a result, their Eval functions remained intact.
var (
	objectCopyCmdMeta = DiceCmdMeta{
		Name:            "OBJECTCOPY",
		Info:            `COPY command copies the value stored at the source key to the destination key.`,
		StoreObjectEval: evalCOPYObject,
		IsMigrated:      true,
		Arity:           -2,
	}
	pfMergeCmdMeta = DiceCmdMeta{
		Name: "PFMERGE",
		Info: `PFMERGE destkey [sourcekey [sourcekey ...]]
		Merges one or more HyperLogLog values into a single key.`,
		IsMigrated:      true,
		Arity:           -2,
		KeySpecs:        KeySpecs{BeginIndex: 1},
		StoreObjectEval: evalPFMERGE,
	}
)

// Single Shard command
// This command type executes within a single shard and no custom logic is required to
// compose the results. Although http and websocket always uses shard - 0 still logic
// remains same in case of RESP as well. As a result following commands are successfully migrated
var (
	jsonsetCmdMeta = DiceCmdMeta{
		Name: "JSON.SET",
		Info: `JSON.SET key path json-string
		Sets a JSON value at the specified key.
		Returns OK if successful.
		Returns encoded error message if the number of arguments is incorrect or the JSON string is invalid.`,
		NewEval:    evalJSONSET,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsongetCmdMeta = DiceCmdMeta{
		Name: "JSON.GET",
		Info: `JSON.GET key [path]
		Returns the encoded RESP value of the key, if present
		Null reply: If the key doesn't exist or has expired.
		Error reply: If the number of arguments is incorrect or the stored value is not a JSON type.`,
		NewEval:    evalJSONGET,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsontoggleCmdMeta = DiceCmdMeta{
		Name: "JSON.TOGGLE",
		Info: `JSON.TOGGLE key [path]
		Toggles Boolean values between true and false at the path.Return
		If the path is enhanced syntax:
    	1.Array of integers (0 - false, 1 - true) that represent the resulting Boolean value at each path.
	    2.If a value is a not a Boolean value, its corresponding return value is null.
		3.NONEXISTENT if the document key does not exist.
		If the path is restricted syntax:
    	1.String ("true"/"false") that represents the resulting Boolean value.
    	2.NONEXISTENT if the document key does not exist.
    	3.WRONGTYPE error if the value at the path is not a Boolean value.`,
		NewEval:    evalJSONTOGGLE,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsontypeCmdMeta = DiceCmdMeta{
		Name: "JSON.TYPE",
		Info: `JSON.TYPE key [path]
		Returns string reply for each path, specified as the value's type.
		Returns RespNIL If the key doesn't exist.
		Error reply: If the number of arguments is incorrect.`,
		NewEval:    evalJSONTYPE,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsonclearCmdMeta = DiceCmdMeta{
		Name: "JSON.CLEAR",
		Info: `JSON.CLEAR key [path]
		Returns an integer reply specifying the number ofmatching JSON arrays and
		objects cleared +number of matching JSON numerical values zeroed.
		Error reply: If the number of arguments is incorrect the key doesn't exist.`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalJSONCLEAR,
	}
	jsondelCmdMeta = DiceCmdMeta{
		Name: "JSON.DEL",
		Info: `JSON.DEL key [path]
		Returns an integer reply specified as the number of paths deleted (0 or more).
		Returns RespZero if the key doesn't exist or key is expired.
		Error reply: If the number of arguments is incorrect.`,
		NewEval:    evalJSONDEL,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsonarrappendCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRAPPEND",
		Info: `JSON.ARRAPPEND key [path] value [value ...]
        Returns an array of integer replies for each path, the array's new size,
        or nil, if the matching JSON value is not an array.`,
		Arity:      -3,
		IsMigrated: true,
		NewEval:    evalJSONARRAPPEND,
	}
	jsonforgetCmdMeta = DiceCmdMeta{
		Name: "JSON.FORGET",
		Info: `JSON.FORGET key [path]
		Returns an integer reply specified as the number of paths deleted (0 or more).
		Returns RespZero if the key doesn't exist or key is expired.
		Error reply: If the number of arguments is incorrect.`,
		NewEval:    evalJSONFORGET,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsonarrlenCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRLEN",
		Info: `JSON.ARRLEN key [path]
		Returns an array of integer replies.
		Returns error response if the key doesn't exist or key is expired or the matching value is not an array.
		Error reply: If the number of arguments is incorrect.`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalJSONARRLEN,
	}
	jsonnummultbyCmdMeta = DiceCmdMeta{
		Name: "JSON.NUMMULTBY",
		Info: `JSON.NUMMULTBY key path value
		Multiply the number value stored at the specified path by a value.`,
		NewEval:    evalJSONNUMMULTBY,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsonobjlenCmdMeta = DiceCmdMeta{
		Name: "JSON.OBJLEN",
		Info: `JSON.OBJLEN key [path]
		Report the number of keys in the JSON object at path in key
		Returns error response if the key doesn't exist or key is expired or the matching value is not an array.
		Error reply: If the number of arguments is incorrect.`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalJSONOBJLEN,
	}
	jsondebugCmdMeta = DiceCmdMeta{
		Name: "JSON.DEBUG",
		Info: `evaluates JSON.DEBUG subcommand based on subcommand
		JSON.DEBUG MEMORY returns memory usage by key in bytes
		JSON.DEBUG HELP displays help message
		`,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalJSONDebug,
	}
	jsonobjkeysCmdMeta = DiceCmdMeta{
		Name: "JSON.OBJKEYS",
		Info: `JSON.OBJKEYS key [path]
		Retrieves the keys of a JSON object stored at path specified.
		Null reply: If the key doesn't exist or has expired.
		Error reply: If the number of arguments is incorrect or the stored value is not a JSON type.`,
		NewEval:    evalJSONOBJKEYS,
		IsMigrated: true,
		Arity:      2,
	}
	jsonarrpopCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRPOP",
		Info: `JSON.ARRPOP key [path [index]]
		Removes and returns an element from the index in the array and updates the array in memory.
		Returns error if key doesn't exist.
		Return nil if array is empty or there is no array at the path.
		It supports negative index and is out of bound safe.
		`,
		Arity:      -2,
		IsMigrated: true,
		NewEval:    evalJSONARRPOP,
	}
	jsoningestCmdMeta = DiceCmdMeta{
		Name: "JSON.INGEST",
		Info: `JSON.INGEST key_prefix json-string
		The whole key is generated by appending a unique identifier to the provided key prefix.
		the generated key is then used to store the provided JSON value at specified path.
		Returns unique identifier if successful.
		Returns encoded error message if the number of arguments is incorrect or the JSON string is invalid.`,
		NewEval:    evalJSONINGEST,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	jsonarrinsertCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRINSERT",
		Info: `JSON.ARRINSERT key path index value [value ...]
		Returns an array of integer replies for each path.
		Returns nil if the matching JSON value is not an array.
		Returns error response if the key doesn't exist or key is expired or the matching value is not an array.
		Error reply: If the number of arguments is incorrect.`,
		NewEval:    evalJSONARRINSERT,
		IsMigrated: true,
		Arity:      -5,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	jsonrespCmdMeta = DiceCmdMeta{
		Name: "JSON.RESP",
		Info: `JSON.RESP key [path]
		Return the JSON present at key`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalJSONRESP,
	}
	jsonarrtrimCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRTRIM",
		Info: `JSON.ARRTRIM key path start stop
		Trim an array so that it contains only the specified inclusive range of elements
		Returns an array of integer replies for each path.
		Returns error response if the key doesn't exist or key is expired.
		Error reply: If the number of arguments is incorrect.`,
		NewEval:    evalJSONARRTRIM,
		IsMigrated: true,
		Arity:      -5,
	}
	incrByFloatCmdMeta = DiceCmdMeta{
		Name: "INCRBYFLOAT",
		Info: `INCRBYFLOAT increments the value of the key in args by the specified increment,
		if the key exists and the value is a number.
		The key should be the first parameter in args, and the increment should be the second parameter.
		If the key does not exist, a new key is created with increment's value.
		If the value at the key is a string, it should be parsable to float64,
		if not INCRBYFLOAT returns an  error response.
		INCRBYFLOAT returns the incremented value for the key after applying the specified increment if there are no errors.`,
		Arity:      2,
		NewEval:    evalINCRBYFLOAT,
		IsMigrated: true,
	}
	clientCmdMeta = DiceCmdMeta{
		Name:       "CLIENT",
		Info:       `This is a container command for client connection commands.`,
		NewEval:    evalCLIENT,
		IsMigrated: true,
		Arity:      -2,
	}
	latencyCmdMeta = DiceCmdMeta{
		Name:       "LATENCY",
		Info:       `This is a container command for latency diagnostics commands.`,
		NewEval:    evalLATENCY,
		IsMigrated: true,
		Arity:      -2,
	}
	bfreserveCmdMeta = DiceCmdMeta{
		Name: "BF.RESERVE",
		Info: `BF.RESERVE command initializes a new bloom filter and allocation it's relevant parameters based on given inputs.
		If no params are provided, it uses defaults.`,
		IsMigrated: true,
		NewEval:    evalBFRESERVE,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfaddCmdMeta = DiceCmdMeta{
		Name: "BF.ADD",
		Info: `BF.ADD adds an element to
		a bloom filter. If the filter does not exists, it will create a new one
		with default parameters.`,
		IsMigrated: true,
		NewEval:    evalBFADD,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfexistsCmdMeta = DiceCmdMeta{
		Name:       "BF.EXISTS",
		Info:       `BF.EXISTS checks existence of an element in a bloom filter.`,
		NewEval:    evalBFEXISTS,
		IsMigrated: true,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfinfoCmdMeta = DiceCmdMeta{
		Name:       "BF.INFO",
		Info:       `BF.INFO returns the parameters and metadata of an existing bloom filter.`,
		NewEval:    evalBFINFO,
		IsMigrated: true,
		Arity:      2,
	}
	setBitCmdMeta = DiceCmdMeta{
		Name:       "SETBIT",
		Info:       "SETBIT sets or clears the bit at offset in the string value stored at key",
		IsMigrated: true,
		NewEval:    evalSETBIT,
	}
	getBitCmdMeta = DiceCmdMeta{
		Name:       "GETBIT",
		Info:       "GETBIT returns the bit value at offset in the string value stored at key",
		IsMigrated: true,
		NewEval:    evalGETBIT,
	}
	bitCountCmdMeta = DiceCmdMeta{
		Name:       "BITCOUNT",
		Info:       "BITCOUNT counts the number of set bits in the string value stored at key",
		Arity:      -1,
		IsMigrated: true,
		NewEval:    evalBITCOUNT,
	}

	persistCmdMeta = DiceCmdMeta{
		Name:       "PERSIST",
		Info:       "PERSIST removes the expiration from a key",
		IsMigrated: true,
		NewEval:    evalPERSIST,
	}

	commandCmdMeta = DiceCmdMeta{
		Name:        "COMMAND",
		Info:        "Evaluates COMMAND <subcommand> command based on subcommand",
		NewEval:     evalCommand,
		IsMigrated:  true,
		Arity:       -1,
		SubCommands: []string{Count, GetKeys, GetKeysandFlags, List, Help, Info, Docs},
	}
	commandCountCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|COUNT",
		Info:       "Returns a count of commands.",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      2,
	}
	commandHelpCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|HELP",
		Info:       "Returns helpful text about the different subcommands",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      2,
	}
	commandInfoCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|INFO",
		Info:       "Returns information about one, multiple or all commands.",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      -2,
	}
	commandListCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|LIST",
		Info:       "Returns a list of command names.",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      -2,
	}
	commandDocsCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|DOCS",
		Info:       "Returns documentary information about one, multiple or all commands.",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      -2,
	}
	commandGetKeysCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|GETKEYS",
		Info:       "Extracts the key names from an arbitrary command.",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      -4,
	}
	commandGetKeysAndFlagsCmdMeta = DiceCmdMeta{
		Name:       "COMMAND|GETKEYSANDFLAGS",
		Info:       "Returns a list of command names.",
		NewEval:    evalCommand,
		IsMigrated: true,
		Arity:      -4,
	}
	jsonArrIndexCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRINDEX",
		Info: `JSON.ARRINDEX key path value [start [stop]]
		Search for the first occurrence of a JSON value in an array`,
		NewEval:    evalJSONARRINDEX,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}

	// Internal command used to spawn request across all shards (works internally with the KEYS command)
	singleKeysCmdMeta = DiceCmdMeta{
		Name:       "SINGLEKEYS",
		Info:       "KEYS command is used to get all the keys in the database. Complexity is O(n) where n is the number of keys in the database.",
		NewEval:    evalKEYS,
		Arity:      1,
		IsMigrated: true,
	}
	pttlCmdMeta = DiceCmdMeta{
		Name: "PTTL",
		Info: `PTTL returns Time-to-Live in millisecs for the queried key in args
		The key should be the only param in args else returns with an error
		Returns
		RESP encoded time (in secs) remaining for the key to expire
		RESP encoded -2 stating key doesn't exist or key is expired
		RESP encoded -1 in case no expiration is set on the key`,
		NewEval:    evalPTTL,
		IsMigrated: true,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	hsetCmdMeta = DiceCmdMeta{
		Name: "HSET",
		Info: `HSET sets the specific fields to their respective values in the
		hash stored at key. If any given field is already present, the previous
		value will be overwritten with the new value
		Returns
		This command returns the number of keys that are stored at given key.
		`,
		NewEval:    evalHSET,
		Arity:      -4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hmsetCmdMeta = DiceCmdMeta{
		Name: "HMSET",
		Info: `HSET sets the specific fields to their respective values in the
		hash stored at key. If any given field is already present, the previous
		value will be overwritten with the new value
		Returns
		This command returns the number of keys that are stored at given key.
		`,
		NewEval:    evalHMSET,
		Arity:      -4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hkeysCmdMeta = DiceCmdMeta{
		Name:       "HKEYS",
		Info:       `HKEYS command is used to retrieve all the keys(or field names) within a hash. Complexity is O(n) where n is the size of the hash.`,
		NewEval:    evalHKEYS,
		Arity:      1,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hsetnxCmdMeta = DiceCmdMeta{
		Name: "HSETNX",
		Info: `Sets field in the hash stored at key to value, only if field does not yet exist.
		If key does not exist, a new key holding a hash is created. If field already exists,
		this operation has no effect.`,
		NewEval:    evalHSETNX,
		Arity:      4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hgetCmdMeta = DiceCmdMeta{
		Name:       "HGET",
		Info:       `Returns the value associated with field in the hash stored at key.`,
		NewEval:    evalHGET,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hmgetCmdMeta = DiceCmdMeta{
		Name:       "HMGET",
		Info:       `Returns the values associated with the specified fields in the hash stored at key.`,
		NewEval:    evalHMGET,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hgetAllCmdMeta = DiceCmdMeta{
		Name: "HGETALL",
		Info: `Returns all fields and values of the hash stored at key. In the returned value,
        every field name is followed by its value, so the length of the reply is twice the size of the hash.`,
		NewEval:    evalHGETALL,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hValsCmdMeta = DiceCmdMeta{
		Name:       "HVALS",
		Info:       `Returns all values of the hash stored at key. The length of the reply is same as the size of the hash.`,
		NewEval:    evalHVALS,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hincrbyCmdMeta = DiceCmdMeta{
		Name: "HINCRBY",
		Info: `Increments the number stored at field in the hash stored at key by increment.
		If key does not exist, a new key holding a hash is created.
		If field does not exist the value is set to 0 before the operation is performed.`,
		Arity:      -4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalHINCRBY,
	}
	hstrLenCmdMeta = DiceCmdMeta{
		Name:       "HSTRLEN",
		Info:       `Returns the length of value associated with field in the hash stored at key.`,
		NewEval:    evalHSTRLEN,
		IsMigrated: true,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}

	hdelCmdMeta = DiceCmdMeta{
		Name: "HDEL",
		Info: `HDEL removes the specified fields from the hash stored at key.
		Specified fields that do not exist within this hash are ignored.
		Deletes the hash if no fields remain.
		If key does not exist, it is treated as an empty hash and this command returns 0.
		Returns
		The number of fields that were removed from the hash, not including specified but non-existing fields.`,
		NewEval:    evalHDEL,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	hscanCmdMeta = DiceCmdMeta{
		Name: "HSCAN",
		Info: `HSCAN is used to iterate over fields and values of a hash.
		It returns a cursor and a list of key-value pairs.
		The cursor is used to paginate through the hash.
		The command returns a cursor value of 0 when all the elements are iterated.`,
		NewEval:    evalHSCAN,
		IsMigrated: true,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	hexistsCmdMeta = DiceCmdMeta{
		Name:       "HEXISTS",
		Info:       `Returns if field is an existing field in the hash stored at key.`,
		NewEval:    evalHEXISTS,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}

	objectCmdMeta = DiceCmdMeta{
		Name: "OBJECT",
		Info: `OBJECT subcommand [arguments [arguments ...]]
		OBJECT command is used to inspect the internals of the DiceDB objects.`,
		NewEval:    evalOBJECT,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 2},
		IsMigrated: true,
	}

	// Internal command used to spawn request across all shards (works internally with Touch command)
	singleTouchCmdMeta = DiceCmdMeta{
		Name: "SINGLETOUCH",
		Info: `TOUCH key1
		Alters the last access time of a key(s).
		A key is ignored if it does not exist.
		This is for one by one counting and for multisharding`,
		NewEval:    evalTouch,
		IsMigrated: true,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}

	lpushCmdMeta = DiceCmdMeta{
		Name:       "LPUSH",
		Info:       "LPUSH pushes values into the left side of the deque",
		NewEval:    evalLPUSH,
		IsMigrated: true,
		Arity:      -3,
	}
	rpushCmdMeta = DiceCmdMeta{
		Name:       "RPUSH",
		Info:       "RPUSH pushes values into the right side of the deque",
		NewEval:    evalRPUSH,
		IsMigrated: true,
		Arity:      -3,
	}
	lpopCmdMeta = DiceCmdMeta{
		Name:       "LPOP",
		Info:       "LPOP pops a value from the left side of the deque",
		NewEval:    evalLPOP,
		IsMigrated: true,
		Arity:      2,
	}
	rpopCmdMeta = DiceCmdMeta{
		Name:       "RPOP",
		Info:       "RPOP pops a value from the right side of the deque",
		NewEval:    evalRPOP,
		IsMigrated: true,
		Arity:      2,
	}
	llenCmdMeta = DiceCmdMeta{
		Name: "LLEN",
		Info: `LLEN key
		Returns the length of the list stored at key. If key does not exist,
		it is interpreted as an empty list and 0 is returned.
		An error is returned when the value stored at key is not a list.`,
		NewEval:    evalLLEN,
		IsMigrated: true,
		Arity:      1,
	}
	// Internal command used to spawn request across all shards (works internally with DBSIZE command)
	singleDBSizeCmdMeta = DiceCmdMeta{
		Name:       "SINGLEDBSIZE",
		Info:       `DBSIZE Return the number of keys in the database`,
		NewEval:    evalDBSize,
		IsMigrated: true,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	bitposCmdMeta = DiceCmdMeta{
		Name: "BITPOS",
		Info: `BITPOS returns the position of the first bit set to 1 or 0 in a string
		 The position is returned, thinking of the string as an array of bits from left to right,
		 where the first byte's most significant bit is at position 0, the second byte's most significant
		 bit is at position 8, and so forth.
		 By default, all the bytes contained in the string are examined. It is possible to look for bits only in a
		 specified interval passing the additional arguments start and end (it is possible to just pass start,
		 the operation will assume that the end is the last byte of the string).
		 By default, the range is interpreted as a range of bytes and not a range of bits, so start=0 and end=2 means
		 to look at the first three bytes.
		 You can use the optional BIT modifier to specify that the range should be interpreted as a range of bits. So
		 start=0 and end=2 means to look at the first three bits.
		 Note that bit positions are returned always as absolute values starting from bit zero even when start and end
		 are used to specify a range.
		 The start and end can contain negative values in order to index bytes starting from the end of the string,
		 where -1 is the last byte, -2 is the penultimate, and so forth. When BIT is specified, -1 is the last bit, -2
		 is the penultimate, and so forth.
		 Returns
		 RESP encoded integer indicating the position of the first bit set to 1 or 0 according to the request.
		 RESP encoded integer if we look for clear bits and the string only contains bits set to 1, the function returns
	     the first bit not part of the string on the right.
		 RESP encoded -1 in case the bit argument is 1 and the string is empty or composed of just zero bytes.
		 RESP encoded -1 if we look for set bits and the string is empty or composed of just zero bytes, -1 is returned.
		 RESP encoded -1 if a clear bit isn't found in the specified range.`,
		IsMigrated: true,
		NewEval:    evalBITPOS,
		Arity:      -2,
	}
	saddCmdMeta = DiceCmdMeta{
		Name: "SADD",
		Info: `SADD key member [member ...]
		Adds the specified members to the set stored at key.
		Specified members that are already a member of this set are ignored
		Non existing keys are treated as empty sets.
		An error is returned when the value stored at key is not a set.`,
		NewEval:    evalSADD,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	smembersCmdMeta = DiceCmdMeta{
		Name: "SMEMBERS",
		Info: `SMEMBERS key
		Returns all the members of the set value stored at key.`,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalSMEMBERS,
	}
	sremCmdMeta = DiceCmdMeta{
		Name: "SREM",
		Info: `SREM key member [member ...]
		Removes the specified members from the set stored at key.
		Non existing keys are treated as empty sets.
		An error is returned when the value stored at key is not a set.`,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalSREM,
	}
	scardCmdMeta = DiceCmdMeta{
		Name: "SCARD",
		Info: `SCARD key
		Returns the number of elements of the set stored at key.
		An error is returned when the value stored at key is not a set.`,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalSCARD,
	}
	pfAddCmdMeta = DiceCmdMeta{
		Name: "PFADD",
		Info: `PFADD key [element [element ...]]
		Adds elements to a HyperLogLog key. Creates the key if it doesn't exist.`,
		NewEval:    evalPFADD,
		IsMigrated: true,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	pfCountCmdMeta = DiceCmdMeta{
		Name: "PFCOUNT",
		Info: `PFCOUNT key [key ...]
		Returns the approximated cardinality of the set(s) observed by the HyperLogLog key(s).`,
		NewEval:    evalPFCOUNT,
		IsMigrated: true,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	jsonStrlenCmdMeta = DiceCmdMeta{
		Name: "JSON.STRLEN",
		Info: `JSON.STRLEN key [path]
		Report the length of the JSON String at path in key`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalJSONSTRLEN,
	}
	hlenCmdMeta = DiceCmdMeta{
		Name: "HLEN",
		Info: `HLEN key
		Returns the number of fields contained in the hash stored at key.`,
		NewEval:    evalHLEN,
		IsMigrated: true,
		Arity:      2,
	}
	jsonnumincrbyCmdMeta = DiceCmdMeta{
		Name:       "JSON.NUMINCRBY",
		Info:       `Increment the number value stored at path by number.`,
		NewEval:    evalJSONNUMINCRBY,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
	}
	dumpkeyCMmdMeta = DiceCmdMeta{
		Name: "DUMP",
		Info: `Serialize the value stored at key and return it to the user.
				The returned value can be synthesized back into the key using the RESTORE command.`,
		NewEval:    evalDUMP,
		IsMigrated: true,
		Arity:      1,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	restorekeyCmdMeta = DiceCmdMeta{
		Name: "RESTORE",
		Info: `Serialize the value stored at key and return it to the user.
				The returned value can be synthesized back into a key using the RESTORE command.`,
		NewEval:    evalRestore,
		IsMigrated: true,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	getRangeCmdMeta = DiceCmdMeta{
		Name:       "GETRANGE",
		Info:       `Returns a substring of the string stored at a key.`,
		IsMigrated: true,
		NewEval:    evalGETRANGE,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	hrandfieldCmdMeta = DiceCmdMeta{
		Name:       "HRANDFIELD",
		Info:       `Returns one or more random fields from a hash.`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalHRANDFIELD,
	}
	appendCmdMeta = DiceCmdMeta{
		Name:       "APPEND",
		Info:       `Appends a string to the value of a key. Creates the key if it doesn't exist.`,
		IsMigrated: true,
		NewEval:    evalAPPEND,
		Arity:      2,
	}
	zaddCmdMeta = DiceCmdMeta{
		Name: "ZADD",
		Info: `ZADD key [NX|XX] [CH] [INCR] score member [score member ...]
		Adds all the specified members with the specified scores to the sorted set stored at key.
		Options: NX, XX, CH, INCR
		Returns the number of elements added to the sorted set, not including elements already existing for which the score was updated.`,
		Arity:      -4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZADD,
	}
	zcountCmdMeta = DiceCmdMeta{
		Name: "ZCOUNT",
		Info: `ZCOUNT key min max
		Counts the number of members in a sorted set with scores between min and max (inclusive).
		Use -inf and +inf for unbounded ranges. Returns 0 if the key does not exist.`,
		Arity:      4,
		IsMigrated: true,
		NewEval:    evalZCOUNT,
	}
	zrangeCmdMeta = DiceCmdMeta{
		Name: "ZRANGE",
		Info: `ZRANGE key start stop [WithScores]
		Returns the specified range of elements in the sorted set stored at key.
		The elements are considered to be ordered from the lowest to the highest score.
		Both start and stop are 0-based indexes, where 0 is the first element, 1 is the next element and so on.
		These indexes can also be negative numbers indicating offsets from the end of the sorted set, with -1 being the last element of the sorted set, -2 the penultimate element and so on.
		Returns the specified range of elements in the sorted set.`,
		Arity:      -4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZRANGE,
	}
	zpopmaxCmdMeta = DiceCmdMeta{
		Name: "ZPOPMAX",
		Info: `ZPOPMAX  key [count]
		Pops count number of elements from the sorted set from highest to lowest and returns those score and member.
		If count is not provided '1' is considered by default.
		The element with the highest score is removed first
		if two elements have same score then the element which is lexicographically higher is popped first`,
		Arity:      -1,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZPOPMAX,
	}
	zpopminCmdMeta = DiceCmdMeta{
		Name: "ZPOPMIN",
		Info: `ZPOPMIN key [count]
		Removes and returns the member with the lowest score from the sorted set at the specified key.
		If multiple members have the same score, the one that comes first alphabetically is returned.
		You can also specify a count to remove and return multiple members at once.
		If the set is empty, it returns an empty result.`,
		Arity:      -1,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZPOPMIN,
	}
	zrankCmdMeta = DiceCmdMeta{
		Name: "ZRANK",
		Info: `ZRANK key member [WITHSCORE]
		Returns the rank of member in the sorted set stored at key, with the scores ordered from low to high.
		The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
		The optional WITHSCORE argument supplements the command's reply with the score of the element returned.`,
		Arity:      -2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZRANK,
	}
	zcardCmdMeta = DiceCmdMeta{
		Name: "ZCARD",
		Info: `ZCARD key
		Returns the sorted set cardinality (number of elements) of the sorted set stored at key.`,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZCARD,
	}
	zremCmdMeta = DiceCmdMeta{
		Name: "ZREM",
		Info: `ZREM key member [member ...]
		Removes the specified members from the sorted set stored at key. Non existing members are ignored.
		An error is returned when key exists and does not hold a sorted set.`,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalZREM,
	}
	bitfieldCmdMeta = DiceCmdMeta{
		Name: "BITFIELD",
		Info: `The command treats a string as an array of bits as well as bytearray data structure,
		and is capable of addressing specific integer fields of varying bit widths
		and arbitrary non (necessary) aligned offset.
		In practical terms using this command you can set, for example,
		a signed 5 bits integer at bit offset 1234 to a specific value,
		retrieve a 31 bit unsigned integer from offset 4567.
		Similarly the command handles increments and decrements of the
		specified integers, providing guaranteed and well specified overflow
		and underflow behavior that the user can configure.
		The following is the list of supported commands.
		GET <encoding> <offset> -- Returns the specified bit field.
		SET <encoding> <offset> <value> -- Set the specified bit field
		and returns its old value.
		INCRBY <encoding> <offset> <increment> -- Increments or decrements
		(if a negative increment is given) the specified bit field and returns the new value.
		There is another subcommand that only changes the behavior of successive
		INCRBY and SET subcommands calls by setting the overflow behavior:
		OVERFLOW [WRAP|SAT|FAIL]`,
		Arity:      -1,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalBITFIELD,
	}
	bitfieldroCmdMeta = DiceCmdMeta{
		Name: "BITFIELD_RO",
		Info: `It is read-only variant of the BITFIELD command.
		It is like the original BITFIELD but only accepts GET subcommand.`,
		Arity:      -1,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalBITFIELDRO,
	}
	hincrbyFloatCmdMeta = DiceCmdMeta{
		Name: "HINCRBYFLOAT",
		Info: `HINCRBYFLOAT increments the specified field of a hash stored at the key,
		and representing a floating point number, by the specified increment.
		If the field does not exist, it is set to 0 before performing the operation.
		If the field contains a value of wrong type or specified increment
		is not parsable as floating point number, then an error occurs.
		`,
		Arity:      -4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalHINCRBYFLOAT,
	}
	geoAddCmdMeta = DiceCmdMeta{
		Name:       "GEOADD",
		Info:       `Adds one or more members to a geospatial index. The key is created if it doesn't exist.`,
		Arity:      -5,
		IsMigrated: true,
		NewEval:    evalGEOADD,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	geoDistCmdMeta = DiceCmdMeta{
		Name:       "GEODIST",
		Info:       `Returns the distance between two members in the geospatial index.`,
		Arity:      -4,
		IsMigrated: true,
		NewEval:    evalGEODIST,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	geoPosCmdMeta = DiceCmdMeta{
		Name:       "GEOPOS",
		Info:       `Returns the latitude and longitude of the members identified by the particular index.`,
		Arity:      -3,
		NewEval:    evalGEOPOS,
		IsMigrated: true,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	geoHashCmdMeta = DiceCmdMeta{
		Name:       "GEOHASH",
		Info:       `Return Geohash strings representing the position of one or more elements representing a geospatial index`,
		Arity:      -2,
		IsMigrated: true,
		NewEval:    evalGEOHASH,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	jsonstrappendCmdMeta = DiceCmdMeta{
		Name: "JSON.STRAPPEND",
		Info: `JSON.STRAPPEND key [path] value
		Append the JSON string values to the string at path
		Returns an array of integer replies for each path, the string's new length, or nil, if the matching JSON value is not a string.
		Error reply: If the value at path is not a string or if the key doesn't exist.`,
		NewEval:    evalJSONSTRAPPEND,
		IsMigrated: true,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	cmsInitByDimCmdMeta = DiceCmdMeta{
		Name:       "CMS.INITBYDIM",
		Info:       `Sets up count min sketch`,
		Arity:      3,
		IsMigrated: true,
		NewEval:    evalCMSINITBYDIM,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	cmsInitByProbCmdMeta = DiceCmdMeta{
		Name:       "CMS.INITBYPROB",
		Info:       `Sets up count min sketch with given error rate and probability`,
		Arity:      3,
		IsMigrated: true,
		NewEval:    evalCMSINITBYPROB,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	cmsInfoCmdMeta = DiceCmdMeta{
		Name:       "CMS.INFO",
		Info:       `Get info about count min sketch`,
		Arity:      1,
		IsMigrated: true,
		NewEval:    evalCMSINFO,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	cmsQueryCmdMeta = DiceCmdMeta{
		Name:       "CMS.QUERY",
		Info:       `Query count min sketch with for given list of keys`,
		Arity:      -2,
		IsMigrated: true,
		NewEval:    evalCMSQuery,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	cmsIncrByCmdMeta = DiceCmdMeta{
		Name:       "CMS.INCRBY",
		Info:       `Increase count of the list of keys to count min sketch`,
		Arity:      -3,
		IsMigrated: true,
		NewEval:    evalCMSIncrBy,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	cmsMergeCmdMeta = DiceCmdMeta{
		Name: "CMS.MERGE",
		Info: `Merges several sketches into one sketch.
				 All sketches must have identical width and depth.
				 Weights can be used to multiply certain sketches. Default weight is 1.`,
		Arity:      -3,
		IsMigrated: true,
		NewEval:    evalCMSMerge,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	linsertCmdMeta = DiceCmdMeta{
		Name: "LINSERT",
		Info: `
		Usage:
			LINSERT key <BEFORE | AFTER> pivot element
		Info:
			Inserts element in the list stored at key either before or after the reference value pivot.
			When key does not exist, it is considered an empty list and no operation is performed.
			An error is returned when key exists but does not hold a list value.
		Returns:
			Integer - the list length after a successful insert operation.
			0 when the key doesn't exist.
			-1 when the pivot wasn't found.
		`,
		NewEval:    evalLINSERT,
		IsMigrated: true,
		Arity:      5,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	lrangeCmdMeta = DiceCmdMeta{
		Name: "LRANGE",
		Info: `
		Usage:
			LRANGE key start stop
		Info:
			Returns the specified elements of the list stored at key.
			The offsets start and stop are zero-based indexes, with 0 being the first element of the list (the head of the list), 1 being the next element and so on.

			These offsets can also be negative numbers indicating offsets starting at the end of the list.
			For example, -1 is the last element of the list, -2 the penultimate, and so on.
			
			Out of range indexes will not produce an error. If start is larger than the end of the list, an empty list is returned.
			If stop is larger than the actual end of the list it will be treated like the last element of the list.
		Returns:
			Array reply: a list of elements in the specified range, or an empty array if the key doesn't exist.
		`,
		NewEval:    evalLRANGE,
		IsMigrated: true,
		Arity:      4,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
)

func init() {
	PreProcessing["COPY"] = evalGetObject
	PreProcessing["GETOBJECT"] = evalGetObject

	DiceCmds["ABORT"] = abortCmdMeta
	DiceCmds["APPEND"] = appendCmdMeta
	DiceCmds["AUTH"] = authCmdMeta
	DiceCmds["BF.ADD"] = bfaddCmdMeta
	DiceCmds["BF.EXISTS"] = bfexistsCmdMeta
	DiceCmds["BF.INFO"] = bfinfoCmdMeta
	DiceCmds["BF.RESERVE"] = bfreserveCmdMeta
	DiceCmds["BITCOUNT"] = bitCountCmdMeta
	DiceCmds["BITFIELD"] = bitfieldCmdMeta
	DiceCmds["BITFIELD_RO"] = bitfieldroCmdMeta
	DiceCmds["BITPOS"] = bitposCmdMeta
	DiceCmds["CLIENT"] = clientCmdMeta
	DiceCmds["COMMAND"] = commandCmdMeta
	DiceCmds["COMMAND|COUNT"] = commandCountCmdMeta
	DiceCmds["COMMAND|GETKEYS"] = commandGetKeysCmdMeta
	DiceCmds["COMMAND|LIST"] = commandListCmdMeta
	DiceCmds["COMMAND|HELP"] = commandHelpCmdMeta
	DiceCmds["COMMAND|INFO"] = commandInfoCmdMeta
	DiceCmds["COMMAND|DOCS"] = commandDocsCmdMeta
	DiceCmds["COMMAND|GETKEYSANDFLAGS"] = commandGetKeysAndFlagsCmdMeta
	DiceCmds["OBJECTCOPY"] = objectCopyCmdMeta
	DiceCmds["DUMP"] = dumpkeyCMmdMeta
	DiceCmds["GEOADD"] = geoAddCmdMeta
	DiceCmds["GEODIST"] = geoDistCmdMeta
	DiceCmds["GEOPOS"] = geoPosCmdMeta
	DiceCmds["GEOHASH"] = geoHashCmdMeta
	DiceCmds["GETBIT"] = getBitCmdMeta
	DiceCmds["GETRANGE"] = getRangeCmdMeta
	DiceCmds["HDEL"] = hdelCmdMeta
	DiceCmds["HELLO"] = helloCmdMeta
	DiceCmds["HEXISTS"] = hexistsCmdMeta
	DiceCmds["HGET"] = hgetCmdMeta
	DiceCmds["HGETALL"] = hgetAllCmdMeta
	DiceCmds["HINCRBY"] = hincrbyCmdMeta
	DiceCmds["HINCRBYFLOAT"] = hincrbyFloatCmdMeta
	DiceCmds["HKEYS"] = hkeysCmdMeta
	DiceCmds["HLEN"] = hlenCmdMeta
	DiceCmds["HMGET"] = hmgetCmdMeta
	DiceCmds["HMSET"] = hmsetCmdMeta
	DiceCmds["HRANDFIELD"] = hrandfieldCmdMeta
	DiceCmds["HSCAN"] = hscanCmdMeta
	DiceCmds["HSET"] = hsetCmdMeta
	DiceCmds["HSETNX"] = hsetnxCmdMeta
	DiceCmds["HSTRLEN"] = hstrLenCmdMeta
	DiceCmds["HVALS"] = hValsCmdMeta
	DiceCmds["INCRBYFLOAT"] = incrByFloatCmdMeta
	DiceCmds["JSON.ARRAPPEND"] = jsonarrappendCmdMeta
	DiceCmds["JSON.ARRINSERT"] = jsonarrinsertCmdMeta
	DiceCmds["JSON.ARRLEN"] = jsonarrlenCmdMeta
	DiceCmds["JSON.ARRPOP"] = jsonarrpopCmdMeta
	DiceCmds["JSON.ARRTRIM"] = jsonarrtrimCmdMeta
	DiceCmds["JSON.CLEAR"] = jsonclearCmdMeta
	DiceCmds["JSON.DEBUG"] = jsondebugCmdMeta
	DiceCmds["JSON.DEL"] = jsondelCmdMeta
	DiceCmds["JSON.FORGET"] = jsonforgetCmdMeta
	DiceCmds["JSON.GET"] = jsongetCmdMeta
	DiceCmds["JSON.INGEST"] = jsoningestCmdMeta
	DiceCmds["JSON.NUMINCRBY"] = jsonnumincrbyCmdMeta
	DiceCmds["JSON.NUMMULTBY"] = jsonnummultbyCmdMeta
	DiceCmds["JSON.OBJKEYS"] = jsonobjkeysCmdMeta
	DiceCmds["JSON.OBJLEN"] = jsonobjlenCmdMeta
	DiceCmds["JSON.RESP"] = jsonrespCmdMeta
	DiceCmds["JSON.SET"] = jsonsetCmdMeta
	DiceCmds["JSON.STRLEN"] = jsonStrlenCmdMeta
	DiceCmds["JSON.TOGGLE"] = jsontoggleCmdMeta
	DiceCmds["JSON.TYPE"] = jsontypeCmdMeta
	DiceCmds["LATENCY"] = latencyCmdMeta
	DiceCmds["LLEN"] = llenCmdMeta
	DiceCmds["LPOP"] = lpopCmdMeta
	DiceCmds["LPUSH"] = lpushCmdMeta
	DiceCmds["OBJECT"] = objectCmdMeta
	DiceCmds["PERSIST"] = persistCmdMeta
	DiceCmds["PFADD"] = pfAddCmdMeta
	DiceCmds["PFCOUNT"] = pfCountCmdMeta
	DiceCmds["PFMERGE"] = pfMergeCmdMeta
	DiceCmds["PTTL"] = pttlCmdMeta
	DiceCmds["RESTORE"] = restorekeyCmdMeta
	DiceCmds["RPOP"] = rpopCmdMeta
	DiceCmds["RPUSH"] = rpushCmdMeta
	DiceCmds["SADD"] = saddCmdMeta
	DiceCmds["SCARD"] = scardCmdMeta
	DiceCmds["SETBIT"] = setBitCmdMeta
	DiceCmds["SLEEP"] = sleepCmdMeta
	DiceCmds["SMEMBERS"] = smembersCmdMeta
	DiceCmds["SREM"] = sremCmdMeta
	DiceCmds["ZADD"] = zaddCmdMeta
	DiceCmds["ZCOUNT"] = zcountCmdMeta
	DiceCmds["ZRANGE"] = zrangeCmdMeta
	DiceCmds["ZPOPMAX"] = zpopmaxCmdMeta
	DiceCmds["ZPOPMIN"] = zpopminCmdMeta
	DiceCmds["ZRANK"] = zrankCmdMeta
	DiceCmds["ZCARD"] = zcardCmdMeta
	DiceCmds["ZREM"] = zremCmdMeta
	DiceCmds["JSON.STRAPPEND"] = jsonstrappendCmdMeta
	DiceCmds["CMS.INITBYDIM"] = cmsInitByDimCmdMeta
	DiceCmds["CMS.INITBYPROB"] = cmsInitByProbCmdMeta
	DiceCmds["CMS.INFO"] = cmsInfoCmdMeta
	DiceCmds["CMS.QUERY"] = cmsQueryCmdMeta
	DiceCmds["CMS.INCRBY"] = cmsIncrByCmdMeta
	DiceCmds["CMS.MERGE"] = cmsMergeCmdMeta
	DiceCmds["LINSERT"] = linsertCmdMeta
	DiceCmds["LRANGE"] = lrangeCmdMeta
	DiceCmds["JSON.ARRINDEX"] = jsonArrIndexCmdMeta

	DiceCmds["SINGLETOUCH"] = singleTouchCmdMeta
	DiceCmds["SINGLEDBSIZE"] = singleDBSizeCmdMeta
	DiceCmds["SINGLEKEYS"] = singleKeysCmdMeta
}

// Function to convert DiceCmdMeta to []interface{}
func convertCmdMetaToSlice(cmdMeta *DiceCmdMeta) []interface{} {
	var result []interface{} = []interface{}{strings.ToLower(cmdMeta.Name), cmdMeta.Arity, cmdMeta.KeySpecs.BeginIndex, cmdMeta.KeySpecs.LastKey, cmdMeta.KeySpecs.Step}
	var subCommandsList []interface{}
	for _, subCommand := range cmdMeta.SubCommands {
		key := cmdMeta.Name + "|" + subCommand
		if val, exists := DiceCmds[key]; exists {
			valCopy := val // Store the value in a variable
			subCommandsList = append(subCommandsList, convertCmdMetaToSlice(&valCopy))
		}
	}

	return append(result, subCommandsList)
}

// Function to convert map[string]DiceCmdMeta{} to []interface{}
func convertDiceCmdsMapToSlice() []interface{} {
	var result []interface{}
	for _, cmdMeta := range DiceCmds {
		result = append(result, convertCmdMetaToSlice(&cmdMeta))
	}
	return result
}

func convertCmdMetaToDocs(cmdMeta *DiceCmdMeta) []interface{} {
	var result []interface{} = []interface{}{"summary", cmdMeta.Info, "arity", cmdMeta.Arity, "beginIndex", cmdMeta.KeySpecs.BeginIndex,
		"lastIndex", cmdMeta.KeySpecs.LastKey, "step", cmdMeta.KeySpecs.Step}
	var subCommandsList []interface{}
	for _, subCommand := range cmdMeta.SubCommands {
		key := cmdMeta.Name + "|" + subCommand
		if val, exists := DiceCmds[key]; exists {
			valCopy := val // Store the value in a variable
			subCommandsList = append(subCommandsList, convertCmdMetaToDocs(&valCopy))
		}
	}

	if len(subCommandsList) != 0 {
		result = append(result, "subcommands", subCommandsList)
	}

	return []interface{}{strings.ToLower(cmdMeta.Name), result}
}

// Function to convert map[string]DiceCmdMeta{} to []interface{}
func convertDiceCmdsMapToDocs() []interface{} {
	var result []interface{}
	// TODO: Add other keys supported as part of COMMAND DOCS, currently only
	// command name and summary supported. This would required adding more metadata to supported commands
	for _, cmdMeta := range DiceCmds {
		result = append(result, strings.ToLower(cmdMeta.Name))
		subResult := []interface{}{"summary", cmdMeta.Info}
		result = append(result, subResult)
	}

	return result
}
