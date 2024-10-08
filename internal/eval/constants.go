package eval

const (
	BYTE = "BYTE"
	BIT  = "BIT"

	AND string = "AND"
	OR  string = "OR"
	XOR string = "XOR"
	NOT string = "NOT"

	Ex              string = "EX"
	Px              string = "PX"
	Pxat            string = "PXAT"
	Exat            string = "EXAT"
	XX              string = "XX"
	NX              string = "NX"
	GT              string = "GT"
	LT              string = "LT"
	KeepTTL         string = "KEEPTTL"
	Sync            string = "SYNC"
	Async           string = "ASYNC"
	Help            string = "HELP"
	Memory          string = "MEMORY"
	Count           string = "COUNT"
	GetKeys         string = "GETKEYS"
	GetKeysAndFlags string = "GETKEYSANDFLAGS"
	List            string = "LIST"
	Info            string = "INFO"
	null            string = "null"
	WithValues      string = "WITHVALUES"
	WithScores      string = "WITHSCORES"
	REV             string = "REV"
	GET             string = "GET"
	SET             string = "SET"
	INCRBY          string = "INCRBY"
	OVERFLOW        string = "OVERFLOW"
	WRAP            string = "WRAP"
	SAT             string = "SAT"
	FAIL            string = "FAIL"
	SIGNED          string = "SIGNED"
	UNSIGNED        string = "UNSIGNED"

	//Flags
	RO             uint64 = (1 << 0)
	RW             uint64 = (1 << 1)
	OW             uint64 = (1 << 2)
	RM             uint64 = (1 << 3)
	ACCESS         uint64 = (1 << 4)
	UPDATE         uint64 = (1 << 5)
	INSERT         uint64 = (1 << 6)
	DELETE         uint64 = (1 << 7)
	NOT_KEY        uint64 = (1 << 8)
	INCOMPLETE     uint64 = (1 << 9)
	VARIABLE_FLAGS uint64 = (1 << 10)
)

func getFlagsNameMap() map[uint64]string {
	return map[uint64]string{
		RO:             "RO",
		RW:             "RW",
		OW:             "OW",
		RM:             "RM",
		ACCESS:         "access",
		UPDATE:         "update",
		INSERT:         "insert",
		DELETE:         "delete",
		NOT_KEY:        "not_key",
		INCOMPLETE:     "incomplete",
		VARIABLE_FLAGS: "variable_flags",
	}
}

func getFlags() []uint64 {
	return []uint64{
		RO,
		RW,
		OW,
		RM,
		ACCESS,
		UPDATE,
		INSERT,
		DELETE,
		NOT_KEY,
		INCOMPLETE,
		VARIABLE_FLAGS,
	}
}
