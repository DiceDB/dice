package eval

import (
	"github.com/dicedb/dice/internal/cmd"
	ds "github.com/dicedb/dice/internal/datastructures"
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
	NewEval func([]string, *dstore.Store) *ds.EvalResponse

	// StoreObjectEval is a specialized evaluation function for commands that operate on an object.
	// It is designed for scenarios where the command and subsequent dependent command requires
	// an object as part of its execution. This function processes the command,
	// evaluates it based on the provided object, and returns an EvalResponse struct
	// Commands that involve object manipulation, is not recommended for general use.
	// Only commands that really requires full object definition to pass across multiple shards
	// should implement this function. e.g. COPY, RENAME etc
	StoreObjectEval func(*cmd.DiceDBCmd, *dstore.Store) *ds.EvalResponse
}

type KeySpecs struct {
	BeginIndex int
	Step       int
	LastKey    int
}

var (
	PreProcessing = map[string]func([]string, *dstore.Store) *ds.EvalResponse{}
	DiceCmds      = map[string]DiceCmdMeta{}
)

var (
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
	}
)
