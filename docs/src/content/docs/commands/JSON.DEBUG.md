---
title: JSON.DEBUG
description: Documentation for the DiceDB command JSON.DEBUG
---

The `JSON.DEBUG` command in DiceDB is part of the DiceDBJSON module, which allows for the manipulation and querying of JSON data stored in DiceDB. This command is primarily used for debugging purposes, providing insights into the internal representation of JSON data within DiceDB.

## Syntax

```bash
JSON.DEBUG <subcommand> <key> [path]
```

## Parameters

| Parameter    | Description                                                                                         | Type   | Required |
| ------------ | --------------------------------------------------------------------------------------------------- | ------ | -------- |
| `subcommand` | The specific debug operation to perform. Currently, the supported subcommand is `MEMORY`.           | String | Yes      |
| `key`        | The key under which the JSON data is stored.                                                        | String | Yes      |
| `path`       | The JSON path to the specific part of the JSON data to debug. Defaults to the root if not provided. | String | No       |

### Subcommands

- `MEMORY`: This subcommand returns the memory usage of the JSON value at the specified path.

## Return Value

| Condition                      | Return Value                                                   |
| ------------------------------ | -------------------------------------------------------------- |
| if `MEMORY` subcommand is used | Memory usage in bytes of the JSON value at the specified path. |

## Behaviour

- For the `MEMORY` subcommand, it will calculate and return the memory usage of the JSON value at the specified path.
- If the path is not provided, it defaults to the root of the JSON data.

## Errors

1. `Invalid Subcommand`:
   - Error Message: `ERR unknown subcommand '<subcommand>'`
   - Occurs when an unsupported subcommand is provided.
2. `Invalid Path`:

   - Error Message: `ERR Path '<path>' does not exist`
   - If the specified path does not exist within the JSON data, DiceDB will return an error.

3. `Wrong Type`:
   - Error Message: `WRONGTYPE Operation against a key holding the wrong kind of value`
   - If the key exists but does not hold JSON data, DiceDB will return an error.

## Example Usage

### Debugging Memory Usage of Entire JSON Data

The `JSON.DEBUG MEMORY` command is used to get the memory usage of the entire JSON data stored under the key `myjson`. The command returns `89`, indicating that the JSON data occupies 89 bytes of memory.

```bash
127.0.0.1:7379> JSON.SET myjson $ '{"a":1}',
OK
127.0.0.1:7379> JSON.DEBUG MEMORY myjson
(integer) 89
```

### Debugging Memory Usage of a Specific Path

The `JSON.DEBUG MEMORY` command is used to get the memory usage of the JSON value at the path `$.a` within the JSON data stored under the key `myjson`. The command returns `16`, indicating that the specified JSON value occupies 16 bytes of memory.

```bash
127.0.0.1:7379> JSON.SET myjson $ '{"a":1,"b":2}',
OK
127.0.0.1:7379> JSON.DEBUG MEMORY myjson $.a
1) (integer) 16
```

### Handling Non-Existent Key

The `JSON.DEBUG MEMORY` command is used on a non-existent key `nonExistentKey`. DiceDB returns an 0 indicating that the key does not exist.

```bash
127.0.0.1:7379> JSON.DEBUG MEMORY nonExistentKey
(integer) 0
```

### Handling Invalid Path

The `JSON.DEBUG MEMORY` command is used on an invalid path `$.nonExistentPath` within the JSON data stored under the key `myjson`. DiceDB returns an error indicating that the specified path does not exist.

```bash
127.0.0.1:7379> JSON.DEBUG MEMORY myjson $.nonExistentPath
(error) ERR Path '$.nonExistentPath' does not exist
```

## Notes

- JSONPath expressions are used to navigate and specify the location within the JSON document. Familiarity with JSONPath syntax is beneficial for effective use of this command.
