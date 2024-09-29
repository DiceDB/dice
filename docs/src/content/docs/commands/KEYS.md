---
title: KEYS
description: Documentation for the DiceDB command KEYS
---

The `KEYS` command in DiceDB is used to find all keys matching a given pattern. This command is useful for retrieving keys that match a specific pattern, which can be helpful for debugging or managing your DiceDB data.

## Syntax

```plaintext
KEYS pattern
```

## Parameters

- `pattern`: A string that represents the pattern to match against the keys in the DiceDB database. The pattern can include special glob-style characters:
  - `*` matches any number of characters (including zero).
  - `?` matches exactly one character.
  - `[abc]` matches any one of the characters inside the brackets.
  - `[a-z]` matches any character in the specified range.

## Return Value

The `KEYS` command returns an array of strings, where each string is a key that matches the specified pattern. If no keys match the pattern, an empty array is returned.

## Example Usage

### Basic Example

```plaintext
DiceDB> SET key1 "value1"
OK
DiceDB> SET key2 "value2"
OK
DiceDB> SET anotherkey "value3"
OK
DiceDB> KEYS key*
1) "key1"
2) "key2"
```

### Using Wildcards

```plaintext
DiceDB> SET key1 "value1"
OK
DiceDB> SET key2 "value2"
OK
DiceDB> SET key3 "value3"
OK
DiceDB> KEYS key?
1) "key1"
2) "key2"
3) "key3"
```

### Using Character Ranges

```plaintext
DiceDB> SET key1 "value1"
OK
DiceDB> SET key2 "value2"
OK
DiceDB> SET key3 "value3"
OK
DiceDB> KEYS key[1-2]
1) "key1"
2) "key2"
```

## Behaviour

When the `KEYS` command is executed, DiceDB scans the entire keyspace to find all keys that match the specified pattern. This operation is performed in a non-blocking manner, but it can still be slow if the keyspace is large. Therefore, it is generally not recommended to use the `KEYS` command in a production environment where performance is critical.

## Error Handling

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

```plaintext
DiceDB> SCAN 0 MATCH key*
1) "0"
2) 1) "key1"
   2) "key2"
```

By following this detailed documentation, you should be able to effectively use the `KEYS` command in DiceDB while understanding its limitations and best practices.

