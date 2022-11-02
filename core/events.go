package core

import "github.com/dicedb/dice/handlers"

func Shutdown(dh *handlers.DiceKVstoreHandler) {
	evalBGREWRITEAOF([]string{}, dh)
}
