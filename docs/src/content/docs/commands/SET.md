---
title: SET
description: The `SET` command in DiceDB is used to set the value of a key. If the key already holds a value, it is overwritten, regardless of its type. This is one of the most fundamental operations in DiceDB as it allows for both creating and updating key-value pairs.
---

The `SET` command in DiceDB is used to set the value of a key. If the key already holds a value, it is overwritten, regardless of its type. This is one of the most fundamental operations in DiceDB as it allows for both creating and updating key-value pairs.

## Syntax

```bash
SET key value [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL] [NX | XX]
```

## Parameters

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The name of the key to be set.                                            | String  | Yes      |
| `value`   | The value to be set for the key.                                          | String  | Yes      |
| `EX`      | Set the specified expire time, in seconds.                                | Integer | No       |
| `EXAT`    | Set the specified Unix time at which the key will expire, in seconds      | Integer | No       |
| `PX`      | Set the specified expire time, in milliseconds.                           | Integer | No       |
| `PXAT`    | Set the specified Unix time at which the key will expire, in milliseconds | Integer | No       |
| `NX`      | Only set the key if it does not already exist.                            | None    | No       |
| `XX`      | Only set the key if it already exists.                                    | None    | No       |
| `KEEPTTL` | Retain the time-to-live associated with the key.                          | None    | No       |

## Return values

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| Command is successful                          | `OK`                                              |
| `NX` or `XX` conditions are not met            | `nil`                                             |
| Syntax or specified constraints are invalid    | error                                             |

## Behaviour

- If the specified key already exists, the `SET` command will overwrite the existing key-value pair with the new value unless the `NX` option is provided.
- If the `NX` option is present, the command will set the key only if it does not already exist. If the key exists, no operation is performed and `nil` is returned.
- If the `XX` option is present, the command will set the key only if it already exists. If the key does not exist, no operation is performed and `nil` is returned.
- Using the `EX`, `EXAT`, `PX` or `PXAT` options together with `KEEPTTL` is not allowed and will result in an error.
- When provided, `EX` sets the expiry time in seconds and `PX` sets the expiry time in milliseconds.
- The `KEEPTTL` option ensures that the key's existing TTL is retained.

## Errors

1. `Wrong type of value or key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

2. `Invalid syntax or conflicting options`:

   - Error Message: `(error) ERR syntax error`
   - Occurs if the command's syntax is incorrect, such as incompatible options like `EX` and `KEEPTTL` used together, or a missing required parameter.

3. `Non-integer value for `EX`or`PX\`\`:

   - Error Message: `(error) ERR value is not an integer or out of range`
   - Occurs when the expiration time provided is not a valid integer.

## Example Usage

### Basic Usage

Setting a key `foo` with the value `bar`

```bash
127.0.0.1:7379> SET foo bar
OK
```

### Using expiration time (in seconds)

Setting a key `foo` with the value `bar` to expire in 10 seconds

```bash
127.0.0.1:7379> SET foo bar EX 10
OK
```

### Using expiration time (in milliseconds)

Setting a key `foo` with the value `bar` to expire in 10000 milliseconds (10 seconds)

```bash
127.0.0.1:7379> SET foo bar PX 10000
OK
```

### Setting only if key does not exist

Setting a key `foo` only if it does not already exist

```bash
127.0.0.1:7379> SET foo bar NX
```

`Response`:

- If the key does not exist: `OK`
- If the key exists: `nil`

### Setting only if key exists

Setting a key `foo` only if it exists

```bash
127.0.0.1:7379> SET foo bar XX
```

`Response`:

- If the key exists: `OK`
- If the key does not exist: `nil`

### Retaining existing TTL

Setting a key `foo` with a value `bar` and retaining existing TTL

```bash
127.0.0.1:7379> SET foo bar KEEPTTL
OK
```

### Invalid usage

Trying to set key `foo` with both `EX` and `KEEPTTL` will result in an error

```bash
127.0.0.1:7379> SET foo bar EX 10 KEEPTTL
(error) ERR syntax error
```
