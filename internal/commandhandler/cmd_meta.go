// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package commandhandler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dicedb/dice/internal/cmd"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/ops"
)

type CmdType int

const (
	// Global represents a command that applies globally across all shards or nodes.
	// This type of command doesn't target a specific shard but affects the entire system.
	Global CmdType = iota

	// SingleShard represents a command that operates on a single shard.
	// This command is scoped to execute on one specific shard, optimizing for shard-local operations.
	SingleShard

	// MultiShard represents a command that operates across multiple shards.
	// This type of command spans more than one shard and may involve coordination between shards.
	MultiShard

	// AllShard represents a command that operates across all available shards.
	// This type of command spans more than one shard and may involve coordination between shards.
	AllShard

	// Custom represents a command that is user-defined or has custom logic.
	// This command type allows for flexibility in executing specific, non-standard operations.
	Custom

	// Watch represents a command that is used to monitor changes or events.
	// This type of command listens for changes on specific keys or resources and responds accordingly.
	Watch

	// Unwatch represents a command that is used to stop monitoring changes or events.
	// This type of command stops listening for changes on specific keys or resources.
	Unwatch
)

// Global commands
const (
	CmdPing  = "PING"
	CmdAbort = "ABORT"
	CmdAuth  = "AUTH"
	CmdEcho  = "ECHO"
	CmdHello = "HELLO"
	CmdSleep = "SLEEP"
)

