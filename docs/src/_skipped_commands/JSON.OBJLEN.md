---
title: JSON.OBJLEN
description: The `JSON.OBJLEN` command in DiceDB retrieves the number of keys stored in the JSON object located at key.
---

The `JSON.OBJLEN` command in DiceDB retrieves the number of keys stored in the JSON object located at key. By default, it counts the keys in the whole JSON object, but you can optionally specify a JSONPath to narrow the operation to a subset of the JSON object.

This functionality is particularly useful for developers working with complex JSON structures who need to quickly gauge the size of those structures.

## Syntax

```bash
JSON.OBJLEN key [path]
```

## Parameters

| Parameter | Description                                             | Type   | Required |
| --------- | ------------------------------------------------------- | ------ | -------- |
| `key`     | The name of the key holding the JSON document.          | String | Yes      |
| `path`    | JSONPath pointing to an array within the JSON document. | String | No       |

## Return values

| Condition                       | Return Value                                                                           |
| ------------------------------- | -------------------------------------------------------------------------------------- |
| Command is successful           | `Integer` denoting the number of keys length of the list at the specified key.         |
| Wrong number of arguments       | Error: `(error) ERR wrong number of arguments for JSON.OBJLEN command`                 |
| Key does not exist              | Error: `(error) ERR could not perform this operation on a key that doesn't exist`      |
| Key is not for a JSON object    | Error: `(error) ERR WRONGTYPE Operation against a key holding the wrong kind of value` |
| Path malformed or doesn't exist | Error: `(error) ERR Path 'foo' does not exist`                                         |

## Behaviour

- Root Path (Default): If no path is provided, JSON.OBJLEN retrieves keys from the root object of the JSON document.
- Path Validation: If the specified path does not point to an object (e.g., if it points to a scalar, array, or does not exist), the command returns `(nil)` for that path.
- Non-existing Key: If the specified key does not exist in the database, an error is returned.
- Invalid JSON Path: If the provided JSONPath expression is invalid, an error message with the details of the parse error is returned.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for JSON.OBJLEN command`
   - Happens if the number of arguments is less or more than required. (It must have at least one argument, or at most two arguments).

2. `Key doesn't exist`:

   - Error Message: `(error) ERR could not perform this operation on a key that doesn't exist`
   - Happens if the specified key does not exist in the DiceDB database.

3. `Key has wrong Dice type`:

   - Error Message: `(error) ERR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Happens if an operation attempted on a key with an incompatible type.

4. `ERR Path 'foo' does not exist`:
   - Error Message: `(error) ERR Path 'foo' does not exist`
   - Happens if the path string provided (ie. 'foo') could not be parsed into a valid JSONPath, or if the JSONPath does not exist in the object.

## Example Usage

### Basic usage

Get number of keys in the Root JSON Object. You can specify the JSON root using the symbol `$`.

```bash
127.0.0.1:7379> JSON.SET a $ '{"name": "Alice", "age": 30, "address": {"city": "Wonderland", "zipcode": "12345"}}'
"OK"
127.0.0.1:7379> JSON.OBJLEN a $
1) 3
```

It returns 3, because there are three root keys in the root JSON object: `name`, `age`, and `address`.

Or, if you don't want to specify a JSON path, it may be omitted. The path defaults to the root, and the result is given as a scalar:

```bash
127.0.0.1:7379> JSON.OBJLEN a
3
```

### Keys inside nested object

To count the number of keys inside a nested object, specify a JSON Path. The root of the JSON object is referred to by the `$` symbol.

```bash
127.0.0.1:7379> JSON.SET b $ '{"name": "Alice", "address": {"city": "Wonderland", "state": "Fantasy", "zipcode": "12345"}}'
"OK"
127.0.0.1:7379> JSON.OBJLEN b $.address
1) 3
```

Here, it returns 3 because it's counting the three keys inside the `$.address` JSON object: `city`, `state`, and `zipcode`.

### When path is not a JSON object

When `path` points to an existing element in a JSON object, but that element is not itself a JSON object, the result is `(nil)`.

```bash
127.0.0.1:7379> JSON.SET c $ '{"name": "Alice", "age": 30}'
"OK"
127.0.0.1:7379> JSON.OBJLEN c $.age
1) (nil)
```

### When path doesn't exist

When `path` does not exist, the result is an empty list or set.

```bash
127.0.0.1:7379> JSON.SET d $ '{"name": "Alice", "address": {"city": "Wonderland"}}'
"OK"
127.0.0.1:7379> JSON.OBJLEN d $.nonexistentPath
(empty list or set)
```

### Invalid Usage: When key doesn't exist

When `key` does not exist, the result is an error.

```bash
127.0.0.1:7379> JSON.OBJLEN nonexistent_key $
(error) ERR could not perform this operation on a key that doesn't exist
```
