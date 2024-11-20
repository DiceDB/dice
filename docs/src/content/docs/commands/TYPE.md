---
title: TYPE
description: The `TYPE` command in DiceDB returns the data type of the value stored at a given key. It's useful for inspecting the contents of a database and for debugging purposes. This command helps in determining how to interact with a specific key-value pair.
---

The `TYPE` command in DiceDB is used to determine the data type of the value stored at a specified key. This command is useful when you need to verify the type of data associated with a key before performing operations on it. It aids in debugging and helps ensure that the correct commands are used for different data types.

## Syntax

```bash
TYPE key
```

## Parameters

| Parameter | Description | Type | Required |
| --------- | ----------- | ---- | -------- |
| `key` | The key to check for its value type | String | Yes |

## Return values

| Condition | Return Value |
| --------- | ------------ |
| Key exists | The type of the value stored at the key (string, list, set, zset, hash, stream) |
| Key does not exist | "none" |

## Behaviour

- The TYPE command examines the value stored at the specified key and returns its data type.
- If the key does not exist, the command returns "none".
- The command does not modify the value or the key in any way; it's a read-only operation.
- The time complexity of this command is O(1), making it efficient for frequent use.

## Errors

1. `Wrong number of arguments`:
   - Error Message: `(error) ERR wrong number of arguments for 'type' command`
   - Occurs when the TYPE command is called without specifying a key or more than 1 arguement.

## Example Usage

### Basic Usage

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> TYPE mykey
string
```

### Checking Different Data Types

```bash
127.0.0.1:7379> LPUSH mylist "element"
(integer) 1
127.0.0.1:7379> TYPE mylist
list

127.0.0.1:7379> SADD myset "element"
(integer) 1
127.0.0.1:7379> TYPE myset
set

127.0.0.1:7379> ZADD myzset 1 "element"
(integer) 1
127.0.0.1:7379> TYPE myzset
zset

127.0.0.1:7379> HSET myhash field value
(integer) 1
127.0.0.1:7379> TYPE myhash
hash

127.0.0.1:7379> GEOADD mygeo 13.361389 38.115556 "Palermo"
(integer) 1
127.0.0.1:7379> TYPE mygeo
zset

127.0.0.1:7379> SET key1 foobar
OK
127.0.0.1:7379> SET key2 abcdef
OK
127.0.0.1:7379> BITOP AND dest key1 key2
(integer) 6
127.0.0.1:7379> TYPE dest
string

```

### Non-existent Key

```bash
127.0.0.1:7379> TYPE nonexistentkey
none
```

## Best Practices

- Use the TYPE command before performing operations on a key to ensure you are using the appropriate commands for the data type.
- Remember that TYPE returns "none" for non-existent keys, which can be useful for checking key existence without modifying data.

## Notes

- The TYPE command is particularly useful in debugging scenarios where you need to verify the structure of your data.
- While TYPE is efficient (O(1) complexity), avoid overusing it in high-performance scenarios where you already know the data types of your keys.
- The TYPE command can be used in combination with other commands to create more robust and type-safe operations in your applications.


