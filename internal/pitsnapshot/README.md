# Point in time Snapshots

A point in time snapshot refers to the copy of the existing data which is representative of
the data in the memory at that specific time. 

## Goals
- Don't affect the throughput performance of the current request processing layer.
- Ability to take multiple snapshot instances simultaneously.
- Ability to snapshot and restore on systems with different shards
- Shouldn't depend on existing data files apart from the in-memory data structures

## Design

### Recap
DiceDB runs multiple `ShardThreads` based on the number of CPUs in the machine and there is
one `Store` object for every thread. Since DiceDB follows a shared nothing architecture, it
should be possible to snapshot and restore data with varying CPU counts.

The `Store` object keeps a reference to the `Object` object in a map. The `Object` object is
where the data is being stored.

### Implementing Copy on write
The snapshotting technique would be similar to the copy-on-write mechanism, ie, additional data
wouldn't have to be stored till the data has to be modified. This means additional memory would
only be required if there are changes to the underyling data.

### Impact on current latency benchmarks
- For reads, there should be minimal latency change since there are no references to the `get`
methods even when snapshotting is running. One thing which may impact the read latency is that
it has to iterate through all the keys, so an implicit lock inside the datastructure may be
required.
- For writes, if a snapshot is going on, then it has to write in 2 places and an additional read
to a map.

### Flow

The initiation flow:
```bash
ShardThread::CallSnapshotter -> Snapshotter::Start -> Store::StartSnapshot -> SnapshotMap::Buffer
-> PITFlusher::Flush
```

When the iteration is over
```bash
Store::StopSnapshot -> SnapshotMap::FlushAllData -> PITFlusher::FlushAllData -> Snapshotter::Close
```

### Changes for ShardThread and Store
The snapshot would start on every `ShardThread` and fetch the `Store` object. Every `Store` object
needs to implement the interface `SnapshotStore` which is contains the `StartSnapshot` and `StopSnapshot`
methods.
The `StartSnapshot` and `StopSnapshot` methods would be called on the store from the snapshotter object.

#### StartSnapshot
When the `StartSnapshot` method is called, the `Store` should keep note of the `SnapshotID` in a map.
There can be multiple instances of snapshots for every store as well.
For any read or write operation which is performed, the `Store` object should check if a snapshot is being
run at that instance. If no snapshot is being run, then continue as usual.
If a snapshot is being run, then for any subsequent write operation, store the previous data in the snapshot's
object, maybe a map. Let's call this the `SnapshotMap`. If there are multiple write operations to the same object
and the data already exists in the `SnapshotMap`, then skip doing anything for the snapshot.
Similarly, for reads, if a snapshot is being run, if the incoming request is from a snapshot layer, then check
if there is anything in the `SnapshotMap` for the key. If no, then return the current value from the `Store`.

It should fetch the list of keys in its store attribute and iterate through them.

#### StopSnapshot
When the iteration through all the keys by the `Store` object is done, the `StopSnapshot` method is called by the
`Store`. The `StopSnapshot` lets the `SnapshotMap` know that there are no more updates coming. The `SnapshotMap`
then talks to the `PITFLusher` to finish syncing all the chunks to disk and then closes the main snapshot
process.

### Point-in-time Flusher
The `PITFlusher` serializes the store updates from the `SnapshotMap` to binary format, currently `gob`.
It serializes and appends to a file.

## Test cases and benchmarks
- Snapshot data less than the buffer size without any subsequent writes
- Snapshot data less than the buffer size with localized subsequent writes
- Snapshot data less than the buffer size with spread out subsequent writes
- Snapshot data more than the buffer size without any subsequent writes
- Snapshot data more than the buffer size with localized subsequent writes
- Snapshot data more than the buffer size with spread out subsequent writes
- Ensure current `get` path is not affected