// Single-shard commands.
const (
	CmdHExists             = "HEXISTS"
	CmdHKeys               = "HKEYS"
	CmdHVals               = "HVALS"
	CmdZPopMin             = "ZPOPMIN"
	CmdZAdd                = "ZADD"
	CmdZRange              = "ZRANGE"
	CmdZRank               = "ZRANK"
	CmdZCount              = "ZCOUNT"
	CmdZRem                = "ZREM"
	CmdZCard               = "ZCARD"
	CmdPFAdd               = "PFADD"
	CmdPFCount             = "PFCOUNT"
	CmdPFMerge             = "PFMERGE"
	CmdTTL                 = "TTL"
	CmdPTTL                = "PTTL"
	CmdIncr                = "INCR"
	CmdIncrBy              = "INCRBY"
	CmdDecr                = "DECR"
	CmdDecrBy              = "DECRBY"
	CmdIncrByFloat         = "INCRBYFLOAT"
	CmdHIncrBy             = "HINCRBY"
	CmdHIncrByFloat        = "HINCRBYFLOAT"
	CmdHRandField          = "HRANDFIELD"
	CmdGetRange            = "GETRANGE"
	CmdAppend              = "APPEND"
	CmdZPopMax             = "ZPOPMAX"
	CmdHLen                = "HLEN"
	CmdHStrLen             = "HSTRLEN"
	CmdHScan               = "HSCAN"
	CmdBFAdd               = "BF.ADD"
	CmdBFReserve           = "BF.RESERVE"
	CmdBFInfo              = "BF.INFO"
	CmdBFExists            = "BF.EXISTS"
	CmdCMSQuery            = "CMS.QUERY"
	CmdCMSInfo             = "CMS.INFO"
	CmdCMSInitByDim        = "CMS.INITBYDIM"
	CmdCMSInitByProb       = "CMS.INITBYPROB"
	CmdCMSMerge            = "CMS.MERGE"
	CmdCMSIncrBy           = "CMS.INCRBY"
	CmdHSet                = "HSET"
	CmdHGet                = "HGET"
	CmdHSetnx              = "HSETNX"
	CmdHDel                = "HDEL"
	CmdHMSet               = "HMSET"
	CmdHMGet               = "HMGET"
	CmdSetBit              = "SETBIT"
	CmdGetBit              = "GETBIT"
	CmdBitCount            = "BITCOUNT"
	CmdBitField            = "BITFIELD"
	CmdBitPos              = "BITPOS"
	CmdBitFieldRO          = "BITFIELD_RO"
	CmdSadd                = "SADD"
	CmdSrem                = "SREM"
	CmdScard               = "SCARD"
	CmdSmembers            = "SMEMBERS"
	CmdDump                = "DUMP"
	CmdRestore             = "RESTORE"
	CmdGeoAdd              = "GEOADD"
	CmdGeoDist             = "GEODIST"
	CmdGeoPos              = "GEOPOS"
	CmdGeoHash             = "GEOHASH"
	CmdClient              = "CLIENT"
	CmdLatency             = "LATENCY"
	CmdDel                 = "DEL"
	CmdExists              = "EXISTS"
	CmdPersist             = "PERSIST"
	CmdTypeOf              = "TYPE"
	CmdObject              = "OBJECT"
	CmdExpire              = "EXPIRE"
	CmdExpireAt            = "EXPIREAT"
	CmdExpireTime          = "EXPIRETIME"
	CmdSet                 = "SET"
	CmdGet                 = "GET"
	CmdGetSet              = "GETSET"
	CmdGetEx               = "GETEX"
	CmdGetDel              = "GETDEL"
	CmdLrange              = "LRANGE"
	CmdLinsert             = "LINSERT"
	CmdJSONArrInsert       = "JSON.ARRINSERT"
	CmdJSONArrTrim         = "JSON.ARRTRIM"
	CmdJSONArrAppend       = "JSON.ARRAPPEND"
	CmdJSONArrLen          = "JSON.ARRLEN"
	CmdJSONArrPop          = "JSON.ARRPOP"
	CmdJSONClear           = "JSON.CLEAR"
	CmdJSONSet             = "JSON.SET"
	CmdJSONObjKeys         = "JSON.OBJKEYS"
	CmdJSONDel             = "JSON.DEL"
	CmdJSONForget          = "JSON.FORGET"
	CmdJSONGet             = "JSON.GET"
	CmdJSONStrlen          = "JSON.STRLEN"
	CmdJSONObjlen          = "JSON.OBJLEN"
	CmdJSONNumIncrBY       = "JSON.NUMINCRBY"
	CmdJSONNumMultBy       = "JSON.NUMMULTBY"
	CmdJSONType            = "JSON.TYPE"
	CmdJSONToggle          = "JSON.TOGGLE"
	CmdJSONNumMultBY       = "JSON.NUMMULTBY"
	CmdJSONDebug           = "JSON.DEBUG"
	CmdJSONResp            = "JSON.RESP"
	CmdLPush               = "LPUSH"
	CmdRPush               = "RPUSH"
	CmdLPop                = "LPOP"
	CmdRPop                = "RPOP"
	CmdLLEN                = "LLEN"
	CmdCommand             = "COMMAND"
	CmdCommandCount        = "COMMAND|COUNT"
	CmdCommandHelp         = "COMMAND|HELP"
	CmdCommandInfo         = "COMMAND|INFO"
	CmdCommandList         = "COMMAND|LIST"
	CmdCommandDocs         = "COMMAND|DOCS"
	CmdCommandGetKeys      = "COMMAND|GETKEYS"
	CmdCommandGetKeysFlags = "COMMAND|GETKEYSANDFLAGS"
	CmdJSONArrIndex        = "JSON.ARRINDEX"
)

// Multi-shard commands.
const (
	CmdMset     = "MSET"
	CmdMget     = "MGET"
	CmdSInter   = "SINTER"
	CmdSDiff    = "SDIFF"
	CmdJSONMget = "JSON.MGET"
	CmdKeys     = "KEYS"
	CmdTouch    = "TOUCH"
	CmdDBSize   = "DBSIZE"
	CmdFlushDB  = "FLUSHDB"
)

// Multi-Step-Multi-Shard commands
const (
	CmdRename = "RENAME"
	CmdCopy   = "COPY"
)

// Watch commands
const (
	CmdGetWatch       = "GET.WATCH"
	CmdGetUnWatch     = "GET.UNWATCH"
	CmdZRangeWatch    = "ZRANGE.WATCH"
	CmdZRangeUnWatch  = "ZRANGE.UNWATCH"
	CmdPFCountWatch   = "PFCOUNT.WATCH"
	CmdPFCountUnWatch = "PFCOUNT.UNWATCH"
)

