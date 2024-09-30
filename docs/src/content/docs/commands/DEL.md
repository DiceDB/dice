---
title: DEL
description: The `DEL` command in DiceDB is used to remove one or more keys from the database. If a given key does not exist, it is ignored. The command returns the number of keys that were removed.
---

The `DEL` command in DiceDB is used to remove one or more keys from the database. If a given key does not exist, it is ignored. The command returns the number of keys that were removed.

## Syntax

```plaintext
DEL key [key ...]
```

## Parameters

- `key`: The key(s) to be removed from the database. Multiple keys can be specified, separated by spaces.

## Return Value

The `DEL` command returns an integer representing the number of keys that were removed.

- `Type`: Integer
- `Description`: The number of keys that were successfully removed.

## Behaviour

When the `DEL` command is executed, DiceDB will attempt to remove the specified keys from the database. The command operates in the following manner:

1. `Key Existence Check`: For each key specified, DiceDB checks if the key exists in the database.
2. `Key Removal`: If a key exists, it is removed from the database.
3. `Count Removal`: The command keeps a count of how many keys were successfully removed.
4. `Return Count`: The total count of removed keys is returned as the result of the command.

## Error Handling

The `DEL` command is generally robust and straightforward, but there are a few scenarios where errors might occur:

1. `Wrong Type of Argument`: If the command is provided with an argument that is not a valid key (e.g., a non-string type), DiceDB will raise a syntax error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'del' command`

2. `No Arguments Provided`: If no keys are provided to the `DEL` command, DiceDB will raise a syntax error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'del' command`

## Example Usage

### Single Key Deletion

```bash
127.0.0.1:7379> DEL mykey
```

`Description`: This command will attempt to delete the key `mykey` from the database. If `mykey` exists, it will be removed, and the command will return `1`. If `mykey` does not exist, the command will return `0`.

### Multiple Keys Deletion

```bash
127.0.0.1:7379> DEL key1 key2 key3
```

`Description`: This command will attempt to delete the keys `key1`, `key2`, and `key3` from the database. The return value will be the number of keys that were successfully removed. For example, if `key1` and `key3` exist but `key2` does not, the command will return `2`.

### Example with Return Values

```bash
127.0.0.1:7379> SET key1 "value1"
127.0.0.1:7379> SET key2 "value2"
127.0.0.1:7379> SET key3 "value3"
127.0.0.1:7379> DEL key1 key2 key4
```

`Description`: In this example:

- `key1`, `key2`, and `key3` are set with some values.
- The `DEL` command attempts to delete `key1`, `key2`, and `key4`.
- Since `key1` and `key2` exist, they will be removed.
- `key4` does not exist, so it will be ignored.
- The command will return `2`, indicating that two keys were removed.
