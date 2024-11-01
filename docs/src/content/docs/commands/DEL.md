---
title: DEL
description: The `DEL` command in DiceDB is used to remove one or more keys from the database. If a given key does not exist, it is ignored. This command is fundamental for data management in DiceDB, allowing for the deletion of key-value pairs. The command returns the number of keys that were removed.


---

The `DEL` command in DiceDB is used to remove one or more keys from the database. If a given key does not exist, it is ignored. This command is fundamental for data management in DiceDB, allowing for the deletion of key-value pairs. The command returns the number of keys that were removed.

## Syntax

```bash
DEL key [key ...]
```

## Parameters

| Parameter | Description                                      | Type   | Required |
|-----------|--------------------------------------------------|--------|----------|
| `key`     | The name of the key(s) to be deleted.            | String | Yes      |

## Return values

| Condition                           | Return Value                                      |
|-------------------------------------|---------------------------------------------------|
| Command is successful               | Integer (number of keys successfully deleted)     |
| No keys match the specified pattern | 0                                                 |
| Syntax or specified constraints are invalid | error                                     |

## Behaviour

When the `DEL` command is executed, DiceDB will attempt to remove the specified keys from the database. The command operates in the following manner:

1. `Key Existence Check`: For each key specified, DiceDB checks if the key exists in the database.
2. `Key Removal`: If a key exists, it is removed from the database along with its associated value, regardless of the value's type.
3. `Count Removal`: The command keeps a count of how many keys were successfully removed.
4. `Ignore Non-existent Keys`: If a specified key does not exist, it is simply ignored and does not affect the count of removed keys.
5. `Return Count`: The total count of removed keys is returned as the result of the command.

## Error Handling

The `DEL` command is generally robust and straightforward, but there are a few scenarios where errors might occur:

1. `Wrong Type of Argument`: If the command is provided with an argument that is not a valid key (e.g., a non-string type), DiceDB will raise a syntax error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'del' command`

2. `No Arguments Provided`: If no keys are provided to the `DEL` command, DiceDB will raise a syntax error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'del' command`


## Example Usage

### Basic Usage

Deleting a single key `foo`:

```bash
127.0.0.1:7379> DEL foo
(integer) 1
```

### Deleting Multiple Keys

Deleting multiple keys `foo`, `bar`, and `baz`:

```bash
127.0.0.1:7379> DEL foo bar baz
(integer) 2
```

In this example, if only `foo` and `bar` existed, the command would return 2, indicating that two keys were successfully deleted.

### Deleting Non-existent Keys

Attempting to delete a non-existent key:

```bash
127.0.0.1:7379> DEL nonexistentkey
(integer) 0
```

### Complex Example

Setting multiple keys and then deleting them:

```bash
127.0.0.1:7379> SET key1 "value1"
OK
127.0.0.1:7379> SET key2 "value2"
OK
127.0.0.1:7379> SET key3 "value3"
OK
127.0.0.1:7379> DEL key1 key2 key4
(integer) 2
```

In this example:
- Three keys are set: `key1`, `key2`, and `key3`.
- The `DEL` command attempts to delete `key1`, `key2`, and `key4`.
- `key1` and `key2` are successfully deleted.
- `key4` doesn't exist, so it's ignored.
- The command returns 2, indicating two keys were deleted.

### Error Example

Calling `DEL` without any arguments:

```bash
127.0.0.1:7379> DEL
(error) ERR wrong number of arguments for 'del' command
```


