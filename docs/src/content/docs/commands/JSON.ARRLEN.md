---
title: JSON.ARRLEN
description: Documentation for the DiceDB command JSON.ARRLEN
---

The `JSON.ARRLEN` command is part of the DiceDBJSON module, which allows you to work with JSON data structures in DiceDB. This command is used to retrieve the length of a JSON array at a specified path within a JSON document.

## Syntax

```plaintext
JSON.ARRLEN <key> [path]
```

## Parameters

- `key`: (Required) The key under which the JSON document is stored.
- `path`: (Optional) The JSONPath to the array within the JSON document. If not provided, the root path (`$`) is assumed.

## Return Value

The command returns an integer representing the length of the JSON array at the specified path. If the path does not exist or does not point to a JSON array, the command returns `null`.

## Behaviour

When the `JSON.ARRLEN` command is executed, DiceDB will:

1. Retrieve the JSON document stored at the specified key.
2. Navigate to the specified path within the JSON document.
3. Determine if the value at the specified path is a JSON array.
4. Return the length of the JSON array if it exists, or `null` if the path does not exist or does not point to a JSON array.

## Error Handling

The following errors may be raised by the `JSON.ARRLEN` command:

- `(error) ERR wrong number of arguments for 'JSON.ARRLEN' command`: This error occurs if the command is called with an incorrect number of arguments.
- `(error) ERR key does not exist`: This error occurs if the specified key does not exist in the DiceDB database.
- `(error) ERR path does not exist`: This error occurs if the specified path does not exist within the JSON document.
- `(error) ERR path is not a JSON array`: This error occurs if the specified path does not point to a JSON array.

## Example Usage

### Example 1: Basic Usage

Assume we have a JSON document stored at key `user:1001`:

```json
{
  "name": "John Doe",
  "emails": ["john.doe@example.com", "johndoe@gmail.com"],
  "age": 30
}
```

To get the length of the `emails` array:

```plaintext
JSON.ARRLEN user:1001 $.emails
```

`Response:`

```plaintext
2
```

### Example 2: Root Path

Assume we have a JSON document stored at key `user:1002`:

```json
[
  "item1",
  "item2",
  "item3"
]
```

To get the length of the root array:

```plaintext
JSON.ARRLEN user:1002
```

`Response:`

```plaintext
3
```

### Example 3: Non-Existent Path

Assume we have a JSON document stored at key `user:1003`:

```json
{
  "name": "Jane Doe",
  "contacts": {
    "phone": "123-456-7890"
  }
}
```

To get the length of a non-existent array:

```plaintext
JSON.ARRLEN user:1003 $.emails
```

`Response:`

```plaintext
(null)
```

### Example 4: Path is Not an Array

Assume we have a JSON document stored at key `user:1004`:

```json
{
  "name": "Alice",
  "age": 25
}
```

To get the length of a non-array path:

```plaintext
JSON.ARRLEN user:1004 $.age
```

`Response:`

```plaintext
(null)
```
