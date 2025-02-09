package pitsnapshot

import (
	"context"
	"fmt"
	"log"
	"os"
)

const (
	FlushFilePath = "/tmp/flusher-%d.ddb"
	Delimiter     = "\n"
)

type (
	PITFlusher struct {
		ctx        context.Context
		snapshotID uint64
		updatesCh  chan []StoreMapUpdate
		exitCh     chan bool

		flushFile *os.File
	}
)

func NewPITFlusher(ctx context.Context, snapshotID uint64, exitCh chan bool) (pf *PITFlusher, err error) {
	pf = &PITFlusher{
		ctx:        ctx,
		snapshotID: snapshotID,
		exitCh:     exitCh,
		updatesCh:  make(chan []StoreMapUpdate, BufferSize),
	}
	if err = pf.setup(); err != nil {
		return
	}
	return
}

func (pf *PITFlusher) setup() (err error) {
	if pf.flushFile, err = os.Create(fmt.Sprintf(FlushFilePath, pf.snapshotID)); err != nil {
		return
	}
	return
}

func (pf *PITFlusher) Start() (err error) {
	for {
		select {
		case <-pf.ctx.Done():
			return
		case updates := <-pf.updatesCh:
			// TODO: Store the failed updates somewhere
			if err = pf.Flush(updates); err != nil {
				log.Println("error in flushing updates", err)
				continue
			}
			continue
		}
	}
	return
}

func (pf *PITFlusher) Flush(updates []StoreMapUpdate) (err error) {
	var (
		serializedUpdates []byte
		storeUpdate       StoreMapUpdate
	)
	if serializedUpdates, err = storeUpdate.Serialize(updates); err != nil {
		return
	}
	serializedUpdates = append(serializedUpdates, Delimiter...)
	if _, err = pf.flushFile.Write(serializedUpdates); err != nil {
	}
	return
}

// Close is called when the overlying SnapshotMap's channel is closed.
// After the SnapshotMap is closed, there is one final push to update all the pending
// data to the flusher
func (pf *PITFlusher) Close() (err error) {
	log.Println("Closing the flusher for snapshot", pf.snapshotID)
	pf.exitCh <- true
	return
}
