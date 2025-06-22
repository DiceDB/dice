package cmd

type KeySpecs struct {
	BeginIndex int
	Step       int
	LastKey    int
	Flags      uint64
}

var (
	// Flags
	RO            uint64 = (1 << 0)
	RW            uint64 = (1 << 1)
	OW            uint64 = (1 << 2)
	RM            uint64 = (1 << 3)
	ACCESS        uint64 = (1 << 4)
	UPDATE        uint64 = (1 << 5)
	INSERT        uint64 = (1 << 6)
	DELETE        uint64 = (1 << 7)
	NOTKEY        uint64 = (1 << 8)
	INCOMPLETE    uint64 = (1 << 9)
	VARIABLEFLAGS uint64 = (1 << 10)
)

func GetFlagsNameMap() map[uint64]string {
	return map[uint64]string{
		RO:            "RO",
		RW:            "RW",
		OW:            "OW",
		RM:            "RM",
		ACCESS:        "access",
		UPDATE:        "update",
		INSERT:        "insert",
		DELETE:        "delete",
		NOTKEY:        "not_key",
		INCOMPLETE:    "incomplete",
		VARIABLEFLAGS: "variable_flags",
	}
}

func GetFlags() []uint64 {
	return []uint64{
		RO,
		RW,
		OW,
		RM,
		ACCESS,
		UPDATE,
		INSERT,
		DELETE,
		NOTKEY,
		INCOMPLETE,
		VARIABLEFLAGS,
	}
}

func SetGetKeys(args []string, ks *KeySpecs) {
	for i := 2; i < len(args); i++ {
		if (len(args[i]) == 3) &&
			(args[i][0] == 'g' || args[i][0] == 'G') &&
			(args[i][1] == 'e' || args[i][1] == 'E') &&
			(args[i][2] == 't' || args[i][2] == 'T') {
			ks.Flags = RW | ACCESS | UPDATE
			return
		}
	}
	ks.Flags = OW | UPDATE
}
