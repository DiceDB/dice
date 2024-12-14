package eval

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/cmd"
	dstore "github.com/dicedb/dice/internal/store"
)

func BenchmarkEvalSet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	evalCmds := make([]*Eval, b.N)
	for i := 0; i < b.N; i++ {
		cmd := &cmd.DiceDBCmd{
			Cmd:  "SET",
			Args: []string{"key", fmt.Sprintf("%d", i)},
		}
		e := NewEval(store, cmd)
		evalCmds[i] = e
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalCmds[i].evalSET()
	}
}

func BenchmarkEvalGet(b *testing.B) {
	store := dstore.NewStore(nil, nil, nil)
	for i := 0; i < 1000; i++ {
		cmd := &cmd.DiceDBCmd{
			Cmd:  "SET",
			Args: []string{fmt.Sprintf("key%d", i), fmt.Sprintf("%d", i)},
		}
		e := NewEval(store, cmd)
		e.evalSET()
	}
	evalCmds := make([]*Eval, b.N)
	for i := 0; i < b.N; i++ {
		randInt := i % 1000
		cmd := &cmd.DiceDBCmd{
			Cmd:  "GET",
			Args: []string{fmt.Sprintf("key%d", randInt)},
		}
		evalCmds[i] = NewEval(store, cmd)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evalCmds[i].evalGET()
	}
}
