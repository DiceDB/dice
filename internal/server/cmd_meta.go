package server

import (
	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/comm"
	"github.com/dicedb/dice/internal/eval"
	"github.com/dicedb/dice/internal/shard"
)

// Type defines the type of DiceDB command based on how it interacts with shards.
// It uses an integer value to represent different command types.
type Type int

// Enum values for Type using iota for auto-increment.
// Global commands don't interact with shards, SingleShard commands interact with one shard,
// MultiShard commands interact with multiple shards, and Custom commands require a direct client connection.
const (
	Global      Type = iota // Global commands don't need to interact with shards.
	SingleShard             // Single-shard commands interact with only one shard.
	MultiShard              // MultiShard commands interact with multiple shards using scatter-gather logic.
	Custom                  // Custom commands involve direct client communication.
)

// CmdMeta stores metadata about DiceDB commands, including how they are processed across shards.
// Type indicates how the command should be handled, while Breakup and Gather provide logic
// for breaking up multishard commands and gathering their responses.
type CmdMeta struct {
	Cmd          string                                                                                  // Command name.
	Breakup      func(mgr *shard.ShardManager, DiceDBCmd *cmd.DiceDBCmd, c *comm.Client) []cmd.DiceDBCmd // Function to break up multishard commands.
	Gather       func(responses ...eval.EvalResponse) []byte                                             // Function to gather responses from shards.
	RespNoShards func(args []string) []byte                                                              // Function for commands that don't interact with shards.
	Type                                                                                                 // Enum indicating the command type.
}

