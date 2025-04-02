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
		dlq        [][]StoreMapUpdate

		totalKeys uint64
		flushFile *os.File
	}
)

func NewPITFlusher(ctx context.Context, snapshotID uint64, exitCh chan bool) (pf *PITFlusher, err error) {
	pf = &PITFlusher{
		ctx:        ctx,
		snapshotID: snapshotID,
		exitCh:     exitCh,
		updatesCh:  make(chan []StoreMapUpdate, 10*BufferSize),
	}
	if err = pf.setup(); err != nil {
		return
	}
	return
}

func (pf *PITFlusher) setup() (err error) {
	filePath := fmt.Sprintf(FlushFilePath, pf.snapshotID)
	if _, err = os.Stat(filePath); os.IsExist(err) {
		return
	}
	if pf.flushFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
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
				//log.Println("error in flushing updates, pushing to DLQ", err)
				pf.dlq = append(pf.dlq, updates)
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
	pf.totalKeys += uint64(len(updates))
	if serializedUpdates, err = storeUpdate.Serialize(updates); err != nil {
		return
	}
	serializedUpdates = append(serializedUpdates, Delimiter...)
	if _, err = pf.flushFile.Write(serializedUpdates); err != nil {
		return
	}
	return
}

func (pf *PITFlusher) clearDLQ() (err error) {
	for {
		select {
		case updates := <-pf.updatesCh:
			if err := pf.Flush(updates); err != nil {
				log.Println("Error in flushing updates while draining", err)
				pf.dlq = append(pf.dlq, updates)
			}
		default:
			// Channel is empty
			goto Done
		}
	}
Done:
	totalKeysInDLQ := 0
	for _, updates := range pf.dlq {
		var (
			update []StoreMapUpdate
		)
		update = updates
		totalKeysInDLQ += len(update)
		if err = pf.Flush(update); err != nil {
			return
		}
	}
	if totalKeysInDLQ > 0 {
		log.Println("Total keys in DLQ", totalKeysInDLQ, len(pf.updatesCh))
	}
	return
}

// Close is called when the overlying SnapshotMap's channel is closed.
// After the SnapshotMap is closed, there is one final push to update all the pending
// data to the flusher
func (pf *PITFlusher) Close() (err error) {
	if err = pf.clearDLQ(); err != nil {
		log.Println("error in clearing DLQ", err)
	}
	//log.Println("Closing the flusher for snapshot", pf.snapshotID, ", total keys flushed", pf.totalKeys)
	pf.flushFile.Close()
	pf.exitCh <- true
	return
}
