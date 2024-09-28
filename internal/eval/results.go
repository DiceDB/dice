package eval

var (
	respNIL []byte = []byte("$-1\r\n") // Represents a nil response in RESP format.
	respOK  []byte = []byte("+OK\r\n") // Indicates a successful command execution.

	// Following varaibles are commented out since not in used (lint-check). Once we start migrating,
	// we will uncomment according to the requirements

	// respQueued     []byte = []byte("+QUEUED\r\n") // Signifies that a command has been queued for execution. //nolint:unused
	// respZero       []byte = []byte(":0\r\n")      // Represents the integer zero in RESP format. //nolint:unused
	// respOne        []byte = []byte(":1\r\n")      // Represents the integer one in RESP format. //nolint:unused
	// respMinusOne   []byte = []byte(":-1\r\n")     // Represents the integer negative one in RESP format. //nolint:unused
	// respMinusTwo   []byte = []byte(":-2\r\n")     // Represents the integer negative two in RESP format. //nolint:unused
	// respEmptyArray []byte = []byte("*0\r\n")      // Represents an empty array in RESP format. //nolint:unused
)
