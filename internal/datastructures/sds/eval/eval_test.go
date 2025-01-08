package eval

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/store"
)

func TestEvalGetBit(t *testing.T) {
	store := store.NewStore(nil, nil, nil)

	op := cmd.DiceDBCmd{
		Cmd:  "SET",
		Args: []string{"key", "foo"},
	}

	e := NewEval(store, &op)
	resp := e.evalSET()
	if resp.Error != nil {
		t.Errorf("Expected nil, got %v", resp.Error)
	}
	op = cmd.DiceDBCmd{
		Cmd:  "GETBIT",
		Args: []string{"key", "128"},
	}
	e = NewEval(store, &op)
	resp = e.evalGETBIT()
	if resp.Error != nil {
		t.Errorf("Expected nil, got %v", resp.Error)
	}
	fmt.Println(resp.Result)
}
func TestEvalSetBit(t *testing.T) {
	store := store.NewStore(nil, nil, nil)

	op := cmd.DiceDBCmd{
		Cmd:  "SETBIT",
		Args: []string{"key", "1", "1"},
	}

	e := NewEval(store, &op)
	resp := e.evalSETBIT()
	if resp.Error != nil {
		t.Errorf("Expected nil, got %v", resp.Error)
	}
	op = cmd.DiceDBCmd{
		Cmd:  "GET",
		Args: []string{"key"},
	}
	e = NewEval(store, &op)
	resp = e.evalGET()
	if resp.Error != nil {
		t.Errorf("Expected nil, got %v", resp.Error)
	}
	fmt.Println(resp.Result)
}
