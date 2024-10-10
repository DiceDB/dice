---
title: COMMAND
description: Documentation for the DiceDB command COMMAND
---

The `COMMAND` command in DiceDB is a powerful introspection tool that provides detailed information about all the DiceDB commands supported by the server. This command can be used to retrieve metadata about commands, such as their arity, flags, first key, last key, and key step. It is particularly useful for clients and developers who need to understand the capabilities and constraints of the DiceDB commands available in their environment.

## Syntex

The `COMMAND` command can be used in multiple forms, supporting various subcommands, each with its own set of parameters:

```
COMMAND <subcommand>
```

## Parameters

- **subcommand**: Optional. Available subcommands include:
  - `COUNT` : Returns the total number of commands in the DiceDB server.
  - `GETKEYS` : Returns the keys from the provided command and arguments.
  - `LIST` : Returns the list of all the commands in the DiceDB server.
  - `INFO` : Returns details about the specified commands.
  - `HELP` : Displays the help section for `COMMAND`, providing information about each available subcommand.

##### For more details on each subcommand, please refer to their respective documentation pages.

## COMMAND (No subcommands)

### Parameters

- **COMMAND**: This form takes no parameters and returns a list of all commands and their metadata supported by the DiceDB server.

### Behavior

This command serves as the default implementation of the COMMAND INFO command when no command name is specified.

### Return Value

Returns an array, where each element is a nested array containing the following details for each command

- **Command Name**: The name of the command.
- **Arity**: An integer representing the number of arguments the command expects.
  - A positive number indicates the exact number of arguments.
  - A negative number indicates that the command accepts a variable number of arguments.
- **Flags** (_Note_: Not supported currently) : An array of flags that describe the command's properties (e.g., `readonly`, `fast`).
- **First Key**: The position of the first key in the argument list (0-based index).
- **Last Key**: The position of the last key in the argument list.
- **Key Step**: The step between keys in the argument list, useful for commands with multiple keys.

### Detailed Return Value Descriptions

- `COMMAND`:
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

### Example Usage

```bash
127.0.0.1:7379> COMMAND
1) 1) "get"
   2) (integer) 2
   3) (integer) 1
   4) (integer) 0
   5) (integer) 0
2) 1) "set"
   2) (integer) -3
   3) (integer) 1
   4) (integer) 0
   5) (integer) 0
...
```
