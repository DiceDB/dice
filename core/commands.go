package core

type DiceCmdMeta struct {
	Name string
	Info string
	Eval func([]string) []byte
}

var (
	diceCmds = map[string]DiceCmdMeta{}

	pingCmdMeta = DiceCmdMeta{
		Name: "PING",
		Info: "ping command info here",
		Eval: evalPING,
	}
	setCmdMeta = DiceCmdMeta{
		Name: "SET",
		Info: "set command info here",
		Eval: evalSET,
	}
	getCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: "get command info here",
		Eval: evalGET,
	}
	ttlCmdMeta = DiceCmdMeta{
		Name: "TTL",
		Info: "ttl command info here",
		Eval: evalTTL,
	}
	delCmdMeta = DiceCmdMeta{
		Name: "DEL",
		Info: "del command info here",
		Eval: evalDEL,
	}
	expireCmdMeta = DiceCmdMeta{
		Name: "EXPIRE",
		Info: "expire command info here",
		Eval: evalEXPIRE,
	}
	helloCmdMeta = DiceCmdMeta{
		Name: "HELLO",
		Info: "hello command info here",
		Eval: evalHELLO,
	}
	bgrewriteaofCmdMeta = DiceCmdMeta{
		Name: "BGREWRITEAOF",
		Info: "bgrewriteaof command info here",
		Eval: evalBGREWRITEAOF,
	}
	incrCmdMeta = DiceCmdMeta{
		Name: "INCR",
		Info: "incr command info here",
		Eval: evalINCR,
	}
	infoCmdMeta = DiceCmdMeta{
		Name: "INFO",
		Info: "info command info here",
		Eval: evalINFO,
	}
	clientCmdMeta = DiceCmdMeta{
		Name: "CLIENT",
		Info: "client command info here",
		Eval: evalCLIENT,
	}
	latencyCmdMeta = DiceCmdMeta{
		Name: "LATENCY",
		Info: "latency command info here",
		Eval: evalLATENCY,
	}
	lruCmdMeta = DiceCmdMeta{
		Name: "LRU",
		Info: "lru command info here",
		Eval: evalLRU,
	}
	sleepCmdMeta = DiceCmdMeta{
		Name: "SLEEP",
		Info: "sleep command info here",
		Eval: evalSLEEP,
	}
	qintinsCmdMeta = DiceCmdMeta{
		Name: "QINTINS",
		Info: "qintins command info here",
		Eval: evalQINTINS,
	}
	qintremCmdMeta = DiceCmdMeta{
		Name: "QINTREM",
		Info: "qintrem command info here",
		Eval: evalQINTREM,
	}
	qintlenCmdMeta = DiceCmdMeta{
		Name: "QINTLEN",
		Info: "qintlen command info here",
		Eval: evalQINTLEN,
	}
	qintpeekCmdMeta = DiceCmdMeta{
		Name: "QINTPEEK",
		Info: "qintpeek command info here",
		Eval: evalQINTPEEK,
	}
	bfinitCmdMeta = DiceCmdMeta{
		Name: "BFINIT",
		Info: "bfinit command info here",
		Eval: evalBFINIT,
	}
	bfaddCmdMeta = DiceCmdMeta{
		Name: "BFADD",
		Info: "bfadd command info here",
		Eval: evalBFADD,
	}
	bfexistsCmdMeta = DiceCmdMeta{
		Name: "BFEXISTS",
		Info: "bfexists command info here",
		Eval: evalBFEXISTS,
	}
	bfinfoCmdMeta = DiceCmdMeta{
		Name: "BFINFO",
		Info: "bfinfo command info here",
		Eval: evalBFINFO,
	}
	qrefinsCmdMeta = DiceCmdMeta{
		Name: "QREFINS",
		Info: "qrefins command info here",
		Eval: evalQREFINS,
	}
	qrefremCmdMeta = DiceCmdMeta{
		Name: "QREFREM",
		Info: "qrefrem command info here",
		Eval: evalQREFREM,
	}
	qreflenCmdMeta = DiceCmdMeta{
		Name: "QREFLEN",
		Info: "qreflen command info here",
		Eval: evalQREFLEN,
	}
	qrefpeekCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: "qrefpeek command info here",
		Eval: evalGET,
	}
	stackintpushCmdMeta = DiceCmdMeta{
		Name: "STACKINTPUSH",
		Info: "stackintpush command info here",
		Eval: evalSTACKINTPUSH,
	}
	stackintpopCmdMeta = DiceCmdMeta{
		Name: "STACKINTPOP",
		Info: "stackintpop command info here",
		Eval: evalSTACKINTPOP,
	}
	stackintlenCmdMeta = DiceCmdMeta{
		Name: "STACKINTLEN",
		Info: "stackintlen command info here",
		Eval: evalSTACKINTLEN,
	}
	stackintpeekCmdMeta = DiceCmdMeta{
		Name: "STACKINTPEEK",
		Info: "stackintpeek command info here",
		Eval: evalSTACKINTPEEK,
	}
	stackrefpushCmdMeta = DiceCmdMeta{
		Name: "STACKREFPUSH",
		Info: "stackrefpush command info here",
		Eval: evalSTACKREFPUSH,
	}
	stackrefpopCmdMeta = DiceCmdMeta{
		Name: "STACKREFPOP",
		Info: "stackrefpop command info here",
		Eval: evalSTACKREFPOP,
	}
	stackreflenCmdMeta = DiceCmdMeta{
		Name: "STACKREFLEN",
		Info: "stackreflen command info here",
		Eval: evalSTACKREFLEN,
	}
	stackrefpeekCmdMeta = DiceCmdMeta{
		Name: "STACKREFPEEK",
		Info: "stackrefpeek command info here",
		Eval: evalSTACKREFPEEK,
	}
	// TODO: Remove this override once we support QWATCH in dice-cli.
	subscribeCmdMeta = DiceCmdMeta{
		Name: "SUBSCRIBE",
		Info: "subscribe command info here",
		Eval: nil,
	}
	qwatchCmdMeta = DiceCmdMeta{
		Name: "QWATCH",
		Info: "qwatch command info here",
		Eval: nil,
	}
	multiCmdMeta = DiceCmdMeta{
		Name: "MULTI",
		Info: "multi command info here",
		Eval: evalMULTI,
	}
	execCmdMeta = DiceCmdMeta{
		Name: "EXEC",
		Info: "exec command info here",
		Eval: nil,
	}
	discardCmdMeta = DiceCmdMeta{
		Name: "DISCARD",
		Info: "discard command info here",
		Eval: nil,
	}
	abortCmdMeta = DiceCmdMeta{
		Name: "ABORT",
		Info: "abort command info here",
		Eval: nil,
	}
)

