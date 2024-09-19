package eval

// Scatter file is used by shardThread to evaluate
// individual commands on each shard. Scatter functions
// returns response as interface and error using
// a structure for each command, which eventiually passed
// till worker from each shard response

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/clientio"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	dstore "github.com/dicedb/dice/internal/store"
)

type EvalResponse struct {
	Result interface{} // Result of the Store operation, for now the type is set to []byte, but this can change in the future.
	Error  error
}

// ScatterPING function is used by shardThread to
// evaluate PING command
func EvalPING(args []string, store *dstore.Store) EvalResponse {
	if len(args) >= 2 {
		return EvalResponse{Result: nil, Error: fmt.Errorf("PING")}
	}

	if len(args) == 0 {
		return EvalResponse{Result: "PONG", Error: nil}
	} else {
		return EvalResponse{Result: args[0], Error: nil}
	}
}

// evalSET puts a new <key, value> pair in db as in the args
// args must contain key and value.
// args can also contain multiple options -
//
//	EX or ex which will set the expiry time(in secs) for the key
//	PX or px which will set the expiry time(in milliseconds) for the key
//	EXAT or exat which will set the specified Unix time at which the key will expire, in seconds (a positive integer).
//	PXAT or PX which will the specified Unix time at which the key will expire, in milliseconds (a positive integer).
//	XX orr xx which will only set the key if it already exists.
//
// Returns encoded error response if at least a <key, value> pair is not part of args
// Returns encoded error response if expiry time value in not integer
// Returns encoded error response if both PX and EX flags are present
// Returns encoded OK RESP once new entry is added
// If the key already exists then the value will be overwritten and expiry will be discarded
func EvalSET(args []string, store *dstore.Store) EvalResponse {
	if len(args) <= 1 {
		return EvalResponse{Result: nil, Error: fmt.Errorf("SET")}
	}

	var key, value string
	var exDurationMs int64 = -1
	var state exDurationState = Uninitialized
	var keepttl bool = false

	key, value = args[0], args[1]
	oType, oEnc := deduceTypeEncoding(value)

	for i := 2; i < len(args); i++ {
		arg := strings.ToUpper(args[i])
		switch arg {
		case Ex, Px:
			if state != Uninitialized {
				return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.SyntaxErr)}
			}
			i++
			if i == len(args) {
				return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.SyntaxErr)}
			}

			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.IntOrOutOfRangeErr)}
			}

			if exDuration <= 0 || exDuration >= maxExDuration {
				return EvalResponse{Result: nil, Error: fmt.Errorf("EXPIRE SET")}
			}

			// converting seconds to milliseconds
			if arg == Ex {
				exDuration *= 1000
			}
			exDurationMs = exDuration
			state = Initialized

		case Pxat, Exat:
			if state != Uninitialized {
				return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.SyntaxErr)}
			}
			i++
			if i == len(args) {
				return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.SyntaxErr)}
			}
			exDuration, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.IntOrOutOfRangeErr)}
			}

			if exDuration < 0 {
				return EvalResponse{Result: nil, Error: fmt.Errorf("SET")}
			}

			if arg == Exat {
				exDuration *= 1000
			}
			exDurationMs = exDuration - utils.GetCurrentTime().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if exDurationMs < 0 {
				exDurationMs = 0
			}
			state = Initialized

		case XX:
			// Get the key from the hash table
			obj := store.Get(key)

			// if key does not exist, return RESP encoded nil
			if obj == nil {
				return EvalResponse{Result: clientio.RespNIL, Error: nil}
			}
		case NX:
			obj := store.Get(key)
			if obj != nil {
				return EvalResponse{Result: clientio.RespNIL, Error: nil}
			}
		case KEEPTTL, Keepttl:
			keepttl = true
		default:
			return EvalResponse{Result: nil, Error: fmt.Errorf(diceerrors.SyntaxErr)}
		}
	}

	// Cast the value properly based on the encoding type
	var storedValue interface{}
	switch oEnc {
	case object.ObjEncodingInt:
		storedValue, _ = strconv.ParseInt(value, 10, 64)
	case object.ObjEncodingEmbStr, object.ObjEncodingRaw:
		storedValue = value
	default:
		return EvalResponse{Result: nil, Error: fmt.Errorf("ERR unsupported encoding: %d", oEnc)}
	}

	// putting the k and value in a Hash Table
	store.Put(key, store.NewObj(storedValue, exDurationMs, oType, oEnc), dstore.WithKeepTTL(keepttl))
	return EvalResponse{Result: clientio.RespOK, Error: nil}
}
