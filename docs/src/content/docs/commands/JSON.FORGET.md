---
title: JSON.FORGET
description: Documentation for the DiceDB command JSON.FORGET
---

The `JSON.FORGET` command in DiceDB is used to delete a specified path from a JSON document stored at a given key. If the path leads to an array element, the element is removed, and the array is reindexed. This is useful for modifying and updating portions of a JSON document.

## Syntax

```bash
JSON.FORGET key path
```

## Parameters

| Parameter | Description                                                                 | Type   | Required |
| --------- | --------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The name of the key under which the JSON document is stored.                | String | Yes      |
| `path`    | The JSONPath expression specifying the part of the JSON document to delete. | String | Yes      |

## Return values

| Condition                           | Return Value                            |
| ----------------------------------- | --------------------------------------- |
| Command successfully deletes a path | `The number of paths deleted (Integer)` |
| Path does not exist                 | `0`                                     |
| Invalid key or path                 | error                                   |

## Behaviour

- The command locates the JSON document stored at the specified key.
- It evaluates the provided JSONPath expression to identify the part of the document to be deleted.
- If the specified path exists, the targeted part is deleted.
- If the path leads to an array element, the element is removed, and the array is reindexed.
- The command returns the number of paths that were successfully deleted.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for 'JSON.FORGET' command`
   - Occurs when the command is called with an incorrect number of arguments.

2. `Key does not exist`:

   - Error Message: `(error) ERR key does not exist`
   - Occurs if the specified key is not present in the database.

3. `Invalid JSONPath expression`:

   - Error Message: `(error) ERR invalid path`
   - Occurs when the provided JSONPath expression is invalid or malformed.

4. `Path does not exist`:

   - Error Message: `(error) ERR path does not exist`
   - Occurs if the specified path does not exist within the JSON document.

## Example Usage

### Basic Usage

Deleting the `age` field from a JSON document stored at key `user:1001`:

```bash
127.0.0.1:7379> JSON.FORGET user:1001 $.age
(integer) 1
```

### Deleting an Element from a JSON Array

Deleting the second element (`"swimming"`) from the `hobbies` array in the JSON document stored at key `user:1002`:

```bash
127.0.0.1:7379> JSON.FORGET user:1002 $.hobbies[1]
(integer) 1
```

updated document:

```json
{
  "name": "Jane Doe",
  "hobbies": ["reading", "hiking"]
}
```

### Deleting a Non-Existent Path

Attempting to delete a non-existent path in the document at key `user:1001`:

```bash
127.0.0.1:7379> JSON.FORGET user:1001 $.nonexistent
(integer) 0
```
