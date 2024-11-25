---
title: EXPIRETIME
description: The `EXPIRETIME` command in DiceDB is used to retrieve the absolute Unix timestamp (in seconds) at which a given key will expire. This command is particularly useful for understanding the exact expiration time of a key, which can help in debugging and managing key lifetimes.
---

The `EXPIRETIME` command in DiceDB is used to retrieve the absolute Unix timestamp (in seconds) at which a given key will expire. This command is particularly useful for understanding the exact expiration time of a key, which can help in debugging and managing key lifetimes.

## Syntax

```bash
EXPIRETIME key
```

## Parameters

| Parameter | Description                                                  | Type   | Required |
| --------- | ------------------------------------------------------------ | ------ | -------- |
| `key`     | The name of the key whose expiration time is to be retrieved | String | Yes      |

## Return Values

| Condition                                 | Return Value                |
| ----------------------------------------- | --------------------------- |
| The key exists and has an expiration time | Unix timestamp (in seconds) |
| The key exists but has no expiration time | -1                          |
| The key does not exist                    | -2                          |

## Behaviour

- DiceDB checks if the specified key exists in the database
- If the key exists and has an associated expiration time, DiceDB returns the absolute Unix timestamp (in seconds)
- If the key exists without an expiration time, the command returns `-1`
- If the key doesn't exist, the command returns `-2`

## Errors

1. `Wrong number of arguments`:
   - Error Message: `(error) ERROR wrong number of arguments for 'expiretime' command`
   - Occurs when the command is called with an incorrect number of arguments

## Example Usage

### Key with Expiration Time

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRE mykey 60
(integer) 1
127.0.0.1:7379> EXPIRETIME mykey
(integer) 1728548993
```

In this example, the key `mykey` is set with a value "Hello" and an expiration time of 60 seconds. The `EXPIRETIME` command returns the Unix timestamp at which `mykey` will expire.

### Key without Expiration Time

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXPIRETIME mykey
(integer) -1
```

In this example, the key `mykey` is set with a value "Hello" but no expiration time is set. The `EXPIRETIME` command returns `-1` indicating that the key does not have an associated expiration time.

### Non-Existent Key

```bash
127.0.0.1:7379> EXPIRETIME nonExistentKey
(integer) -2
```

In this example, the key `nonExistentKey` does not exist in the database. The `EXPIRETIME` command returns `-2` indicating that the key does not exist.

## Alternatives

- Use [`TTL`](/commands/ttl) to get relative expiration times
- Use [`PTTL`](/commands/pttl) to get relative expiration times in milliseconds