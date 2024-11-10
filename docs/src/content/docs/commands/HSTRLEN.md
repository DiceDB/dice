---
title: HSTRLEN
description: Documentation for the DiceDB command HSTRLEN
---

The `HSTRLEN` command in DiceDB is used to obtain the string length of value associated with field in the hash stored at a specified key.

## Syntax

```bash
HSTRLEN key field
```

## Parameters

| Parameter | Description                                                                             | Type   | Required |
| --------- | --------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the hash, which consists of the field whose string length you want to obtain | String | Yes      |
| `field`   | The field present in the hash whose length you want to obtain                           | String | Yes      |

## Return Value

| Condition                         | Return Value  |
| --------------------------------- | ------------- |
| If specified key and field exists | string length |
| If key doesn't exist              | `0`           |
| If field doesn't exist            | `0`           |

## Behaviour

- DiceDB checks if the specified key exists.
- If the key exists, is associated with a hash and specified field exists in the hash, DiceDB returns the string length of value associated with specified field in the hash.
- If the key does not exist, DiceDB returns `0`.
- If the key exists and specified field does not exist in the key, DiceDB returns `0`.

## Errors

1. `Wrong type of key`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-hash value.

2. `Wrong number of arguments`:

   - Error Message: `(error) -ERR wrong number of arguments for 'hstrlen' command`
   - Occurs if key or field isn't specified in the command.

## Example Usage

### Basic Usage

Creating hash `myhash` with two fields `field1` and `field2`. Getting string length of value in `field1`.

```bash
127.0.0.1:7379> HSET myhash field1 "helloworld" field2 "value2"
(integer) 1

127.0.0.1:7379> HSTRLEN myhash field1
(integer) 10
```

### Invalid Usage on non-existent key

Getting string length from a non-existent key `nonExistentHash`.

```bash
127.0.0.1:7379> HSTRLEN nonExistentHash field1
(integer) 0
```

### Invalid Usage on non-hash key

Getting string length from a key `mystring` associated with a non-hash type.

```bash
127.0.0.1:7379> SET mystring "This is a string"
OK

127.0.0.1:7379> HSTRLEN mystring field1
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

## Notes

- The `HSTRLEN` command has a constant-time operation, meaning its execution time is O(1), regardless of the number of fields in the hash.
