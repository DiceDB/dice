---
title: COMMAND INFO
description: Documentation for the DiceDB command COMMAND INFO
---

## Introduction

The `COMMAND INFO` command is used to retrieve detailed information about one or more specified commands in the DiceDB server. For each command, it returns an array containing the command's name, arity (number of arguments), command flags, and key-related information.

### Syntax

```
COMMAND INFO command-name [command-name ...]
```

### Parameters

- **`command-name`**: One or more command names for which the information is requested. You can pass multiple command names to retrieve their details.

## Behavior

The `COMMAND INFO` command retrieves detailed information about one or more specified commands in the DiceDB server. The command operates as follows:

1. **Input Arguments**: The command takes a variable number of arguments, where each argument is expected to be the name of a command for which information is requested.
2. **Default Behavior**: If no command names are provided, the command will return default information about all available commands.
3. **Command Metadata Retrieval**: The command iterates over the predefined command metadata (`DiceCmds`) and stores the metadata in a map for quick lookup.
4. **Command Name Lookup**: For each provided command name, the command checks if it exists in the metadata map:
   - If the command exists, its metadata is appended to the result list.
   - If the command name is incorrect or not supported:
     - If multiple command names are provided, `nil` is appended in its place to indicate the absence of valid information.
     - If a single command name is specified, the command will return an error: `(error) ERR invalid command specified`.
5. **Result Encoding**: Finally, the result list, which contains the metadata for the valid command names and `nil` for any invalid ones, is encoded and returned as the output.

### Note:

1. If a valid command name is specified, its corresponding metadata is returned.
2. If a command name is incorrect or not supported, `nil` is returned in its place for multiple commands, while an error is returned for a single invalid command name.

## Return Values

Returns an array of arrays, where each sub-array contains the following information for the specified commands:

- **Command Name**: The name of the command.
- **Arity**: An integer representing the number of arguments the command expects.
  - A positive number indicates the exact number of arguments.
  - A negative number indicates that the command accepts a variable number of arguments.
- **Flags** (_Note_: Not supported currently) : An array of flags that describe the command's properties (e.g., `readonly`, `fast`).
- **First Key**: The position of the first key in the argument list (0-based index).
- **Last Key**: The position of the last key in the argument list.
- **Key Step**: The step between keys in the argument list, useful for commands with multiple keys.

The structure of the returned data is as follows:

```bash
[
  [
    "command-name",
    arity,
    [
      "flag1",
      "flag2",
      ...
    ],
    first-key,
    last-key,
    key-step
  ],
  ...
]
```

## Errors

- **Error: unknown subcommand**: This error may occur if the subcommand is misspelled or not recognized by the server.

  - `(error) ERR unknown subcommand 'subcommand-name'. Try COMMAND HELP.`

- **Error: invalid command specified**: This error may occur when a single command name is passed in the `COMMAND INFO` command, and that command is not supported.
  - `(error) ERR invalid command specified`

## Example Usage

### Basic Usage : Get command info for `SET` and `MGET`

```bash
127.0.0.1:7379> COMMAND INFO SET MGET
[
  [
    "SET",
    -3,
    1,
    0,
    0
  ],
  [
    "MGET",
    -2,
    1,
    -1,
    1
  ]
]
```

### Usage Example when mixture of valid and invalid commands

In this example, we request information for two commands: one valid (`SET`) and one invalid (`UNKNOWNCOMMAND`).

````bash
127.0.0.1:7379> COMMAND INFO SET UNKNOWNCOMMAND
[
  [
    "SET",
    -3,
    1,
    0,
    0
  ],
  (nil)
]
````

### When the command name passed is incorrect or not supported

```bash
127.0.0.1:7379> COMMAND INFO UNKNOWNCOMMAND
(error) ERR invalid command specified
````
