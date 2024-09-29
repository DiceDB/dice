---
title: PFMERGE
description: Documentation for the DiceDB command PFMERGE
---

The `PFMERGE` command in DiceDB is used to merge multiple HyperLogLog data structures into a single HyperLogLog. HyperLogLog is a probabilistic data structure used for estimating the cardinality of a set, i.e., the number of unique elements in a dataset. This command is particularly useful when you have multiple HyperLogLogs and you want to combine them to get an estimate of the total unique elements across all of them.

## Syntax

```
PFMERGE destkey sourcekey [sourcekey ...]
```

## Parameters

- `destkey`: The key where the merged HyperLogLog will be stored. If this key already exists, it will be overwritten.
- `sourcekey`: One or more keys of the HyperLogLogs that you want to merge. These keys must already exist and contain HyperLogLog data structures.

## Return Value

- `Simple String Reply`: Returns `OK` if the merge operation is successful.

## Behaviour

When the `PFMERGE` command is executed, DiceDB will:

1. Retrieve the HyperLogLog data structures from the specified `sourcekey` keys.
2. Merge these HyperLogLogs into a single HyperLogLog.
3. Store the resulting HyperLogLog in the `destkey`.
4. If the `destkey` already exists, its previous value will be overwritten with the new merged HyperLogLog.

## Error Handling

The `PFMERGE` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If any of the `sourcekey` keys or the `destkey` key contains a value that is not a HyperLogLog, DiceDB will return an error.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

2. `Non-Existent Key Error`: If any of the `sourcekey` keys do not exist, DiceDB will treat them as empty HyperLogLogs and proceed with the merge operation without raising an error.

## Example Usage

### Example 1: Basic Usage

Suppose you have three HyperLogLogs stored at keys `hll1`, `hll2`, and `hll3`, and you want to merge them into a new HyperLogLog stored at key `hll_merged`.

```sh
127.0.0.1:6379> PFADD hll1 "a" "b" "c"
(integer) 1
127.0.0.1:6379> PFADD hll2 "c" "d" "e"
(integer) 1
127.0.0.1:6379> PFADD hll3 "e" "f" "g"
(integer) 1
127.0.0.1:6379> PFMERGE hll_merged hll1 hll2 hll3
OK
127.0.0.1:6379> PFCOUNT hll_merged
(integer) 7
```

### Example 2: Overwriting Existing Key

If the `destkey` already exists, it will be overwritten by the merged HyperLogLog.

```sh
127.0.0.1:6379> PFADD hll_merged "x" "y" "z"
(integer) 1
127.0.0.1:6379> PFMERGE hll_merged hll1 hll2 hll3
OK
127.0.0.1:6379> PFCOUNT hll_merged
(integer) 7
```

### Example 3: Handling Non-Existent Source Keys

If a `sourcekey` does not exist, DiceDB will treat it as an empty HyperLogLog.

```sh
127.0.0.1:6379> PFMERGE hll_merged hll1 hll2 non_existent_key
OK
127.0.0.1:6379> PFCOUNT hll_merged
(integer) 5
```
