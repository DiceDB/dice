---
title: EXISTS
description: The `EXISTS` command in DiceDB is used to determine if one or more specified keys exist in the database. It returns the number of keys that exist among the specified ones.
---

The `EXISTS` command in DiceDB is used to determine if one or more specified keys exist in the database. It returns the number of keys that exist among the specified ones.

## Syntax
```
EXISTS key [key ...]
```

## Parameters

| Parameter | Description                                    | Type   | Required |
|-----------|------------------------------------------------|--------|----------|
| `key`     | The key(s) to check for existence. One or more keys can be specified, separated by spaces. | String | Yes      |

## Return values

| Condition                                      | Return Value                                      |
|------------------------------------------------|---------------------------------------------------|
| None of the specified keys exist               | `0`                                               |
| One or more specified keys exist               | Integer representing the count of keys that exist |

## Behaviour
- The `EXISTS` command checks whether the specified keys are present in the database.
- Returns 1 or 0, or for multiple keys returns the count of existing keys.
- The command performs a read-only operation and does not modify the database.

## Errors
1. `Wrong number of arguments`:
   - Error Message: `(error) ERR wrong number of arguments for 'exists' command`
   - Occurs when no key is provided.

2. `Wrong type of value or key`:
   - Error Message: `(error) WRONGTYPE Operation against a key holding the wrong kind of value`
   - Occurs when attempting to use the command on a key that contains a non-string value.

## Example Usage

### Single Key Check
Checking if a key `mykey` exists in the database:

```bash
127.0.0.1:7379> SET mykey "Hello"
OK
127.0.0.1:7379> EXISTS mykey
(integer) 1
```

### Multiple Keys Check
Checking if multiple keys (`key1`, `key2`, `key3`) exist in the database:

```bash
127.0.0.1:7379> SET key1 "value1"
OK
127.0.0.1:7379> SET key2 "value2"
OK
127.0.0.1:7379> EXISTS key1 key2 key3
(integer) 2
```
In this case, `key1` and `key2` exist, but `key3` does not.

### Non-Existent Key
Checking if a non-existent key (`nonExistentKey`) is present in the database:

```bash
127.0.0.1:7379> EXISTS nonExistentKey
(integer) 0
```
### All Non-Existent Keys

Checking if all non-existent keys return 0:

```bash
127.0.0.1:7379> EXISTS nonExistentKey1 nonExistentKey2
(integer) 0
```

### Empty Command

Providing no keys should trigger an error:

```bash
127.0.0.1:7379> EXISTS
(error) ERR wrong number of arguments for 'exists' command
```
