package cmd

import (
	"strconv"
	"strings"

	diceerrors "github.com/dicedb/dice/internal/errors"
	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/server/utils"
	"github.com/dicedb/dice/internal/store"
)

// Represents the state of expiration flag processing.
type exDurationState int

const (
	// States to track whether expiration is being set
	Uninitialized exDurationState = iota // No expiration flag set
	Initialized                          // Expiration flag processed
	
	Ex      = "EX"    // Expiry in seconds
	Px      = "PX"    // Expiry in milliseconds
	Exat    = "EXAT"  // Expiry at a specific timestamp (seconds)
	Pxat    = "PXAT"  // Expiry at a specific timestamp (milliseconds)
	Persist = "PERSIST" // Remove expiration

	// Maximum allowed expiration duration to prevent overflow
	maxExDuration = 9223372036854775
)

var cGETEX = &DiceDBCommand{
	Name:      "GETEX",
	HelpShort: "GETEX retrieves the value of a key and optionally updates its expiration (EX, PX, EXAT, PXAT, PERSIST).",
	Eval:      evalGETEX,
}

func init() {
	commandRegistry.AddCommand(cGETEX)
}

func evalGETEX(c *Cmd, s *store.Store) (*CmdRes, error) {
	if len(c.C.Args) < 1 {
		return cmdResNil, diceerrors.ErrWrongArgumentCount("GETEX")
	}

	key := c.C.Args[0]

	var expiryDurationMs int64 = -1 // Stores new expiration in milliseconds (-1 means unchanged)
	state := Uninitialized          // Tracks whether an expiration flag is set
	persist := false                // Flag to indicate expiration should be removed

	for i := 1; i < len(c.C.Args); i++ {
		arg := strings.ToUpper(c.C.Args[i])
		switch arg {
		case Ex, Px:
			// Ensure only one expiration flag is provided
			if state != Uninitialized {
				return nil, diceerrors.ErrSyntax
			}
			i++
			if i == len(c.C.Args) {
				return nil, diceerrors.ErrSyntax
			}

			expiry, err := strconv.ParseInt(c.C.Args[i], 10, 64)
			if err != nil || expiry <= 0 || expiry > maxExDuration {
				return nil, diceerrors.ErrInvalidExpireTime("GETEX")
			}

			// Convert seconds to milliseconds if EX is used
			if arg == Ex {
				expiry *= 1000
			}
			expiryDurationMs = expiry
			state = Initialized

		case Exat, Pxat:
			// Ensure only one expiration flag is provided
			if state != Uninitialized {
				return nil, diceerrors.ErrSyntax
			}
			i++
			if i == len(c.C.Args) {
				return nil, diceerrors.ErrSyntax
			}

			expiryTimestamp, err := strconv.ParseInt(c.C.Args[i], 10, 64)
			if err != nil || expiryTimestamp < 0 || expiryTimestamp > maxExDuration {
				return nil, diceerrors.ErrInvalidExpireTime("GETEX")
			}

			// Convert seconds to milliseconds if EXAT is used
			if arg == Exat {
				expiryTimestamp *= 1000
			}

			expiryDurationMs = expiryTimestamp - utils.GetCurrentTime().UnixMilli()
			// If the expiry time is in the past, set exDurationMs to 0
			// This will be used to signal immediate expiration
			if expiryDurationMs < 0 {
				expiryDurationMs = 0
			}
			state = Initialized

		case Persist:
			// Ensure only one expiration flag is provided
			if state != Uninitialized {
				return nil, diceerrors.ErrSyntax
			}
			persist = true
			state = Initialized

		default:
			return nil, diceerrors.ErrIntegerOutOfRange
		}
	}

	// Retrieve the key's object from the store
	obj := s.Get(key)
	if obj == nil {
		return cmdResNil, nil
	}

	// Ensure the object is not of a type that doesn't support expiration
	if object.AssertType(obj.Type, object.ObjTypeSet) == nil ||
		object.AssertType(obj.Type, object.ObjTypeJSON) == nil {
		return nil, diceerrors.ErrWrongTypeOperation
	}

	getResp, err := evalGET(c, s)
	if err != nil {
		return nil, err
	}

	// Apply expiration updates only if a flag was provided
	if state == Initialized {
		if persist {
			// Remove expiration if PERSIST flag is used
			store.DelExpiry(obj, s)
		} else {
			// Set new expiration
			s.SetExpiry(obj, expiryDurationMs)
		}
	}

	return getResp, nil
}
