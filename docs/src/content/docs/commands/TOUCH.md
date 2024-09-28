---
title: TOUCH
description: Documentation for the DiceDB command TOUCH
---

The `TOUCH` command in DiceDB is used to update the last access time of one or more keys without modifying their values. This can be particularly useful for cache management, where you want to keep certain keys from expiring by marking them as recently used.

## Syntax

```plaintext
TOUCH key [key ...]
```

## Parameters

- `key`: The key whose last access time you want to update. You can specify multiple keys separated by spaces.

## Return Value

The `TOUCH` command returns an integer representing the number of keys that were successfully touched (i.e., the keys that exist and had their last access time updated).

## Behaviour

When the `TOUCH` command is executed, DiceDB will:

1. Check if each specified key exists in the database.
1. If a key exists, its last access time will be updated to the current time.
1. If a key does not exist, it will be ignored.
1. The command will return the count of keys that were successfully touched.

## Error Handling

The `TOUCH` command can raise the following errors:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if the key exists but is not of a type that supports the `TOUCH` operation (e.g., a key holding a non-string value).

## Example Usage

### Single Key

```plaintext
SET mykey "Hello"
TOUCH mykey
```

In this example, the `TOUCH` command updates the last access time of the key `mykey`.

### Multiple Keys

```plaintext
SET key1 "value1"
SET key2 "value2"
TOUCH key1 key2 key3
```

In this example, the `TOUCH` command attempts to update the last access time for `key1`, `key2`, and `key3`. Since `key3` does not exist, it will be ignored. The command will return `2`, indicating that two keys were successfully touched.

## Error Handling Example

### Wrong Type Error

```plaintext
LPUSH mylist "element"
TOUCH mylist
```

In this example, the `TOUCH` command will raise a `WRONGTYPE` error because `mylist` is a list, not a string.

## Notes

- The `TOUCH` command is useful for cache management and eviction policies, especially in scenarios where you want to prevent certain keys from expiring by marking them as recently accessed.
- The command does not modify the value of the key, only its last access time.