// CmdMetaMap is a map that associates command names with their corresponding metadata.
var (
	CmdMetaMap = map[string]CmdMeta{}

	// Metadata for global commands that don't interact with shards.
	// PING is an example of global command.
	pingCmdMeta = CmdMeta{
		Cmd:  "PING",
		Type: Global,
	}

	// Metadata for single-shard commands that only interact with one shard.
	// These commands don't require breakup and gather logic.
	setCmdMeta = CmdMeta{
		Cmd:  "SET",
		Type: SingleShard,
	}
	expireCmdMeta = CmdMeta{
		Cmd:  "EXPIRE",
		Type: SingleShard,
	}
	expireAtCmdMeta = CmdMeta{
		Cmd:  "EXPIREAT",
		Type: SingleShard,
	}
	expireTimeCmdMeta = CmdMeta{
		Cmd:  "EXPIRETIME",
		Type: SingleShard,
	}
	getCmdMeta = CmdMeta{
		Cmd:  "GET",
		Type: SingleShard,
	}
	getsetCmdMeta = CmdMeta{
		Cmd:  "GETSET",
		Type: SingleShard,
	}
	setexCmdMeta = CmdMeta{
		Cmd:  "SETEX",
		Type: SingleShard,
	}
	saddCmdMeta = CmdMeta{
		Cmd:  "SADD",
		Type: SingleShard,
	}
	sremCmdMeta = CmdMeta{
		Cmd:  "SREM",
		Type: SingleShard,
	}
	scardCmdMeta = CmdMeta{
		Cmd:  "SCARD",
		Type: SingleShard,
	}
	smembersCmdMeta = CmdMeta{
		Cmd: "SMEMBERS",
	}

	jsonArrAppendCmdMeta = CmdMeta{
		Cmd:  "JSON.ARRAPPEND",
		Type: SingleShard,
	}
	jsonArrLenCmdMeta = CmdMeta{
		Cmd:  "JSON.ARRLEN",
		Type: SingleShard,
	}
	jsonArrPopCmdMeta = CmdMeta{
		Cmd:  "JSON.ARRPOP",
		Type: SingleShard,
	}
	jsonDebugCmdMeta = CmdMeta{
		Cmd:  "JSON.DEBUG",
		Type: SingleShard,
	}
	jsonRespCmdMeta = CmdMeta{
		Cmd:  "JSON.RESP",
		Type: SingleShard,
	}

	getrangeCmdMeta = CmdMeta{
		Cmd:  "GETRANGE",
		Type: SingleShard,
	}
	hexistsCmdMeta = CmdMeta{
		Cmd:  "HEXISTS",
		Type: SingleShard,
	}
	hkeysCmdMeta = CmdMeta{
		Cmd:  "HKEYS",
		Type: SingleShard,
	}

	hvalsCmdMeta = CmdMeta{
		Cmd:  "HVALS",
		Type: SingleShard,
	}
	zaddCmdMeta = CmdMeta{
		Cmd:  "ZADD",
		Type: SingleShard,
	}
	zcountCmdMeta = CmdMeta{
		Cmd:  "ZCOUNT",
		Type: SingleShard,
	}
	zrangeCmdMeta = CmdMeta{
		Cmd:  "ZRANGE",
		Type: SingleShard,
	}
	appendCmdMeta = CmdMeta{
		Cmd:  "APPEND",
		Type: SingleShard,
	}
	zpopminCmdMeta = CmdMeta{
		Cmd:  "ZPOPMIN",
		Type: SingleShard,
	}
	zrankCmdMeta = CmdMeta{
		Cmd:  "ZRANK",
		Type: SingleShard,
	}
	zcardCmdMeta = CmdMeta{
		Cmd:  "ZCARD",
		Type: SingleShard,
	}
	zremCmdMeta = CmdMeta{
		Cmd:  "ZREM",
		Type: SingleShard,
	}
	pfaddCmdMeta = CmdMeta{
		Cmd:  "PFADD",
		Type: SingleShard,
	}
	pfcountCmdMeta = CmdMeta{
		Cmd:  "PFCOUNT",
		Type: SingleShard,
	}
	pfmergeCmdMeta = CmdMeta{
		Cmd:  "PFMERGE",
		Type: SingleShard,
	}
	ttlCmdMeta = CmdMeta{
		Cmd:  "TTL",
		Type: SingleShard,
	}
	pttlCmdMeta = CmdMeta{
		Cmd:  "PTTL",
		Type: SingleShard,
	}
	setbitCmdMeta = CmdMeta{
		Cmd:  "SETBIT",
		Type: SingleShard,
	}
	getbitCmdMeta = CmdMeta{
		Cmd:  "GETBIT",
		Type: SingleShard,
	}
	bitcountCmdMeta = CmdMeta{
		Cmd:  "BITCOUNT",
		Type: SingleShard,
	}
	bitfieldCmdMeta = CmdMeta{
		Cmd:  "BITFIELD",
		Type: SingleShard,
	}
	bitposCmdMeta = CmdMeta{
		Cmd:  "BITPOS",
		Type: SingleShard,
	}
	bitfieldroCmdMeta = CmdMeta{
		Cmd:  "BITFIELD_RO",
		Type: SingleShard,
	}
	delCmdMeta = CmdMeta{
		Cmd:  "DEL",
		Type: SingleShard,
	}
	existsCmdMeta = CmdMeta{
		Cmd:  "EXISTS",
		Type: SingleShard,
	}
	persistCmdMeta = CmdMeta{
		Cmd:  "PERSIST",
		Type: SingleShard,
	}
	typeCmdMeta = CmdMeta{
		Cmd:  "TYPE",
		Type: SingleShard,
	}

	jsonclearCmdMeta = CmdMeta{
		Cmd:  "JSON.CLEAR",
		Type: SingleShard,
	}

	jsonstrlenCmdMeta = CmdMeta{
		Cmd:  "JSON.STRLEN",
		Type: SingleShard,
	}

	jsonobjlenCmdMeta = CmdMeta{
		Cmd:  "JSON.OBJLEN",
		Type: SingleShard,
	}
	hlenCmdMeta = CmdMeta{
		Cmd:  "HLEN",
		Type: SingleShard,
	}
	hstrlenCmdMeta = CmdMeta{
		Cmd:  "HSTRLEN",
		Type: SingleShard,
	}
	hscanCmdMeta = CmdMeta{
		Cmd:  "HSCAN",
		Type: SingleShard,
	}

	jsonarrinsertCmdMeta = CmdMeta{
		Cmd:  "JSON.ARRINSERT",
		Type: SingleShard,
	}

	jsonarrtrimCmdMeta = CmdMeta{
		Cmd:  "JSON.ARRTRIM",
		Type: SingleShard,
	}

	jsonobjkeystCmdMeta = CmdMeta{
		Cmd:  "JSON.OBJKEYS",
		Type: SingleShard,
	}

	incrCmdMeta = CmdMeta{
		Cmd:  "INCR",
		Type: SingleShard,
	}
	incrByCmdMeta = CmdMeta{
		Cmd:  "INCRBY",
		Type: SingleShard,
	}
	decrCmdMeta = CmdMeta{
		Cmd:  "DECR",
		Type: SingleShard,
	}
	decrByCmdMeta = CmdMeta{
		Cmd:  "DECRBY",
		Type: SingleShard,
	}
	incrByFloatCmdMeta = CmdMeta{
		Cmd:  "INCRBYFLOAT",
		Type: SingleShard,
	}
	hincrbyCmdMeta = CmdMeta{
		Cmd:  "HINCRBY",
		Type: SingleShard,
	}
	hincrbyfloatCmdMeta = CmdMeta{
		Cmd:  "HINCRBYFLOAT",
		Type: SingleShard,
	}
	hrandfieldCmdMeta = CmdMeta{
		Cmd:  "HRANDFIELD",
		Type: SingleShard,
	}
	zpopmaxCmdMeta = CmdMeta{
		Cmd:  "ZPOPMAX",
		Type: SingleShard,
	}
	bfaddCmdMeta = CmdMeta{
		Cmd:  "BF.ADD",
		Type: SingleShard,
	}
	bfreserveCmdMeta = CmdMeta{
		Cmd:  "BF.RESERVE",
		Type: SingleShard,
	}
	bfexistsCmdMeta = CmdMeta{
		Cmd:  "BF.EXISTS",
		Type: SingleShard,
	}
	bfinfoCmdMeta = CmdMeta{
		Cmd:  "BF.INFO",
		Type: SingleShard,
	}
	cmsInitByDimCmdMeta = CmdMeta{
		Cmd:  "CMS.INITBYDIM",
		Type: SingleShard,
	}
	cmsInitByProbCmdMeta = CmdMeta{
		Cmd:  "CMS.INITBYPROB",
		Type: SingleShard,
	}
	cmsInfoCmdMeta = CmdMeta{
		Cmd:  "CMS.INFO",
		Type: SingleShard,
	}
	cmsIncrByCmdMeta = CmdMeta{
		Cmd:  "CMS.INCRBY",
		Type: SingleShard,
	}
	cmsQueryCmdMeta = CmdMeta{
		Cmd:  "CMS.QUERY",
		Type: SingleShard,
	}
	cmsMergeCmdMeta = CmdMeta{
		Cmd:  "CMS.MERGE",
		Type: SingleShard,
	}
	getexCmdMeta = CmdMeta{
		Cmd:  "GETEX",
		Type: SingleShard,
	}
	getdelCmdMeta = CmdMeta{
		Cmd:  "GETDEL",
		Type: SingleShard,
	}
	hsetCmdMeta = CmdMeta{
		Cmd:  "HSET",
		Type: SingleShard,
	}
	hgetCmdMeta = CmdMeta{
		Cmd:  "HGET",
		Type: SingleShard,
	}
	hsetnxCmdMeta = CmdMeta{
		Cmd:  "HSETNX",
		Type: SingleShard,
	}
	hdelCmdMeta = CmdMeta{
		Cmd:  "HDEL",
		Type: SingleShard,
	}
	hmsetCmdMeta = CmdMeta{
		Cmd:  "HMSET",
		Type: SingleShard,
	}
	hmgetCmdMeta = CmdMeta{
		Cmd:  "HMGET",
		Type: SingleShard,
	}
	lrangeCmdMeta = CmdMeta{
		Cmd:  "LRANGE",
		Type: SingleShard,
	}
	linsertCmdMeta = CmdMeta{
		Cmd:  "LINSERT",
		Type: SingleShard,
	}
	lpushCmdMeta = CmdMeta{
		Cmd:  "LPUSH",
		Type: SingleShard,
	}
	rpushCmdMeta = CmdMeta{
		Cmd:  "RPUSH",
		Type: SingleShard,
	}
	lpopCmdMeta = CmdMeta{
		Cmd:  "LPOP",
		Type: SingleShard,
	}
	rpopCmdMeta = CmdMeta{
		Cmd:  "RPOP",
		Type: SingleShard,
	}
	llenCmdMeta = CmdMeta{
		Cmd:  "LLEN",
		Type: SingleShard,
	}
	jsonForgetCmdMeta = CmdMeta{
		Cmd:  "JSON.FORGET",
		Type: SingleShard,
	}
	jsonDelCmdMeta = CmdMeta{
		Cmd:  "JSON.DEL",
		Type: SingleShard,
	}
	jsonToggleCmdMeta = CmdMeta{
		Cmd:  "JSON.TOGGLE",
		Type: SingleShard,
	}
	jsonNumIncrByCmdMeta = CmdMeta{
		Cmd:  "JSON.NUMINCRBY",
		Type: SingleShard,
	}
	jsonNumMultByCmdMeta = CmdMeta{
		Cmd:  "JSON.NUMMULTBY",
		Type: SingleShard,
	}
	jsonSetCmdMeta = CmdMeta{
		Cmd:  "JSON.SET",
		Type: SingleShard,
	}
	jsonGetCmdMeta = CmdMeta{
		Cmd:  "JSON.GET",
		Type: SingleShard,
	}
	jsonTypeCmdMeta = CmdMeta{
		Cmd:  "JSON.TYPE",
		Type: SingleShard,
	}
	jsonIngestCmdMeta = CmdMeta{
		Cmd:  "JSON.INGEST",
		Type: SingleShard,
	}
	jsonArrStrAppendCmdMeta = CmdMeta{
		Cmd:  "JSON.STRAPPEND",
		Type: SingleShard,
	}
	hGetAllCmdMeta = CmdMeta{
		Cmd:  "HGETALL",
		Type: SingleShard,
	}
	dumpCmdMeta = CmdMeta{
		Cmd:  "DUMP",
		Type: SingleShard,
	}
	restoreCmdMeta = CmdMeta{
		Cmd:  "RESTORE",
		Type: SingleShard,
	}
	geoaddCmdMeta = CmdMeta{
		Cmd:  "GEOADD",
		Type: SingleShard,
	}
	geodistCmdMeta = CmdMeta{
		Cmd:  "GEODIST",
		Type: SingleShard,
	}
	clientCmdMeta = CmdMeta{
		Cmd:  "CLIENT",
		Type: SingleShard,
	}
	latencyCmdMeta = CmdMeta{
		Cmd:  "LATENCY",
		Type: SingleShard,
	}
	flushDBCmdMeta = CmdMeta{
		Cmd:  "FLUSHDB",
		Type: MultiShard,
	}
	objectCmdMeta = CmdMeta{
		Cmd:  "OBJECT",
		Type: SingleShard,
	}
	commandCmdMeta = CmdMeta{
		Cmd:  "COMMAND",
		Type: SingleShard,
	}
	CmdCommandCountMeta = CmdMeta{
		Cmd:  "COMMAND|COUNT",
		Type: SingleShard,
	}
	CmdCommandHelp = CmdMeta{
		Cmd:  "COMMAND|HELP",
		Type: SingleShard,
	}
	CmdCommandInfo = CmdMeta{
		Cmd:  "COMMAND|INFO",
		Type: SingleShard,
	}
	CmdCommandList = CmdMeta{
		Cmd:  "COMMAND|LIST",
		Type: SingleShard,
	}
	CmdCommandDocs = CmdMeta{
		Cmd:  "COMMAND|DOCS",
		Type: SingleShard,
	}
	CmdCommandGetKeys = CmdMeta{
		Cmd:  "COMMAND|GETKEYS",
		Type: SingleShard,
	}
	CmdCommandGetKeysFlags = CmdMeta{
		Cmd:  "COMMAND|GETKEYSANDFLAGS",
		Type: SingleShard,
	}

	// Metadata for multishard commands would go here.
	// These commands require both breakup and gather logic.

	// Metadata for custom commands requiring specific client-side logic would go here.
)

