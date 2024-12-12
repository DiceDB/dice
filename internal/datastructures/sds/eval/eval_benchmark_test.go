package eval

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/cmd"
	dstore "github.com/dicedb/dice/internal/store"
)

func BenchmarkEvalSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < b.N; i++ {
		cmd := &cmd.DiceDBCmd{
			Cmd:  "SET",
			Args: []string{"key", fmt.Sprintf("%d", i)},
		}
		e := NewEval(store, cmd)
		e.evalSET()
	}
}
