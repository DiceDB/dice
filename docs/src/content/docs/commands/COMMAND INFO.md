---
title: COMMAND INFO
description: Documentation for the DiceDB command COMMAND INFO
---

## Introduction

The `COMMAND INFO` command is used to retrieve detailed information about one or more specified commands in the DiceDB server. For each command, it returns an array containing the command's name, arity (number of arguments), command flags, and key-related information.

## Syntax

```bash
COMMAND INFO command-name [command-name ...]
```

## Parameters

- **`command-name`**: One or more command names for which the information is requested. You can pass multiple command names to retrieve their details.

## Behavior

The `COMMAND INFO` command retrieves detailed information about one or more specified commands in the DiceDB server. The command operates as follows:

1. **Input Arguments**: The command takes a variable number of arguments, where each argument is expected to be the name of a command for which information is requested.
2. **Default Behavior**: If no command names are provided, the command will return default information about all available commands.
3. **Command Metadata Retrieval**: The command iterates over the predefined command metadata (`DiceCmds`) and stores the metadata in a map for quick lookup.
4. **Command Name Lookup**: For each provided command name, the command checks if it exists in the metadata map:
   - If the command exists, its metadata is appended to the result list.
   - If the command name is incorrect or not supported `nil` is appended in its place to indicate the absence of valid information.
5. **Result Encoding**: Finally, the result list, which contains the metadata for the valid command names and `nil` for any invalid ones, is encoded and returned as the output.

### Note:

1. If a valid command name is specified, its corresponding metadata is returned.
2. If a command name is incorrect or not supported, `nil` is returned in its place.

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

No error is raised, as this command supports a variable number of arguments.

## Example Usage

### Get command info for `SET` and `MGET`

```bash
127.0.0.1:7379> COMMAND INFO SET MGET
1) 1) "SET"
   2) (integer) -3
   3) (integer) 1
   4) (integer) 0
   5) (integer) 0
2) 1) "MGET"
   2) (integer) -2
   3) (integer) 1
   4) (integer) -1
   5) (integer) 1
```

### Usage Example when mixture of valid and invalid commands

In this example, we request information for two commands: one valid (`SET`) and one invalid (`UNKNOWNCOMMAND`).

```bash
127.0.0.1:7379> COMMAND INFO SET UNKNOWNCOMMAND
1) 1) "SET"
   2) (integer) -3
   3) (integer) 1
   4) (integer) 0
   5) (integer) 0
2) (nil)
```

### Incorrect Usage

An error is thrown when the command name passed to the `COMMAND INFO` command is incorrect or not supported in DiceDB.

```bash
127.0.0.1:7379> COMMAND INFO UNKNOWNCOMMAND
1) (nil)
```
