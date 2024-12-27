---
title: HGETALL
description: Documentation for the DiceDB command HGETALL
---

The `HGETALL` command in DiceDB is used to retrieve all the fields and values of a hash stored at a specified key. This command is particularly useful when you need to fetch the entire hash in one go, rather than fetching individual fields one by one.

## Syntax

```bash
HGETALL key
```

## Parameters

| Parameter | Description                                                                | Type   | Required |
| --------- | -------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key of the hash from which you want to retrieve all fields and values. | String | Yes      |

## Return values

| Condition                      | Return Value     |
| ------------------------------ | ---------------- |
| The `key` exists and is a hash | Array of strings |
| The `key` does not exist       | Empty array      |

## Behaviour

When the `HGETALL` command is executed:

1. DiceDB checks if the specified key exists.
2. If the key exists and is of type hash, DiceDB retrieves all the fields and their corresponding values.
3. If the key does not exist, DiceDB returns an empty array.
4. If the key exists but is not of type hash, an error is returned.

## Errors

The `HGETALL` command can raise the following errors:

1. `Non-hash type or wrong data type`:

   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs if the key holds a non-hash data structure.

2. `Incorrect Argument Count`:

   - Error Message: `(error) ERR wrong number of arguments for 'hgetall' command`
   - Occurs if the command is not provided with the correct number of arguments (i.e., more than one).

## Example Usage

### Retrieving all fields and values from an existing hash

```bash
127.0.0.1:7379> HSET user:1000 name "John Doe" age "30" country "USA"
(integer) 3
127.0.0.1:7379> HGETALL user:1000
1) "name"
2) "John Doe"
3) "age"
4) "30"
5) "country"
6) "USA"
```

### Retrieving from a non-existing key

```bash
127.0.0.1:7379> HGETALL user:2000
(empty array)
```

### Error Example

Key is not a hash

```bash
127.0.0.1:7379> SET user:3000 "This is a string"
OK
127.0.0.1:7379> HGETALL user:3000
(error) ERROR WRONGTYPE Operation against a key holding the wrong kind of value
```

Invalid number of arguments are passed

```bash
127.0.0.1:7379> HGETALL user:3000 helloworld
(error) ERROR wrong number of arguments for 'hgetall' command
```
