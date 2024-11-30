---
title: JSON.SET
description: The `JSON.SET` command in DiceDB is used to set the JSON value at a specified key and path. This command allows storing, updating, and querying JSON documents in DiceDB.
---

The `JSON.SET` command in DiceDB is used to set the JSON value at a specified key and path. This command allows storing, updating, and querying JSON documents in DiceDB.

## Syntax

```bash
JSON.SET <key> <path> <json> [NX | XX]
```

## Parameters

| Parameter | Description                                                                                                                                                    | Type   | Required |
| --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| `key`     | The key under which the JSON document is stored. If the key does not exist, it will be created                                                                 | String | Yes      |
| `path`    | The path within the JSON document where the value should be set. The path should be specified in JSONPath format. Use `$` to refer to the root of the document | String | Yes      |
| `json`    | The JSON value to set at the specified path. This should be a valid JSON string                                                                                | JSON   | Yes      |
| `NX`      | Optional flag. Only set the value if the key does not already exist                                                                                            | Flag   | No       |
| `XX`      | Optional flag. Only set the value if the key already exists                                                                                                    | Flag   | No       |

## Return Values

| Condition                                   | Return Value  |
| ------------------------------------------- | ------------- |
| Command is successful                       | returns `OK`  |
| NX or XX conditions not met                 | returns `nil` |
| Syntax or specified constraints are invalid | error         |

## Behaviour

When the `JSON.SET` command is executed, the following behaviors are observed:

1. `Key Creation`: If the specified key does not exist and no flags are provided, a new key is created with the given JSON value.
2. `Path Creation`: If the specified path does not exist within the JSON document, it will be created.
3. `Conditional Set`: If the `NX` flag is provided, the value will only be set if the key does not already exist. If the `XX` flag is provided, the value will only be set if the key already exists.
4. `Overwrite`: If the key and path already exist, the existing value will be overwritten with the new JSON value.

## Errors

1. `Incorrect number of arguments`
   - `Error Message`: `(error) ERR wrong number of arguments for 'JSON.SET' command`
   - Occurs when the command is called with an incorrect number of arguments.
2. `Invalid JSON`: If the provided JSON value is not a valid JSON string, an error will be raised.
   - `Error Message`: `(error) expected value at line 1 column 1`
   - Occurs when the JSON string is malformed or contains syntax errors.
3. `Invalid JSONPath expression`:
   1. If the object is being created and the path is not root:
      - Error Message: `ERR new objects must be created at the root`
   1. If the specified path is static but does not exist/cannot be created
      - Error Message: `(error) Err wrong static path`
   1. If the specified path does not exist and cannot be created
      - Error Message: `Error occurred on position {}, "$.. <<<<----", expected one of the following: <string>, '*'`
4. `NX/XX Conflict`: If both `NX` and `XX` flags are provided, an error will be raised.
   - `Error Message`: `(error) ERR syntax error`
   - Occurs when both `NX` and `XX` flags are provided in the same command.

## Example Usage

### Basic Usage

Set a JSON value at the root of a new key:

```bash
127.0.0.1:7379> JSON.SET user:1001 $ '{"name": "John Doe", "age": 30}'
OK
```

### Set Value at a Specific Path

Set a JSON value at a specific path within an existing JSON document:

```bash
127.0.0.1:7379> JSON.SET user:1001 $.address '{"city": "New York", "zip": "10001"}'
OK
```

### Conditional Set with NX Flag

Set a JSON value only if the key does not already exist:

```bash
127.0.0.1:7379> JSON.SET user:1002 $ '{"name": "Jane Doe", "age": 25}' NX
OK
127.0.0.1:7379> JSON.SET user:1002 $ '{"name": "Jane Doe", "age": 30}' NX
(nil)
```

### Conditional Set with XX Flag

Set a JSON value only if the key already exists:

```bash
127.0.0.1:7379> JSON.SET user:1001 $.age 31 XX
OK
```

## JSONPath Support

`JSON.SET` leverages the powerful JSONPath syntax to accurately pinpoint the target location for value insertion or modification. Key JSONPath concepts include:

- Root: Represented by `$`, indicates the root of the JSON document.
- Child Operators:
  - `.`: Accesses object properties.
  - `[]`: Accesses array elements.

###
