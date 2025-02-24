---
title: KEYS
description: Documentation for the DiceDB command KEYS
---

The `KEYS` command in DiceDB is used to find all keys matching a given pattern. This command is useful for retrieving keys that match a specific pattern, which can be helpful for debugging or managing your DiceDB data.

## Syntax

```bash
KEYS pattern
```

## Parameters

| Parameter | Description                                                                     | Type   | Required |
| --------- | ------------------------------------------------------------------------------- | ------ | -------- |
| `pattern` | A string representing the pattern to match against the keys in DiceDB database. | String | Yes      |

Supported glob-style characters:

- `*` matches any number of characters (including zero).
- `?` matches exactly one character.
- `[abc]` matches any one of the characters inside the brackets.
- `[a-z]` matches any character in the specified range.

Use `\` to escape special characters if you want to match them verbatim.

## Return values

| Condition                                   | Return Value                                                                   |
| ------------------------------------------- | ------------------------------------------------------------------------------ |
| Command is successful                       | Array of strings, where each string is key that matches the specified pattern. |
| Command is successful but key not found     | `(empty list or set)`                                                          |
| Syntax or specified constraints are invalid | error                                                                          |

## Example usage

### Basic example

```bash
127.0.0.1:7379> SET key1 "value1"
OK
127.0.0.1:7379> SET key2 "value2"
OK
127.0.0.1:7379> SET anotherkey "value3"
OK
127.0.0.1:7379> KEYS key*
1) "key1"
2) "key2"
```

### Using wildcards

```bash
127.0.0.1:7379> SET key1 "value1"
OK
127.0.0.1:7379> SET key2 "value2"
OK
127.0.0.1:7379> SET key3 "value3"
OK
127.0.0.1:7379> KEYS key?
1) "key3"
2) "key1"
3) "key2"
```

### Using character ranges

```bash
127.0.0.1:7379> SET key1 "value1"
OK
127.0.0.1:7379> SET key2 "value2"
OK
127.0.0.1:7379> SET key3 "value3"
OK
127.0.0.1:7379> KEYS key[1-2]
1) "key1"
2) "key2"
```

### Using \ to escape special characters

```bash
127.0.0.1:7379> SET key1 "value1"
OK
127.0.0.1:7379> SET key2 "value2"
OK
127.0.0.1:7379> SET key3 "value3"
OK
127.0.0.1:7379> KEYS key?
1) "key3"
2) "key*"
3) "key?"
127.0.0.1:7379> KEYS key\?
1) "key?"
```

## Behaviour

When the `KEYS` command is executed, DiceDB scans the entire keyspace to find all keys that match the specified pattern. This operation is performed in a non-blocking manner, but it can still be slow if the keyspace is large. Therefore, it is generally not recommended to use the `KEYS` command in a production environment where performance is critical.

Additionally, the ordering of the output keys can be different if you run the same command subsequently.

## Errors

The `KEYS` command is straightforward and does not have many error conditions. However, there are a few scenarios where errors might occur:

1. `Invalid Pattern`: If the pattern is not a valid string, DiceDB will return an error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'keys' command`

2. `Memory Issues`: If the keyspace is extremely large, the `KEYS` command might consume a significant amount of memory, potentially leading to memory-related errors.

   - `Error Message`: This is more of a system-level issue and might not return a specific DiceDB error message but could lead to performance degradation or crashes.

## Best Practices

- `Avoid in Production`: Due to its potential to slow down the server, avoid using the `KEYS` command in a production environment. Instead, consider using the `SCAN` command, which is more efficient for large keyspaces.
- `Use Specific Patterns`: When using the `KEYS` command, try to use the most specific pattern possible to minimize the number of keys returned and reduce the load on the server.

## Alternatives

- `SCAN`: The `SCAN` command is a cursor-based iterator that allows you to incrementally iterate over the keyspace without blocking the server. It is a more efficient alternative to `KEYS` for large datasets.

```bash
127.0.0.1:7379> SCAN 0 MATCH key*
1) "0"
2) 1) "key1"
   2) "key2"
```
