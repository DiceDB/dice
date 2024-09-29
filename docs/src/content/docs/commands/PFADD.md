---
title: PFADD
description: Documentation for the DiceDB command PFADD
---

The `PFADD` command in DiceDB is used to add elements to a HyperLogLog data structure. HyperLogLog is a probabilistic data structure used for estimating the cardinality of a set, i.e., the number of unique elements in a dataset. The `PFADD` command helps in maintaining this data structure by adding new elements to it.

## Syntax

```
PFADD key element [element ...]
```

## Parameters

- `key`: The name of the HyperLogLog data structure to which the elements will be added. If the key does not exist, a new HyperLogLog structure is created.
- `element`: One or more elements to be added to the HyperLogLog. Multiple elements can be specified, separated by spaces.

## Return Value

- `Integer reply`: The command returns `1` if at least one internal register was altered, `0` otherwise. This indicates whether the HyperLogLog was modified by the addition of the new elements.

## Behaviour

When the `PFADD` command is executed, the following steps occur:

1. `Key Existence Check`: DiceDB checks if the specified key exists.
   - If the key does not exist, a new HyperLogLog data structure is created.
   - If the key exists but is not a HyperLogLog, an error is returned.
2. `Element Addition`: The specified elements are added to the HyperLogLog.
3. `Register Update`: The internal registers of the HyperLogLog are updated based on the hash values of the added elements.
4. `Cardinality Estimation`: The HyperLogLog uses the updated registers to maintain an estimate of the cardinality of the set.

## Error Handling

- `Wrong Type Error`: If the key exists but is not a HyperLogLog, DiceDB will return an error:
  ```
  (error) WRONGTYPE Operation against a key holding the wrong kind of value
  ```
- `Syntax Error`: If the command is not used with the correct syntax, DiceDB will return a syntax error:
  ```
  (error) ERR wrong number of arguments for 'pfadd' command
  ```

## Example Usage

### Basic Example

Add a single element to a HyperLogLog:

```shell
> PFADD myhyperloglog "element1"
(integer) 1
```

### Adding Multiple Elements

Add multiple elements to a HyperLogLog:

```shell
> PFADD myhyperloglog "element1" "element2" "element3"
(integer) 1
```

### Checking if the HyperLogLog was Modified

If the elements are already present and do not alter the internal registers:

```shell
> PFADD myhyperloglog "element1"
(integer) 0
```

### Error Example

Attempting to add elements to a key that is not a HyperLogLog:

```shell
> SET mykey "notahyperloglog"
OK
> PFADD mykey "element1"
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `PFADD` command is part of the HyperLogLog family of commands in DiceDB, which also includes `PFCOUNT` and `PFMERGE`.
- HyperLogLog is a probabilistic data structure, so it provides an approximate count of unique elements with a standard error of 0.81%.
- The `PFADD` command is useful for applications that need to count unique items in a large dataset efficiently, such as unique visitor counts, unique search queries, etc.

By understanding and using the `PFADD` command effectively, you can leverage DiceDB's powerful HyperLogLog data structure to manage and estimate the cardinality of large sets with minimal memory usage.

