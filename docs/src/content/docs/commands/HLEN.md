---
title: HLEN
description: Documentation for the DiceDB command HLEN
---

The `HLEN` command in DiceDB is used to obtain the number of fields contained within a hash stored at a specified key. This command is particularly useful for understanding the size of a hash and for performing operations that depend on the number of fields in a hash.

## Syntax

```
HLEN key
```

## Parameters

- `key`: The key associated with the hash for which the number of fields is to be retrieved. This parameter is mandatory.

## Return Value

The `HLEN` command returns an integer representing the number of fields in the hash stored at the specified key. If the key does not exist, it returns `0`.

## Behaviour

When the `HLEN` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists and is associated with a hash, DiceDB counts the number of fields in the hash and returns this count.
3. If the key does not exist, DiceDB returns `0`.
4. If the key exists but is not associated with a hash, an error is returned.

## Error Handling

The `HLEN` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the specified key exists but is not associated with a hash. For example, if the key is associated with a string, list, set, or any other data type, this error will be raised.

## Example Usage

### Example 1: Basic Usage

```DiceDB
> HSET myhash field1 "value1" field2 "value2"
(integer) 2

> HLEN myhash
(integer) 2
```

In this example, the hash `myhash` is created with two fields, `field1` and `field2`. The `HLEN` command returns `2`, indicating that there are two fields in the hash.

### Example 2: Non-Existent Key

```DiceDB
> HLEN nonExistentHash
(integer) 0
```

In this example, the key `nonExistentHash` does not exist. The `HLEN` command returns `0`, indicating that there are no fields in the hash because the hash itself does not exist.

### Example 3: Key with Wrong Data Type

```DiceDB
> SET mystring "This is a string"
OK

> HLEN mystring
(error) WRONGTYPE Operation against a key holding the wrong kind of value
```

In this example, the key `mystring` is associated with a string value. When the `HLEN` command is executed on this key, it raises an error because the key is not associated with a hash.

## Additional Notes

- The `HLEN` command is a constant-time operation, meaning its execution time is O(1) regardless of the number of fields in the hash.
- This command is useful for quickly determining the size of a hash without needing to retrieve all the fields and values.

By understanding the `HLEN` command, you can efficiently manage and interact with hash data structures in DiceDB, ensuring that your applications can handle hash-based data effectively.

