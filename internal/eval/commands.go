package eval

import dstore "github.com/dicedb/dice/internal/store"

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
}

type KeySpecs struct {
	BeginIndex int
	Step       int
	LastKey    int
}

var (
	DiceCmds = map[string]DiceCmdMeta{}

	echoCmdMeta = DiceCmdMeta{
		Name:  "ECHO",
		Info:  `ECHO returns the string given as argument.`,
		Eval:  evalECHO,
		Arity: 1,
	}

	pingCmdMeta = DiceCmdMeta{
		Name:  "PING",
		Info:  `PING returns with an encoded "PONG" If any message is added with the ping command,the message will be returned.`,
		Arity: -1,
		// TODO: Move this to true once compatible with HTTP server
		IsMigrated: false,
		Eval:       evalPING,
	}

	setCmdMeta = DiceCmdMeta{
		Name: "SET",
		Info: `SET puts a new <key, value> pair in db as in the args
		args must contain key and value.
		args can also contain multiple options -
		EX or ex which will set the expiry time(in secs) for the key
		Returns encoded error response if at least a <key, value> pair is not part of args
		Returns encoded error response if expiry tme value in not integer
		Returns encoded OK RESP once new entry is added
		If the key already exists then the value will be overwritten and expiry will be discarded`,
		Arity:      -3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalSET,
	}
	getCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: `GET returns the value for the queried key in args
		The key should be the only param in args
		The RESP value of the key is encoded and then returned
		GET returns RespNIL if key is expired or it does not exist`,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalGET,
	}

	getSetCmdMeta = DiceCmdMeta{
		Name:       "GETSET",
		Info:       `GETSET returns the previous string value of a key after setting it to a new value.`,
		Arity:      2,
		IsMigrated: true,
		NewEval:    evalGETSET,
	}

	authCmdMeta = DiceCmdMeta{
		Name: "AUTH",
		Info: `AUTH returns with an encoded "OK" if the user is authenticated.
		If the user is not authenticated, it returns with an encoded error message`,
		Eval: nil,
	}
	getDelCmdMeta = DiceCmdMeta{
		Name: "GETDEL",
		Info: `GETDEL returns the value for the queried key in args
		The key should be the only param in args And If the key exists, it will be deleted before its value is returned.
		The RESP value of the key is encoded and then returned
		GETDEL returns RespNIL if key is expired or it does not exist`,
		Eval:     evalGETDEL,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	msetCmdMeta = DiceCmdMeta{
		Name: "MSET",
		Info: `MSET sets multiple keys to multiple values in the db
		args should contain an even number of elements
		each pair of elements will be treated as <key, value> pair
		Returns encoded error response if the number of arguments is not even
		Returns encoded OK RESP once all entries are added`,
		Eval:     evalMSET,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 2, LastKey: -1},
	}
	jsonsetCmdMeta = DiceCmdMeta{
		Name: "JSON.SET",
		Info: `JSON.SET key path json-string
		Sets a JSON value at the specified key.
		Returns OK if successful.
		Returns encoded error message if the number of arguments is incorrect or the JSON string is invalid.`,
		Eval:     evalJSONSET,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsongetCmdMeta = DiceCmdMeta{
		Name: "JSON.GET",
		Info: `JSON.GET key [path]
		Returns the encoded RESP value of the key, if present
		Null reply: If the key doesn't exist or has expired.
		Error reply: If the number of arguments is incorrect or the stored value is not a JSON type.`,
		Eval:     evalJSONGET,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonMGetCmdMeta = DiceCmdMeta{
		Name: "JSON.MGET",
		Info: `JSON.MGET key..key [path]
		Returns the encoded RESP value of the key, if present
		Null reply: If the key doesn't exist or has expired.
		Error reply: If the number of arguments is incorrect or the stored value is not a JSON type.`,
		Eval:     evalJSONMGET,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
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
		Eval:     evalJSONTOGGLE,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsontypeCmdMeta = DiceCmdMeta{
		Name: "JSON.TYPE",
		Info: `JSON.TYPE key [path]
		Returns string reply for each path, specified as the value's type.
		Returns RespNIL If the key doesn't exist.
		Error reply: If the number of arguments is incorrect.`,
		Eval:     evalJSONTYPE,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
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
		Eval:     evalJSONDEL,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonarrappendCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRAPPEND",
		Info: `JSON.ARRAPPEND key [path] value [value ...]
        Returns an array of integer replies for each path, the array's new size,
        or nil, if the matching JSON value is not an array.`,
		Eval:  evalJSONARRAPPEND,
		Arity: -3,
	}
	jsonforgetCmdMeta = DiceCmdMeta{
		Name: "JSON.FORGET",
		Info: `JSON.FORGET key [path]
		Returns an integer reply specified as the number of paths deleted (0 or more).
		Returns RespZero if the key doesn't exist or key is expired.
		Error reply: If the number of arguments is incorrect.`,
		Eval:     evalJSONFORGET,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonarrlenCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRLEN",
		Info: `JSON.ARRLEN key [path]
		Returns an array of integer replies.
		Returns error response if the key doesn't exist or key is expired or the matching value is not an array.
		Error reply: If the number of arguments is incorrect.`,
		Eval:     evalJSONARRLEN,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonnummultbyCmdMeta = DiceCmdMeta{
		Name: "JSON.NUMMULTBY",
		Info: `JSON.NUMMULTBY key path value
		Multiply the number value stored at the specified path by a value.`,
		Eval:     evalJSONNUMMULTBY,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1},
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
		Eval:     evalJSONDebug,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonobjkeysCmdMeta = DiceCmdMeta{
		Name: "JSON.OBJKEYS",
		Info: `JSON.OBJKEYS key [path]
		Retrieves the keys of a JSON object stored at path specified.
		Null reply: If the key doesn't exist or has expired.
		Error reply: If the number of arguments is incorrect or the stored value is not a JSON type.`,
		Eval:  evalJSONOBJKEYS,
		Arity: 2,
	}
	jsonarrpopCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRPOP",
		Info: `JSON.ARRPOP key [path [index]]
		Removes and returns an element from the index in the array and updates the array in memory.
		Returns error if key doesn't exist.
		Return nil if array is empty or there is no array at the path.
		It supports negative index and is out of bound safe.
		`,
		Eval:  evalJSONARRPOP,
		Arity: -2,
	}
	jsoningestCmdMeta = DiceCmdMeta{
		Name: "JSON.INGEST",
		Info: `JSON.INGEST key_prefix json-string
		The whole key is generated by appending a unique identifier to the provided key prefix.
		the generated key is then used to store the provided JSON value at specified path.
		Returns unique identifier if successful.
		Returns encoded error message if the number of arguments is incorrect or the JSON string is invalid.`,
		Eval:     evalJSONINGEST,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonarrinsertCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRINSERT",
		Info: `JSON.ARRINSERT key path index value [value ...]
		Returns an array of integer replies for each path.
		Returns nil if the matching JSON value is not an array.
		Returns error response if the key doesn't exist or key is expired or the matching value is not an array.
		Error reply: If the number of arguments is incorrect.`,
		Eval:     evalJSONARRINSERT,
		Arity:    -5,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonrespCmdMeta = DiceCmdMeta{
		Name: "JSON.RESP",
		Info: `JSON.RESP key [path]
		Return the JSON in key in Redis serialization protocol specification form`,
		Eval:     evalJSONRESP,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonarrtrimCmdMeta = DiceCmdMeta{
		Name: "JSON.ARRTRIM",
		Info: `JSON.ARRTRIM key path start stop
		Trim an array so that it contains only the specified inclusive range of elements
		Returns an array of integer replies for each path.
		Returns error response if the key doesn't exist or key is expired.
		Error reply: If the number of arguments is incorrect.`,
		Eval:  evalJSONARRTRIM,
		Arity: -5,
	}
	ttlCmdMeta = DiceCmdMeta{
		Name: "TTL",
		Info: `TTL returns Time-to-Live in secs for the queried key in args
		The key should be the only param in args else returns with an error
		Returns
		RESP encoded time (in secs) remaining for the key to expire
		RESP encoded -2 stating key doesn't exist or key is expired
		RESP encoded -1 in case no expiration is set on the key`,
		Eval:     evalTTL,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	delCmdMeta = DiceCmdMeta{
		Name: "DEL",
		Info: `DEL deletes all the specified keys in args list
		returns the count of total deleted keys after encoding`,
		Eval:     evalDEL,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1, LastKey: -1},
	}
	expireCmdMeta = DiceCmdMeta{
		Name: "EXPIRE",
		Info: `EXPIRE sets a expiry time(in secs) on the specified key in args
		args should contain 2 values, key and the expiry time to be set for the key
		The expiry time should be in integer format; if not, it returns encoded error response
		Returns RespOne if expiry was set on the key successfully.
		Once the time is lapsed, the key will be deleted automatically`,
		Eval:     evalEXPIRE,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	helloCmdMeta = DiceCmdMeta{
		Name:  "HELLO",
		Info:  `HELLO always replies with a list of current server and connection properties, such as: versions, modules loaded, client ID, replication role and so forth`,
		Eval:  evalHELLO,
		Arity: -1,
	}
	bgrewriteaofCmdMeta = DiceCmdMeta{
		Name:  "BGREWRITEAOF",
		Info:  `Instruct Dice to start an Append Only File rewrite process. The rewrite will create a small optimized version of the current Append Only File.`,
		Eval:  EvalBGREWRITEAOF,
		Arity: 1,
	}
	incrCmdMeta = DiceCmdMeta{
		Name: "INCR",
		Info: `INCR increments the value of the specified key in args by 1,
		if the key exists and the value is integer format.
		The key should be the only param in args.
		If the key does not exist, new key is created with value 0,
		the value of the new key is then incremented.
		The value for the queried key should be of integer format,
		if not INCR returns encoded error response.
		evalINCR returns the incremented value for the key if there are no errors.`,
		NewEval:    evalINCR,
		IsMigrated: true,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
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
	infoCmdMeta = DiceCmdMeta{
		Name: "INFO",
		Info: `INFO creates a buffer with the info of total keys per db
		Returns the encoded buffer as response`,
		Eval:  evalINFO,
		Arity: -1,
	}
	clientCmdMeta = DiceCmdMeta{
		Name:  "CLIENT",
		Info:  `This is a container command for client connection commands.`,
		Eval:  evalCLIENT,
		Arity: -2,
	}
	latencyCmdMeta = DiceCmdMeta{
		Name:  "LATENCY",
		Info:  `This is a container command for latency diagnostics commands.`,
		Eval:  evalLATENCY,
		Arity: -2,
	}
	lruCmdMeta = DiceCmdMeta{
		Name: "LRU",
		Info: `LRU deletes all the keys from the LRU
		returns encoded RESP OK`,
		Eval:  evalLRU,
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
	// TODO: Remove this override once we support QWATCH in dice-cli.
	subscribeCmdMeta = DiceCmdMeta{
		Name: "SUBSCRIBE",
		Info: `SUBSCRIBE(or QWATCH) adds the specified key to the watch list for the caller client.
		Every time a key in the watch list is modified, the client will be sent a response
		containing the new value of the key along with the operation that was performed on it.
		Contains only one argument, the key to be watched.`,
		Eval:  nil,
		Arity: 1,
	}
	qwatchCmdMeta = DiceCmdMeta{
		Name: "Q.WATCH",
		Info: `Q.WATCH adds the specified key to the watch list for the caller client.
		Every time a key in the watch list is modified, the client will be sent a response
		containing the new value of the key along with the operation that was performed on it.
		Contains only one argument, the key to be watched.`,
		Eval:  nil,
		Arity: 1,
	}
	qUnwatchCmdMeta = DiceCmdMeta{
		Name: "Q.UNWATCH",
		Info: `Unsubscribes or QUnwatches the client from the given key's watch session.
		It removes the key from the watch list for the caller client.`,
		Eval:  nil,
		Arity: 1,
	}
	MultiCmdMeta = DiceCmdMeta{
		Name: "MULTI",
		Info: `MULTI marks the start of the transaction for the client.
		All subsequent commands fired will be queued for atomic execution.
		The commands will not be executed until EXEC is triggered.
		Once EXEC is triggered it executes all the commands in queue,
		and closes the MULTI transaction.`,
		Eval:  evalMULTI,
		Arity: 1,
	}
	ExecCmdMeta = DiceCmdMeta{
		Name:  "EXEC",
		Info:  `EXEC executes commands in a transaction, which is initiated by MULTI`,
		Eval:  nil,
		Arity: 1,
	}
	DiscardCmdMeta = DiceCmdMeta{
		Name:  "DISCARD",
		Info:  `DISCARD discards all the commands in a transaction, which is initiated by MULTI`,
		Eval:  nil,
		Arity: 1,
	}
	abortCmdMeta = DiceCmdMeta{
		Name:  "ABORT",
		Info:  "Quit the server",
		Eval:  nil,
		Arity: 1,
	}
	setBitCmdMeta = DiceCmdMeta{
		Name: "SETBIT",
		Info: "SETBIT sets or clears the bit at offset in the string value stored at key",
		Eval: evalSETBIT,
	}
	getBitCmdMeta = DiceCmdMeta{
		Name: "GETBIT",
		Info: "GETBIT returns the bit value at offset in the string value stored at key",
		Eval: evalGETBIT,
	}
	bitCountCmdMeta = DiceCmdMeta{
		Name:  "BITCOUNT",
		Info:  "BITCOUNT counts the number of set bits in the string value stored at key",
		Eval:  evalBITCOUNT,
		Arity: -1,
	}
	bitOpCmdMeta = DiceCmdMeta{
		Name: "BITOP",
		Info: "BITOP performs bitwise operations between multiple keys",
		Eval: evalBITOP,
	}
	commandCmdMeta = DiceCmdMeta{
		Name:        "COMMAND <subcommand>",
		Info:        "Evaluates COMMAND <subcommand> command based on subcommand",
		Eval:        evalCommand,
		Arity:       -1,
		SubCommands: []string{Count, GetKeys, List, Help, Info},
	}
	keysCmdMeta = DiceCmdMeta{
		Name: "KEYS",
		Info: "KEYS command is used to get all the keys in the database. Complexity is O(n) where n is the number of keys in the database.",
		Eval: evalKeys,
	}
	MGetCmdMeta = DiceCmdMeta{
		Name: "MGET",
		Info: `The MGET command returns an array of RESP values corresponding to the provided keys.
		For each key, if the key is expired or does not exist, the response will be RespNIL;
		otherwise, the response will be the RESP value of the key.
		`,
		Eval:     evalMGET,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1, LastKey: -1},
	}
	persistCmdMeta = DiceCmdMeta{
		Name: "PERSIST",
		Info: "PERSIST removes the expiration from a key",
		Eval: evalPersist,
	}
	copyCmdMeta = DiceCmdMeta{
		Name:  "COPY",
		Info:  `COPY command copies the value stored at the source key to the destination key.`,
		Eval:  evalCOPY,
		Arity: -2,
	}
	decrCmdMeta = DiceCmdMeta{
		Name: "DECR",
		Info: `DECR decrements the value of the specified key in args by 1,
		if the key exists and the value is integer format.
		The key should be the only param in args.
		If the key does not exist, new key is created with value 0,
		the value of the new key is then decremented.
		The value for the queried key should be of integer format,
		if not DECR returns encoded error response.
		evalDECR returns the decremented value for the key if there are no errors.`,
		NewEval:    evalDECR,
		IsMigrated: true,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
	}
	decrByCmdMeta = DiceCmdMeta{
		Name: "DECRBY",
		Info: `DECRBY decrements the value of the specified key in args by the specified decrement,
		if the key exists and the value is in integer format.
		The key should be the first parameter in args, and the decrement should be the second parameter.
		If the key does not exist, new key is created with value 0,
		the value of the new key is then decremented by specified decrement.
		The value for the queried key should be of integer format,
		if not, DECRBY returns an encoded error response.
		evalDECRBY returns the decremented value for the key after applying the specified decrement if there are no errors.`,
		NewEval:    evalDECRBY,
		IsMigrated: true,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
	}
	existsCmdMeta = DiceCmdMeta{
		Name: "EXISTS",
		Info: `EXISTS key1 key2 ... key_N
		Return value is the number of keys existing.`,
		Eval: evalEXISTS,
	}
	renameCmdMeta = DiceCmdMeta{
		Name:  "RENAME",
		Info:  "Renames a key and overwrites the destination",
		Eval:  evalRename,
		Arity: 3,
	}
	getexCmdMeta = DiceCmdMeta{
		Name: "GETEX",
		Info: `Get the value of key and optionally set its expiration.
		GETEX is similar to GET, but is a write command with additional options.`,
		Eval:     evalGETEX,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	pttlCmdMeta = DiceCmdMeta{
		Name: "PTTL",
		Info: `PTTL returns Time-to-Live in millisecs for the queried key in args
		The key should be the only param in args else returns with an error
		Returns
		RESP encoded time (in secs) remaining for the key to expire
		RESP encoded -2 stating key doesn't exist or key is expired
		RESP encoded -1 in case no expiration is set on the key`,
		Eval:     evalPTTL,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hsetCmdMeta = DiceCmdMeta{
		Name: "HSET",
		Info: `HSET sets the specific fields to their respective values in the
		hash stored at key. If any given field is already present, the previous
		value will be overwritten with the new value
		Returns
		This command returns the number of keys that are stored at given key.
		`,
		Eval:     evalHSET,
		Arity:    -4,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hmsetCmdMeta = DiceCmdMeta{
		Name: "HMSET",
		Info: `HSET sets the specific fields to their respective values in the
		hash stored at key. If any given field is already present, the previous
		value will be overwritten with the new value
		Returns
		This command returns the number of keys that are stored at given key.
		`,
		Eval:     evalHMSET,
		Arity:    -4,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hkeysCmdMeta = DiceCmdMeta{
		Name:     "HKEYS",
		Info:     `HKEYS command is used to retrieve all the keys(or field names) within a hash. Complexity is O(n) where n is the size of the hash.`,
		Eval:     evalHKEYS,
		Arity:    1,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hsetnxCmdMeta = DiceCmdMeta{
		Name: "HSETNX",
		Info: `Sets field in the hash stored at key to value, only if field does not yet exist.
		If key does not exist, a new key holding a hash is created. If field already exists,
		this operation has no effect.`,
		Eval:     evalHSETNX,
		Arity:    4,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hgetCmdMeta = DiceCmdMeta{
		Name:     "HGET",
		Info:     `Returns the value associated with field in the hash stored at key.`,
		Eval:     evalHGET,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hmgetCmdMeta = DiceCmdMeta{
		Name:     "HMGET",
		Info:     `Returns the values associated with the specified fields in the hash stored at key.`,
		Eval:     evalHMGET,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hgetAllCmdMeta = DiceCmdMeta{
		Name: "HGETALL",
		Info: `Returns all fields and values of the hash stored at key. In the returned value,
        every field name is followed by its value, so the length of the reply is twice the size of the hash.`,
		Eval:     evalHGETALL,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hValsCmdMeta = DiceCmdMeta{
		Name:     "HVALS",
		Info:     `Returns all values of the hash stored at key. The length of the reply is same as the size of the hash.`,
		Eval:     evalHVALS,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
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
		Name:     "HSTRLEN",
		Info:     `Returns the length of value associated with field in the hash stored at key.`,
		Eval:     evalHSTRLEN,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hdelCmdMeta = DiceCmdMeta{
		Name: "HDEL",
		Info: `HDEL removes the specified fields from the hash stored at key.
		Specified fields that do not exist within this hash are ignored.
		Deletes the hash if no fields remain.
		If key does not exist, it is treated as an empty hash and this command returns 0.
		Returns
		The number of fields that were removed from the hash, not including specified but non-existing fields.`,
		Eval:     evalHDEL,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hscanCmdMeta = DiceCmdMeta{
		Name: "HSCAN",
		Info: `HSCAN is used to iterate over fields and values of a hash.
		It returns a cursor and a list of key-value pairs.
		The cursor is used to paginate through the hash.
		The command returns a cursor value of 0 when all the elements are iterated.`,
		Eval:     evalHSCAN,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	hexistsCmdMeta = DiceCmdMeta{
		Name:     "HEXISTS",
		Info:     `Returns if field is an existing field in the hash stored at key.`,
		Eval:     evalHEXISTS,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}

	objectCmdMeta = DiceCmdMeta{
		Name: "OBJECT",
		Info: `OBJECT subcommand [arguments [arguments ...]]
		OBJECT command is used to inspect the internals of the Redis objects.`,
		Eval:     evalOBJECT,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 2},
	}
	touchCmdMeta = DiceCmdMeta{
		Name: "TOUCH",
		Info: `TOUCH key1 key2 ... key_N
		Alters the last access time of a key(s).
		A key is ignored if it does not exist.`,
		Eval:     evalTOUCH,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	expiretimeCmdMeta = DiceCmdMeta{
		Name: "EXPIRETIME",
		Info: `EXPIRETIME returns the absolute Unix timestamp (since January 1, 1970) in seconds
		at which the given key will expire`,
		Eval:     evalEXPIRETIME,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	expireatCmdMeta = DiceCmdMeta{
		Name: "EXPIREAT",
		Info: `EXPIREAT sets a expiry time(in unix-time-seconds) on the specified key in args
		args should contain 2 values, key and the expiry time to be set for the key
		The expiry time should be in integer format; if not, it returns encoded error response
		Returns RespOne if expiry was set on the key successfully.
		Once the time is lapsed, the key will be deleted automatically`,
		Eval:     evalEXPIREAT,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
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
	dbSizeCmdMeta = DiceCmdMeta{
		Name:  "DBSIZE",
		Info:  `DBSIZE Return the number of keys in the database`,
		Eval:  evalDBSIZE,
		Arity: 1,
	}
	flushdbCmdMeta = DiceCmdMeta{
		Name:  "FLUSHDB",
		Info:  `FLUSHDB deletes all the keys of the currently selected DB`,
		Eval:  evalFLUSHDB,
		Arity: -1,
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
		Eval:  evalBITPOS,
		Arity: -2,
	}
	saddCmdMeta = DiceCmdMeta{
		Name: "SADD",
		Info: `SADD key member [member ...]
		Adds the specified members to the set stored at key.
		Specified members that are already a member of this set are ignored
		Non existing keys are treated as empty sets.
		An error is returned when the value stored at key is not a set.`,
		Eval:     evalSADD,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	smembersCmdMeta = DiceCmdMeta{
		Name: "SMEMBERS",
		Info: `SMEMBERS key
		Returns all the members of the set value stored at key.`,
		Eval:     evalSMEMBERS,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	sremCmdMeta = DiceCmdMeta{
		Name: "SREM",
		Info: `SREM key member [member ...]
		Removes the specified members from the set stored at key.
		Non existing keys are treated as empty sets.
		An error is returned when the value stored at key is not a set.`,
		Eval:     evalSREM,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	scardCmdMeta = DiceCmdMeta{
		Name: "SCARD",
		Info: `SCARD key
		Returns the number of elements of the set stored at key.
		An error is returned when the value stored at key is not a set.`,
		Eval:     evalSCARD,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	sdiffCmdMeta = DiceCmdMeta{
		Name: "SDIFF",
		Info: `SDIFF key1 [key2 ... key_N]
		Returns the members of the set resulting from the difference between the first set and all the successive sets.
		Non existing keys are treated as empty sets.`,
		Eval:     evalSDIFF,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	sinterCmdMeta = DiceCmdMeta{
		Name: "SINTER",
		Info: `SINTER key1 [key2 ... key_N]
		Returns the members of the set resulting from the intersection of all the given sets.
		Non existing keys are treated as empty sets.`,
		Eval:     evalSINTER,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1},
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
	pfMergeCmdMeta = DiceCmdMeta{
		Name: "PFMERGE",
		Info: `PFMERGE destkey [sourcekey [sourcekey ...]]
		Merges one or more HyperLogLog values into a single key.`,
		NewEval:    evalPFMERGE,
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
		Eval:  evalHLEN,
		Arity: 2,
	}
	selectCmdMeta = DiceCmdMeta{
		Name:  "SELECT",
		Info:  `Select the logical database having the specified zero-based numeric index. New connections always use the database 0`,
		Eval:  evalSELECT,
		Arity: 1,
	}
	jsonnumincrbyCmdMeta = DiceCmdMeta{
		Name:     "JSON.NUMINCRBY",
		Info:     `Increment the number value stored at path by number.`,
		Eval:     evalJSONNUMINCRBY,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	dumpkeyCMmdMeta = DiceCmdMeta{
		Name: "DUMP",
		Info: `Serialize the value stored at key in a Redis-specific format and return it to the user.
				The returned value can be synthesized back into a Redis key using the RESTORE command.`,
		Eval:     evalDUMP,
		Arity:    1,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	restorekeyCmdMeta = DiceCmdMeta{
		Name: "RESTORE",
		Info: `Serialize the value stored at key in a Redis-specific format and return it to the user.
				The returned value can be synthesized back into a Redis key using the RESTORE command.`,
		Eval:     evalRestore,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	typeCmdMeta = DiceCmdMeta{
		Name:  "TYPE",
		Info:  `Returns the string representation of the type of the value stored at key. The different types that can be returned are: string, list, set, zset, hash and stream.`,
		Eval:  evalTYPE,
		Arity: 1,

		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	incrbyCmdMeta = DiceCmdMeta{
		Name: "INCRBY",
		Info: `INCRBY increments the value of the specified key in args by increment integer specified,
		if the key exists and the value is integer format.
		The key and the increment integer should be the only param in args.
		If the key does not exist, new key is created with value 0,
		the value of the new key is then incremented.
		The value for the queried key should be of integer format,
		if not INCRBY returns encoded error response.
		evalINCRBY returns the incremented value for the key if there are no errors.`,
		NewEval:    evalINCRBY,
		IsMigrated: true,
		Arity:      2,
		KeySpecs:   KeySpecs{BeginIndex: 1, Step: 1},
	}
	getRangeCmdMeta = DiceCmdMeta{
		Name:       "GETRANGE",
		Info:       `Returns a substring of the string stored at a key.`,
		IsMigrated: true,
		NewEval:    evalGETRANGE,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
	}
	setexCmdMeta = DiceCmdMeta{
		Name: "SETEX",
		Info: `SETEX puts a new <key, value> pair in along with expity
		args must contain key and value and expiry.
		Returns encoded error response if <key,exp,value> is not part of args
		Returns encoded error response if expiry time value in not integer
		Returns encoded OK RESP once new entry is added
		If the key already exists then the value and expiry will be overwritten`,
		Arity:      3,
		KeySpecs:   KeySpecs{BeginIndex: 1},
		IsMigrated: true,
		NewEval:    evalSETEX,
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
		Arity:    -1,
		KeySpecs: KeySpecs{BeginIndex: 1},
		Eval:     evalBITFIELD,
	}
	bitfieldroCmdMeta = DiceCmdMeta{
		Name: "BITFIELD_RO",
		Info: `It is read-only variant of the BITFIELD command. 
		It is like the original BITFIELD but only accepts GET subcommand.`,
		Arity:    -1,
		KeySpecs: KeySpecs{BeginIndex: 1},
		Eval:     evalBITFIELDRO,
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
		Name:     "GEOADD",
		Info:     `Adds one or more members to a geospatial index. The key is created if it doesn't exist.`,
		Arity:    -5,
		Eval:     evalGEOADD,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	geoDistCmdMeta = DiceCmdMeta{
		Name:     "GEODIST",
		Info:     `Returns the distance between two members in the geospatial index.`,
		Arity:    -4,
		Eval:     evalGEODIST,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	jsonstrappendCmdMeta = DiceCmdMeta{
		Name: "JSON.STRAPPEND",
		Info: `JSON.STRAPPEND key [path] value
		Append the JSON string values to the string at path
		Returns an array of integer replies for each path, the string's new length, or nil, if the matching JSON value is not a string. 
		Error reply: If the value at path is not a string or if the key doesn't exist.`,
		Eval:     evalJSONSTRAPPEND,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
)

func init() {
	DiceCmds["ABORT"] = abortCmdMeta
	DiceCmds["APPEND"] = appendCmdMeta
	DiceCmds["AUTH"] = authCmdMeta
	DiceCmds["BF.ADD"] = bfaddCmdMeta
	DiceCmds["BF.EXISTS"] = bfexistsCmdMeta
	DiceCmds["BF.INFO"] = bfinfoCmdMeta
	DiceCmds["BF.RESERVE"] = bfreserveCmdMeta
	DiceCmds["BGREWRITEAOF"] = bgrewriteaofCmdMeta
	DiceCmds["BITCOUNT"] = bitCountCmdMeta
	DiceCmds["BITFIELD"] = bitfieldCmdMeta
	DiceCmds["BITOP"] = bitOpCmdMeta
	DiceCmds["BITFIELD_RO"] = bitfieldroCmdMeta
	DiceCmds["BITPOS"] = bitposCmdMeta
	DiceCmds["CLIENT"] = clientCmdMeta
	DiceCmds["COMMAND"] = commandCmdMeta
	DiceCmds["COPY"] = copyCmdMeta
	DiceCmds["DBSIZE"] = dbSizeCmdMeta
	DiceCmds["DECR"] = decrCmdMeta
	DiceCmds["DECRBY"] = decrByCmdMeta
	DiceCmds["DEL"] = delCmdMeta
	DiceCmds["DISCARD"] = DiscardCmdMeta
	DiceCmds["DUMP"] = dumpkeyCMmdMeta
	DiceCmds["ECHO"] = echoCmdMeta
	DiceCmds["EXEC"] = ExecCmdMeta
	DiceCmds["EXISTS"] = existsCmdMeta
	DiceCmds["EXPIRE"] = expireCmdMeta
	DiceCmds["EXPIREAT"] = expireatCmdMeta
	DiceCmds["EXPIRETIME"] = expiretimeCmdMeta
	DiceCmds["FLUSHDB"] = flushdbCmdMeta
	DiceCmds["GEOADD"] = geoAddCmdMeta
	DiceCmds["GEODIST"] = geoDistCmdMeta
	DiceCmds["GET"] = getCmdMeta
	DiceCmds["GETBIT"] = getBitCmdMeta
	DiceCmds["GETDEL"] = getDelCmdMeta
	DiceCmds["GETEX"] = getexCmdMeta
	DiceCmds["GETRANGE"] = getRangeCmdMeta
	DiceCmds["GETSET"] = getSetCmdMeta
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
	DiceCmds["INCR"] = incrCmdMeta
	DiceCmds["INCRBYFLOAT"] = incrByFloatCmdMeta
	DiceCmds["INCRBY"] = incrbyCmdMeta
	DiceCmds["INFO"] = infoCmdMeta
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
	DiceCmds["JSON.MGET"] = jsonMGetCmdMeta
	DiceCmds["JSON.NUMINCRBY"] = jsonnumincrbyCmdMeta
	DiceCmds["JSON.NUMMULTBY"] = jsonnummultbyCmdMeta
	DiceCmds["JSON.OBJKEYS"] = jsonobjkeysCmdMeta
	DiceCmds["JSON.OBJLEN"] = jsonobjlenCmdMeta
	DiceCmds["JSON.RESP"] = jsonrespCmdMeta
	DiceCmds["JSON.SET"] = jsonsetCmdMeta
	DiceCmds["JSON.STRLEN"] = jsonStrlenCmdMeta
	DiceCmds["JSON.TOGGLE"] = jsontoggleCmdMeta
	DiceCmds["JSON.TYPE"] = jsontypeCmdMeta
	DiceCmds["KEYS"] = keysCmdMeta
	DiceCmds["LATENCY"] = latencyCmdMeta
	DiceCmds["LLEN"] = llenCmdMeta
	DiceCmds["LPOP"] = lpopCmdMeta
	DiceCmds["LPUSH"] = lpushCmdMeta
	DiceCmds["LRU"] = lruCmdMeta
	DiceCmds["MGET"] = MGetCmdMeta
	DiceCmds["MSET"] = msetCmdMeta
	DiceCmds["MULTI"] = MultiCmdMeta
	DiceCmds["OBJECT"] = objectCmdMeta
	DiceCmds["PERSIST"] = persistCmdMeta
	DiceCmds["PFADD"] = pfAddCmdMeta
	DiceCmds["PFCOUNT"] = pfCountCmdMeta
	DiceCmds["PFMERGE"] = pfMergeCmdMeta
	DiceCmds["PING"] = pingCmdMeta
	DiceCmds["PTTL"] = pttlCmdMeta
	DiceCmds["Q.UNWATCH"] = qUnwatchCmdMeta
	DiceCmds["Q.WATCH"] = qwatchCmdMeta
	DiceCmds["RENAME"] = renameCmdMeta
	DiceCmds["RESTORE"] = restorekeyCmdMeta
	DiceCmds["RPOP"] = rpopCmdMeta
	DiceCmds["RPUSH"] = rpushCmdMeta
	DiceCmds["SADD"] = saddCmdMeta
	DiceCmds["SCARD"] = scardCmdMeta
	DiceCmds["SDIFF"] = sdiffCmdMeta
	DiceCmds["SELECT"] = selectCmdMeta
	DiceCmds["SET"] = setCmdMeta
	DiceCmds["SETBIT"] = setBitCmdMeta
	DiceCmds["SETEX"] = setexCmdMeta
	DiceCmds["SINTER"] = sinterCmdMeta
	DiceCmds["SLEEP"] = sleepCmdMeta
	DiceCmds["SMEMBERS"] = smembersCmdMeta
	DiceCmds["SREM"] = sremCmdMeta
	DiceCmds["SUBSCRIBE"] = subscribeCmdMeta
	DiceCmds["TOUCH"] = touchCmdMeta
	DiceCmds["TTL"] = ttlCmdMeta
	DiceCmds["TYPE"] = typeCmdMeta
	DiceCmds["ZADD"] = zaddCmdMeta
	DiceCmds["ZRANGE"] = zrangeCmdMeta
	DiceCmds["ZPOPMIN"] = zpopminCmdMeta
	DiceCmds["ZRANK"] = zrankCmdMeta
	DiceCmds["JSON.STRAPPEND"] = jsonstrappendCmdMeta
}

// Function to convert DiceCmdMeta to []interface{}
func convertCmdMetaToSlice(cmdMeta *DiceCmdMeta) []interface{} {
	return []interface{}{cmdMeta.Name, cmdMeta.Arity, cmdMeta.KeySpecs.BeginIndex, cmdMeta.KeySpecs.LastKey, cmdMeta.KeySpecs.Step}
}

// Function to convert map[string]DiceCmdMeta{} to []interface{}
func convertDiceCmdsMapToSlice() []interface{} {
	var result []interface{}
	for _, cmdMeta := range DiceCmds {
		result = append(result, convertCmdMetaToSlice(&cmdMeta))
	}
	return result
}
