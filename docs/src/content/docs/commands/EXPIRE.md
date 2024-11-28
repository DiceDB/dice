---
title: EXPIRE
description: The `EXPIRE` command in DiceDB is used to set a timeout on a specified key. After the timeout has expired, the key will automatically be deleted. This command is useful for implementing time-based expiration of keys, ensuring that data does not persist longer than necessary.
---

The `EXPIRE` command in DiceDB is used to set a timeout on a specified key. After the timeout has expired, the key will automatically be deleted. This command is useful for implementing time-based expiration of keys, ensuring that data does not persist longer than necessary.

## Syntax

```bash
EXPIRE key seconds [NX | XX | GT | LT]
```

## Parameters

| Parameter | Description                                                        | Type    | Required |
| --------- | ------------------------------------------------------------------ | ------- | -------- |
| `key`     | The key to set the timeout on. Must be an existing key.            | String  | Yes      |
| `seconds` | Timeout duration in seconds. Must be a positive integer.           | Integer | Yes      |
| `NX`      | Set expiry only if the key does not already have an expiry.        | None    | No       |
| `XX`      | Set expiry only if the key already has an expiry.                  | None    | No       |
| `GT`      | Set expiry only if the new expiry is greater than the current one. | None    | No       |
| `LT`      | Set expiry only if the new expiry is less than the current one.    | None    | No       |

## Return Values

| Condition                                                              | Return Value |
| ---------------------------------------------------------------------- | ------------ |
| Timeout was successfully set.                                          | `1`          |
| Timeout was not set (e.g., key does not exist, or conditions not met). | `0`          |

## Behaviour

- When a key exists, DiceDB sets a timeout on it. The key will be automatically deleted after the specified seconds.
- If the key doesn't exist, no timeout is set, and the command returns `0`.
- If the key already has a timeout, the existing timeout is replaced with the new one.
- Conditional flags (NX, XX, GT, LT) control when the expiry can be set based on existing timeouts.

## Errors

1. `Syntax Error`:
   - Error Message: `(error) ERROR wrong number of arguments for 'expire' command`
   - Returned if the command is issued with an incorrect number of arguments.

## Example Usage

### Basic Usage

This example demonstrates the fundamental usage of the EXPIRE command. First, we set a key with a value, then set it to expire in 10 seconds. The TTL command shows the remaining time to live.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
127.0.0.1:7379> TTL mykey
(integer) 10
```

### Using Conditional Flags

This example shows how to use the NX and XX flags to conditionally set expiration times. NX sets the expiry only when there isn't one, while XX updates an existing expiry.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRE mykey 10 NX
(integer) 1
127.0.0.1:7379> EXPIRE mykey 20 XX
(integer) 1
```

### Replacing an Existing Timeout

This example illustrates how EXPIRE can replace an existing timeout with a new value. The final TTL command confirms the updated expiration time.

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRE mykey 10
(integer) 1
127.0.0.1:7379> EXPIRE mykey 20
(integer) 1
127.0.0.1:7379> TTL mykey
(integer) 20
```

### Invalid Key

This example shows what happens when trying to set an expiration on a non-existent key. The command returns 0 to indicate failure.

```bash
127.0.0.1:7379> EXPIRE non_existent_key 10
(integer) 0
```

## Best Practices

- Use [`TTL`](/commands/ttl) command to check remaining time before expiration
- Consider using [`PERSIST`](/commands/persist) command to remove expiration if needed
- Choose appropriate conditional flags (NX, XX, GT, LT) based on your use case
- Ensure timeout values are appropriate for your application's needs

## Alternatives

- Use [`EXPIREAT`](/commands/expireat) command for more precise expiration control based on Unix timestamps