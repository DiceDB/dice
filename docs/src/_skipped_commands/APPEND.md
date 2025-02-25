---
title: APPEND
description: The `APPEND` command in DiceDB is used to either set the value of a key or append a value to an existing key. This command allows for both creating and updating key-value pairs.
---

The `APPEND` command in DiceDB is used to either set the value of a key or append a value to an existing key and returns the length of the value stored at the specified key after appending. This command allows for both creating and updating key-value pairs.

## Syntax

```bash
APPEND key value
```

## Parameters

| Parameter | Description                      | Type   | Required |
| --------- | -------------------------------- | ------ | -------- |
| `key`     | The name of the key to be set.   | String | Yes      |
| `value`   | The value to be set for the key. | String | Yes      |

## Return values

| Condition                  | Return Value          |
| -------------------------- | --------------------- |
| if key is set successfully | length of the string. |

## Behaviour

- If the specified key does not exist, the `APPEND` command will create a new key-value pair.
- If the specified key already exists, the `APPEND` command will append the value to the existing value of the key.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid syntax or conflicting options`:

   - Error Message: `(error) ERROR wrong number of arguments for 'append' command`
   - If the number of arguments are not exactly equal to 2.

## Example Usage

### Basic Usage

Setting a key `foo` with the value `bar` using `APPEND`

```bash
127.0.0.1:7379> APPEND foo bar
(integer) 3
```

Appending to key `foo` that contains `bar` with `baz`

```bash
127.0.0.1:7379> SET foo bar
OK
127.0.0.1:7379> APPEND foo baz
(integer) 6
127.0.0.1:7379> GET foo
"barbaz"
```

Appending "1" to key `bmkey` that contains a bitmap equivalent of `42`

```bash
127.0.0.1:7379> SETBIT bmkey 2 1
(integer) 0
127.0.0.1:7379> SETBIT bmkey 3 1
(integer) 0
127.0.0.1:7379> SETBIT bmkey 5 1
(integer) 0
127.0.0.1:7379> SETBIT bmkey 10 1
(integer) 0
127.0.0.1:7379> SETBIT bmkey 11 1
(integer) 0
127.0.0.1:7379> SETBIT bmkey 14 1
(integer) 0
127.0.0.1:7379> GET bmkey
"42"
127.0.0.1:7379> APPEND bmkey 1
(integer) 3
127.0.0.1:7379> GET bmkey
"421"
```

### Invalid usage

Trying to use `APPEND` without giving the value

```bash
127.0.0.1:7379> APPEND foo
(error) ERROR wrong number of arguments for 'append' command
```

Trying to use `APPEND` on a invalid data type.

```bash
127.0.0.1:7379> LPUSH foo bar
127.0.0.1:7379> APPEND foo baz
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```
