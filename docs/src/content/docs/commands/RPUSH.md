---
title: RPUSH
description: Documentation for the DiceDB command RPUSH
---

The `RPUSH` command is used in DiceDB to insert one or multiple values at the tail (right end) of a list. If the list does not exist, it is created as an empty list before performing the push operations. This command is useful for maintaining a list of elements where new elements are added to the end of the list.

## Syntax

```
RPUSH key value [value ...]
```

## Parameters

- `key`: The name of the list where the values will be inserted. If the list does not exist, a new list will be created.
- `value`: One or more values to be inserted at the tail of the list. Multiple values can be specified, separated by spaces.

## Return Value

- `Integer`: The length of the list after the push operation.

## Behaviour

When the `RPUSH` command is executed, the specified values are inserted at the tail of the list identified by the given key. If the key does not exist, a new list is created and the values are inserted. If the key exists but is not a list, an error is returned.

## Error Handling

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error is returned if the key exists but is not associated with a list data type.

## Example Usage

### Example 1: Basic Usage

```DiceDB
RPUSH mylist "hello"
```

`Description`: Inserts the value "hello" at the tail of the list `mylist`. If `mylist` does not exist, it is created.

`Return Value`: `1` (since the list now contains one element)

### Example 2: Inserting Multiple Values

```DiceDB
RPUSH mylist "world" "foo" "bar"
```

`Description`: Inserts the values "world", "foo", and "bar" at the tail of the list `mylist`.

`Return Value`: `4` (since the list now contains four elements: "hello", "world", "foo", "bar")

### Example 3: Handling Non-List Key

```DiceDB
SET mykey "notalist"
RPUSH mykey "value"
```

`Description`: Attempts to insert the value "value" at the tail of `mykey`, which is not a list but a string.

`Return Value`: Error `WRONGTYPE Operation against a key holding the wrong kind of value`

## Detailed Behaviour

1. `List Creation`: If the specified key does not exist, DiceDB creates a new list and then inserts the values.
2. `Appending Values`: The values are appended to the tail of the list in the order they are specified.
3. `Atomicity`: The `RPUSH` operation is atomic. If multiple clients issue `RPUSH` commands concurrently, DiceDB ensures that the list remains consistent and the values are appended in the order the commands were received.

## Error Handling

### Error Scenarios

1. `Non-List Key`: If the key exists but is not a list, DiceDB returns an error.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`
   - `Example`:
     ```DiceDB
     SET mykey "notalist"
     RPUSH mykey "value"
     ```
     `Result`: `WRONGTYPE Operation against a key holding the wrong kind of value`

2. `Invalid Syntax`: If the command is not used with the correct syntax, DiceDB returns a syntax error.

   - `Error Message`: `ERR wrong number of arguments for 'rpush' command`
   - `Example`:
     ```DiceDB
     RPUSH mylist
     ```
     `Result`: `ERR wrong number of arguments for 'rpush' command`

## Conclusion

The `RPUSH` command is a powerful and flexible way to manage lists in DiceDB, allowing for efficient insertion of elements at the tail of a list. Proper understanding of its parameters, return values, and error handling mechanisms ensures effective utilization of this command in various DiceDB-based applications.

