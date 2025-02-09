package pitsnapshot

import (
	"context"
	"log"
	"time"

	"github.com/dicedb/dice/internal/store"
)

const (
	BufferSize = 1000
)

type (
	SnapshotStore interface {
		StartSnapshot(uint64, store.Snapshotter) error
		StopSnapshot(uint64) error
	}
	PointInTimeSnapshot struct {
		ctx        context.Context
		cancelFunc context.CancelFunc

		ID uint64

		store          SnapshotStore
		totalStoreShot uint8

		SnapshotMap *SnapshotMap

		flusher *PITFlusher

		StartedAt time.Time
		EndedAt   time.Time

		exitCh chan bool
	}
)

func NewPointInTimeSnapshot(ctx context.Context, store SnapshotStore) (pit *PointInTimeSnapshot, err error) {
	pit = &PointInTimeSnapshot{
		ctx:       ctx,
		ID:        uint64(time.Now().Nanosecond()),
		StartedAt: time.Now(),

		store:  store,
		exitCh: make(chan bool, 5),
	}
	pit.ctx, pit.cancelFunc = context.WithCancel(ctx)
	if pit.flusher, err = NewPITFlusher(pit.ctx, pit.ID, pit.exitCh); err != nil {
		return
	}
	if pit.SnapshotMap, err = NewSnapshotMap(pit.ctx, pit.flusher); err != nil {
		return
	}
	return
}

func (pit *PointInTimeSnapshot) processStoreUpdates() (err error) {
	for {
		select {
		case <-pit.ctx.Done():
			pit.Close()
			return
		case <-pit.exitCh:
			pit.Close()
			return
		}
	}
	return
}

func (pit *PointInTimeSnapshot) Run() (err error) {
	go pit.flusher.Start()
	if err = pit.store.StartSnapshot(pit.ID, pit.SnapshotMap); err != nil {
		return
	}
	go pit.processStoreUpdates()
	return
}

func (pit *PointInTimeSnapshot) Close() (err error) {
	pit.EndedAt = time.Now()
	pit.cancelFunc()
	log.Println("Closing snapshot", pit.ID, ". Total time taken",
		pit.EndedAt.Sub(pit.StartedAt), "for total keys", pit.SnapshotMap.totalKeys)
	return
}
