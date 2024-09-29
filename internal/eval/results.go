package eval

type RespType int

const (
	RespNIL        RespType = iota
	RespOK                  // OK
	RespQueued              //[]byte("+QUEUED\r\n") // Signifies that a command has been queued for execution. //nolint:unused
	RespZero                //[]byte(":0\r\n")      // Represents the integer zero in RESP format. //nolint:unused
	RespOne                 //[]byte(":1\r\n")      // Represents the integer one in RESP format. //nolint:unused
	RespMinusOne            //[]byte(":-1\r\n")     // Represents the integer negative one in RESP format. //nolint:unused
	RespMinusTwo            //[]byte(":-2\r\n")     // Represents the integer negative two in RESP format. //nolint:unused
	RespEmptyArray          //[]byte("*0\r\n")      // Represents an empty array in RESP format. //nolint:unused
)
