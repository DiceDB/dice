---
title: SDIFF
description: Documentation for the DiceDB command SDIFF
---

# DiceDB Command: SDIFF

## Description

The `SDIFF` command in DiceDB is used to compute the difference between multiple sets. It returns the members of the set resulting from the difference between the first set and all the successive sets. This command is useful when you need to find elements that are unique to the first set compared to other sets.

## Syntax

```
SDIFF key1 [key2 ... keyN]
```

## Parameters

- `key1`: The key of the first set.
- `key2 ... keyN`: The keys of the subsequent sets to be compared with the first set. These parameters are optional, but at least one key must be provided.

## Return Value

The `SDIFF` command returns an array of elements that are present in the first set but not in any of the subsequent sets. If the first set does not exist, it is considered an empty set. If none of the sets exist, an empty array is returned.

## Behaviour

When the `SDIFF` command is executed:

1. DiceDB retrieves the set associated with `key1`.
2. DiceDB retrieves the sets associated with `key2` through `keyN`.
3. DiceDB computes the difference by removing elements found in `key2` through `keyN` from the set found in `key1`.
4. The resulting set, which contains elements unique to `key1`, is returned.

## Error Handling

- `Wrong Type Error`: If any of the keys provided do not hold a set, DiceDB will return an error of type `WRONGTYPE`. This error indicates that the operation against a key holding the wrong kind of value was attempted.
- `Syntax Error`: If no keys are provided, DiceDB will return a syntax error indicating that at least one key must be specified.

## Example Usage

### Example 1: Basic Usage

```DiceDB
SADD set1 "a" "b" "c"
SADD set2 "c" "d" "e"
SADD set3 "a" "f"

SDIFF set1 set2 set3
```

`Output:`

```
1) "b"
```

In this example, the difference between `set1` and the union of `set2` and `set3` is computed. The element "b" is unique to `set1`.

### Example 2: Single Set

```DiceDB
SADD set1 "a" "b" "c"

SDIFF set1
```

`Output:`

```
1) "a"
2) "b"
3) "c"
```

In this example, since only one set is provided, the command returns all elements of `set1`.

### Example 3: Non-Existent Sets

```DiceDB
SDIFF set1 set2
```

`Output:`

```
(empty array)
```

In this example, since neither `set1` nor `set2` exist, the command returns an empty array.

## Error Handling Examples

### Example 1: Wrong Type Error

```DiceDB
SET not_a_set "value"

SDIFF not_a_set
```

`Output:`

```
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, `not_a_set` is not a set, so DiceDB returns a `WRONGTYPE` error.

### Example 2: Syntax Error

```DiceDB
SDIFF
```

`Output:`

```
(error) ERR wrong number of arguments for 'sdiff' command
```

In this example, no keys are provided, so DiceDB returns a syntax error.
