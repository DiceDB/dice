---
title: SET
description: The `SET` command in DiceDB is used to set the value of a key. If the key already holds a value, it is overwritten, regardless of its type. This is one of the most fundamental operations in DiceDB as it allows for both creating and updating key-value pairs.
---
<!-- description in 2 to 3 sentences, following is an example -->
The `SET` command in DiceDB is used to set the value of a key. If the key already holds a value, it is overwritten, regardless of its type. This is one of the most fundamental operations in DiceDB as it allows for both creating and updating key-value pairs.

## Syntax

```bash
SET key value [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL] [NX | XX]
```

## Parameters
<!-- add all parameters, small description, type and required -->

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The name of the key to be set.                                            | String  | Yes      |
|           |                                                                           |         |       |


## Return values
<!-- add all scenarios -->

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
|                                                |                                                   |

## Behaviour

- How does the command execute
- goes here
- Kind of explaining the underlying algorithm

## Errors
<!-- sample errors, please update for commands -->
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
<!-- examples here are for set, please update them for the command -->
Setting a key `foo` with the value `bar`

```bash
127.0.0.1:7379> SET foo bar
OK
```
<!-- Please use detailed scenarios and edges cases if possible -->
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
<!-- Optional: Used when additional information is to conveyed to users -->
### Notes