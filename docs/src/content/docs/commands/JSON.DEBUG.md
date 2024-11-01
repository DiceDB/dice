---
title: JSON.DEBUG
description: Documentation for the DiceDB command JSON.DEBUG
---

The `JSON.DEBUG` command in DiceDB is part of the DiceDBJSON module, which allows for the manipulation and querying of JSON data stored in DiceDB. This command is primarily used for debugging purposes, providing insights into the internal representation of JSON data within DiceDB.

## Parameters

### Syntax

```bash
JSON.DEBUG <subcommand> <key> [path]
```

### Parameters Description

- `subcommand`: (Required) The specific debug operation to perform. Currently, the supported subcommand is `MEMORY`.
- `key`: (Required) The key under which the JSON data is stored.
- `path`: (Optional) The JSON path to the specific part of the JSON data to debug. Defaults to the root if not provided.

### Subcommands

- `MEMORY`: This subcommand returns the memory usage of the JSON value at the specified path.

## Return Value

The return value of the `JSON.DEBUG` command depends on the subcommand used:

- `MEMORY`: Returns an integer representing the memory usage in bytes of the JSON value at the specified path.

## Behaviour

When the `JSON.DEBUG` command is executed, DiceDB will perform the specified debug operation on the JSON data stored at the given key and path. For the `MEMORY` subcommand, it will calculate and return the memory usage of the JSON value at the specified path. If the path is not provided, it defaults to the root of the JSON data.

## Error Handling

The `JSON.DEBUG` command can raise errors in the following scenarios:

1. `Invalid Subcommand`: If an unsupported subcommand is provided, DiceDB will return an error.

   - `Error Message`: `ERR unknown subcommand '<subcommand>'`

2. `Non-Existent Key`: If the specified key does not exist in the DiceDB database, DiceDB will return an error.

   - `Error Message`: `ERR no such key`

3. `Invalid Path`: If the specified path does not exist within the JSON data, DiceDB will return an error.

   - `Error Message`: `ERR path '<path>' does not exist`

4. `Wrong Type`: If the key exists but does not hold JSON data, DiceDB will return an error.

   - `Error Message`: `WRONGTYPE Operation against a key holding the wrong kind of value`

## Example Usage

### Example 1: Debugging Memory Usage of Entire JSON Data

```bash
127.0.0.1:7379> JSON.DEBUG MEMORY myjson
(integer) 256
```

In this example, the `JSON.DEBUG MEMORY` command is used to get the memory usage of the entire JSON data stored under the key `myjson`. The command returns `256`, indicating that the JSON data occupies 256 bytes of memory.

### Example 2: Debugging Memory Usage of a Specific Path

```bash
127.0.0.1:7379> JSON.DEBUG MEMORY myjson $.store.book[0]
(integer) 64
```

In this example, the `JSON.DEBUG MEMORY` command is used to get the memory usage of the JSON value at the path `$.store.book[0]` within the JSON data stored under the key `myjson`. The command returns `64`, indicating that the specified JSON value occupies 64 bytes of memory.

### Example 3: Handling Non-Existent Key

```bash
127.0.0.1:7379> JSON.DEBUG MEMORY nonExistentKey
(error) ERR no such key
```

In this example, the `JSON.DEBUG MEMORY` command is used on a non-existent key `nonExistentKey`. DiceDB returns an error indicating that the key does not exist.

### Example 4: Handling Invalid Path

```bash
127.0.0.1:7379> JSON.DEBUG MEMORY myjson $.nonExistentPath
(error) ERR path '$.nonExistentPath' does not exist
```

In this example, the `JSON.DEBUG MEMORY` command is used on an invalid path `$.nonExistentPath` within the JSON data stored under the key `myjson`. DiceDB returns an error indicating that the specified path does not exist.
