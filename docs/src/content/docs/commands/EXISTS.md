---
title: EXISTS
description: Documentation for the DiceDB command EXISTS
---

The `EXISTS` command in DiceDB is used to determine if one or more keys exist in the database. This command is useful for checking the presence of keys before performing operations that depend on their existence.

## Syntax

```plaintext
EXISTS key [key ...]
```

## Parameters

- `key`: The key(s) to check for existence. You can specify one or more keys separated by spaces.

## Return Value

The command returns an integer representing the number of keys that exist among the specified keys.

- If none of the specified keys exist, the command returns `0`.
- If one or more of the specified keys exist, the command returns the count of those keys.

## Behaviour

When the `EXISTS` command is executed, DiceDB checks the existence of each specified key in the database. The command does not modify the database in any way; it only performs a read operation to check for the presence of the keys.

- If a single key is specified, the command returns `1` if the key exists and `0` if it does not.
- If multiple keys are specified, the command returns the count of keys that exist among the specified keys.

## Error Handling

The `EXISTS` command is straightforward and typically does not raise errors under normal usage. However, there are a few scenarios where errors might occur:

1. `Wrong Type of Arguments`: If the command is not provided with at least one key, DiceDB will return a syntax error.

   - `Error Message`: `(error) ERR wrong number of arguments for 'exists' command`

2. `Non-String Keys`: If the keys provided are not strings, DiceDB will raise a type error.

   - `Error Message`: `(error) ERR value is not a valid string`

## Example Usage

### Single Key Check

```plaintext
SET mykey "Hello"
EXISTS mykey
```

`Output:`

```plaintext
(integer) 1
```

In this example, the key `mykey` exists, so the command returns `1`.

### Multiple Keys Check

```plaintext
SET key1 "value1"
SET key2 "value2"
EXISTS key1 key2 key3
```

`Output:`

```plaintext
(integer) 2
```

In this example, `key1` and `key2` exist, but `key3` does not. Therefore, the command returns `2`.

### Non-Existent Key

```plaintext
EXISTS nonExistentKey
```

`Output:`

```plaintext
(integer) 0
```

In this example, the key `nonExistentKey` does not exist, so the command returns `0`.
