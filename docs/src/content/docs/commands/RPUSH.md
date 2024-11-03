---
title: RPUSH
description: Documentation for the DiceDB command RPUSH
---

The `RPUSH` command is used in DiceDB to insert one or multiple values at the tail (right end) of a list. If the list does not exist, it is created as an empty list before performing the push operations. This command is useful for maintaining a list of elements where new elements are added to the end of the list.

## Syntax

```bash
RPUSH key value [value ...]
```

## Parameters

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The name of the list where the values will be inserted. If the list does not exist, a new list will be created.                                           | String  | Yes      |
| `value`   | One or more values to be inserted at the tail of the list. Multiple values can be specified, separated by spaces.                                          | String  | Yes      |


## Return Value

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| Command is successful                          | integer                                           |
| Syntax or specified constraints are invalid    | error                                             |

## Behaviour

- When the `RPUSH` command is executed, the specified values are inserted at the tail of the list identified by the given key. 
- If the key does not exist, a new list is created and the values are inserted.
- If the key exists but is not a list, an error is returned.
- The `RPUSH` operation is atomic. If multiple clients issue `RPUSH` commands concurrently, DiceDB ensures that the list remains consistent and the values are appended in the order the commands were received.


## Errors

1. `Non-List Key`: If the key exists but is not a list, DiceDB returns an error.
   - Error Message: `WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when trying to use the command on a key that does not contain a list.


2. `Invalid Syntax`: If the command is not used with the correct syntax, DiceDB returns a syntax error.
   - Error Message: `ERR wrong number of arguments for 'rpush' command`
   - Occurs if the command's syntax is incorrect, like wrong number of arguments.


## Example Usage

### Basic Usage

Inserting a value `hello` into `mylist`
```bash
127.0.0.1:7379> RPUSH mylist "hello"
(integer) 1
```

### Inserting Multiple Values

Inserting multiple values `world`, `foo`, `bar` into `mylist`
```bash
127.0.0.1:7379> RPUSH mylist "world" "foo" "bar"
(integer) 4
```

### Invalid Usage

Trying to insert value `val` into a non-list type `mykey`

```bash
127.0.0.1:7379> SET mykey "notalist"
127.0.0.1:7379> RPUSH mykey "val"
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```
Trying to run `rpush` with only 1 argument

```bash
127.0.0.1:7379> RPUSH mylist
(error) ERROR wrong number of arguments 
```