// init initializes the CmdMetaMap map by associating each command name with its corresponding metadata.
func init() {
	// Global commands.
	CmdMetaMap["PING"] = pingCmdMeta

	// Single-shard commands.
	CmdMetaMap["SET"] = setCmdMeta
	CmdMetaMap["EXPIRE"] = expireCmdMeta
	CmdMetaMap["EXPIREAT"] = expireAtCmdMeta
	CmdMetaMap["EXPIRETIME"] = expireTimeCmdMeta
	CmdMetaMap["GET"] = getCmdMeta
	CmdMetaMap["GETSET"] = getsetCmdMeta
	CmdMetaMap["SETEX"] = setexCmdMeta

	CmdMetaMap["SADD"] = saddCmdMeta
	CmdMetaMap["SREM"] = sremCmdMeta
	CmdMetaMap["SCARD"] = scardCmdMeta
	CmdMetaMap["SMEMBERS"] = smembersCmdMeta

	CmdMetaMap["JSON.ARRAPPEND"] = jsonArrAppendCmdMeta
	CmdMetaMap["JSON.ARRLEN"] = jsonArrLenCmdMeta
	CmdMetaMap["JSON.ARRPOP"] = jsonArrPopCmdMeta
	CmdMetaMap["JSON.DEBUG"] = jsonDebugCmdMeta
	CmdMetaMap["JSON.RESP"] = jsonRespCmdMeta

	CmdMetaMap["GETRANGE"] = getrangeCmdMeta
	CmdMetaMap["APPEND"] = appendCmdMeta
	CmdMetaMap["JSON.CLEAR"] = jsonclearCmdMeta
	CmdMetaMap["JSON.STRLEN"] = jsonstrlenCmdMeta
	CmdMetaMap["JSON.OBJLEN"] = jsonobjlenCmdMeta
	CmdMetaMap["HEXISTS"] = hexistsCmdMeta
	CmdMetaMap["HKEYS"] = hkeysCmdMeta
	CmdMetaMap["HVALS"] = hvalsCmdMeta
	CmdMetaMap["JSON.ARRINSERT"] = jsonarrinsertCmdMeta
	CmdMetaMap["JSON.ARRTRIM"] = jsonarrtrimCmdMeta
	CmdMetaMap["JSON.OBJKEYS"] = jsonobjkeystCmdMeta
	CmdMetaMap["ZADD"] = zaddCmdMeta
	CmdMetaMap["ZCOUNT"] = zcountCmdMeta
	CmdMetaMap["ZRANGE"] = zrangeCmdMeta
	CmdMetaMap["ZRANK"] = zrankCmdMeta
	CmdMetaMap["ZCARD"] = zcardCmdMeta
	CmdMetaMap["ZREM"] = zremCmdMeta
	CmdMetaMap["PFADD"] = pfaddCmdMeta
	CmdMetaMap["ZPOPMIN"] = zpopminCmdMeta
	CmdMetaMap["PFCOUNT"] = pfcountCmdMeta
	CmdMetaMap["PFMERGE"] = pfmergeCmdMeta
	CmdMetaMap["DEL"] = delCmdMeta
	CmdMetaMap["EXISTS"] = existsCmdMeta
	CmdMetaMap["PERSIST"] = persistCmdMeta
	CmdMetaMap["TYPE"] = typeCmdMeta
	CmdMetaMap["HLEN"] = hlenCmdMeta
	CmdMetaMap["HSTRLEN"] = hstrlenCmdMeta
	CmdMetaMap["HSCAN"] = hscanCmdMeta
	CmdMetaMap["INCR"] = incrCmdMeta
	CmdMetaMap["INCRBY"] = incrByCmdMeta
	CmdMetaMap["INCR"] = incrCmdMeta
	CmdMetaMap["DECR"] = decrCmdMeta
	CmdMetaMap["DECRBY"] = decrByCmdMeta
	CmdMetaMap["INCRBYFLOAT"] = incrByFloatCmdMeta
	CmdMetaMap["HINCRBY"] = hincrbyCmdMeta
	CmdMetaMap["HINCRBYFLOAT"] = hincrbyfloatCmdMeta
	CmdMetaMap["HRANDFIELD"] = hrandfieldCmdMeta
	CmdMetaMap["PFADD"] = pfaddCmdMeta
	CmdMetaMap["ZPOPMIN"] = zpopminCmdMeta
	CmdMetaMap["PFCOUNT"] = pfcountCmdMeta
	CmdMetaMap["PFMERGE"] = pfmergeCmdMeta
	CmdMetaMap["TTL"] = ttlCmdMeta
	CmdMetaMap["PTTL"] = pttlCmdMeta
	CmdMetaMap["HINCRBY"] = hincrbyCmdMeta
	CmdMetaMap["HINCRBYFLOAT"] = hincrbyfloatCmdMeta
	CmdMetaMap["HRANDFIELD"] = hrandfieldCmdMeta
	CmdMetaMap["PFADD"] = pfaddCmdMeta
	CmdMetaMap["PFCOUNT"] = pfcountCmdMeta
	CmdMetaMap["PFMERGE"] = pfmergeCmdMeta
	CmdMetaMap["HINCRBY"] = hincrbyCmdMeta
	CmdMetaMap["HINCRBYFLOAT"] = hincrbyfloatCmdMeta
	CmdMetaMap["HRANDFIELD"] = hrandfieldCmdMeta
	CmdMetaMap["ZPOPMAX"] = zpopmaxCmdMeta
	CmdMetaMap["BF.ADD"] = bfaddCmdMeta
	CmdMetaMap["BF.RESERVE"] = bfreserveCmdMeta
	CmdMetaMap["BF.EXISTS"] = bfexistsCmdMeta
	CmdMetaMap["BF.INFO"] = bfinfoCmdMeta
	CmdMetaMap["CMS.INITBYDIM"] = cmsInitByDimCmdMeta
	CmdMetaMap["CMS.INITBYPROB"] = cmsInitByProbCmdMeta
	CmdMetaMap["CMS.INFO"] = cmsInfoCmdMeta
	CmdMetaMap["CMS.INCRBY"] = cmsIncrByCmdMeta
	CmdMetaMap["CMS.QUERY"] = cmsQueryCmdMeta
	CmdMetaMap["CMS.MERGE"] = cmsMergeCmdMeta
	CmdMetaMap["GETEX"] = getexCmdMeta
	CmdMetaMap["GETDEL"] = getdelCmdMeta
	CmdMetaMap["HSET"] = hsetCmdMeta
	CmdMetaMap["HGET"] = hgetCmdMeta
	CmdMetaMap["HSETNX"] = hsetnxCmdMeta
	CmdMetaMap["HDEL"] = hdelCmdMeta
	CmdMetaMap["HMSET"] = hmsetCmdMeta
	CmdMetaMap["HMGET"] = hmgetCmdMeta
	CmdMetaMap["SETBIT"] = setbitCmdMeta
	CmdMetaMap["GETBIT"] = getbitCmdMeta
	CmdMetaMap["BITCOUNT"] = bitcountCmdMeta
	CmdMetaMap["BITFIELD"] = bitfieldCmdMeta
	CmdMetaMap["BITPOS"] = bitposCmdMeta
	CmdMetaMap["BITFIELD_RO"] = bitfieldroCmdMeta
	CmdMetaMap["LRANGE"] = lrangeCmdMeta
	CmdMetaMap["LINSERT"] = linsertCmdMeta
	CmdMetaMap["LPUSH"] = lpushCmdMeta
	CmdMetaMap["RPUSH"] = rpushCmdMeta
	CmdMetaMap["LPOP"] = lpopCmdMeta
	CmdMetaMap["RPOP"] = rpopCmdMeta
	CmdMetaMap["LLEN"] = llenCmdMeta
	CmdMetaMap["JSON.FORGET"] = jsonForgetCmdMeta
	CmdMetaMap["JSON.DEL"] = jsonDelCmdMeta
	CmdMetaMap["JSON.TOGGLE"] = jsonToggleCmdMeta
	CmdMetaMap["JSON.NUMINCRBY"] = jsonNumIncrByCmdMeta
	CmdMetaMap["JSON.NUMMULTBY"] = jsonNumMultByCmdMeta
	CmdMetaMap["JSON.SET"] = jsonSetCmdMeta
	CmdMetaMap["JSON.GET"] = jsonGetCmdMeta
	CmdMetaMap["JSON.TYPE"] = jsonTypeCmdMeta
	CmdMetaMap["JSON.INGEST"] = jsonIngestCmdMeta
	CmdMetaMap["JSON.STRAPPEND"] = jsonArrStrAppendCmdMeta
	CmdMetaMap["HGETALL"] = hGetAllCmdMeta
	CmdMetaMap["DUMP"] = dumpCmdMeta
	CmdMetaMap["RESTORE"] = restoreCmdMeta
	CmdMetaMap["GEOADD"] = geoaddCmdMeta
	CmdMetaMap["GEODIST"] = geodistCmdMeta
	CmdMetaMap["CLIENT"] = clientCmdMeta
	CmdMetaMap["LATENCY"] = latencyCmdMeta
	CmdMetaMap["FLUSHDB"] = flushDBCmdMeta
	CmdMetaMap["OBJECT"] = objectCmdMeta
	CmdMetaMap["COMMAND"] = commandCmdMeta
	CmdMetaMap["COMMAND|COUNT"] = CmdCommandCountMeta
	CmdMetaMap["COMMAND|HELP"] = CmdCommandHelp
	CmdMetaMap["COMMAND|INFO"] = CmdCommandInfo
	CmdMetaMap["COMMAND|LIST"] = CmdCommandList
	CmdMetaMap["COMMAND|DOCS"] = CmdCommandDocs
	CmdMetaMap["COMMAND|GETKEYS"] = CmdCommandGetKeys
	CmdMetaMap["COMMAND|GETKEYSANDFLAGS"] = CmdCommandGetKeysFlags
}