type CmdMeta struct {
	CmdType
	Cmd                string
	CmdHandlerFunction func([]string) []byte

	// decomposeCommand is a function that takes a DiceDB command and breaks it down into smaller,
	// manageable DiceDB commands for each shard processing. It returns a slice of DiceDB commands.
	decomposeCommand func(h *BaseCommandHandler, ctx context.Context, DiceDBCmd *cmd.DiceDBCmd) ([]*cmd.DiceDBCmd, error)

	// composeResponse is a function that combines multiple responses from the execution of commands
	// into a single response object. It accepts a variadic parameter of EvalResponse objects
	// and returns a unified response interface. It is used in the command type "MultiShard"
	composeResponse func(responses ...ops.StoreResponse) interface{}

	// preProcessingReq indicates whether the command requires preprocessing before execution.
	// If set to true, it signals that a preliminary step (such as fetching values from shards)
	// is necessary before the main command is executed. This is important for commands that depend
	// on the current state of data in the database.
	preProcessing bool

	// preProcessResponse is a function that handles the preprocessing of a DiceDB command by
	// preparing the necessary operations (e.g., fetching values from shards) before the command
	// is executed. It takes the CommandHandler and the original DiceDB command as parameters and
	// ensures that any required information is retrieved and processed in advance. Use this when set
	// preProcessingReq = true.
	preProcessResponse func(h *BaseCommandHandler, DiceDBCmd *cmd.DiceDBCmd) error

	// ReadOnly indicates whether the command modifies the database state.
	// If true, the command only reads data and doesn't need to be logged in WAL.
	ReadOnly bool
}

