package core

type DiceCmdMeta struct {
	Name string
	Info string
	Docs string
	Eval func([]string) []byte
}

var (
	diceCmds = map[string]DiceCmdMeta{}

	pingCmdMeta = DiceCmdMeta{
		Name: "PING",
		Info: "ping command info here",
		Docs: "ping command docs here",
		Eval: evalPING,
	}
	setCmdMeta = DiceCmdMeta{
		Name: "SET",
		Info: "set command info here",
		Docs: "set command docs here",
		Eval: evalSET,
	}
	getCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: "get command info here",
		Docs: "get command docs here",
		Eval: evalGET,
	}
	ttlCmdMeta = DiceCmdMeta{
		Name: "TTL",
		Info: "ttl command info here",
		Docs: "ttl command docs here",
		Eval: evalTTL,
	}
	delCmdMeta = DiceCmdMeta{
		Name: "DEL",
		Info: "del command info here",
		Docs: "del command docs here",
		Eval: evalDEL,
	}
	expireCmdMeta = DiceCmdMeta{
		Name: "EXPIRE",
		Info: "expire command info here",
		Docs: "expire command docs here",
		Eval: evalEXPIRE,
	}
	helloCmdMeta = DiceCmdMeta{
		Name: "HELLO",
		Info: "hello command info here",
		Docs: "hello command docs here",
		Eval: evalHELLO,
	}
	bgrewriteaofCmdMeta = DiceCmdMeta{
		Name: "BGREWRITEAOF",
		Info: "bgrewriteaof command info here",
		Docs: "bgrewriteaof command docs here",
		Eval: evalBGREWRITEAOF,
	}
	incrCmdMeta = DiceCmdMeta{
		Name: "INCR",
		Info: "incr command info here",
		Docs: "incr command docs here",
		Eval: evalINCR,
	}
	infoCmdMeta = DiceCmdMeta{
		Name: "INFO",
		Info: "info command info here",
		Docs: "info command docs here",
		Eval: evalINFO,
	}
	clientCmdMeta = DiceCmdMeta{
		Name: "CLIENT",
		Info: "client command info here",
		Docs: "client command docs here",
		Eval: evalCLIENT,
	}
	latencyCmdMeta = DiceCmdMeta{
		Name: "LATENCY",
		Info: "latency command info here",
		Docs: "latency command docs here",
		Eval: evalLATENCY,
	}
	lruCmdMeta = DiceCmdMeta{
		Name: "LRU",
		Info: "lru command info here",
		Docs: "lru command docs here",
		Eval: evalLRU,
	}
	sleepCmdMeta = DiceCmdMeta{
		Name: "SLEEP",
		Info: "sleep command info here",
		Docs: "sleep command docs here",
		Eval: evalSLEEP,
	}
	qintinsCmdMeta = DiceCmdMeta{
		Name: "QINTINS",
		Info: "qintins command info here",
		Docs: "qintins command docs here",
		Eval: evalQINTINS,
	}
	qintremCmdMeta = DiceCmdMeta{
		Name: "QINTREM",
		Info: "qintrem command info here",
		Docs: "qintrem command docs here",
		Eval: evalQINTREM,
	}
	qintlenCmdMeta = DiceCmdMeta{
		Name: "QINTLEN",
		Info: "qintlen command info here",
		Docs: "qintlen command docs here",
		Eval: evalQINTLEN,
	}
	qintpeekCmdMeta = DiceCmdMeta{
		Name: "QINTPEEK",
		Info: "qintpeek command info here",
		Docs: "qintpeek command docs here",
		Eval: evalQINTPEEK,
	}
	bfinitCmdMeta = DiceCmdMeta{
		Name: "BFINIT",
		Info: "bfinit command info here",
		Docs: "bfinit command docs here",
		Eval: evalBFINIT,
	}
	bfaddCmdMeta = DiceCmdMeta{
		Name: "BFADD",
		Info: "bfadd command info here",
		Docs: "bfadd command docs here",
		Eval: evalBFADD,
	}
	bfexistsCmdMeta = DiceCmdMeta{
		Name: "BFEXISTS",
		Info: "bfexists command info here",
		Docs: "bfexists command docs here",
		Eval: evalBFEXISTS,
	}
	bfinfoCmdMeta = DiceCmdMeta{
		Name: "BFINFO",
		Info: "bfinfo command info here",
		Docs: "bfinfo command docs here",
		Eval: evalBFINFO,
	}
	qrefinsCmdMeta = DiceCmdMeta{
		Name: "QREFINS",
		Info: "qrefins command info here",
		Docs: "qrefins command docs here",
		Eval: evalQREFINS,
	}
	qrefremCmdMeta = DiceCmdMeta{
		Name: "QREFREM",
		Info: "qrefrem command info here",
		Docs: "qrefrem command docs here",
		Eval: evalQREFREM,
	}
	qreflenCmdMeta = DiceCmdMeta{
		Name: "QREFLEN",
		Info: "qreflen command info here",
		Docs: "qreflen command docs here",
		Eval: evalQREFLEN,
	}
	qrefpeekCmdMeta = DiceCmdMeta{
		Name: "GET",
		Info: "qrefpeek command info here",
		Docs: "qrefpeek command docs here",
		Eval: evalGET,
	}
	stackintpushCmdMeta = DiceCmdMeta{
		Name: "STACKINTPUSH",
		Info: "stackintpush command info here",
		Docs: "stackintpush command docs here",
		Eval: evalSTACKINTPUSH,
	}
	stackintpopCmdMeta = DiceCmdMeta{
		Name: "STACKINTPOP",
		Info: "stackintpop command info here",
		Docs: "stackintpop command docs here",
		Eval: evalSTACKINTPOP,
	}
	stackintlenCmdMeta = DiceCmdMeta{
		Name: "STACKINTLEN",
		Info: "stackintlen command info here",
		Docs: "stackintlen command docs here",
		Eval: evalSTACKINTLEN,
	}
	stackintpeekCmdMeta = DiceCmdMeta{
		Name: "STACKINTPEEK",
		Info: "stackintpeek command info here",
		Docs: "stackintpeek command docs here",
		Eval: evalSTACKINTPEEK,
	}
	stackrefpushCmdMeta = DiceCmdMeta{
		Name: "STACKREFPUSH",
		Info: "stackrefpush command info here",
		Docs: "stackrefpush command docs here",
		Eval: evalSTACKREFPUSH,
	}
	stackrefpopCmdMeta = DiceCmdMeta{
		Name: "STACKREFPOP",
		Info: "stackrefpop command info here",
		Docs: "stackrefpop command docs here",
		Eval: evalSTACKREFPOP,
	}
	stackreflenCmdMeta = DiceCmdMeta{
		Name: "STACKREFLEN",
		Info: "stackreflen command info here",
		Docs: "stackreflen command docs here",
		Eval: evalSTACKREFLEN,
	}
	stackrefpeekCmdMeta = DiceCmdMeta{
		Name: "STACKREFPEEK",
		Info: "stackrefpeek command info here",
		Docs: "stackrefpeek command docs here",
		Eval: evalSTACKREFPEEK,
	}
	// TODO: Remove this override once we support QWATCH in dice-cli.
	subscribeCmdMeta = DiceCmdMeta{
		Name: "SUBSCRIBE",
		Info: "subscribe command info here",
		Docs: "subscribe command docs here",
		Eval: nil,
	}
	qwatchCmdMeta = DiceCmdMeta{
		Name: "QWATCH",
		Info: "qwatch command info here",
		Docs: "qwatch command docs here",
		Eval: nil,
	}
	multiCmdMeta = DiceCmdMeta{
		Name: "MULTI",
		Info: "multi command info here",
		Docs: "multi command docs here",
		Eval: evalMULTI,
	}
	execCmdMeta = DiceCmdMeta{
		Name: "EXEC",
		Info: "exec command info here",
		Docs: "exec command docs here",
		Eval: nil,
	}
	discardCmdMeta = DiceCmdMeta{
		Name: "DISCARD",
		Info: "discard command info here",
		Docs: "discard command docs here",
		Eval: nil,
	}
	abortCmdMeta = DiceCmdMeta{
		Name: "ABORT",
		Info: "abort command info here",
		Docs: "abort command docs here",
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
