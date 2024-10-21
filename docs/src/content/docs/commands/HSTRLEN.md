---
title: HSTRLEN
description: Documentation for the DiceDB command HSTRLEN
---

The `HSTRLEN` command in DiceDB is used to obtain the string length of value associated with field in the hash stored at a specified key. 

## Syntax

```
HSTRLEN key field
```

## Parameters

| Parameter       | Description                                      | Type    | Required |
|-----------------|--------------------------------------------------|---------|----------|
| `key`           | The key of the hash, which consists of the field whose string length you want to obtain  | String  | Yes      |
| `field`         | The field present in the hash whose length you want to obtain      | String  | Yes      |

## Return Value

The `HSTRLEN` command returns an integer representing the string length of fields in the hash stored at the specified key. If the key does not exist, it returns `0`.

## Behaviour

When the `HSTRLEN` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists, is associated with a hash and specified field exists in the hash, DiceDB returns the string length of value associated with specified field in the hash.
3. If the key does not exist, DiceDB returns `0`.
4. If the key exists and specified field does not exist in the key, DiceDB returns `0`.

## Error Handling

The `HSTRLEN` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the specified key exists but is not associated with a hash. For example, if the key is associated with a string, list, set, or any other data type, this error will be raised.
- `-ERR wrong number of arguments for 'hstrlen' command`: This error occurs if key or field value isn't specified in the command.

## Example Usage

### Example 1: Basic Usage

```DiceDB
> HSET myhash field1 "helloworld" field2 "value2"
(integer) 1

> HSTRLEN myhash field1
(integer) 10
```

In this example, the hash `myhash` is created with two fields, `field1` and `field2`. The `HSTRLEN` command returns `10`, indicating that string length of `helloworld`.

### Example 2: Non-Existent Key

```DiceDB
> HSTRLEN nonExistentHash field1
(integer) 0
```

In this example, the key `nonExistentHash` does not exist. The `HSTRLEN` command returns `0`, indicating that `field1` does not exist in the specified hash.

### Example 3: Key with Wrong Data Type

```DiceDB
> SET mystring "This is a string"
OK

> HSTRLEN mystring field1
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key `mystring` is associated with a string value. When the `HSTRLEN` command is executed on this key, it raises an error because the key is not associated with a hash.

## Additional Notes

- The `HSTRLEN` command has a constant-time operation, meaning its execution time is O(1), regardless of the number of fields in the hash.
