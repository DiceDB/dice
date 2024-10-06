---
title: EXPIREAT
description: Documentation for the DiceDB command EXPIREAT
---

The `EXPIREAT` command is used to set the expiration time of a key in DiceDB. Unlike the `EXPIRE` command, which sets the expiration time in seconds from the current time, `EXPIREAT` sets the expiration time as an absolute Unix timestamp (in seconds). This allows for more precise control over when a key should expire.

## Syntax

```
EXPIREAT key timestamp
```

## Parameters

| Parameter           | Description                                                                                     | Type    | Required |
| ------------------- | ----------------------------------------------------------------------------------------------- | ------- | -------- |
| `key`               | The name of the key to be set.                                                                  | String  | Yes      |
| `unix-time-seconds` | The unix-time-seconds timestamp at which the key should expire. This is an integer.             | Integer | Yes      |
| NX                  | Set the expiration only if the key does not already have an expiration time.                    | None    | No       |
| XX                  | Set the expiration only if the key already has an expiration time.                              | None    | No       |
| GT                  | Set the expiration only if the new expiration time is greater than or equal to the current one. | None    | No       |
| LT                  | Set the expiration only if the new expiration time is less than the current one.                | None    | No       |


## Return Value

| Condition                                          | Return Value |
| -------------------------------------------------- | ------------ |
| Command is successful                              | `1`          |
| Key does not exist or the timeout could not be set | `0`          |
| Syntax or specified constraints are invalid        | error        |

## Behaviour

- When the `EXPIREAT` command is executed, DiceDB will set the expiration time of the specified key to the given Unix timestamp.
- If the key already has an expiration time, it will be overwritten with the new timestamp.
- If the key does not exist, the command will return `0` and no expiration time will be set.
- 

## Error Handling

1. `Wrong number of arguments`: 
   - Error Message: `(error) ERROR wrong number of arguments for 'expireat' command`
   - Occurs if the command is called with too little arguments
2. `Invalid timestamp`: 
   - Error Message: `(error) ERROR value is not an integer or out of range`
   - If the provided timestamp is not a valid integer
3. `Invalid format of Unix Timestamp`: 
   - Error Message: `(error) ERROR invalid expire time in 'expireat' command`
   - If the expiry time is not in the allowed range.

## Example Usage

### Setting an Expiration Time

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIREAT mykey 17282126871
(integer) 1
```

In this example, the key `mykey` is set to expire at the Unix timestamp `17282126871`.

### Checking the expiration time

This will return the remaining time to live for the key `mykey` in seconds.

```bash
127.0.0.1:7379> EXPIREAT mykey 17282126871
(integer) 1
127.0.0.1:7379> TTL mykey
(integer) 15553913293
```

### Key does not exist

This will return `0` as the key does not exist.

```bash
127.0.0.1:7379> EXPIREAT nonexistingkey 17282126871
(integer) 0
```

### Setting an EXPIRYTIME only if not exists

```bash
127.0.0.1:7379> set key value
OK
127.0.0.1:7379> EXPIREAT key 1728212987222 NX
(integer) 1
127.0.0.1:7379> EXPIREAT key 1728212987222 NX
(integer) 0
```

### Setting an EXPIRYTIME only if it already has one

```bash
127.0.0.1:7379> set key value
OK                                                                           
127.0.0.1:7379> EXPIREAT key 12345677777 XX
(integer) 0
127.0.0.1:7379> EXPIREAT key 12345677777
(integer) 1
127.0.0.1:7379> EXPIREAT key 123456777722 XX 
(integer) 1
```

### Setting an EXPIRYTIME only if the new expiry time is greater than or equal to the current one

```bash 
127.0.0.1:7379> set key value
OK                             
127.0.0.1:7379> expireat key 12334444444
(integer) 1
127.0.0.1:7379> expireat key 12334444424 GT 
(integer) 0
127.0.0.1:7379> expireat key 12334444524 GT 
(integer) 1
```

### Setting an EXPIRYTIME only if the new expiry time is less than or equal to the current one

```bash
127.0.0.1:7379> set key value
OK          
127.0.0.1:7379> expireat key 12334444444
(integer) 1
127.0.0.1:7379> expireat key 12334444444 LT
(integer) 1
127.0.0.1:7379> expireat key 12334444445 LT
(integer) 0
127.0.0.1:7379> expireat key 12334444442 LT
(integer) 1
```

## Additional Notes

- The `EXPIREAT` command is useful when you need to synchronize the expiration of keys across multiple DiceDB instances or when you need to set an expiration time based on an external event that provides a Unix timestamp.
- The timestamp should be in seconds. If you have a timestamp in milliseconds, you need to convert it to seconds before using it with `EXPIREAT`.
- There is an arbritrary limit to the size of the `unix-time-seconds` of [9223372036854775](https://github.com/DiceDB/dice/blob/b74dc8ffd5e518eaa9b82020d2b25a592c6472d4/internal/eval/eval.go#L69).

## Related Commands

- `EXPIRE`: Sets the expiration time of a key in seconds from the current time.
- `PEXPIREAT`: Sets the expiration time of a key as an absolute Unix timestamp in milliseconds.
- `TTL`: Returns the remaining time to live of a key in seconds.
- `PTTL`: Returns the remaining time to live of a key in milliseconds.

By understanding and using the `EXPIREAT` command, you can effectively manage the lifecycle of keys in your DiceDB database, ensuring that data is available only as long as it is needed.