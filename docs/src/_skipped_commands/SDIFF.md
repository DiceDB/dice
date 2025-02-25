---
title: SDIFF
description: Documentation for the DiceDB command SDIFF
---

The `SDIFF` command in DiceDB is used to compute the difference between multiple sets. It returns the members of the set resulting from the difference between the first set and all the successive sets. This command is useful when you need to find elements that are unique to the first set compared to other sets.

## Syntax

```bash
SDIFF key1 [key2 ... keyN]
```

## Parameters

| Parameter    | Description                                                        | Type   | Required |
| ------------ | ------------------------------------------------------------------ | ------ | -------- |
| `key1`       | The key of the first set.                                          | String | Yes      |
| `key2..keyN` | The keys of the subsequent sets to be compared with the first set. | String | No       |

## Return Values

The `SDIFF` command returns an array of elements that are present in the first set but not in any of the subsequent sets. If the first set does not exist, it is considered an empty set. If none of the sets exist, an empty array is returned.

| Condition                                                        | Return Value        |
| ---------------------------------------------------------------- | ------------------- |
| Elements present in the first set but not in any subsequent sets | `Array of elements` |
| First set does not exist                                         | `Empty array`       |
| None of the sets exist                                           | `Empty array`       |
| Syntax or specified constraints are invalid                      | error               |

## Behaviour

- DiceDB retrieves the set associated with `key1`.
- DiceDB retrieves the sets associated with `key2` through `keyN`.
- DiceDB computes the difference by removing elements found in `key2` through `keyN` from the set found in `key1`.
- The resulting set, containing elements unique to `key1`, is returned.

## Errors

- `Wrong Type Error`:

  - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
  - If any of the keys provided do not hold a set, DiceDB will return an error of type `WRONGTYPE`. This error indicates that the operation against a key holding the wrong kind of value was attempted.

- `Wrong number of arguments`:

  - Error Message: `(error) ERR wrong number of arguments for 'sdiff' command`
  - If no keys are provided, DiceDB will return a syntax error indicating that at least one key must be specified.

- `Syntax Error`: If no keys are provided, DiceDB will return a syntax error indicating that at least one key must be specified.

## Example Usage

### Basic Usage

In this example, the difference between set1 and the union of set2 and set3 is computed. The element “b” is unique to set1.

```bash
127.0.0.1:7379> SADD set1 "a" "b" "c"
127.0.0.1:7379> SADD set2 "c" "d" "e"
127.0.0.1:7379> SADD set3 "a" "f"
127.0.0.1:7379> SDIFF set1 set2 set3

"b"
```

### Single Set

In this example, since only one set is provided, the command returns all elements of `set1`.

```bash
127.0.0.1:7379> SADD set1 "a" "b" "c"
127.0.0.1:7379> SDIFF set1
1) "a"
2) "b"
3) "c"
```

### Non-Existent Sets

In this example, since neither `set1` nor `set2` exist, the command returns an empty array.

```bash
127.0.0.1:7379> SDIFF set1 set2
(empty array)
```

### Wrong Type Error

In this example, `not_a_set` is not a set, so DiceDB returns a `WRONGTYPE` error.

```bash
127.0.0.1:7379> SET not_a_set "value"
127.0.0.1:7379> SDIFF not_a_set
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Syntax Error

In this example, no keys are provided, so DiceDB returns a syntax error.

```bash
127.0.0.1:7379> SDIFF
(error) ERR wrong number of arguments for 'sdiff' command

```