var CommandsMeta = map[string]CmdMeta{
	// Single-shard commands.
	CmdSet: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdExpire: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdExpireAt: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdExpireTime: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdGet: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdGetSet: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdGetEx: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdGetDel: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdSadd: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdSrem: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdScard: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdSmembers: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHExists: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHKeys: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHVals: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONArrAppend: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONArrInsert: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONArrTrim: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONArrLen: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONArrPop: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONClear: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONSet: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONObjKeys: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONDel: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdDel: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdExists: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdPersist: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdTypeOf: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONForget: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONGet: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONStrlen: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONObjlen: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONNumIncrBY: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONNumMultBy: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONType: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONToggle: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdJSONDebug: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONResp: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdGetRange: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdPFAdd: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdPFCount: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdTTL: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdPTTL: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHLen: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHStrLen: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHScan: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHIncrBy: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHIncrByFloat: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHRandField: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdSetBit: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdGetBit: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdBitCount: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdBitField: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdBitPos: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdBitFieldRO: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdLrange: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdLinsert: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdLPush: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdRPush: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdLPop: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdRPop: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdLLEN: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCMSQuery: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCMSInfo: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCMSIncrBy: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdCMSInitByDim: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdCMSInitByProb: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdCMSMerge: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHSet: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHGet: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdHSetnx: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHDel: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHMSet: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdHMGet: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	// Sorted set commands
	CmdZAdd: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdZCount: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdZRank: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdZRange: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdZCard: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdZRem: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdAppend: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdIncr: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdIncrBy: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdDecr: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdDecrBy: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdIncrByFloat: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdZPopMin: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdZPopMax: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	// Bloom Filter
	CmdBFAdd: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdBFInfo: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdBFExists: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdBFReserve: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdDump: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdRestore: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	// geoCommands
	CmdGeoAdd: {
		CmdType:  SingleShard,
		ReadOnly: false,
	},
	CmdGeoDist: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdGeoPos: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdGeoHash: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdClient: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdLatency: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdObject: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommand: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandCount: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandHelp: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandInfo: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandList: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandDocs: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandGetKeys: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdCommandGetKeysFlags: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},
	CmdJSONArrIndex: {
		CmdType:  SingleShard,
		ReadOnly: true,
	},

	// Multi-shard commands.
	CmdRename: {
		CmdType:            MultiShard,
		preProcessing:      true,
		preProcessResponse: preProcessRename,
		decomposeCommand:   (*BaseCommandHandler).decomposeRename,
		composeResponse:    composeRename,
		ReadOnly:           false,
	},

	CmdCopy: {
		CmdType:            MultiShard,
		preProcessing:      true,
		preProcessResponse: customProcessCopy,
		decomposeCommand:   (*BaseCommandHandler).decomposeCopy,
		composeResponse:    composeCopy,
		ReadOnly:           false,
	},

	CmdPFMerge: {
		CmdType:            MultiShard,
		preProcessing:      true,
		preProcessResponse: preProcessPFMerge,
		decomposeCommand:   (*BaseCommandHandler).decomposePFMerge,
		composeResponse:    composePFMerge,
		ReadOnly:           false,
	},

	CmdMset: {
		CmdType:          MultiShard,
		decomposeCommand: (*BaseCommandHandler).decomposeMSet,
		composeResponse:  composeMSet,
		ReadOnly:         false,
	},

	CmdMget: {
		CmdType:          MultiShard,
		decomposeCommand: (*BaseCommandHandler).decomposeMGet,
		composeResponse:  composeMGet,
		ReadOnly:         true,
	},

	CmdSInter: {
		CmdType:          MultiShard,
		decomposeCommand: (*BaseCommandHandler).decomposeSInter,
		composeResponse:  composeSInter,
		ReadOnly:         true,
	},

	CmdSDiff: {
		CmdType:          MultiShard,
		decomposeCommand: (*BaseCommandHandler).decomposeSDiff,
		composeResponse:  composeSDiff,
		ReadOnly:         true,
	},

	CmdJSONMget: {
		CmdType:          MultiShard,
		decomposeCommand: (*BaseCommandHandler).decomposeJSONMget,
		composeResponse:  composeJSONMget,
		ReadOnly:         true,
	},
	CmdTouch: {
		CmdType:          MultiShard,
		decomposeCommand: (*BaseCommandHandler).decomposeTouch,
		composeResponse:  composeTouch,
		ReadOnly:         false,
	},
	CmdDBSize: {
		CmdType:          AllShard,
		decomposeCommand: (*BaseCommandHandler).decomposeDBSize,
		composeResponse:  composeDBSize,
		ReadOnly:         true,
	},
	CmdKeys: {
		CmdType:          AllShard,
		decomposeCommand: (*BaseCommandHandler).decomposeKeys,
		composeResponse:  composeKeys,
		ReadOnly:         true,
	},
	CmdFlushDB: {
		CmdType:          AllShard,
		decomposeCommand: (*BaseCommandHandler).decomposeFlushDB,
		composeResponse:  composeFlushDB,
		ReadOnly:         false,
	},

	// Custom commands.
	CmdAbort: {
		CmdType:  Custom,
		ReadOnly: true,
	},
	CmdAuth: {
		CmdType:  Custom,
		ReadOnly: true,
	},
	CmdEcho: {
		CmdType:  Custom,
		ReadOnly: true,
	},
	CmdPing: {
		CmdType:  Custom,
		ReadOnly: true,
	},

	// Watch commands
	CmdGetWatch: {
		CmdType:  Watch,
		ReadOnly: true,
	},
	CmdZRangeWatch: {
		CmdType:  Watch,
		ReadOnly: true,
	},
	CmdPFCountWatch: {
		CmdType:  Watch,
		ReadOnly: true,
	},

	// Unwatch commands
	CmdGetUnWatch: {
		CmdType:  Unwatch,
		ReadOnly: true,
	},
	CmdZRangeUnWatch: {
		CmdType:  Unwatch,
		ReadOnly: true,
	},
	CmdPFCountUnWatch: {
		CmdType:  Unwatch,
		ReadOnly: true,
	},
}

func init() {
	for c, meta := range CommandsMeta {
		if err := validateCmdMeta(c, meta); err != nil {
			slog.Error("error validating command metadata %s: %v", c, err)
		}
	}
}

// validateCmdMeta ensures that the metadata for each command is properly configured
func validateCmdMeta(c string, meta CmdMeta) error {
	switch meta.CmdType {
	case Global:
		if meta.CmdHandlerFunction == nil {
			slog.Debug("global command %s must have CmdHandlerFunction function", slog.String("command", c))
			return diceerrors.ErrInternalServer
		}
	case MultiShard, AllShard:
		if meta.decomposeCommand == nil || meta.composeResponse == nil {
			return fmt.Errorf("multi-shard command %s must have both decomposeCommand and composeResponse implemented", c)
		}
	case SingleShard, Watch, Unwatch, Custom:
		// No specific validations for these types currently
	default:
		return fmt.Errorf("unknown command type for %s", c)
	}

	return nil
}
