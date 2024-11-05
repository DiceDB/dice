---
title: PFADD
description: The `PFADD` command in DiceDB is used to add elements to a HyperLogLog data structure. HyperLogLog is a probabilistic data structure used for estimating the cardinality of a set, i.e., the number of unique elements in a dataset.
---

The `PFADD` command in DiceDB is used to add elements to a HyperLogLog data structure. HyperLogLog is a probabilistic data structure used for estimating the cardinality of a set, i.e., the number of unique elements in a dataset.

## Syntax

```bash
PFADD key element [element ...]
```

## Parameters

| Parameter | Description                                                                                              | Type   | Required |
| --------- | -------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The name of the HyperLogLog data structure. If it does not exist, a new one is created.                  | String | Yes      |
| `element` | One or more elements to add to the HyperLogLog. Multiple elements can be specified, separated by spaces. | String | Yes      |

## Return values

| Condition                                  | Return Value |
| ------------------------------------------ | ------------ |
| At least one internal register was altered | `1`          |
| No internal register was altered           | `0`          |

## Behaviour

- The command first checks if the specified key exists.
- If the key does not exist, a new HyperLogLog data structure is created.
- If the key exists but is not a HyperLogLog, an error is returned.
- The specified elements are added to the HyperLogLog, and the internal registers are updated based on the hash values of these elements.
- The HyperLogLog maintains an estimate of the cardinality of the set using these updated registers.

## Errors

1. `Wrong type error`:

   - Error Message: `(error) WRONGTYPE Key is not a valid HyperLogLog string value`
   - Occurs when trying to use the command on a key that is not a HyperLogLog.

2. `Syntax error`:

   - Error Message: `(error) wrong number of arguments for 'pfadd' command`
   - Occurs when the command syntax is incorrect or missing required parameters.

## Example Usage

### Basic Usage

Adding a single element to a HyperLogLog:

```bash
127.0.0.1:7379> PFADD myhyperloglog "element1"
(integer) 1
```

### Adding Multiple Elements

Adding multiple elements to a HyperLogLog:

```bash
127.0.0.1:7379> PFADD myhyperloglog "element1" "element2" "element3"
(integer) 1
```

### Checking if the HyperLogLog was Modified

If the elements do not alter the internal registers:

```bash
127.0.0.1:7379> PFADD myhyperloglog "element1"
(integer) 0
```

### Invalid Usage

Attempting to add elements to a key that is not a HyperLogLog:

```bash
127.0.0.1:7379> SET mykey "notahyperloglog"
OK
127.0.0.1:7379> PFADD mykey "element1"
(error) WRONGTYPE Key is not a valid HyperLogLog string value
```

---
