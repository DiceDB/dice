---
title: EXPIREAT
description: Documentation for the DiceDB command EXPIREAT
---

The `EXPIREAT` command is used to set the expiration time of a key in DiceDB. Unlike the `EXPIRE` command, which sets the expiration time in seconds from the current time, `EXPIREAT` sets the expiration time as an absolute Unix timestamp (in seconds). This allows for more precise control over when a key should expire.

## Syntax

```plaintext
EXPIREAT key timestamp
```

## Parameters

- `key`: The key for which the expiration time is to be set. This is a string.
- `timestamp`: The Unix timestamp (in seconds) at which the key should expire. This is an integer.

## Return Value

- `Integer reply`: Returns `1` if the timeout was set successfully, and `0` if the key does not exist or the timeout could not be set.

## Behaviour

When the `EXPIREAT` command is executed, DiceDB will set the expiration time of the specified key to the given Unix timestamp. If the key already has an expiration time, it will be overwritten with the new timestamp. If the key does not exist, the command will return `0` and no expiration time will be set.

## Error Handling

- `Wrong number of arguments`: If the command is called with an incorrect number of arguments, DiceDB will return an error message: `ERR wrong number of arguments for 'expireat' command`.
- `Invalid timestamp`: If the provided timestamp is not a valid integer, DiceDB will return an error message: `ERR value is not an integer or out of range`.
- `Non-string key`: If the key is not a string, DiceDB will return an error message: `WRONGTYPE Operation against a key holding the wrong kind of value`.

## Example Usage

### Setting an Expiration Time

```plaintext
SET mykey "Hello"
EXPIREAT mykey 1672531199
```

In this example, the key `mykey` is set to expire at the Unix timestamp `1672531199`.

### Checking the Expiration Time

```plaintext
TTL mykey
```

This command can be used to check the remaining time to live for the key `mykey`.

### Key Does Not Exist

```plaintext
EXPIREAT nonexistingkey 1672531199
```

In this example, since `nonexistingkey` does not exist, the command will return `0`.

## Additional Notes

- The `EXPIREAT` command is useful when you need to synchronize the expiration of keys across multiple DiceDB instances or when you need to set an expiration time based on an external event that provides a Unix timestamp.
- The timestamp should be in seconds. If you have a timestamp in milliseconds, you need to convert it to seconds before using it with `EXPIREAT`.

## Related Commands

- `EXPIRE`: Sets the expiration time of a key in seconds from the current time.
- `PEXPIREAT`: Sets the expiration time of a key as an absolute Unix timestamp in milliseconds.
- `TTL`: Returns the remaining time to live of a key in seconds.
- `PTTL`: Returns the remaining time to live of a key in milliseconds.

By understanding and using the `EXPIREAT` command, you can effectively manage the lifecycle of keys in your DiceDB database, ensuring that data is available only as long as it is needed.

