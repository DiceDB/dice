---
title: JSON.NUMINCRBY
description: Documentation for the DiceDB command JSON.NUMINCRBY
---

The `JSON.NUMINCRBY` command is part of the DiceDBJSON module, which allows for manipulation of JSON data stored in DiceDB. This command increments a numeric value stored at a specified path within a JSON document by a given amount. It is particularly useful for atomic operations on numeric fields within JSON objects.

## Syntax

```bash
JSON.NUMINCRBY <key> <path> <increment>
```

## Parameters

| Parameter   | Description                                                                             | Type           | Required |
| ----------- | --------------------------------------------------------------------------------------- | -------------- | -------- |
| `key`       | The key under which the JSON document is stored.                                        | String         | Yes      |
| `path`      | The JSONPath expression specifying the location of the numeric value to be incremented. | String         | Yes      |
| `increment` | The numeric value by which to increment the target value.                               | Floating Point | Yes      |

## Return Value

| Condition                                  | Return Value                          |
| ------------------------------------------ | ------------------------------------- |
| Command is successful                      | Array of new value for update objects |
| Key does not exist or JSON path is invalid | error                                 |

## Behaviour

- If the specified key exists, specified path is valid and the value is numeric/floating point, `JSON.NUMINCRBY` command will increment the current value with the specified value and return the new value.
- If the path does not exist or does not contain a numeric value, an error will is raised.

## Errors

1. `Key does not exist`:

   - Error Message: `(error) ERROR could not perform this operation on a key that doesn't exist`
   - Occurs when the specified key does not exist in the DiceDB database.

2. `Invalid JSON path`:

   - Error Message: `(error) ERROR invalid JSONPath`
   - Occurs when the provided JSONPath expression is not valid.

## Example Usage

Create a document:

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30, "balance": 100.50, "account": {"id": 0, "lien": 0, "balance": 100.50}}'
"OK"
```

### Basic Usage

Incrementing the value of `age` object by 1

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1001 $.age 1
"[31]"
```

### Incrementing a Floating-Point Value

Incrementing the value of `balance` object by 25.75

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1001 $.balance 25.75
"[126.25]"
```

### Incrementing values for all matching objects recursively

Finding and recursively incrementing the values of both `balance` objects by 25.75

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1001 $..balance 25.75
"[126.25,126.25]"
```

### Handling Non-Numeric Value

Incrementing a non-numeric value

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1001 $.name 5
"[null]"
```

### Handling Non-Existent Path

Incrementing a non-existent path

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1001 $.nonexistent 10
"[]"
```

### Invalid Usage: Handling Invalid Path

Trying to increment an invalid path will result in an error

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1001 . 5
(error) ERROR invalid JSONPath
```

### Invalid Usage: Handling Non-existent Key

Trying to increment a path for a non-existent key will result in an error

```bash
127.0.0.1:7379> JSON.NUMINCRBY user:1002 . 5
(error) ERROR could not perform this operation on a key that doesn't exist
```
