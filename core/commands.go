package core

type DiceCmdMeta struct {
	Name string
	Info string
	Eval func([]string) []byte
}

var (
	diceCmds = map[string]DiceCmdMeta{}

	pingCmdMeta = DiceCmdMeta{
		Name: "PING",
		Info: `PING returns with an encoded "PONG" If any message is added with the ping command,the message will be returned.`,
		Eval: evalPING,
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
		Eval: evalSET,
	}
	getCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: `GET returns the value for the queried key in args
		The key should be the only param in args
		The RESP value of the key is encoded and then returned
		GET returns RESP_NIL if key is expired or it does not exist`,
		Eval: evalGET,
	}
	jsonsetCmdMeta = DiceCmdMeta{
		Name: "JSON.SET",
		Info: `JSON.SET key path json-string
		Sets a JSON value at the specified key.
		Returns OK if successful.
		Returns encoded error message if the number of arguments is incorrect or the JSON string is invalid.`,
		Eval: evalJSONSET,
	}

	jsongetCmdMeta = DiceCmdMeta{
		Name: "JSON.GET",
		Info: `JSON.GET key [path]
		Returns the encoded RESP value of the key, if present
		Null reply: If the key doesn't exist or has expired.
		Error reply: If the number of arguments is incorrect or the stored value is not a JSON type.`,
		Eval: evalJSONGET,
	}
	ttlCmdMeta = DiceCmdMeta{
		Name: "TTL",
		Info: `TTL returns Time-to-Live in secs for the queried key in args
		 The key should be the only param in args else returns with an error
		 Returns	
		 RESP encoded time (in secs) remaining for the key to expire
		 RESP encoded -2 stating key doesn't exist or key is expired
		 RESP encoded -1 in case no expiration is set on the key`,
		Eval: evalTTL,
	}
	delCmdMeta = DiceCmdMeta{
		Name: "DEL",
		Info: `DEL deletes all the specified keys in args list
		returns the count of total deleted keys after encoding`,
		Eval: evalDEL,
	}
	expireCmdMeta = DiceCmdMeta{
		Name: "EXPIRE",
		Info: `EXPIRE sets a expiry time(in secs) on the specified key in args
		args should contain 2 values, key and the expiry time to be set for the key
		The expiry time should be in integer format; if not, it returns encoded error response
		Returns RESP_ONE if expiry was set on the key successfully.
		Once the time is lapsed, the key will be deleted automatically`,
		Eval: evalEXPIRE,
	}
	helloCmdMeta = DiceCmdMeta{
		Name: "HELLO",
		Info: `HELLO always replies with a list of current server and connection properties, such as: versions, modules loaded, client ID, replication role and so forth`,
		Eval: evalHELLO,
	}
	bgrewriteaofCmdMeta = DiceCmdMeta{
		Name: "BGREWRITEAOF",
		Info: `Instruct Dice to start an Append Only File rewrite process. The rewrite will create a small optimized version of the current Append Only File.`,
		Eval: evalBGREWRITEAOF,
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
		Eval: evalINCR,
	}
	infoCmdMeta = DiceCmdMeta{
		Name: "INFO",
		Info: `INFO creates a buffer with the info of total keys per db
		Returns the encoded buffer as response`,
		Eval: evalINFO,
	}
	clientCmdMeta = DiceCmdMeta{
		Name: "CLIENT",
		Info: `This is a container command for client connection commands.`,
		Eval: evalCLIENT,
	}
	latencyCmdMeta = DiceCmdMeta{
		Name: "LATENCY",
		Info: `This is a container command for latency diagnostics commands.`,
		Eval: evalLATENCY,
	}
	lruCmdMeta = DiceCmdMeta{
		Name: "LRU",
		Info: `LRU deletes all the keys from the LRU
		returns encoded RESP OK`,
		Eval: evalLRU,
	}
	sleepCmdMeta = DiceCmdMeta{
		Name: "SLEEP",
		Info: `SLEEP sets db to sleep for the specified number of seconds.
		The sleep time should be the only param in args.
		Returns error response if the time param in args is not of integer format.
		SLEEP returns RESP_OK after sleeping for mentioned seconds`,
		Eval: evalSLEEP,
	}
	qintinsCmdMeta = DiceCmdMeta{
		Name: "QINTINS",
		Info: `QINTINS inserts the provided integer in the key identified by key
		first argument will be the key, that should be of type "QINT"
		second argument will be the integer value
		if the key does not exist, QINTINS will also create the integer queue`,
		Eval: evalQINTINS,
	}
	qintremCmdMeta = DiceCmdMeta{
		Name: "QINTREM",
		Info: `QINTREM removes the element from the QINT identified by key
		first argument will be the key, that should be of type "QINT"
		if the key does not exist, QINTREM returns nil otherwise it
		returns the integer value popped from the queue
		if we remove from the empty queue, nil is returned`,
		Eval: evalQINTREM,
	}
	qintlenCmdMeta = DiceCmdMeta{
		Name: "QINTLEN",
		Info: `QINTLEN returns the length of the QINT identified by key
		returns the integer value indicating the length of the queue
		if the key does not exist, the response is 0`,
		Eval: evalQINTLEN,
	}
	qintpeekCmdMeta = DiceCmdMeta{
		Name: "QINTPEEK",
		Info: `QINTPEEK peeks into the QINT and returns 5 elements without popping them
		returns the array of integers as the response.
		if the key does not exist, then we return an empty array`,
		Eval: evalQINTPEEK,
	}
	bfinitCmdMeta = DiceCmdMeta{
		Name: "BFINIT",
		Info: `BFINIT command initializes a new bloom filter and allocation it's relevant parameters based on given inputs.
		If no params are provided, it uses defaults.`,
		Eval: evalBFINIT,
	}
	bfaddCmdMeta = DiceCmdMeta{
		Name: "BFADD",
		Info: `BFADD adds an element to
		a bloom filter. If the filter does not exists, it will create a new one
		with default parameters.`,
		Eval: evalBFADD,
	}
	bfexistsCmdMeta = DiceCmdMeta{
		Name: "BFEXISTS",
		Info: `BFEXISTS checks existance of an element in a bloom filter.`,
		Eval: evalBFEXISTS,
	}
	bfinfoCmdMeta = DiceCmdMeta{
		Name: "BFINFO",
		Info: `BFINFO returns the parameters and metadata of an existing bloom filter.`,
		Eval: evalBFINFO,
	}
	qrefinsCmdMeta = DiceCmdMeta{
		Name: "QREFINS",
		Info: `QREFINS inserts the reference of the provided key identified by key
		first argument will be the key, that should be of type "QREF"
		second argument will be the key that needs to be added to the queueref
		if the queue does not exist, QREFINS will also create the queueref
		returns 1 if the key reference was inserted
		returns 0 otherwise`,
		Eval: evalQREFINS,
	}
	qrefremCmdMeta = DiceCmdMeta{
		Name: "QREFREM",
		Info: `QREFREM removes the element from the QREF identified by key
		first argument will be the key, that should be of type "QREF"
		if the key does not exist, QREFREM returns nil otherwise it
		returns the RESP encoded value of the key reference from the queue
		if we remove from the empty queue, nil is returned`,
		Eval: evalQREFREM,
	}
	qreflenCmdMeta = DiceCmdMeta{
		Name: "QREFLEN",
		Info: `QREFLEN returns the length of the QREF identified by key
		returns the integer value indicating the length of the queue
		if the key does not exist, the response is 0`,
		Eval: evalQREFLEN,
	}
	qrefpeekCmdMeta = DiceCmdMeta{
		Name: "QREFPEEK",
		Info: `QREFPEEK peeks into the QREF and returns 5 elements without popping them
		returns the array of resp encoded values as the response.
		if the key does not exist, then we return an empty array`,
		Eval: evalQREFPEEK,
	}
	stackintpushCmdMeta = DiceCmdMeta{
		Name: "STACKINTPUSH",
		Info: `STACKINTPUSH pushes the provided integer in the key identified by key
		first argument will be the key, that should be of type "STACKINT"
		second argument will be the integer value
		if the key does not exist, STACKINTPUSH will also create the integer stack`,
		Eval: evalSTACKINTPUSH,
	}
	stackintpopCmdMeta = DiceCmdMeta{
		Name: "STACKINTPOP",
		Info: `STACKINTPOP pops the element from the STACKINT identified by key
		first argument will be the key, that should be of type "STACKINT"
		if the key does not exist, STACKINTPOP returns nil otherwise it
		returns the integer value popped from the stack
		if we remove from the empty stack, nil is returned`,
		Eval: evalSTACKINTPOP,
	}
	stackintlenCmdMeta = DiceCmdMeta{
		Name: "STACKINTLEN",
		Info: `STACKINTLEN returns the length of the STACKINT identified by key
		returns the integer value indicating the length of the stack
		if the key does not exist, the response is 0`,
		Eval: evalSTACKINTLEN,
	}
	stackintpeekCmdMeta = DiceCmdMeta{
		Name: "STACKINTPEEK",
		Info: `STACKINTPEEK peeks into the DINT and returns 5 elements without popping them
		returns the array of integers as the response.
		if the key does not exist, then we return an empty array`,
		Eval: evalSTACKINTPEEK,
	}
	stackrefpushCmdMeta = DiceCmdMeta{
		Name: "STACKREFPUSH",
		Info: `STACKREFPUSH inserts the reference of the provided key identified by key
		first argument will be the key, that should be of type "STACKREF"
		second argument will be the key that needs to be added to the stackref
		if the stack does not exist, STACKREFPUSH will also create the stackref
		returns 1 if the key reference was inserted
		returns 0 otherwise`,
		Eval: evalSTACKREFPUSH,
	}
	stackrefpopCmdMeta = DiceCmdMeta{
		Name: "STACKREFPOP",
		Info: `STACKREFPOP removes the element from the DREF identified by key
		first argument will be the key, that should be of type "STACKREF"
		if the key does not exist, STACKREFPOP returns nil otherwise it
		returns the RESP encoded value of the key reference from the stack
		if we remove from the empty stack, nil is returned`,
		Eval: evalSTACKREFPOP,
	}
	stackreflenCmdMeta = DiceCmdMeta{
		Name: "STACKREFLEN",
		Info: `STACKREFLEN returns the length of the STACKREF identified by key
		returns the integer value indicating the length of the stack
		if the key does not exist, the response is 0`,
		Eval: evalSTACKREFLEN,
	}
	stackrefpeekCmdMeta = DiceCmdMeta{
		Name: "STACKREFPEEK",
		Info: `STACKREFPEEK peeks into the STACKREF and returns 5 elements without popping them
		returns the array of resp encoded values as the response.
		if the key does not exist, then we return an empty array`,
		Eval: evalSTACKREFPEEK,
	}
	// TODO: Remove this override once we support QWATCH in dice-cli.
	subscribeCmdMeta = DiceCmdMeta{
		Name: "SUBSCRIBE",
		Info: `SUBSCRIBE(or QWATCH) adds the specified key to the watch list for the caller client.
		Every time a key in the watch list is modified, the client will be sent a response
		containing the new value of the key along with the operation that was performed on it.
		Contains only one argument, the key to be watched.`,
		Eval: nil,
	}
	qwatchCmdMeta = DiceCmdMeta{
		Name: "QWATCH",
		Info: `QWATCH adds the specified key to the watch list for the caller client.
		Every time a key in the watch list is modified, the client will be sent a response
		containing the new value of the key along with the operation that was performed on it.
		Contains only one argument, the key to be watched.`,
		Eval: nil,
	}
	multiCmdMeta = DiceCmdMeta{
		Name: "MULTI",
		Info: `MULTI marks the start of the transaction for the client.
		All subsequent commands fired will be queued for atomic execution.
		The commands will not be executed until EXEC is triggered.
		Once EXEC is triggered it executes all the commands in queue,
		and closes the MULTI transaction.`,
		Eval: evalMULTI,
	}
	execCmdMeta = DiceCmdMeta{
		Name: "EXEC",
		Info: `EXEC executes commands in a transaction, which is initiated by MULTI`,
		Eval: nil,
	}
	discardCmdMeta = DiceCmdMeta{
		Name: "DISCARD",
		Info: `DISCARD discards all the commands in a transaction, which is initiated by MULTI`,
		Eval: nil,
	}
	abortCmdMeta = DiceCmdMeta{
		Name: "ABORT",
		Info: "Quit the server",
		Eval: nil,
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
		Name: "COMMAND <subcommand>",
		Info: "Evaluates COMMAND <subcommand> command based on subcommand",
		Eval: evalCommand,
	}
)

func init() {
	diceCmds["PING"] = pingCmdMeta
	diceCmds["SET"] = setCmdMeta
	diceCmds["GET"] = getCmdMeta
	diceCmds["JSON.SET"] = jsonsetCmdMeta
	diceCmds["JSON.GET"] = jsongetCmdMeta
	diceCmds["TTL"] = ttlCmdMeta
	diceCmds["DEL"] = delCmdMeta
	diceCmds["EXPIRE"] = expireCmdMeta
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
	diceCmds["MULTI"] = multiCmdMeta
	diceCmds["EXEC"] = execCmdMeta
	diceCmds["DISCARD"] = discardCmdMeta
	diceCmds["ABORT"] = abortCmdMeta
	diceCmds["COMMAND"] = commandCmdMeta
	diceCmds["SETBIT"] = setBitCmdMeta
	diceCmds["GETBIT"] = getBitCmdMeta
	diceCmds["BITCOUNT"] = bitCountCmdMeta
	diceCmds["BITOP"] = bitOpCmdMeta
}
