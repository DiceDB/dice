package DiceErrors

import (
	"fmt"
	"strings"
)

const (
	ArityErr           = "wrong number of arguments for '%s' command"
	SyntaxErr          = "syntax error"
	ExpiryErr          = "invalid expire time in '%s' command"
	AuthErr            = "AUTH failed"
	IntOrOutOfRangeErr = "value is not an integer or out of range"
	ValOutOfRangeErr   = "value is out of range"
	ElementPeekErr     = "number of elements to peek should be a positive number less than %d"
	NoKeyErr           = "no such key"
	ErrDefault         = "ERR %s"
	WrongTypeErr       = "WRONGTYPE Operation against a key holding the wrong kind of value"
	SameObjectErr      = "ERR source and destination objects are the same"
	OutOfRangeErr      = "ERR index out of range"
	NoScriptErr        = "NOSCRIPT No matching script. Please use EVAL."
	LoadingErr         = "LOADING Redis is loading the dataset in memory"
	SlowEvalErr        = "BUSY Redis is busy running a script. You can only call SCRIPT KILL or SHUTDOWN NOSAVE."
	SlowScriptErr      = "BUSY Redis is busy running a script. You can only call FUNCTION KILL or SHUTDOWN NOSAVE."
	SlowModuleErr      = "BUSY Redis is busy running a module command."
	MasterDownErr      = "MASTERDOWN Link with MASTER is down and replica-serve-stale-data is set to 'no'."
	BgSaveErr          = "MISCONF Redis is configured to save RDB snapshots, but it's currently unable to persist to disk. Commands that may modify the data set are disabled, because this instance is configured to report errors during writes if RDB snapshotting fails (stop-writes-on-bgsave-error option). Please check the Redis logs for details about the RDB error."
	RoSlaveErr         = "READONLY You can't write against a read only replica."
	NoAuthErr          = "NOAUTH Authentication required."
	OOMErr             = "OOM command not allowed when used memory > 'maxmemory'."
	ExecAbortErr       = "EXECABORT Transaction discarded because of previous errors."
	NoReplicasErr      = "NOREPLICAS Not enough good replicas to write."
	BusyKeyErr         = "BUSYKEY Target key name already exists."
)

type DiceError struct {
	message string
}

func newDiceErr(message string) *DiceError {
	return &DiceError{
		message: message,
	}
}

func (d *DiceError) toEncodedMessage() []byte {
	return []byte(fmt.Sprintf("-%s\r\n", d.message))
}

// NewErrWithMessage If the error code is already passed in the string,
// the error code provided is used, otherwise the string "-ERR " for the generic
// error code is automatically added. Note that 's' must NOT end with \r\n.
func NewErrWithMessage(errMsg string) []byte {
	// If the string already starts with "-..." then the error code
	// is provided by the caller. Otherwise, we use "-ERR".
	if len(errMsg) == 0 || errMsg[0] != '-' {
		errMsg = fmt.Sprintf(ErrDefault, errMsg)
	}

	return newDiceErr(errMsg).toEncodedMessage()
}

func NewErrWithFormattedMessage(errMsgFmt string, args ...interface{}) []byte {
	if len(args) > 0 {
		errMsgFmt = fmt.Sprintf(errMsgFmt, args)
	}

	return NewErrWithMessage(errMsgFmt)
}

func NewErrArity(cmdName string) []byte {
	return NewErrWithFormattedMessage(ArityErr, strings.ToLower(cmdName))
}

func NewErrExpireTime(cmdName string) []byte {
	return NewErrWithFormattedMessage(ExpiryErr, strings.ToLower(cmdName))
}