func init() {
	diceCmds["PING"] = pingCmdMeta
	diceCmds["SET"] = setCmdMeta
	diceCmds["GET"] = getCmdMeta
	diceCmds["TTL"] = ttlCmdMeta
	diceCmds["DEL"] = delCmdMeta
	diceCmds["EXPIRE"] = expireCmdMeta
	diceCmds["HELLO"] = helloCmdMeta
	diceCmds["BGREWRITEAOF"] = bgrewriteaofCmdMeta
	diceCmds["INCR"] = incrCmdMeta
	diceCmds["INFO"] = infoCmdMeta
	diceCmds["CLIENT"] = clientCmdMeta
	diceCmds["LATENCY"] = latencyCmdMeta
	diceCmds["LRU"] = lruCmdMeta
	diceCmds["SLEEP"] = sleepCmdMeta
	diceCmds["QINTINS"] = qintinsCmdMeta
	diceCmds["QINTREM"] = qintremCmdMeta
	diceCmds["QINTLEN"] = qintlenCmdMeta
	diceCmds["QINTPEEK"] = qintpeekCmdMeta
	diceCmds["BFINIT"] = bfinitCmdMeta
	diceCmds["BFADD"] = bfaddCmdMeta
	diceCmds["BFEXISTS"] = bfexistsCmdMeta
	diceCmds["BFINFO"] = bfinfoCmdMeta
	diceCmds["QREFINS"] = qrefinsCmdMeta
	diceCmds["QREFREM"] = qrefremCmdMeta
	diceCmds["QREFLEN"] = qreflenCmdMeta
	diceCmds["QREFPEEK"] = qrefpeekCmdMeta
	diceCmds["STACKINTPUSH"] = stackintpushCmdMeta
	diceCmds["STACKINTPOP"] = stackintpopCmdMeta
	diceCmds["STACKINTLEN"] = stackintlenCmdMeta
	diceCmds["STACKINTPEEK"] = stackintpeekCmdMeta
	diceCmds["STACKREFPUSH"] = stackrefpushCmdMeta
	diceCmds["STACKREFPOP"] = stackrefpopCmdMeta
	diceCmds["STACKREFLEN"] = stackreflenCmdMeta
	diceCmds["STACKREFPEEK"] = stackrefpeekCmdMeta
	diceCmds["SUBSCRIBE"] = subscribeCmdMeta
	diceCmds["QWATCH"] = qwatchCmdMeta
	diceCmds["MULTI"] = multiCmdMeta
	diceCmds["EXEC"] = execCmdMeta
	diceCmds["DISCARD"] = discardCmdMeta
	diceCmds["ABORT"] = abortCmdMeta
}
