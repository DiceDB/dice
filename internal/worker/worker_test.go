package worker

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/dicedb/dice/internal/cmd"
	"github.com/dicedb/dice/internal/ops"
	"github.com/dicedb/dice/internal/shard"
	dstore "github.com/dicedb/dice/internal/store"
)

type MockIOHandler struct{}
func (m *MockIOHandler) Read(ctx context.Context) ([]byte, error) { 
	return []byte("mock data"), nil 
}

func (m *MockIOHandler) Write(ctx context.Context, data interface{}) error {
	 return nil 
}

func (m *MockIOHandler) Close() error {
	return nil
}

type MockParser struct{}
func (m *MockParser) Parse(data []byte) ([]*cmd.DiceDBCmd, error) { 
	return []*cmd.DiceDBCmd{{}}, nil 
}

func setupMockWorker() *BaseWorker {
	mockIOHandler := &MockIOHandler{}
	mockParser := &MockParser{}
	responseChan := make(chan *ops.StoreResponse, 1000)
	preprocessingChan := make(chan *ops.StoreResponse, 1000)
	globalErrChan := make(chan error, 100)
	queryWatchChan := make(chan dstore.QueryWatchEvent, 100)
	cmdWatchChan := make(chan dstore.CmdWatchEvent, 100)
	errorChan := make(chan error)
	mockShardManager := shard.NewShardManager(
		uint8(runtime.NumCPU()),
		queryWatchChan,
		cmdWatchChan,
		globalErrChan,
	)

	
	return NewWorker("mock-worker", responseChan, preprocessingChan, mockIOHandler, mockParser, mockShardManager, errorChan)
}

func BenchmarkBaseWorkerStart(b *testing.B) {
	worker := setupMockWorker()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = worker.Start(ctx)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := &cmd.DiceDBCmd{
			Cmd: "PING",
		}
		select {
		case worker.adhocReqChan <- cmd:
		case <-time.After(1 * time.Second):
			b.Error("Timeout waiting to send command to worker")
		}
	}

	cancel()
	b.StopTimer()
}