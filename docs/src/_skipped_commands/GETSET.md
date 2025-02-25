---
title: GETSET
description: Documentation for the DiceDB command GETSET
sidebar:
  badge:
    text: Deprecated
    variant: danger
---

The `GETSET` command in DiceDB is a powerful atomic operation that combines the functionality of `GET` and `SET` commands. It retrieves the current value of a key and simultaneously sets a new value for that key. This command is particularly useful when you need to update a value and also need to know the previous value in a single atomic operation.

## Syntax

```bash
GETSET key value
```

## Parameters

| Parameter | Description                                       | Type   | Required |
| --------- | ------------------------------------------------- | ------ | -------- |
| `key`     | The key whose value you want to retrieve and set. | String | Yes      |
| `value`   | The new value to set for the specified key.       | String | Yes      |

## Return values

| Condition                                    | Return Value   |
| -------------------------------------------- | -------------- |
| The old value stored at the specifiied `key` | A string value |
| The key does not exist                       | `nil`          |

## Behaviour

When the `GETSET` command is executed, the following sequence of actions occurs:

1. The current value of the specified key is retrieved.
2. The specified key is updated with the new value.
3. If the specified key had an existing `TTL` , it is reset.
4. The old value is returned to the client.

This operation is atomic, meaning that no other commands can be executed on the key between the get and set operations.

## Errors

The `GETSET` command can raise errors in the following scenarios:

1. `Wrong Type Error`: If the key exists but is not a string, DiceDB will return an error.
   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
2. `Syntax Error`: If the command is not provided with exactly two arguments (key and value), DiceDB will return a syntax error.
   - Error Message: `(error) ERROR wrong number of arguments for 'getset' command`

## Examples

### Basic Example

```bash
127.0.0.1:7379> SET mykey "Hello"
127.0.0.1:7379> GETSET mykey "World"
"Hello"
```

- The initial value of `mykey` is set to "Hello".
- The `GETSET` command retrieves the current value "Hello" and sets the new value "World".
- The old value "Hello" is returned.

### Example with Non-Existent Key

```bash
127.0.0.1:7379> GETSET newkey "NewValue"
(nil)
```

- The key `newkey` does not exist.
- The `GETSET` command sets the value of `newkey` to "NewValue".
- Since the key did not exist before, `nil` is returned.

### Example with Key having pre-existing TTL

```bash
127.0.0.1:7379> SET newkey "test"
OK
127.0.0.1:7379> EXPIRE newkey 60
1
127.0.0.1:7379> TTL newkey
55
127.0.0.1:7379> GETSET newkey "new value"
"test"
127.0.0.1:7379> TTL newkey
(integer) -1
```

- The `newkey` used in the `GETSET` command had an existing `TTL` set to expire in 60 seconds
- When `GETSET` is executed on the mentioned key, it updates the value and resets the `TTL` on the key.
- Hence, the `TTL` on `newkey` post `GETSET` returns `-1` , suggesting that the key exists without any `TTL` configured

### Wrong Type

```bash
127.0.0.1:7379> LPUSH mylist "item"
127.0.0.1:7379> GETSET mylist "NewValue"
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

- The key `mylist` is a list, not a string.
- The `GETSET` command cannot operate on a list and returns a `WRONGTYPE` error.

### Syntax Error

```bash
127.0.0.1:7379> GETSET mykey
(error) ERROR wrong number of arguments for 'getset' command
```

- The `GETSET` command requires exactly two arguments: a key and a value.
- Since only one argument is provided, DiceDB returns a syntax error.
