---
title: JSON.OBJKEYS
description: The `JSON.OBJKEYS` command in DiceDB retrieves the keys of a JSON object located at a specified path within the document stored under the given key. This command is useful when you want to list the fields within a JSON object stored in a database.
---

The `JSON.OBJKEYS` command in DiceDB allows users to access the keys of a JSON object stored at a specific path within a document identified by a given key. By executing this command, users can easily retrieve a list of the fields present in the JSON object, making it a valuable tool for exploring and managing the structure of JSON data stored in the database.

This functionality is particularly useful for developers working with complex JSON structures who need to quickly identify and manipulate the various attributes within their data.

## Syntax

```bash
JSON.OBJKEYS key [path]
```

## Parameters

| Parameter | Description                                             | Type   | Required |
| --------- | ------------------------------------------------------- | ------ | -------- |
| `key`     | The name of the key holding the JSON document.          | String | Yes      |
| `path`    | JSONPath pointing to an array within the JSON document. | String | No       |

## Return Values

| Condition                                              | Return Value                                                                                            |
| ------------------------------------------------------ | ------------------------------------------------------------------------------------------------------- |
| Success                                                | ([]String) `Array of strings containing the keys present within the JSON object at the specified path.` |
| Key does not exist                                     | Error: `(error) ERR could not perform this operation on a key that doesn't exist`                       |
| Wrong number of arguments                              | Error: `(error) ERR wrong number of arguments for JSON.OBJKEYS command`                                 |
| Key has wrong type                                     | Error: `(error) ERR Existing key has wrong Dice type`                                                   |
| Operation attempted on a key with an incompatible type | Error: `(error) ERR WRONGTYPE Operation against a key holding the wrong kind of value`                  |

## Behaviour

- Root Path (Default): If no path is provided, JSON.OBJKEYS retrieves keys from the root object of the JSON document.
- Path Validation: If the specified path does not point to an object (e.g., if it points to a scalar, array, or does not exist), the command returns nil for that path.
- Non-existing Key: If the specified key does not exist in the database, an error is returned.
- Invalid JSON Path: If the provided JSONPath expression is invalid, an error message with the details of the parse error is returned.

## Errors

1. `Wrong number of arguments`:

   - Error Message: `(error) ERR wrong number of arguments for JSON.OBJKEYS command`
   - Raised if the number of arguments are less or more than expected.

2. `Key doesn't exist`:

   - Error Message: `(error) ERR could not perform this operation on a key that doesn't exist`
   - Raised if the specified key does not exist in the DiceDB database.

3. `Key has wrong Dice type`:

   - Error Message: `(error) ERR Existing key has wrong Dice type`
   - Raised if thevalue of the specified key doesn't match the specified value in DiceDb

4. `Path doesn't exist`:

   - Error Message: `(error) ERR WRONGTYPE Operation against a key holding the wrong kind of value`
   - Raised if an operation attempted on a key with an incompatible type.

## Example Usage

### Basic usage

Retrieving Keys of the Root Object

```bash
127.0.0.1:7379> JSON.SET a $ '{"name": "Alice", "age": 30, "address": {"city": "Wonderland", "zipcode": "12345"}}'
"OK"
127.0.0.1:7379> JSON.OBJKEYS a $
1) "name"
2) "age"
3) "address"
```

### Fetching keys of nested object

Retrieving Keys of a Nested Object

```bash
127.0.0.1:7379> JSON.SET b $ '{"name": "Alice", "partner": {"name": "Bob", "age": 28}}'
"OK"
127.0.0.1:7379> JSON.OBJKEYS b $.partner
1) "name"
2) "age"
```

### path pointing to non-object type

Error When Path Points to a Non-Object Type

```bash
127.0.0.1:7379> JSON.SET c $ '{"name": "Alice", "age": 30}'
"OK"
127.0.0.1:7379> JSON.OBJKEYS c $.age
(nil)
```

### When path doesn't exist

Error When Path Does Not Exist

```bash
127.0.0.1:7379> JSON.SET d $ '{"name": "Alice", "address": {"city": "Wonderland"}}'
"OK"
127.0.0.1:7379> JSON.OBJKEYS d $.nonexistentPath
(empty list or set)
```

### When key doesn't exist

Error When Key Does Not Exist

```bash
127.0.0.1:7379> JSON.OBJKEYS nonexistent_key $
(error) ERROR could not perform this operation on a key that doesn't exist
```
