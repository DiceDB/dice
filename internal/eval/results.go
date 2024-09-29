package eval

type (
	RespNIL        string // (nil)
	RespOK         string // OK
	RespQueued     string //[]byte("+QUEUED\r\n") // Signifies that a command has been queued for execution. //nolint:unused
	RespZero       string //[]byte(":0\r\n")      // Represents the integer zero in RESP format. //nolint:unused
	RespOne        string //[]byte(":1\r\n")      // Represents the integer one in RESP format. //nolint:unused
	RespMinusOne   string //[]byte(":-1\r\n")     // Represents the integer negative one in RESP format. //nolint:unused
	RespMinusTwo   string //[]byte(":-2\r\n")     // Represents the integer negative two in RESP format. //nolint:unused
	RespEmptyArray string //[]byte("*0\r\n")      // Represents an empty array in RESP format. //nolint:unused
)
