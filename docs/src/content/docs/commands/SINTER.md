---
title: SINTER
description: Documentation for the DiceDB command SINTER
---

The `SINTER` command in DiceDB is used to compute the intersection of multiple sets. This command returns the members that are common to all the specified sets. If any of the sets do not exist, they are considered to be empty sets. The result of the intersection will be an empty set if at least one of the sets is empty.

## Parameters

- `key [key ...]`: One or more keys corresponding to the sets you want to intersect. At least one key must be provided.

## Return Value

- `Array of elements`: The command returns an array of elements that are present in all the specified sets. If no common elements are found, an empty array is returned.

## Behaviour

When the `SINTER` command is executed, DiceDB performs the following steps:

1. `Fetch Sets`: It retrieves the sets associated with the provided keys.
2. `Intersection Calculation`: It computes the intersection of these sets.
3. `Return Result`: It returns the members that are common to all the sets.

If any of the specified keys do not exist, they are treated as empty sets. The intersection of any set with an empty set is always an empty set.

## Error Handling

- `Wrong Type Error`: If any of the specified keys exist but are not of the set data type, DiceDB will return an error.
  - `Error Message`: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
- `No Key Error`: If no keys are provided, DiceDB will return an error.
  - `Error Message`: `(error) ERR wrong number of arguments for 'sinter' command`

## Example Usage

### Example 1: Basic Intersection

```shell
# Add elements to sets
SADD set1 "a" "b" "c"
SADD set2 "b" "c" "d"
SADD set3 "c" "d" "e"

# Compute intersection
SINTER set1 set2 set3
```

`Expected Output:`

```shell
1) "c"
```

### Example 2: Intersection with Non-Existent Set

```shell
# Add elements to sets
SADD set1 "a" "b" "c"
SADD set2 "b" "c" "d"

# Compute intersection with a non-existent set
SINTER set1 set2 set3
```

`Expected Output:`

```shell
(empty array)
```

### Example 3: Intersection with Empty Set

```shell
# Add elements to sets
SADD set1 "a" "b" "c"
SADD set2 "b" "c" "d"

# Create an empty set
SADD set3

# Compute intersection
SINTER set1 set2 set3
```

`Expected Output:`

```shell
(empty array)
```

### Example 4: Error Handling - Wrong Type

```shell
# Add elements to sets
SADD set1 "a" "b" "c"

# Create a string key
SET stringKey "value"

# Attempt to compute intersection with a non-set key
SINTER set1 stringKey
```

`Expected Output:`

```shell
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Example 5: Error Handling - No Keys Provided

```shell
# Attempt to compute intersection without providing any keys
SINTER
```

`Expected Output:`

```shell
(error) ERR wrong number of arguments for 'sinter' command
```
