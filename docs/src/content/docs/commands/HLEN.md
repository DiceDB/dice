---
title: HLEN
description: Documentation for the DiceDB command HLEN
---

The `HLEN` command in DiceDB is used to obtain the number of fields contained within a hash stored at a specified key. This command is particularly useful for understanding the size of a hash and for performing operations that depend on the number of fields in a hash.

## Syntax

```bash
HLEN key
```

## Parameters

| Parameter | Description                                                                        | Type   | Required |
| --------- | ---------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key associated with the hash for which the number of fields is to be retrieved | String | Yes      |

## Return Value

| Condition               | Return Value                        |
| ----------------------- | ----------------------------------- |
| If specified key exists | number of fields in the hash at key |
| If key doesn't exist    | `0`                                 |

## Behaviour

- DiceDB checks if the specified key exists.
- If the key exists and is associated with a hash, DiceDB counts the number of fields in the hash and returns this count.
- If the key does not exist, DiceDB returns `0`.
- If the key exists but is not associated with a hash, an error is returned.

## Errors

1. `Wrong type of key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-hash value.

2. `Wrong number of arguments`:

   - Error Message: `(error) -ERR wrong number of arguments for 'HLEN' command`
   - Occurs if key isn't specified in the command.

## Example Usage

### Basic Usage

Creating hash `myhash` with two fields `field1` and `field2`. Getting hash length of `myhash`.

```bash
127.0.0.1:7379> HSET myhash field1 "value1" field2 "value2"
(integer) 2

127.0.0.1:7379> HLEN myhash
(integer) 2
```

### Invalid Usage on non-existent key

Getting hash length from a non-existent hash key `nonExistentHash`.

```bash
127.0.0.1:7379> HLEN nonExistentHash
(integer) 0
```

### Invalid Usage on non-hash key

Getting hash length from a key `mystring` associated with a non-hash type.

```bash
127.0.0.1:7379> SET mystring "This is a string"
OK

127.0.0.1:7379> HLEN mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `HLEN` command is a constant-time operation, meaning its execution time is O(1) regardless of the number of fields in the hash.
- This command is useful for quickly determining the size of a hash without needing to retrieve all the fields and values.

By understanding the `HLEN` command, you can efficiently manage and interact with hash data structures in DiceDB, ensuring that your applications can handle hash-based data effectively.
