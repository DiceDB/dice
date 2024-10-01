---
title: JSON.FORGET
description: Documentation for the DiceDB command JSON.FORGET
---

The `JSON.FORGET` command is part of the DiceDBJSON module, which allows you to manipulate JSON data stored in DiceDB. This command is used to delete a specified path from a JSON document stored at a given key. If the path leads to an array element, the element is removed, and the array is reindexed.

## Parameters

- `key`: (String) The key under which the JSON document is stored.
- `path`: (String) The JSONPath expression specifying the part of the JSON document to be deleted. The path must be a valid JSONPath expression.

## Return Value

- `Integer`: The number of paths that were deleted. If the specified path does not exist, the command returns `0`.

## Behaviour

When the `JSON.FORGET` command is executed, the following actions occur:

1. The command locates the JSON document stored at the specified key.
2. It evaluates the provided JSONPath expression to identify the part of the document to be deleted.
3. If the path is valid and exists, the specified part of the JSON document is removed.
4. If the path leads to an array element, the element is removed, and the array is reindexed.
5. The command returns the number of paths that were successfully deleted.

## Error Handling

The `JSON.FORGET` command can raise the following errors:

- `(error) ERR wrong number of arguments for 'JSON.FORGET' command`: This error occurs if the command is called with an incorrect number of arguments.
- `(error) ERR key does not exist`: This error occurs if the specified key does not exist in the DiceDB database.
- `(error) ERR invalid path`: This error occurs if the provided JSONPath expression is invalid or malformed.
- `(error) ERR path does not exist`: This error occurs if the specified path does not exist within the JSON document.

## Example Usage

### Example 1: Deleting a Field from a JSON Document

Suppose we have a JSON document stored at the key `user:1001`:

```json
{
  "name": "John Doe",
  "age": 30,
  "address": {
    "street": "123 Main St",
    "city": "Anytown"
  }
}
```

To delete the `age` field from this document, you would use the following command:

```sh
JSON.FORGET user:1001 $.age
```

`Expected Output:`

```sh
(integer) 1
```

### Example 2: Deleting an Element from a JSON Array

Suppose we have a JSON document stored at the key `user:1002`:

```json
{
  "name": "Jane Doe",
  "hobbies": ["reading", "swimming", "hiking"]
}
```

To delete the second element (`"swimming"`) from the `hobbies` array, you would use the following command:

```sh
JSON.FORGET user:1002 $.hobbies[1]
```

`Expected Output:`

```sh
(integer) 1
```

The updated JSON document would be:

```json
{
  "name": "Jane Doe",
  "hobbies": ["reading", "hiking"]
}
```

### Example 3: Deleting a Non-Existent Path

Suppose we have the same JSON document as in Example 1. If you attempt to delete a non-existent path:

```sh
JSON.FORGET user:1001 $.nonexistent
```

`Expected Output:`

```sh
(integer) 0
```
