package core

import (
	"errors"
	"testing"

	"github.com/dicedb/dice/testutils"
	"gotest.tools/v3/assert"
)

func TestEvalSET(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []byte
	}{
		{
			name: "sanity check",
			args: testutils.ParseCommand("SET k v")[1:],
			want: RESP_OK,
		},
		{
			name: "EX check",
			args: testutils.ParseCommand("SET k v EX 1")[1:],
			want: RESP_OK,
		},
		{
			name: "PX check",
			args: testutils.ParseCommand("SET k v PX 1000")[1:],
			want: RESP_OK,
		},
		{
			name: "EX PX error check",
			args: testutils.ParseCommand("SET k v EX 1 PX 1000")[1:],
			want: Encode(errors.New("ERR syntax error"), false),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalSET(tt.args)
			assert.DeepEqual(t, tt.want, got)
		})
	}
}

func BenchmarkEvalSET(b *testing.B) {
	for i := 0; i < b.N; i++ {
		evalSET(testutils.ParseCommand("SET k v EX 1")[1:])
	}
}
