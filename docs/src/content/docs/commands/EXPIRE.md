---
title: EXPIRE
description: The `EXPIRE` command in DiceDB is used to set a timeout on a specified key. After the timeout has expired, the key will automatically be deleted. This command is useful for implementing time-based expiration of keys, ensuring that data does not persist longer than necessary.
---

The `EXPIRE` command in DiceDB is used to set a timeout on a specified key. After the timeout has expired, the key will automatically be deleted. This command is useful for implementing time-based expiration of keys, ensuring that data does not persist longer than necessary.

## Syntax

```
EXPIRE key seconds [NX | XX | GT | LT]
```

## Parameters

| Parameter | Description                                                               | Type    | Required |
|-----------|---------------------------------------------------------------------------|---------|----------|
| `key`     | The key to set the timeout on. Must be an existing key.                   | String  | Yes      |
| `seconds` | Timeout duration in seconds. Must be a positive integer.                  | Integer | Yes      |
| `NX`      | Set expiry only if the key does not already have an expiry.               | None    | No       |
| `XX`      | Set expiry only if the key already has an expiry.                         | None    | No       |
| `GT`      | Set expiry only if the new expiry is greater than the current one.        | None    | No       |
| `LT`      | Set expiry only if the new expiry is less than the current one.           | None    | No       |


## Return Value

The `EXPIRE` command returns an integer value:

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| Timeout was successfully set.                  | `1`                                              |
| Timeout was not set (e.g., key does not exist, or conditions not met).| `0`                                             |

## Behaviour

When the `EXPIRE` command is issued:

1. If the specified key exists, DiceDB sets a timeout on the key. The key will be automatically deleted after the specified number of seconds.
2. If the key does not exist, no timeout is set, and the command returns `0`.
3. If the key already has a timeout, the existing timeout is replaced with the new one specified by the `EXPIRE` command.

## Error Handling

The `EXPIRE` command can raise errors in the following scenarios:

1. `Syntax Error`: If the command is not used with the correct number of arguments, DiceDB will return a syntax error.
    - Error Message: `(error) ERROR wrong number of arguments for 'expire' command`
    - Returned if the command is issued with an incorrect number of arguments.

## Example Usage

### Setting a Timeout on a Key

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```
```bash
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
```

In this example, the key `mykey` is set with the value "Hello". The `EXPIRE` command sets a timeout of 10 seconds on `mykey`. After 10 seconds, `mykey` will be automatically deleted.

### Checking if Timeout was Set

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```
```bash
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
```
```bash
127.0.0.1:7379> TTL mykey
(integer) 10
```

The `TTL` command shows the remaining time to live for mykey, which is 10 seconds.

### Setting Expiry with Conditions (NX and XX)
```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```
```bash
127.0.0.1:7379> EXPIRE mykey 10 NX
(integer) 1
```
```bash
127.0.0.1:7379> EXPIRE mykey 20 XX
(integer) 1
```

The `NX` option sets the expiry only if there was no expiry set, and the `XX` option updates it because there was an existing expiry.

### Replacing an Existing Timeout


```bash
127.0.0.1:7379> SET mykey "Hello"
OK
```

```bash
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
```

```bash
127.0.0.1:7379> EXPIRE mykey 20
(integer) 1
```

```bash
127.0.0.1:7379> TTL mykey
(integer) 20
```

The initial `EXPIRE` command sets a timeout of 10 seconds. The subsequent `EXPIRE` command replaces the existing timeout with a new timeout of 20 seconds.

### Attempting to Set Timeout on a Non-Existent Key

```bash
127.0.0.1:7379> EXPIRE non_existent_key 10
(integer) 0
```

The command returns `0`, indicating that the key does not exist and no timeout was set.

## Error Handling Examples

### Syntax Error

```bash
127.0.0.1:7379> EXPIRE mykey
(error) ERROR wrong number of arguments for 'expire' command
```

This example shows a syntax error due to missing the `seconds` argument.

## Additional Notes

- The `EXPIRE` command is often used in conjunction with other commands like `SET`, `GET`, and `DEL` to manage the lifecycle of keys in DiceDB.
- The timeout can be removed by using the `PERSIST` command, which removes the expiration from the key.
- The `TTL` command can be used to check the remaining time to live of a key with an expiration.
