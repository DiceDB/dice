---
title: JSON.ARRLEN
description: The `JSON.ARRLEN` command in DiceDB is used to retrieve the length of a JSON array at a specified path within a JSON document.
---

The `JSON.ARRLEN` command in DiceDB is used to retrieve the length of a JSON array at a specified path within a JSON document.

## Syntax

```
JSON.ARRLEN <key> [path]
```

## Parameters

| Parameter | Description                                                           | Type   | Required |
| --------- | --------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key under which the JSON document is stored.                      | String | Yes      |
| `path`    | The JSONPath to the array within the JSON document. Defaults to root. | String | No       |

## Return values

| Condition                                 | Return Value                                                                    |
| ----------------------------------------- | ------------------------------------------------------------------------------- |
| JSON array is found at the specified path | Length of the JSON array                                                        |
| JSONPath contains `*` wildcard            | Indicates the length of each key in the JSON. Returns `nil` for non-array keys. |

## Behaviour

- If the specified `key` does not exist, the command returns `null`.
- If `path` is not specified, the command assumes the root path (`$`).
- If the specified path exists and points to a JSON array, the command returns the length of the array.
- If the path does not exist or does not point to a JSON array, an error is returned.
- Multiple results are represented as a list in case of wildcards, where each array length is returned in order of appearance.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERROR wrong number of arguments for 'JSON.ARRLEN' command`
   - Occurs when the command is executed with less than one argument or with an invalid number of parameters.

2. `Invalid JSONPath`:

   - Error Message: `(error) ERROR Invalid JSONPath`
   - Occurs when the specified JSONPath is not valid or incorrect

3. `Path does not exist`:

   - Error Message: `(error) ERROR Path 'NON_EXISTANT_PATH' does not exist`
   - Occurs when the provided JSON path does not exist in the document.

4. `Path is not a JSON array`:
   - Error Message: `(error) ERROR Path 'NON_ARRAY_PATH' does not exist or not array`
   - Occurs when the specified path does not point to a JSON array.

## Examples Usage

### Basic Usage

Setting a key `user:1001` with a JSON document:

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name":"John Doe","emails":["john.doe@example.com","johndoe@gmail.com"],"age":30}'
OK
```

To get the length of the `emails` array:

```bash
127.0.0.1:7379> JSON.ARRLEN user:1001 $.emails
(integer) 2
```

### Root Path

Setting a key `user:1002` with a root array:

```bash
127.0.0.1:7379> JSON.SET user:1002 $ '["item1", "item2", "item3"]'
OK
```

To get the length of the root array:

```bash
127.0.0.1:7379> JSON.ARRLEN user:1002
(integer) 3
```

### Non-Existent Path

Setting a key `user:1003` with a JSON document:

```bash
127.0.0.1:7379> JSON.SET user:1003 $ '{"name": "Jane Doe","contacts":{"phone":"123-456-7890"}}'
OK
```

To get the length of a non-existent `emails` array:

```bash
127.0.0.1:7379> JSON.ARRLEN user:1003 $.emails
(error) ERROR Path '$.emails' does not exist
```

### Path is Not an Array

Setting a key `user:1004` with a JSON document:

```bash
127.0.0.1:7379> JSON.SET user:1004 $ '{"name": "Alice","age": 25}'
OK
```

To get the length of a non-array path:

```bash
127.0.0.1:7379> JSON.ARRLEN user:1004 $.age
(error) ERROR Path '$.age' does not exist or not array
```

### Path with Wildcards

Setting a key `user:1005` with a complex JSON document:

```bash
127.0.0.1:7379> JSON.SET user:1005 $ '{"age": 13,"high": 1.60,"pet": null,"language": ["python", "golang"],"partner": {"name": "tom"}}'
OK
```

To get the length of the `language` array using a wildcard path:

```bash
127.0.0.1:7379> JSON.ARRLEN user:1005 $.*
1) (nil)
1) (nil)
1) (nil)
4) (integer) 2
5) (nil)
```
