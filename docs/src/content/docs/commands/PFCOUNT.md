---
title: PFCOUNT
description: Documentation for the DiceDB command PFCOUNT
---

The `PFCOUNT` command in DiceDB is used to return the approximate cardinality (i.e., the number of unique elements) of the set(s) stored in the specified HyperLogLog data structure(s). HyperLogLog is a probabilistic data structure used for estimating the cardinality of a set with a high degree of accuracy and minimal memory usage.

## Syntax

```bash
PFCOUNT key [key ...]
```

## Parameters

| Parameter | Description                                                                                                               | Type   | Required |
| --------- | ------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key(s) of the HyperLogLog data structure(s) whose cardinality you want to estimate. You can specify one or more keys. | String | Yes      |

## Return Values

| Condition                                                   | Return Value                                          |
| ----------------------------------------------------------- | ----------------------------------------------------- |
| The specified `key` exists and contains a valid HyperLogLog | The estimated number of unique elements as an integer |
| The `key` does not exist or is not a valid HyperLogLog      | `0`                                                   |

## Behaviour

When the `PFCOUNT` command is executed, DiceDB will:

1. Retrieve the HyperLogLog data structure(s) associated with the specified key(s).
2. Estimate the cardinality of the set(s) represented by the HyperLogLog data structure(s).
3. If multiple keys are specified, DiceDB will merge the HyperLogLog data structures and return the cardinality of the union of the sets.

## Errors

The `PFCOUNT` command can raise errors in the following scenarios:

1. `Non-Existent Key`: If the specified key does not exist, DiceDB will treat it as an empty HyperLogLog and return a cardinality of 0.

   ```bash
   127.0.0.1:7379> PFCOUNT non_existent_key
   (integer) 0
   ```

2. `Wrong Type of Key`: If the specified key exists but is not a HyperLogLog data structure, DiceDB will return a type error.

   `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

   ```bash
   127.0.0.1:7379> SET mykey "value"
   OK
   127.0.0.1:7379> PFCOUNT mykey
   (error) WRONGTYPE Operation against a key holding the wrong kind of value
   ```

3. `Invalid Arguments`: If the command is called with no arguments, DiceDB will return a syntax error.

   `Error Message`: `ERR wrong number of arguments for 'pfcount' command`

   ```bash
   127.0.0.1:7379> PFCOUNT
   (error) ERR wrong number of arguments for 'pfcount' command
   ```

## Example usage

### Single Key

```bash
127.0.0.1:7379> PFADD hll1 "foo" "bar" "baz"
(integer) 1
127.0.0.1:7379> PFCOUNT hll1
(integer) 3
```

### Multiple Keys

```bash
127.0.0.1:7379> PFADD hll1 "foo" "bar"
(integer) 1
127.0.0.1:7379> PFADD hll2 "baz" "qux"
(integer) 1
127.0.0.1:7379> PFCOUNT hll1 hll2
(integer) 4
```

## Notes

- The `PFCOUNT` command provides an approximate count, not an exact count. The error rate is typically less than 1%.
- HyperLogLog is particularly useful for large datasets where an exact count is not feasible due to memory constraints.

By understanding the `PFCOUNT` command, you can efficiently estimate the cardinality of large sets with minimal memory usage, making it a powerful tool for various applications such as analytics and monitoring.
