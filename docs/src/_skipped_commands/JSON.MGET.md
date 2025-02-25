---
title: JSON.MGET
description: Documentation for the DiceDB command JSON.MGET
---

# JSON.MGET

The `JSON.MGET` command in DiceDB is used to retrieve the values of specific JSON keys from multiple JSON documents stored at different keys. This command is particularly useful when you need to fetch the same JSON path from multiple JSON objects in a single operation, thereby reducing the number of round trips to the DiceDB server.

## Syntax

```bash
JSON.MGET key [key ...] path
```

## Parameters

| Parameter | Description                                                                                                                | Type   | Required |
| --------- | -------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | One or more keys from which the JSON values will be retrieved. These keys should point to JSON documents stored in DiceDB. | String | Yes      |
| `path`    | The JSON path to retrieve from each of the specified keys. The path should be specified in JSONPath format.                | String | Yes      |

## Return values

| Condition                                                                 | Return Value                                                                                |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| Command is successful                                                     | An array of JSON values corresponding to the specified path from each of the provided keys. |
| If a key does not exist or the path does not exist within a JSON document | The corresponding entry in the returned array will be `nil`.                                |

## Behaviour

When the `JSON.MGET` command is executed, DiceDB will:

1. Iterate over each provided key.
2. Retrieve the JSON document stored at each key.
3. Extract the value at the specified JSON path from each document.
4. Return an array of the extracted values.

If a key does not exist or the specified path is not found within a JSON document, `nil` will be returned for that key.

Here’s the revised error section for the `JSON.MGET` command documentation in the specified format:

---

## Errors

1. `Key does not exist`:

   - **Error Message**: `(error) ERROR could not perform this operation on a key that doesn't exist`
   - This error occurs when the specified key does not exist in the DiceDB database.

2. `Invalid JSON path`:
   - **Error Message**: `(error) ERROR invalid JSONPath`
   - This error occurs when the provided JSONPath expression is not valid.

## Example Usage

### Basic Usage

Assume we have the following JSON documents stored in DiceDB:

```bash
127.0.0.1:7379> JSON.SET user:1 $ '{"name": "Alice", "age": 30}'
OK
127.0.0.1:7379> JSON.SET user:2 $ '{"name": "Bob", "age": 25}'
OK
127.0.0.1:7379> JSON.SET user:3 $ '{"name": "Charlie", "age": 35}'
OK
```

To retrieve the `name` field from each of these JSON documents:

```bash
127.0.0.1:7379> JSON.MGET user:1 user:2 user:3 $.name
1) "\"Alice\""
2) "\"Bob\""
3) "\"Charlie\""
```

### Handling Non-Existent Keys and Paths

Assume we have the following JSON documents stored in DiceDB:

```bash
127.0.0.1:7379> JSON.SET user:1 $ '{"name": "Alice", "age": 30}'
OK
127.0.0.1:7379> JSON.SET user:2 $ '{"name": "Bob", "age": 25}'
OK
```

To retrieve the `address` field, which does not exist in the documents:

```bash
127.0.0.1:7379> JSON.MGET user:1 user:2 user:3 $.address
1) (nil)
2) (nil)
3) (nil)
```

In this case, `user:3` does not exist, and the `address` field does not exist in `user:1` and `user:2`, so `nil` is returned for each key.
