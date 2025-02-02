---
title: SINTER
description: The `SINTER` command in DiceDB is used to compute the intersection of multiple sets. This command returns the members that are common to all the specified sets. If any of the sets do not exist, they are considered to be empty sets. The result of the intersection will be an empty set if at least one of the sets is empty.
---

The `SINTER` command in DiceDB is used to compute the intersection of multiple sets. This command returns the members that are common to all the specified sets. If any of the sets do not exist, they are considered to be empty sets. The result of the intersection will be an empty set if at least one of the sets is empty.

## Syntax

```bash
SINTER key [key ...]
```

## Parameters

| Parameter       | Description                                                                                    | Type   | Required |
| --------------- | ---------------------------------------------------------------------------------------------- | ------ | -------- |
| `key [key ...]` | One or more identifier keys representing sets to intersect. At least one key must be provided. | String | Yes      |

## Return Values

| Condition                           | Return Value                                                              |
| ----------------------------------- | ------------------------------------------------------------------------- |
| Common elements exist               | array of elements (as strings) that are present in all the specified sets |
| No common elements exist            | `(empty array)`                                                           |
| Invalid syntax or no specified keys | error                                                                     |

## Behaviour

When the `SINTER` command is executed, DiceDB performs the following steps:

1. `Fetch Sets`: It retrieves the sets associated with the provided keys.
2. `Intersection Calculation`: It computes the intersection of these sets.
3. `Return Result`: It returns the members that are common to all the sets.

If any of the specified keys do not exist, they are treated as empty sets. The intersection of any set with an empty set is always an empty set.

## Error Handling

- `Wrong Type Error`:
  - `Error Message`: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
  - If any of the specified keys exist but are not of the set data type, DiceDB will return an error.
- `No Key Error`:
  - `Error Message`: `(error) ERR wrong number of arguments for 'sinter' command`
  - If no keys are provided, DiceDB will return an error.

## Example Usage

### Basic Intersection

```bash
# Add elements to sets
127.0.0.1:7379> SADD set1 "a" "b" "c"
(integer) 3
127.0.0.1:7379> SADD set2 "b" "c" "d"
(integer) 3
127.0.0.1:7379> SADD set3 "c" "d" "e"
(integer) 3

# Compute intersection
127.0.0.1:7379> SINTER set1 set2 set3
1) "c"
```

### Intersection with Non-Existent Set

```bash
# Add elements to sets
127.0.0.1:7379> SADD set1 "a" "b" "c"
(integer) 3
127.0.0.1:7379> SADD set2 "b" "c" "d"
(integer) 3

# Compute intersection with a non-existent set
127.0.0.1:7379> SINTER set1 set2 set3
(empty array)
```

Note: By default, non-existent keys (such as set3 in the example above) are treated like empty sets. There's no built-in way to create an empty set.

### Error Handling - Wrong Type

```bash
# Add elements to sets
127.0.0.1:7379> SADD set1 "a" "b" "c"
(integer) 3
# Create a string key
127.0.0.1:7379> SET stringKey "value"
OK

# Attempt to compute intersection with a non-set key
SINTER set1 stringKey
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

### Error Handling - No Keys Provided

```bash
# Attempt to compute intersection without providing any keys
127.0.0.1:7379> SINTER
(error) ERR wrong number of arguments for 'sinter' command
```
