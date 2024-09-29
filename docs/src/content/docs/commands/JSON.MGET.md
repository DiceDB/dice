---
title: JSON.MGET
description: Documentation for the DiceDB command JSON.MGET
---

The `JSON.MGET` command in DiceDB is used to retrieve the values of specific JSON keys from multiple JSON documents stored at different keys. This command is particularly useful when you need to fetch the same JSON path from multiple JSON objects in a single operation, thereby reducing the number of round trips to the DiceDB server.

## Syntax

```plaintext
JSON.MGET key [key ...] path
```

## Parameters

- `key [key ...]`: One or more keys from which the JSON values will be retrieved. These keys should point to JSON documents stored in DiceDB.
- `path`: The JSON path to retrieve from each of the specified keys. The path should be specified in JSONPath format.

## Return Value

The command returns an array of JSON values corresponding to the specified path from each of the provided keys. If a key does not exist or the path does not exist within a JSON document, the corresponding entry in the returned array will be `null`.

## Behaviour

When the `JSON.MGET` command is executed, DiceDB will:

1. Iterate over each provided key.
2. Retrieve the JSON document stored at each key.
3. Extract the value at the specified JSON path from each document.
4. Return an array of the extracted values.

If a key does not exist or the specified path is not found within a JSON document, `null` will be returned for that key.

## Error Handling

The following errors may be raised by the `JSON.MGET` command:

- `WRONGTYPE Operation against a key holding the wrong kind of value`: This error occurs if one of the specified keys does not hold a JSON document.
- `ERR syntax error`: This error occurs if the command syntax is incorrect, such as missing parameters or incorrect JSON path format.

## Example Usage

### Example 1: Basic Usage

Assume we have the following JSON documents stored in DiceDB:

```plaintext
DiceDB> JSON.SET user:1 $ '{"name": "Alice", "age": 30}'
OK
DiceDB> JSON.SET user:2 $ '{"name": "Bob", "age": 25}'
OK
DiceDB> JSON.SET user:3 $ '{"name": "Charlie", "age": 35}'
OK
```

To retrieve the `name` field from each of these JSON documents:

```plaintext
DiceDB> JSON.MGET user:1 user:2 user:3 $.name
1) "\"Alice\""
2) "\"Bob\""
3) "\"Charlie\""
```

### Example 2: Handling Non-Existent Keys and Paths

Assume we have the following JSON documents stored in DiceDB:

```plaintext
DiceDB> JSON.SET user:1 $ '{"name": "Alice", "age": 30}'
OK
DiceDB> JSON.SET user:2 $ '{"name": "Bob", "age": 25}'
OK
```

To retrieve the `address` field, which does not exist in the documents:

```plaintext
DiceDB> JSON.MGET user:1 user:2 user:3 $.address
1) (nil)
2) (nil)
3) (nil)
```

In this case, `user:3` does not exist, and the `address` field does not exist in `user:1` and `user:2`, so `null` is returned for each key.

## Notes

- The JSON path should be specified in JSONPath format, starting with `$` to represent the root of the JSON document.
- The command is useful for batch retrieval of the same JSON path from multiple JSON documents, optimizing performance by reducing the number of commands sent to the DiceDB server.

By understanding the `JSON.MGET` command, you can efficiently retrieve specific JSON values from multiple documents stored in DiceDB, making your data retrieval operations more streamlined and performant.

