---
title: PFMERGE
description: Documentation for the DiceDB command PFMERGE
---

The `PFMERGE` command in DiceDB is used to merge multiple HyperLogLog data structures into a single HyperLogLog. HyperLogLog is a probabilistic data structure used for estimating the cardinality of a set, i.e., the number of unique elements in a dataset. This command is particularly useful when you have multiple HyperLogLogs and you want to combine them to get an estimate of the total unique elements across all of them.

## Syntax

```bash
PFMERGE destkey sourcekey [sourcekey ...]
```

## Parameters

| Parameter   | Description                                                                                                              | Type         | Required |
| ----------- | ------------------------------------------------------------------------------------------------------------------------ | ------------ | -------- |
| `destkey`   | The key where the merged HyperLogLog will be stored. If this key already exists, it will be overwritten.                 | String       | Yes      |
| `sourcekey` | One or more keys of the HyperLogLogs that you want to merge. These keys must already exist and contain HyperLogLog data. | List[String] | Yes      |

## Return Values

| Condition                                   | Return Value |
| ------------------------------------------- | ------------ |
| Command is successful                       | `OK`         |
| Syntax or specified constraints are invalid | `(error)`    |

## Behaviour

- If the `destkey` already exists, the `PFMERGE` command will overwrite the existing value with the new merged HyperLogLog.
- The command retrieves the HyperLogLog data structures from the specified `sourcekey` keys.
- These `sourcekey` keys are merged into a single HyperLogLog and stored in the `destkey`.
- If the `sourcekey` keys are not valid HyperLogLogs, an error is returned.

## Errors

The `PFMERGE` command can raise errors in the following scenarios:

1. `Wrong Type Error`:

   - `(error)`: `WRONGTYPE Operation against a key holding the wrong kind of value`
   - If any of the `sourcekey` keys or the `destkey` key contains a value that is not a HyperLogLog, DiceDB will return an error.

2. `Non-Existent Key Error`:
   - If any of the `sourcekey` keys do not exist, DiceDB will treat them as empty HyperLogLogs and proceed with the merge operation without raising an error.

## Example Usage

### Basic Usage

Suppose you have three HyperLogLogs stored at keys `hll1`, `hll2`, and `hll3`, and you want to merge them into a new HyperLogLog stored at key `hll_merged`.

```bash
127.0.0.1:7379> PFADD hll1 "a" "b" "c"
(integer) 1
127.0.0.1:7379> PFADD hll2 "c" "d" "e"
(integer) 1
127.0.0.1:7379> PFADD hll3 "e" "f" "g"
(integer) 1
127.0.0.1:7379> PFMERGE hll_merged hll1 hll2 hll3
OK
127.0.0.1:7379> PFCOUNT hll_merged
(integer) 7
```

### Overwriting Existing Key

If the `destkey` already exists, it will be overwritten by the merged HyperLogLog.

```bash
127.0.0.1:7379> PFADD hll_merged "x" "y" "z"
(integer) 1
127.0.0.1:7379> PFMERGE hll_merged hll1 hll2 hll3
OK
127.0.0.1:7379> PFCOUNT hll_merged
(integer) 7
```

### Handling Non-Existent Source Keys

If a `sourcekey` does not exist, DiceDB will treat it as an empty HyperLogLog.

```bash
127.0.0.1:7379> PFMERGE hll_merged hll1 hll2 non_existent_key
OK
127.0.0.1:7379> PFCOUNT hll_merged
(integer) 5
```

### Invalid Usage

if a `sourcekey` exists and is not of type HyperLogLog, the command will result in an error

```bash
127.0.0.1:7379> PFMERGE hll_merged not_hyperLogLog
(error) WRONGTYPE Key is not a valid HyperLogLog string value
```
