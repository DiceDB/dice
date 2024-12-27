---
title: SETEX
description: The SETEX command in DiceDB is used to set the value of a key and its expiration time in seconds. This command is atomic and is commonly used to create time-sensitive key-value pairs.
sidebar:
  badge:
    text: Deprecated
    variant: danger
---

The SETEX command in DiceDB is used to set the value of a key and its expiration time in seconds. This command is atomic and is commonly used to create time-sensitive key-value pairs.

## Syntax

```bash
SETEX key seconds value
```

## Parameters

| Parameter | Description                             | Type    | Required |
| --------- | --------------------------------------- | ------- | -------- |
| `key`     | The name of the key to be set.          | String  | Yes      |
| `seconds` | Expiration time for the key in seconds. | Integer | Yes      |
| `value`   | The value to be set for the key.        | Integer | No       |

## Return values

| Condition                                   | Return Value |
| ------------------------------------------- | ------------ |
| Command is successful                       | `OK`         |
| Syntax or specified constraints are invalid | error        |

## Behaviour

- The SETEX command sets the value of a key and specifies its expiration time in seconds.
- If the specified key already exists, the value is overwritten, and the new expiration time is set.
- If the key does not exist, it is created with the specified expiration time.
- If the provided expiration time is invalid or not an integer, the command will return an error.
- This command is equivalent to using SET key value EX seconds but provides a more concise and dedicated syntax.

## Errors

1. `Missing or invalid expiration time`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs if the expiration time is not a valid positive integer.

2. `Missing required arguments`:

   - Error Message: `(error) ERR wrong number of arguments for 'SETEX' command`
   - Occurs if any of the required arguments (key, seconds, or value) are not provided.

## Example Usage

### Basic Usage

Set a key `foo` with the value `bar` to expire in `10` seconds:

```bash
127.0.0.1:7379> SETEX foo 10 bar
OK
```

Set a key `foo` with the value `new_value`, overwriting the existing value and resetting the expiration time:

```bash
127.0.0.1:7379> SETEX foo 20 new_value
OK
```

### Invalid usage

Setting a key with an invalid expiration time will result in an error:

```bash
127.0.0.1:7379> SETEX foo -10 bar
(error) ERROR invalid expire time in 'setex' command
```

Attempting to use the command with missing arguments will result in an error:

```bash
127.0.0.1:7379> SETEX foo 10
(error) ERROR wrong number of arguments for 'setex' command
```

### Notes:

`SETEX` can be replaced via [`SET`](/commands/set) with `EX` option.
