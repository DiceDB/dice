package core

type DiceCmdMeta struct {
	Name  string
	Info  string
	Eval  func([]string, *Store) []byte
	Arity int // number of arguments, it is possible to use -N to say >= N
	KeySpecs
}

type KeySpecs struct {
	BeginIndex int
	Step       int
	LastKey    int
}

var (
	diceCmds = map[string]DiceCmdMeta{}

	pingCmdMeta = DiceCmdMeta{
		Name:  "PING",
		Info:  `PING returns with an encoded "PONG" If any message is added with the ping command,the message will be returned.`,
		Eval:  evalPING,
		Arity: -1,
	}
	authCmdMeta = DiceCmdMeta{
		Name: "AUTH",
		Info: `AUTH returns with an encoded "OK" if the user is authenticated.
		If the user is not authenticated, it returns with an encoded error message`,
		Eval: nil,
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
		Eval:     evalSET,
		Arity:    -3,
		KeySpecs: KeySpecs{BeginIndex: 1},
	}
	getCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: `GET returns the value for the queried key in args
		The key should be the only param in args
		The RESP value of the key is encoded and then returned
		GET returns RespNIL if key is expired or it does not exist`,
		Eval:     evalGET,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1},
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
		Eval:  evalBGREWRITEAOF,
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
		Eval:     evalINCR,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
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
	qintinsCmdMeta = DiceCmdMeta{
		Name: "QINTINS",
		Info: `QINTINS inserts the provided integer in the key identified by key
		first argument will be the key, that should be of type "QINT"
		second argument will be the integer value
		if the key does not exist, QINTINS will also create the integer queue`,
		Eval:     evalQINTINS,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	qintremCmdMeta = DiceCmdMeta{
		Name: "QINTREM",
		Info: `QINTREM removes the element from the QINT identified by key
		first argument will be the key, that should be of type "QINT"
		if the key does not exist, QINTREM returns nil otherwise it
		returns the integer value popped from the queue
		if we remove from the empty queue, nil is returned`,
		Eval:     evalQINTREM,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	qintlenCmdMeta = DiceCmdMeta{
		Name: "QINTLEN",
		Info: `QINTLEN returns the length of the QINT identified by key
		returns the integer value indicating the length of the queue
		if the key does not exist, the response is 0`,
		Eval:     evalQINTLEN,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	qintpeekCmdMeta = DiceCmdMeta{
		Name: "QINTPEEK",
		Info: `QINTPEEK peeks into the QINT and returns 5 elements without popping them
		returns the array of integers as the response.
		if the key does not exist, then we return an empty array`,
		Eval:     evalQINTPEEK,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfinitCmdMeta = DiceCmdMeta{
		Name: "BFINIT",
		Info: `BFINIT command initializes a new bloom filter and allocation it's relevant parameters based on given inputs.
		If no params are provided, it uses defaults.`,
		Eval:     evalBFINIT,
		Arity:    -2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfaddCmdMeta = DiceCmdMeta{
		Name: "BFADD",
		Info: `BFADD adds an element to
		a bloom filter. If the filter does not exists, it will create a new one
		with default parameters.`,
		Eval:     evalBFADD,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfexistsCmdMeta = DiceCmdMeta{
		Name:     "BFEXISTS",
		Info:     `BFEXISTS checks existence of an element in a bloom filter.`,
		Eval:     evalBFEXISTS,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	bfinfoCmdMeta = DiceCmdMeta{
		Name:  "BFINFO",
		Info:  `BFINFO returns the parameters and metadata of an existing bloom filter.`,
		Eval:  evalBFINFO,
		Arity: 2,
	}
	qrefinsCmdMeta = DiceCmdMeta{
		Name: "QREFINS",
		Info: `QREFINS inserts the reference of the provided key identified by key
		first argument will be the key, that should be of type "QREF"
		second argument will be the key that needs to be added to the queueref
		if the queue does not exist, QREFINS will also create the queueref
		returns 1 if the key reference was inserted
		returns 0 otherwise`,
		Eval:     evalQREFINS,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	qrefremCmdMeta = DiceCmdMeta{
		Name: "QREFREM",
		Info: `QREFREM removes the element from the QREF identified by key
		first argument will be the key, that should be of type "QREF"
		if the key does not exist, QREFREM returns nil otherwise it
		returns the RESP encoded value of the key reference from the queue
		if we remove from the empty queue, nil is returned`,
		Eval:     evalQREFREM,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	qreflenCmdMeta = DiceCmdMeta{
		Name: "QREFLEN",
		Info: `QREFLEN returns the length of the QREF identified by key
		returns the integer value indicating the length of the queue
		if the key does not exist, the response is 0`,
		Eval:     evalQREFLEN,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	qrefpeekCmdMeta = DiceCmdMeta{
		Name: "QREFPEEK",
		Info: `QREFPEEK peeks into the QREF and returns 5 elements without popping them
		returns the array of resp encoded values as the response.
		if the key does not exist, then we return an empty array`,
		Eval:     evalQREFPEEK,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackintpushCmdMeta = DiceCmdMeta{
		Name: "STACKINTPUSH",
		Info: `STACKINTPUSH pushes the provided integer in the key identified by key
		first argument will be the key, that should be of type "STACKINT"
		second argument will be the integer value
		if the key does not exist, STACKINTPUSH will also create the integer stack`,
		Eval:     evalSTACKINTPUSH,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackintpopCmdMeta = DiceCmdMeta{
		Name: "STACKINTPOP",
		Info: `STACKINTPOP pops the element from the STACKINT identified by key
		first argument will be the key, that should be of type "STACKINT"
		if the key does not exist, STACKINTPOP returns nil otherwise it
		returns the integer value popped from the stack
		if we remove from the empty stack, nil is returned`,
		Eval:     evalSTACKINTPOP,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackintlenCmdMeta = DiceCmdMeta{
		Name: "STACKINTLEN",
		Info: `STACKINTLEN returns the length of the STACKINT identified by key
		returns the integer value indicating the length of the stack
		if the key does not exist, the response is 0`,
		Eval:     evalSTACKINTLEN,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackintpeekCmdMeta = DiceCmdMeta{
		Name: "STACKINTPEEK",
		Info: `STACKINTPEEK peeks into the DINT and returns 5 elements without popping them
		returns the array of integers as the response.
		if the key does not exist, then we return an empty array`,
		Eval:     evalSTACKINTPEEK,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackrefpushCmdMeta = DiceCmdMeta{
		Name: "STACKREFPUSH",
		Info: `STACKREFPUSH inserts the reference of the provided key identified by key
		first argument will be the key, that should be of type "STACKREF"
		second argument will be the key that needs to be added to the stackref
		if the stack does not exist, STACKREFPUSH will also create the stackref
		returns 1 if the key reference was inserted
		returns 0 otherwise`,
		Eval:     evalSTACKREFPUSH,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackrefpopCmdMeta = DiceCmdMeta{
		Name: "STACKREFPOP",
		Info: `STACKREFPOP removes the element from the DREF identified by key
		first argument will be the key, that should be of type "STACKREF"
		if the key does not exist, STACKREFPOP returns nil otherwise it
		returns the RESP encoded value of the key reference from the stack
		if we remove from the empty stack, nil is returned`,
		Eval:     evalSTACKREFPOP,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackreflenCmdMeta = DiceCmdMeta{
		Name: "STACKREFLEN",
		Info: `STACKREFLEN returns the length of the STACKREF identified by key
		returns the integer value indicating the length of the stack
		if the key does not exist, the response is 0`,
		Eval:     evalSTACKREFLEN,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
	}
	stackrefpeekCmdMeta = DiceCmdMeta{
		Name: "STACKREFPEEK",
		Info: `STACKREFPEEK peeks into the STACKREF and returns 5 elements without popping them
		returns the array of resp encoded values as the response.
		if the key does not exist, then we return an empty array`,
		Eval:     evalSTACKREFPEEK,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
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
		Name: "QWATCH",
		Info: `QWATCH adds the specified key to the watch list for the caller client.
		Every time a key in the watch list is modified, the client will be sent a response
		containing the new value of the key along with the operation that was performed on it.
		Contains only one argument, the key to be watched.`,
		Eval:  nil,
		Arity: 1,
	}
	qUnwatchCmdMeta = DiceCmdMeta{
		Name: "QUNWATCH",
		Info: `Unsubscribes or QUnwatches the client from the given key's watch session.
		It removes the key from the watch list for the caller client.`,
		Eval:  nil,
		Arity: 1,
	}
	multiCmdMeta = DiceCmdMeta{
		Name: "MULTI",
		Info: `MULTI marks the start of the transaction for the client.
		All subsequent commands fired will be queued for atomic execution.
		The commands will not be executed until EXEC is triggered.
		Once EXEC is triggered it executes all the commands in queue,
		and closes the MULTI transaction.`,
		Eval:  evalMULTI,
		Arity: 1,
	}
	execCmdMeta = DiceCmdMeta{
		Name:  "EXEC",
		Info:  `EXEC executes commands in a transaction, which is initiated by MULTI`,
		Eval:  nil,
		Arity: 1,
	}
	discardCmdMeta = DiceCmdMeta{
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
		Name: "BITCOUNT",
		Info: "BITCOUNT counts the number of set bits in the string value stored at key",
		Eval: evalBITCOUNT,
	}
	bitOpCmdMeta = DiceCmdMeta{
		Name: "BITOP",
		Info: "BITOP performs bitwise operations between multiple keys",
		Eval: evalBITOP,
	}
	commandCmdMeta = DiceCmdMeta{
		Name:  "COMMAND <subcommand>",
		Info:  "Evaluates COMMAND <subcommand> command based on subcommand",
		Eval:  evalCommand,
		Arity: -1,
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
		Eval:     evalDECR,
		Arity:    2,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
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
		Eval:     evalDECRBY,
		Arity:    3,
		KeySpecs: KeySpecs{BeginIndex: 1, Step: 1},
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
		Name:  "LPUSH",
		Info:  "LPUSH pushes values into the left side of the deque",
		Eval:  evalLPUSH,
		Arity: -3,
	}
	rpushCmdMeta = DiceCmdMeta{
		Name:  "RPUSH",
		Info:  "RPUSH pushes values into the right side of the deque",
		Eval:  evalRPUSH,
		Arity: -3,
	}
	lpopCmdMeta = DiceCmdMeta{
		Name:  "LPOP",
		Info:  "LPOP pops a value from the left side of the deque",
		Eval:  evalLPOP,
		Arity: 2,
	}
	rpopCmdMeta = DiceCmdMeta{
		Name:  "RPOP",
		Info:  "RPOP pops a value from the right side of the deque",
		Eval:  evalRPOP,
		Arity: 2,
	}
	dbSizeCmdMeta = DiceCmdMeta{
		Name:  "DBSIZE",
		Info:  `DBSIZE Return the number of keys in the database`,
		Eval:  evalDBSIZE,
		Arity: 1,
	}
	getSetCmdMeta = DiceCmdMeta{
		Name:  "GETSET",
		Info:  `GETSET returns the previous string value of a key after setting it to a new value.`,
		Eval:  evalGETSET,
		Arity: 2,
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
)

func init() {
	diceCmds["PING"] = pingCmdMeta
	diceCmds["AUTH"] = authCmdMeta
	diceCmds["SET"] = setCmdMeta
	diceCmds["GET"] = getCmdMeta
	diceCmds["MSET"] = msetCmdMeta
	diceCmds["JSON.SET"] = jsonsetCmdMeta
	diceCmds["JSON.GET"] = jsongetCmdMeta
	diceCmds["JSON.TYPE"] = jsontypeCmdMeta
	diceCmds["TTL"] = ttlCmdMeta
	diceCmds["DEL"] = delCmdMeta
	diceCmds["EXPIRE"] = expireCmdMeta
	diceCmds["EXPIRETIME"] = expiretimeCmdMeta
	diceCmds["EXPIREAT"] = expireatCmdMeta
	diceCmds["HELLO"] = helloCmdMeta
	diceCmds["BGREWRITEAOF"] = bgrewriteaofCmdMeta
	diceCmds["INCR"] = incrCmdMeta
	diceCmds["INFO"] = infoCmdMeta
	diceCmds["CLIENT"] = clientCmdMeta
	diceCmds["LATENCY"] = latencyCmdMeta
	diceCmds["LRU"] = lruCmdMeta
	diceCmds["SLEEP"] = sleepCmdMeta
	diceCmds["QINTINS"] = qintinsCmdMeta
	diceCmds["QINTREM"] = qintremCmdMeta
	diceCmds["QINTLEN"] = qintlenCmdMeta
	diceCmds["QINTPEEK"] = qintpeekCmdMeta
	diceCmds["BFINIT"] = bfinitCmdMeta
	diceCmds["BFADD"] = bfaddCmdMeta
	diceCmds["BFEXISTS"] = bfexistsCmdMeta
	diceCmds["BFINFO"] = bfinfoCmdMeta
	diceCmds["QREFINS"] = qrefinsCmdMeta
	diceCmds["QREFREM"] = qrefremCmdMeta
	diceCmds["QREFLEN"] = qreflenCmdMeta
	diceCmds["QREFPEEK"] = qrefpeekCmdMeta
	diceCmds["STACKINTPUSH"] = stackintpushCmdMeta
	diceCmds["STACKINTPOP"] = stackintpopCmdMeta
	diceCmds["STACKINTLEN"] = stackintlenCmdMeta
	diceCmds["STACKINTPEEK"] = stackintpeekCmdMeta
	diceCmds["STACKREFPUSH"] = stackrefpushCmdMeta
	diceCmds["STACKREFPOP"] = stackrefpopCmdMeta
	diceCmds["STACKREFLEN"] = stackreflenCmdMeta
	diceCmds["STACKREFPEEK"] = stackrefpeekCmdMeta
	diceCmds["SUBSCRIBE"] = subscribeCmdMeta
	diceCmds["QWATCH"] = qwatchCmdMeta
	diceCmds["QUNWATCH"] = qUnwatchCmdMeta
	diceCmds["MULTI"] = multiCmdMeta
	diceCmds["EXEC"] = execCmdMeta
	diceCmds["DISCARD"] = discardCmdMeta
	diceCmds["ABORT"] = abortCmdMeta
	diceCmds["COMMAND"] = commandCmdMeta
	diceCmds["SETBIT"] = setBitCmdMeta
	diceCmds["GETBIT"] = getBitCmdMeta
	diceCmds["BITCOUNT"] = bitCountCmdMeta
	diceCmds["BITOP"] = bitOpCmdMeta
	diceCmds["KEYS"] = keysCmdMeta
	diceCmds["MGET"] = MGetCmdMeta
	diceCmds["PERSIST"] = persistCmdMeta
	diceCmds["COPY"] = copyCmdMeta
	diceCmds["DECR"] = decrCmdMeta
	diceCmds["EXISTS"] = existsCmdMeta
	diceCmds["GETDEL"] = getDelCmdMeta
	diceCmds["DECRBY"] = decrByCmdMeta
	diceCmds["RENAME"] = renameCmdMeta
	diceCmds["GETEX"] = getexCmdMeta
	diceCmds["PTTL"] = pttlCmdMeta
	diceCmds["OBJECT"] = objectCmdMeta
	diceCmds["TOUCH"] = touchCmdMeta
	diceCmds["LPUSH"] = lpushCmdMeta
	diceCmds["RPOP"] = rpopCmdMeta
	diceCmds["RPUSH"] = rpushCmdMeta
	diceCmds["LPOP"] = lpopCmdMeta
	diceCmds["DBSIZE"] = dbSizeCmdMeta
	diceCmds["GETSET"] = getSetCmdMeta
	diceCmds["FLUSHDB"] = flushdbCmdMeta
	diceCmds["BITPOS"] = bitposCmdMeta
}
